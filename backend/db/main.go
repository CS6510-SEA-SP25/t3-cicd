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
	var host string = os.Getenv("DB_HOST")
	var user string = os.Getenv("DB_USER")
	var password string = os.Getenv("DB_PASSWORD")

	if host == "" {
		host = "localhost"
	}
	if user == "" {
		user = "root"
	}
	if password == "" {
		password = "root"
	}

	// Capture connection properties.
	cfg := mysql.Config{
		User:      user,
		Passwd:    password,
		Net:       "tcp",
		Addr:      host + ":3306",
		DBName:    "CicdApplication",
		ParseTime: true,
	}
	// Get a database handle.
	var err error
	fmt.Printf("cfg.FormatDSN() %v", cfg.FormatDSN())
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
