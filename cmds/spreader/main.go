package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/socket"
	"github.com/12urenloop/gonny-the-station-chef/internal/wshandlers"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func main() {
	db := db.New()

	recvSocket, err := socket.NewRecv()
	if err != nil {
		log.Fatalf("Failed to open socket listener: %v", err)
	}

	// TODO: should listen and when returned, start listening again until a value is send to a channel
	go recvSocket.Listen()

	app := fiber.New()

	app.Get("/buffer_size", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"size": recvSocket.BufferSize,
		})
	})

	app.Post("/buffer_size/:size", func(c *fiber.Ctx) error {
		param := struct {
			Size uint `params:"size"`
		}{}
		err := c.ParamsParser(&param)

		if err != nil {
			log.Printf("Failed to parse params: %+v\n", err)
			return c.SendStatus(400)
		}

		recvSocket.BufferSize = param.Size

		return c.JSON(fiber.Map{
			"size": recvSocket.BufferSize,
		})
	})

	app.Use("/detections", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/detections", websocket.New(func(c *websocket.Conn) {
		initMsg := struct {
			LastId uint64 `json:"lastId"`
		}{}

		if err := c.ReadJSON(&initMsg); err != nil {
			log.Printf("Failed to read initial WS msg: %+v\n", err)
			c.Close()
			return
		}

		c.Locals("db", db)
		go wshandlers.Writer(c, recvSocket, initMsg.LastId)
		wshandlers.Receiver(c)
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
