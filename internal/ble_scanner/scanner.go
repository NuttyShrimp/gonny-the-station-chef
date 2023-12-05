package blescanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/12urenloop/gonny-the-station-chef/internal/db"
	"github.com/12urenloop/gonny-the-station-chef/internal/socket"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/pkg/errors"
)

const ZEUS_MAC_PREFIX = "5a:45:55:53"

type Scanner struct {
	db         *db.DB
	ctx        context.Context
	socket     *socket.Send
	recvSocket *socket.Recv
}

type BatonData struct {
	ManufacturId      uint16
	UptimeMs          uint64
	BatteryPercentage uint8
}

func New(db *db.DB) *Scanner {
	d, err := linux.NewDevice()
	if err != nil {
		log.Fatalf("can't create new device: %v\n", err)
	}

	ble.SetDefaultDevice(d)
	ctx := context.Background()

	// Gotta have a listener before we can dial into it
	recvSocket, err := socket.NewRecv()

	if err != nil {
		log.Fatalf("Error creating unix socket: %v\n", err)
	}

	sendSocket, err := socket.NewSend()

	if err != nil {
		log.Fatalf("Error dialing unix socket: %v\n", err)
	}

	scanner := Scanner{
		db:         db,
		ctx:        ctx,
		socket:     sendSocket,
		recvSocket: recvSocket,
	}
	return &scanner
}

func (S *Scanner) scanFilter(a ble.Advertisement) bool {
	return strings.HasPrefix(a.Addr().String(), ZEUS_MAC_PREFIX)
}

func (S *Scanner) handleAdvertisment(a ble.Advertisement) {
	advData := a.ManufacturerData()
	if len(advData) == 25 {
		// Old baton
		return
	}
	if len(advData) != 11 {
		// Fake baton
		return
	}

	batonData := BatonData{}
	if err := binary.Read(bytes.NewReader(advData), binary.BigEndian, &batonData); err != nil {
		log.Printf("Failed to parse manufacturer data: %v\n", err)
		return
	}

	detection := db.Detection{
		DetectionTime:     time.Now(),
		Mac:               a.Addr().String(),
		Rssi:              a.RSSI(),
		UptimeMs:          batonData.UptimeMs,
		BatteryPercentage: batonData.BatteryPercentage,
	}

	go S.socket.NotifyChange(&detection)
	go S.db.InsertDetection(&detection)
}

func (S *Scanner) Scan() {
	// Stop on kill
	ctx, stop := signal.NotifyContext(S.ctx, syscall.SIGINT, syscall.SIGTERM)

	chkErr(ble.Scan(ctx, true, S.handleAdvertisment, S.scanFilter))
	fmt.Println("finished scanning")
	stop()
}

func chkErr(err error) {
	switch errors.Cause(err) {
	case nil:
	case context.DeadlineExceeded:
		return
	case context.Canceled:
		fmt.Println("Bluetooth scannig got canceled")
	default:
		log.Fatalf(err.Error())
	}
}

func (S *Scanner) Close() {
	S.socket.Close()
	S.recvSocket.Close()
}
