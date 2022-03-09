package main

import (
	"Proxy/ProxyServer"
	"Proxy/Repeater"
	myConfig "Proxy/config"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/jackc/pgx"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	"log"
	"time"
)

func DBConn(Username, DBName, Password, DBHost, DBPort string) (*pgx.ConnPool, error) {
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

func main() {
	viper.AddConfigPath("./config/")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	var config myConfig.Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	dbc := config.DB
	dbConn, err := DBConn(dbc.Username, dbc.DBName, dbc.Password, dbc.Host, dbc.Port)
	if err != nil {
		log.Fatal(err)
	}

	server := ProxyServer.Server{
		Addr:         config.Proxy.Addr(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		DB:           dbConn,
	}

	fmt.Printf("Start proxy server on port %s\n", config.Proxy.Port)
	go func() {
		log.Fatal(server.ListenAndServe())
	}()

	r := router.New()

	db := Repeater.NewRepoPostgres(dbConn)

	Repeater.SetRepeaterRouting(r, &Repeater.Handlers{
		Repo: db,
	})

	fmt.Printf("Start repeater server on port %s\n", config.Repeater.Port)
	log.Fatal(fasthttp.ListenAndServe(config.Repeater.Addr(), r.Handler))
}
