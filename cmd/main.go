package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Hello, world")
	router := createRouter()

	server := http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("server did not start")
	}
}
