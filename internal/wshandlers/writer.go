package wshandlers

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/socket"
	"github.com/gofiber/contrib/websocket"
)

func Writer(c *websocket.Conn, recvSocket *socket.Recv, lastId uint64) {
	if lastId > recvSocket.LastValue {
		lastId = recvSocket.LastValue
	}

	closeChan := make(chan bool)

	db := c.Locals("db").(*db.DB)

out:
	for {
		select {
		case <-closeChan:
			break out
		default:
			{
				if c.Conn == nil {
					closeChan <- true
				}
				oldId := lastId
				newId := recvSocket.LastValue
				if newId == oldId {
					continue
				}

				log.Printf("Sending detections between %d and %d\n", oldId+1, newId)
				detections, err := db.GetDetectionsBetweenIds(oldId+1, newId)

				if err != nil {
					log.Printf("Failed fetching detections: %+v\n", err)
				}

				err = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err != nil {
					log.Printf("Failed set write deadline on WS: %+v\n", err)
					continue
				}

				if err = c.WriteJSON(detections); err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						log.Println("Failed to write data to websocket: deadline exceeded")
						continue
					}
					if errors.Is(err, io.ErrClosedPipe) || errors.Is(err, os.ErrDeadlineExceeded) || strings.Contains(err.Error(), "broken pipe") {
						// Handle connection closure
						log.Println("Connection closed")
						closeChan <- true
						break out
					}
					// Do some error recovery/restart procedure
					log.Printf("Failed to send detections over websocket: %+v\n", err)
					continue
				}
				lastId = newId

				// Do not spam the loop
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
