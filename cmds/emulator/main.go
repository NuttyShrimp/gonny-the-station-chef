package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/socket"
)

func main() {
	// Open DB conn
	db := db.New()

	sendSocket, err := socket.NewSend()
	if err != nil {
		log.Fatalf("Error dialing unix socket: %v\n", err)
	}

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

mainLoop:
	for {
		select {
		case <-c:
			break mainLoop
		default:
			{

				detection := generateRandomDetection()

				id, err := db.InsertDetection(detection)
				if err != nil {
					log.Fatalf("Failed to insert detection: %v", err)
				}
				go sendSocket.NotifyChange(id)
				log.Printf("Inserted detection with id: %d\n", id)

				time.Sleep(time.Duration(randInt(10, 500)) * time.Millisecond)
			}
		}
	}

	sendSocket.Close()
}

func generateRandomDetection() *db.Detection {
	detection := db.Detection{
		DetectionTime:     time.Now(),
		Mac:               fmt.Sprintf("5a:45:55:53:00:%d%d", randInt(0, 10), randInt(0, 10)),
		Rssi:              randInt(-120, -40),
		UptimeMs:          0,
		BatteryPercentage: uint8(randInt(0, 101)),
	}

	return &detection
}

// [min, max[
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
