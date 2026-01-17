package main

import (
	"log"
	"net/http"
)

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("unable to read config: %v", err)
	}

	srv := newServer(cfg)

	log.Println("mee yahta api listening on :3000")
	if err := http.ListenAndServe(":3000", srv.routes()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
