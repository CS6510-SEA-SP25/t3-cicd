//nolint:errcheck
package db

import (
	"crypto/tls"
	"crypto/x509"
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
	var port string = os.Getenv("DB_PORT")
	var dbName string = os.Getenv("DB_NAME")
	var password string = os.Getenv("DB_PASSWORD")
	var sslMode string = os.Getenv("DB_SSL_MODE") // e.g., "true", "false"
	var sslCA string = os.Getenv("DB_SSL_CA")     // Path to CA cert

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
		Addr:      host + ":" + port,
		DBName:    dbName,
		ParseTime: true,
	}

	// Configure SSL if enabled
	if sslMode == "true" {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(sslCA)
		if err != nil {
			log.Fatal("Failed to read CA cert:", err)
		}
		if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
			log.Fatal("Failed to append CA cert")
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCertPool,
		}

		dbTLSConfig := "custom"
		mysql.RegisterTLSConfig(dbTLSConfig, tlsConfig)
		cfg.TLSConfig = dbTLSConfig
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
