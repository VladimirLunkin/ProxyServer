package main

import (
	"Proxy/ProxyServer"
	"fmt"
	"log"
	"time"
)

func main() {
	server := ProxyServer.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("starting server at :8080")
	log.Fatal(server.ListenAndServe())
}
