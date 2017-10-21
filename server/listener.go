package server

import (
	"net"
)

type listener struct {
	conn *net.Conn
	ch   chan struct{}
}

// newListener 返回一个连接到指定 net.Conn 的 listener
func newListener(conn *net.Conn) (lis *listener, err error) {
	return &listener{
		conn: conn,
	}, nil
}

func (l *listener) Accept() (conn net.Conn, err error) {
	if l.conn != nil {
		conn := l.conn
		l.conn = nil
		return *conn, err
	}
	_ = <-l.ch
	return
}

func (l *listener) Close() (err error) {
	if l.conn == nil {
		return l.Close()
	}
	return
}

func (l *listener) Addr() (addr net.Addr) {
	return
}
