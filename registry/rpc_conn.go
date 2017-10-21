package registry

import (
	"fmt"
	"github.com/pkg/errors"
	pb "github.com/wangkechun/vv/proto"
	"io"
	"net"
	"qiniupkg.com/x/log.v7"
	"sync"
)

type rpcConn struct {
	conn   net.Conn
	header *pb.ProtoHeader
}

func (r rpcConn) String() string {
	return fmt.Sprintf("%+v", r.header)
}

// Register 组件
type Register struct {
	mux        sync.RWMutex
	laddr      string
	listenConn map[string]rpcConn
}

// Config Register 配置
type Config struct {
	Addr string
}

// New 返回一个register
func New(cfg Config) *Register {
	r := &Register{
		laddr:      cfg.Addr,
		listenConn: make(map[string]rpcConn),
	}
	return r
}

// TODO: 如何管理关闭的连接
func (r *Register) handleConn(conn net.Conn) (err error) {
	header, err := readHeader(conn)
	if err != nil {
		return errors.Wrap(err, "registry: readHeader")
	}
	rpcConn := rpcConn{conn: conn, header: header}
	log.Info("new connection", header, conn.RemoteAddr())
	ConnKind := rpcConn.header.ConnKind
	token := rpcConn.header.Token
	if ConnKind == pb.ProtoHeader_LISTEN {
		r.mux.Lock()
		r.listenConn[token] = rpcConn
		r.mux.Unlock()
	} else if ConnKind == pb.ProtoHeader_DIAL {
		dialConn := rpcConn
		r.mux.RLock()
		listenConn, ok := r.listenConn[token]
		r.mux.RUnlock()
		if !ok {
			log.Warn("dial request not found listener, close", dialConn)
			dialConn.conn.Close()
			return
		}
		r.mux.Lock()
		delete(r.listenConn, token)
		r.mux.Unlock()
		go io.Copy(dialConn.conn, listenConn.conn)
		go io.Copy(listenConn.conn, dialConn.conn)
		log.Info("connection create success", dialConn, "->", listenConn)
	}
	return
}

// Run 启动
func (r *Register) Run() (err error) {
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
