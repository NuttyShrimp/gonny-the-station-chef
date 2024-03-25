package main

import (
	blescanner "github.com/12urenloop/gonny-the-station-chef/internal/ble_scanner"
	"github.com/12urenloop/gonny-the-station-chef/internal/db"
)

func main() {
	// Open DB conn
	db := db.New()

	scanner := blescanner.New(db)

	scanner.Scan()
}
