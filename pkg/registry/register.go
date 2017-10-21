package registry

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/wangkechun/vv/pkg/header"
	pb "github.com/wangkechun/vv/pkg/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	log.Info("tcp close")
	return r.Conn.Close()
}

func (r rpcConn) String() string {
	return fmt.Sprintf("%+v", r.header)
}

// Register 组件
type Register struct {
	mux    sync.Mutex
	laddr  string
	laddr2 string
	// 如果需要某个server发起一个新连接，则写一个msg
	user  map[string]chan struct{}
	conns []rpcConn
	// 如果服务端发起了一个新连接，写到这里
	server map[string]chan newConnMsg
}

type newConnMsg struct {
	index int
}

// Config Register 配置
type Config struct {
	Addr  string
	Addr2 string
}

// New 返回一个register
func New(cfg Config) *Register {
	r := &Register{
		laddr:  cfg.Addr,
		laddr2: cfg.Addr2,
		user:   make(map[string]chan struct{}, 0),
		server: make(map[string]chan newConnMsg, 0),
		conns:  make([]rpcConn, 0),
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
	log.Info("new connection", header, conn.RemoteAddr())
	ConnKind := rpcCn.header.ConnKind
	user := rpcCn.header.User
	if ConnKind == pb.ProtoHeader_LISTEN {
		log.Info("x")
		r.mux.Lock()
		r.conns = append(r.conns, rpcCn)
		ch, ok := r.server[user]
		if !ok {
			ch = make(chan newConnMsg, 1)
			r.server[user] = ch
		}
		r.mux.Unlock()
		log.Info("r.server.user <-")
		ch <- newConnMsg{index: len(r.conns) - 1}
		log.Info("x")
	} else if ConnKind == pb.ProtoHeader_DIAL {
		log.Info("dial")
		r.mux.Lock()
		ch, ok := r.user[user]
		if !ok {
			ch = make(chan struct{}, 1)
			r.user[user] = ch
		}
		r.mux.Unlock()
		// 通知需要server起一个新连接
		log.Info("send dial", user)
		ch <- struct{}{}
		log.Info("send ok")

		r.mux.Lock()
		ch2, ok := r.server[user]
		if !ok {
			ch2 = make(chan newConnMsg)
			r.server[user] = ch2
		}
		r.mux.Unlock()
		log.Info("recv dial")
		msg := <-ch2
		log.Info("recv ok")

		r.mux.Lock()
		conn := r.conns[msg.index]
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
	ch, ok := r.user[in.User]
	if !ok {
		ch = make(chan struct{}, 1)
		r.user[in.User] = ch
	}
	r.mux.Unlock()
	for {
		log.Info("wait")
		<-ch
		log.Info("recv, dial")
		err := stream.Send(&pb.OpenListenReply{})
		if err != nil {
			return err
		}
		log.Info("recv, dial")
	}
}

// Run 启动
func (r *Register) Run() (err error) {
	go func() {
		// rpc 端口
		lis, err := net.Listen("tcp", r.laddr2)
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
	ln, err := net.Listen("tcp", r.laddr)
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
