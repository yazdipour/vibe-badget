package main

import (
	"log"
	"net/http"
	"os"

	"github.com/sh-yazdipour/vibe-wallet/internal/db"
	"github.com/sh-yazdipour/vibe-wallet/internal/httpapi"
	"github.com/sh-yazdipour/vibe-wallet/internal/store"
)

func main() {
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "vibe-wallet.db"
	}
	d, err := db.Open(path)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer d.Close()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	h := httpapi.NewServer(store.New(d), staticFS())
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, h))
}
