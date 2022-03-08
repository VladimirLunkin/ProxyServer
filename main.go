package main

import (
	"Proxy/ProxyServer"
	"fmt"
	"github.com/jackc/pgx"
	"log"
	"time"
)

func DBInit(Username, DBName, Password, DBHost, DBPort string) (*pgx.ConnPool, error) {
	ConnStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		Username,
		DBName,
		Password,
		DBHost,
		DBPort)

	pgxConnectionConfig, err := pgx.ParseConnectionString(ConnStr)
	if err != nil {
		log.Fatalf("Invalid config string: %s", err)
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgxConnectionConfig,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	})
	if err != nil {
		log.Fatalf("Error %s occurred during connection to database", err)
	}

	return pool, nil
}

var (
	username = "user_proxy"
	dbname   = "proxy"
	password = "password"
	host     = "localhost"
	port     = "5432"
)

func main() {
	dbConn, err := DBInit(username, dbname, password, host, port)
	if err != nil {
		log.Fatal(err)
	}

	server := ProxyServer.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		DB:           dbConn,
	}

	log.Println("starting server at :8080")
	log.Fatal(server.ListenAndServe())
}
