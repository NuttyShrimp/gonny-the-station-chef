package socket

import (
	"encoding/binary"
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

func (S *Send) NotifyChange(id uint64) error {
	idArr := make([]byte, 8)
	binary.BigEndian.PutUint64(idArr, id)
	_, err := S.conn.Write(idArr)
	return err
}
