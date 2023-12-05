package socket

import (
	"net"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
)

type Send struct {
	conn net.Conn
}

func NewSend() (*Send, error) {
	c, err := net.Dial("unix", "/tmp/gonny.sock")
	if err != nil {
		return nil, err
	}

	s := Send{
		conn: c,
	}
	return &s, nil
}

func (S *Send) Close() {
	S.conn.Close()
}

func (S *Send) NotifyChange(*db.Detection) error {
	_, err := S.conn.Write([]byte("ping"))
	return err
}
