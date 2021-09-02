package httpclient

import (
	"context"
	"net"
	"sync/atomic"
)

func (c *client) nextConnID() int64 {
	return atomic.AddInt64(&c.currentConnID, 1)
}

func (c *client) dialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	connID := c.nextConnID()
	dialerThing := &net.Dialer{
		Timeout: c.dialTimeout,
	}
	if !c.disableKeepAlive {
		dialerThing.KeepAlive = c.keepAliveTimeout
	}
	dc := dialerThing.DialContext
	c.log.Printf("Dialing conn %d to %s %s", connID, network, addr)
	conn, err := dc(ctx, network, addr)
	if err != nil {
		c.log.Printf(
			"Dialing conn %d to %s %s failed with %s",
			connID, network, addr, err.Error())
		return conn, err
	}
	onClose := func() {
		c.log.Printf("Closing conn %d to %s %s", connID, network, addr)
	}
	return newConn(conn, onClose), err
}

// below the connection wrapper to keep track of what is happening on TCP level
func newConn(conn net.Conn, onClose func()) net.Conn {
	return &connWrapper{conn, onClose}
}

type connWrapper struct {
	net.Conn
	onClose func()
}

func (c *connWrapper) Close() error {
	c.onClose()
	return c.Conn.Close()
}
