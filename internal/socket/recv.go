package socket

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Recv struct {
	conn       net.Conn
	LastValue  uint64
	BufferSize uint
}

func NewRecv() (*Recv, error) {
	c, err := net.Dial("unix", "/tmp/gonny.sock")
	if err != nil {
		return nil, err
	}

	recv := Recv{
		conn:       c,
		BufferSize: 4096,
	}
	return &recv, nil
}

func (R *Recv) Close() {
	R.conn.Close()
}

func (R *Recv) Listen() {
	for {
		time.Sleep(1 * time.Second)

		err := R.conn.SetDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			log.Printf("Failed to set deadline on unix socket: %+v\n", err)
			continue
		}

		payload := make([]byte, R.BufferSize)
		log.Printf("Listening on unix socket\n")
		nr, err := R.conn.Read(payload)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				log.Println("Failed to read data from unix socket: deadline exceeded")
				continue
			}
			if errors.Is(err, io.EOF) {
				// Handle connection closure
				log.Println("Connection closed")
				break
			}
			log.Printf("Failed to read data from unix socket: %+v\n", err)
		}

		if nr < 1 {
			continue
		}

		start := nr - 8
		if start < 0 {
			start = 0
		}

		id := binary.BigEndian.Uint64(payload[start:nr])
		log.Printf("Received payload: %d\n", id)
		R.LastValue = id
	}
}
