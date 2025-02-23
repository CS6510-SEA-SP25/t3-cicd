package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

var Instance *sql.DB

func Init() {
	// Capture connection properties.
	cfg := mysql.Config{
		// User:   os.Getenv("DBUSER"),
		User:   "root",
		Passwd: os.Getenv("DB_PASSWORD"),
		// Passwd:    "16032002",
		Net:       "tcp",
		Addr:      "127.0.0.1:3306",
		DBName:    "CicdApplication",
		ParseTime: true,
	}
	// Get a database handle.
	var err error
	Instance, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := Instance.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Database connected!")
}
