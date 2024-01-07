package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/socket"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type InitMessage struct {
	LastId uint64 `json:"lastId"`
}

func main() {
	db := db.New()

	recvSocket, err := socket.NewRecv()
	if err != nil {
		log.Fatalf("Failed to open socket listener: %v", err)
	}

	// TODO: should listen and when returned, start listening again until a value is send to a channel
	go recvSocket.Listen()

	app := fiber.New()

	app.Use("/detections", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/detections", websocket.New(func(c *websocket.Conn) {
		initMsg := InitMessage{}

		if err := c.ReadJSON(&initMsg); err != nil {
			log.Fatalf("Failed to read initial WS msg: %+v\n", err)
			c.Close()
			return
		}

		lastId := initMsg.LastId
		if lastId > recvSocket.LastValue {
			lastId = recvSocket.LastValue
		}

		for {
			oldId := lastId
			newId := recvSocket.LastValue
			if newId == oldId {
				continue
			}

			log.Printf("Sending detections between %d and %d\n", oldId+1, newId)
			detections, err := db.GetDetectionsBetweenIds(oldId+1, newId)
			log.Printf("Fetched detections between %d and %d\n", oldId+1, newId)

			if err != nil {
				log.Fatalf("Failed fetching detections: %+v\n", err)
			}

			err = c.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				log.Fatalf("Failed set write deadline on WS: %+v\n", err)
				continue
			}

			if err = c.WriteJSON(detections); err != nil {
				// Do some error recovery/restart procedure
				log.Fatalf("Failed to send detections over websocket: %+v\n", err)
				continue
			}
			lastId = newId
		}
	}))

	go func() {
		if err := app.Listen(":3000"); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	<-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	recvSocket.Close()

	fmt.Println("Fiber was successful shutdown.")
}
