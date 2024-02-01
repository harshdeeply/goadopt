package main

import (
	"log"
)

func main() {
	db, err := DBConnect()
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	server := NewAPIServer(":8080", db)
	server.Run()
}
