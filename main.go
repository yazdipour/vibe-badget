package main

import (
	"log"
	"os"

	"github.com/sh-yazdipour/vibe-badget/internal/db"
)

func main() {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "vibe-badget.db"
	}
	d, err := db.Open(path)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer d.Close()
	log.Println("db ready:", path)
}
