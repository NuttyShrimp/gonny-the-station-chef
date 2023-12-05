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

func (db *DB) InsertDetection(detection *Detection) {
	fmt.Println("Inserting detection")
	// db.conn.Exec("INSERT INTO ")
}
