package wshandlers

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/gofiber/contrib/websocket"
)

func Writer(c *websocket.Conn, lastId int64) {
	closeChan := make(chan bool)

	db := c.Locals("db").(*db.DB)

	lastDbId, err := db.GetLastDetectionId()
	if err != nil {
		log.Printf("Failed to get last detection id: %+v\n", err)
		closeChan <- true
	}

	// TODO: SQL query
	if lastId > lastDbId {
		lastId = lastDbId
	}

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

				lastDbId, err := db.GetLastDetectionId()
				if err != nil {
					log.Printf("Failed to get last detection id: %+v\n", err)
					closeChan <- true
				}

				if lastDbId == lastId {
					continue
				}

				log.Printf("Sending detections between %d and %d\n", lastId+1, lastDbId)
				detections, err := db.GetDetectionsBetweenIds(lastId+1, lastDbId)

				if err != nil {
					log.Printf("Failed fetching detections: %+v\n", err)
					continue
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
				lastId = lastDbId

				// Do not spam the loop
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}
