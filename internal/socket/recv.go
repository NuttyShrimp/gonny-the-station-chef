package socket

import "net"

type Recv struct {
	conn net.Conn
}

func NewRecv() (*Recv, error) {
	c, err := net.Dial("unix", "/tmp/gonny.sock")
	if err != nil {
		return nil, err
	}

	recv := Recv{
		conn: c,
	}
	return &recv, nil
}

func (R *Recv) Close() {
	R.conn.Close()
}
