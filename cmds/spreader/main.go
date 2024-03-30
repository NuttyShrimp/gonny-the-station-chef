package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/logger"
	"github.com/12urenloop/gonny-the-station-chef/internal/wshandlers"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogger()

	db := db.New()

	app := fiber.New()

	app.Use("/detections", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/detections", websocket.New(func(c *websocket.Conn) {
		initMsg := struct {
			LastId int64 `json:"lastId"`
		}{}

		if err := c.ReadJSON(&initMsg); err != nil {
			logrus.Errorf("Failed to read initial WS msg: %+v\n", err)
			c.Close()
			return
		}

		c.Locals("db", db)
		go wshandlers.Writer(c, initMsg.LastId)
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

	fmt.Println("Fiber was successful shutdown.")
}
