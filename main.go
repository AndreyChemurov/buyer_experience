package main

import (
	"buyer_experience/internal/database"
	internalhttp "buyer_experience/internal/http"
	"log"
	"net/http"
)

func main() {
	if err := database.CreateTables(); err != nil {
		panic(err)
	}

	http.HandleFunc("/subscribe", internalhttp.Subscribe)
	http.HandleFunc("/", internalhttp.NotFound)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
