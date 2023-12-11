package socket

import (
	"encoding/binary"
	"net"

	"github.com/gofiber/fiber/v2/log"
)

type Recv struct {
	conn       net.Listener
	NotifyChan chan uint64
}

func NewRecv() (*Recv, error) {
	c, err := net.Listen("unix", "/tmp/gonny.sock")
	if err != nil {
		return nil, err
	}

	recv := Recv{
		conn:       c,
		NotifyChan: make(chan uint64),
	}
	return &recv, nil
}

func (R *Recv) Close() {
	R.conn.Close()
}

func (R *Recv) Listen() {
	for {
		fd, err := R.conn.Accept()
		if err != nil {
			log.Fatalf("Failed to accept unix conn socket: %v", fd)
			return
		}

		payload := []byte{}
		fd.Read(payload)

		id := binary.BigEndian.Uint64(payload)

		R.NotifyChan <- id
	}
}
