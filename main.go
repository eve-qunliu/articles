package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eve-qunliu/articles/config"
	"github.com/eve-qunliu/articles/database"
	"github.com/eve-qunliu/articles/handlers"
)

func main() {
	cfg := config.NewConfig()
	dbProvider, err := database.NewProvider(cfg)

	if err != nil {
		log.Fatal("Cannot create data provider")
	}

	srv := &http.Server{
		Handler:      handlers.NewHandler(cfg, dbProvider),
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
