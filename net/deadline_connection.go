package net

import (
	"net"
	"time"
)

// DeadlineConnection is an extension over the standard net.Conn,
// that makes it more convenient to read/write within a configurable timeout.
// It is essential in network communication where the network is unreliable.
type DeadlineConnection struct {
	net.Conn
	ReadDeadlineTimeout  time.Duration
	WriteDeadlineTimeout time.Duration
}

func (d *DeadlineConnection) Read(b []byte) (int, error) {
	err := d.Conn.SetReadDeadline(time.Now().Add(d.ReadDeadlineTimeout))
	if err != nil {
		return 0, err
	}

	return d.Conn.Read(b)
}

func (d *DeadlineConnection) Write(b []byte) (int, error) {
	err := d.Conn.SetWriteDeadline(time.Now().Add(d.WriteDeadlineTimeout))
	if err != nil {
		return 0, err
	}

	return d.Conn.Write(b)
}
