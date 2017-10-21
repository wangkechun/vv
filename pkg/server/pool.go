package server

import (
	"github.com/pkg/errors"
	"github.com/wangkechun/vv/pkg/header"
	pb "github.com/wangkechun/vv/pkg/proto"
	"github.com/wangkechun/vv/pkg/token"
	"net"
	"qiniupkg.com/x/log.v7"
	"time"
)

type rlisten struct {
	conn   net.Conn
	readed bool
}

func (c *rlisten) Accept() (conn net.Conn, err error) {
	log.Info("pre Accept")
	if !c.readed {
		c.readed = true
		log.Info("Accept ok")
		return c.conn, nil
	}
	// log.Info("x")
	// defer log.Info("ret")
	// ch := make(chan struct{}, 0)
	// ch <- struct{}{}
	// log.Info("x2")
	time.Sleep(time.Second * 100)
	return
}

func (c *rlisten) Close() (err error) {
	return c.conn.Close()
}

func (c *rlisten) Addr() (addr net.Addr) {
	return c.conn.RemoteAddr()
}

// func newPool(cfg Config, num int) connPool {
// 	pool := connPool{cfg: cfg, ch: make(chan struct{}, num)}
// 	for i := 0; i < defaultConnectionNum; i++ {
// 		pool.ch <- struct{}{}
// 	}
// 	return pool
// }

// type connPool struct {
// 	cfg Config
// 	ch  chan struct{}
// }

// func (c *connPool) Get() (conn net.Conn, err error) {
// 	log.Info("start get")

// 	<-c.ch
// 	conn, err = net.Dial("tcp", c.cfg.RegistryAddr)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to connect registry")
// 	}

// 	err = header.WriteHeader(conn, &pb.ProtoHeader{
// 		Version:    "1",
// 		Token:      token.GetServerToken(),
// 		ServerKind: pb.ProtoHeader_SERVER,
// 		ConnKind:   pb.ProtoHeader_LISTEN,
// 	})
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to connect registry: write header")
// 	}
// 	log.Info("conn get")
// 	return &xConn{conn, c}, nil
// }

// func (c *connPool) Release() {
// 	log.Info("conn release")
// 	c.ch <- struct{}{}
// }

// func (c *connPool) Accept() (conn net.Conn, err error) {
// 	return c.Get()
// }

// func (c *connPool) Close() (err error) {
// 	return
// }

// func (c *connPool) Addr() (addr net.Addr) {
// 	return
// }

// type xConn struct {
// 	net.Conn
// 	pool *connPool
// }

// func (x *xConn) Close() error {
// 	x.pool.Release()
// 	return x.Conn.Close()
// }
