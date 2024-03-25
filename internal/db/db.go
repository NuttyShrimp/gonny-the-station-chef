package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/12urenloop/gonny-the-station-chef/internal/utils"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	host     = utils.GetEnvOrFallback("DATABASE_HOST", "localhost")
	port, _  = strconv.ParseUint(utils.GetEnvOrFallback("DATABASE_PORT", "5432"), 10, 16)
	user     = utils.GetEnvOrFallback("DATABASE_USER", "ronny")
	password = utils.GetEnvOrFallback("DATABASE_PASSWORD", "ronnydbpassword")
	dbname   = utils.GetEnvOrFallback("DATABASE_DB", "ronny")
	psqldsn  = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
)

type DB struct {
	conn *gorm.DB
}

func New() *DB {
	conn, err := gorm.Open(postgres.Open(psqldsn), &gorm.Config{})

	if err != nil {
		// TODO: properly handle this
		os.Exit(1)
	}

	db := DB{
		conn: conn,
	}

	err = conn.AutoMigrate(&Detection{})
	if err != nil {
		log.Fatalf("Failed to do auto migration for tables: %+v\n", err)
	}

	return &db
}

func (db *DB) InsertDetection(detection *Detection) (int64, error) {
	err := db.conn.Create(detection).Error
	return detection.ID, err
}

func (db *DB) GetDetectionsBetweenIds(a, b int64) (*[]Detection, error) {
	detections := []Detection{}

	err := db.conn.Where("id BETWEEN ? AND ?", a, b).Find(&detections).Error

	return &detections, err
}

func (db *DB) GetLastDetectionId() (int64, error) {
	var detection Detection
	err := db.conn.Last(&detection).Error
	return detection.ID, err
}
