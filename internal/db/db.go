package db

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"github.com/12urenloop/gonny-the-station-chef/internal/utils"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

var (
	host     = utils.GetEnvOrFallback("DATABASE_HOST", "localhost")
	port, _  = strconv.ParseUint(utils.GetEnvOrFallback("DATABASE_PORT", "5432"), 10, 16)
	user     = utils.GetEnvOrFallback("DATABASE_USER", "ronny")
	password = utils.GetEnvOrFallback("DATABASE_PASSWORD", "ronnydbpassword")
	dbname   = utils.GetEnvOrFallback("DATABASE_DB", "ronny")
	psqlconn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
)

type DB struct {
	conn *sql.DB
}

func New() *DB {
	conn, err := sql.Open("pgx", psqlconn)

	if err != nil {
		log.Fatalf("Cannot create db connection: %v\n", err)
	}

	db := DB{
		conn: conn,
	}

	return &db
}

func (db *DB) Close() {
	db.conn.Close()
}

func (db *DB) InsertDetection(detection *Detection) (uint64, error) {
	fmt.Println("Inserting detection")
	id := uint64(0)
	err := db.conn.QueryRow("INSERT INTO detections (detection_time, mac, rssi, baton_uptime_ms, battery_percentage) VALUES ($1, $2, $3, $4, $5) RETURNING id", detection.DetectionTime, detection.Mac, detection.Rssi, detection.UptimeMs, detection.BatteryPercentage).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (db *DB) GetDetectionsBetweenIds(a, b uint64) (*[]Detection, error) {
	detections := []Detection{}
	rows, err := db.conn.Query("SELECT id, detection_time, mac, rssi, baton_uptime_ms, battery_percentage WHERE id BETWEEN $1 AND $2", a, b)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	detection := Detection{}
	for rows.Next() {
		err := rows.Scan(&detection.Id, &detection.DetectionTime, &detection.Mac, &detection.Rssi, &detection.UptimeMs, &detection.BatteryPercentage)
		if err != nil {
			return nil, err
		}

		detections = append(detections, detection)
	}
	return &detections, nil
}
