package registry

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/wangkechun/vv/pkg/header"
	pb "github.com/wangkechun/vv/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
	"qiniupkg.com/x/log.v7"
	"sync"
	"time"
)

type rpcConn struct {
	net.Conn
	closed bool
	header *pb.ProtoHeader
}

func (r *rpcConn) Close() error {
	r.closed = true
	log.Info("tcp close", r.Conn.RemoteAddr())
	return r.Conn.Close()
}

func (r rpcConn) String() string {
	return fmt.Sprintf("%+v", r.header)
}

// Register 组件
type Register struct {
	cfg Config
	mux sync.Mutex
	// 如果客户端需要某个服务端发起一个新连接与其联通，则在这里面写个消息
	clientRequire map[string]chan struct{}
	// 所有连接 TODO:回收
	conns []rpcConn
	// 如果服务端成功发起了一个新连接，写到这里
	serverNewConn map[string]chan newConnMsg
}

type newConnMsg struct {
	connIndex int
}

// Config Register 配置
type Config struct {
	RegistryAddrTCP string
	RegistryAddrRPC string
}

// New 返回一个register
func New(cfg Config) *Register {
	r := &Register{
		cfg:           cfg,
		clientRequire: make(map[string]chan struct{}, 0),
		serverNewConn: make(map[string]chan newConnMsg, 0),
		conns:         make([]rpcConn, 0),
	}
	return r
}

func proxy(name string, dst io.Writer, src io.Reader, errCh chan error) {
	n, err := io.Copy(dst, src)
	log.Printf("[DEBUG] socks: Copied %d bytes to %s", n, name)
	time.Sleep(10 * time.Millisecond)
	errCh <- err
}

// TODO: 如何管理关闭超时的连接
func (r *Register) handleConn(conn net.Conn) (err error) {
	log.Info("new connection", conn.RemoteAddr())
	header, err := header.ReadHeader(conn)
	if err != nil {
		return errors.Wrap(err, "registry: readHeader")
	}
	rpcCn := rpcConn{Conn: conn, header: header}
	ConnKind := rpcCn.header.ConnKind
	user := rpcCn.header.User
	if ConnKind == pb.ProtoHeader_LISTEN {
		r.mux.Lock()
		r.conns = append(r.conns, rpcCn)
		ch, ok := r.serverNewConn[user]
		if !ok {
			ch = make(chan newConnMsg, 1)
			r.serverNewConn[user] = ch
		}
		r.mux.Unlock()
		log.Info("LISTEN new conn, user", user)
		ch <- newConnMsg{connIndex: len(r.conns) - 1}
	} else if ConnKind == pb.ProtoHeader_DIAL {
		r.mux.Lock()
		ch, ok := r.clientRequire[user]
		if !ok {
			ch = make(chan struct{}, 1)
			r.clientRequire[user] = ch
		}
		r.mux.Unlock()
		// 通知需要server起一个新连接
		log.Info("DIAL new conn, user", user)
		ch <- struct{}{}

		r.mux.Lock()
		ch2, ok := r.serverNewConn[user]
		if !ok {
			ch2 = make(chan newConnMsg)
			r.serverNewConn[user] = ch2
		}
		r.mux.Unlock()
		msg := <-ch2
		log.Info("DIAL wait conn success", user)

		r.mux.Lock()
		conn := r.conns[msg.connIndex]
		r.mux.Unlock()

		go func(dialConn, listenConn rpcConn) {
			errCh := make(chan error, 2)
			defer dialConn.Close()
			defer listenConn.Close()
			go proxy("target", dialConn, listenConn, errCh)
			go proxy("target", listenConn, dialConn, errCh)
			select {
			case e := <-errCh:
				log.Info(e)
				return
			}
		}(rpcCn, conn)

		log.Info("connection create success", rpcCn, "->", conn)
	}
	return
}

func (r *Register) OpenListen(in *pb.OpenListenRequest, stream pb.VvRegistry_OpenListenServer) (err error) {
	log.Info("open listen", in.User)
	var ch chan struct{}
	r.mux.Lock()
	ch, ok := r.clientRequire[in.User]
	if !ok {
		ch = make(chan struct{}, 1)
		r.clientRequire[in.User] = ch
	}
	r.mux.Unlock()
	for {
		<-ch
		log.Info("OpenListen recv")
		err := stream.Send(&pb.OpenListenReply{})
		if err != nil {
			return err
		}
	}
}

// Run 启动
func (r *Register) Run() (err error) {
	go func() {
		// rpc 端口
		lis, err := net.Listen("tcp", r.cfg.RegistryAddrRPC)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterVvRegistryServer(s, r)
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	// tcp 端口
	ln, err := net.Listen("tcp", r.cfg.RegistryAddrTCP)
	if err != nil {
		return errors.Wrap(err, "registry: net.Listen")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return errors.Wrap(err, "registry: net.Accept")
		}
		go r.handleConn(conn)
	}
}
