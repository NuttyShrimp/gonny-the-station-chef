package socket

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Send struct {
	conn     net.Listener
	channels []*net.Conn
}

func NewSend() (*Send, error) {
	c, err := net.Listen("unix", "/tmp/gonny.sock")
	if err != nil {
		return nil, err
	}

	s := Send{
		conn:     c,
		channels: []*net.Conn{},
	}

	go s.AcceptConnections()

	return &s, nil
}

func (S *Send) removeChannel(chanIdx int) {
	newChannels := S.channels[:chanIdx]
	if (chanIdx + 1) < len(S.channels) {
		newChannels = append(newChannels, S.channels[chanIdx+1:]...)
	}
	S.channels = newChannels
}

func (S *Send) Close() {
	for chanIdx := range S.channels {
		(*S.channels[chanIdx]).Close()
	}
	S.channels = []*net.Conn{}

	S.conn.Close()
}

func (S *Send) NotifyChange(id int64) {
	idArr := make([]byte, 8)
	binary.BigEndian.PutUint64(idArr, uint64(id))

	log.Printf("Sending id: %d\n", id)
	for chanIdx := range S.channels {
		fd := *S.channels[chanIdx]

		err := fd.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			S.removeChannel(chanIdx)
		}

		n, err := fd.Write(idArr)
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "broken pipe") {
				S.removeChannel(chanIdx)
				continue
			}
			log.Printf("Failed to write to unix socket: %+v\n", err)
		}
		log.Printf("Wrote %d bytes: %+v to unix socket\n", n, idArr)
	}
}

func (S *Send) AcceptConnections() {
	for {
		fd, err := S.conn.Accept()
		if err != nil {
			log.Fatalf("Failed to accept unix conn socket: %v error: %v", fd, err)
			return
		}
		log.Printf("Accepted unix conn socket\n")
		S.channels = append(S.channels, &fd)
	}
}
