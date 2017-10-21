package server

import (
	"net"
)

type listen struct {
	conn   net.Conn
	readed bool
}

func (c *listen) Accept() (conn net.Conn, err error) {
	if !c.readed {
		c.readed = true
		return c.conn, nil
	}
	ch := make(chan struct{}, 0)
	<-ch
	return
}

func (c *listen) Close() (err error) {
	return c.conn.Close()
}

func (c *listen) Addr() (addr net.Addr) {
	return c.conn.RemoteAddr()
}
