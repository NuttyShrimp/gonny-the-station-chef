package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

		// TODO: Resend all messages after the lastId from the initMsg

		// The lastId we passed
		lastId := uint64(initMsg.LastId)

		for {
			newId := recvSocket.LastValue
			if newId == lastId {
				continue
			}

			detections, err := db.GetDetectionsBetweenIds(lastId, newId)

			if err != nil {
				log.Fatalf("Failed fetching detections: %+v\n", err)
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

	_ = <-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	db.Close()
	recvSocket.Close()

	fmt.Println("Fiber was successful shutdown.")
}
