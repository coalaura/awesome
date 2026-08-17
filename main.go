package main

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/coalaura/plain"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	log = plain.New(plain.WithDate(plain.RFC3339Local))

	config   *Config
	database *Database
)

func main() {
	var err error

	log.Println("Loading config...")

	config, err = LoadConfig()
	log.MustFail(err)

	log.Println("Loading database...")

	database, err = LoadDatabase()
	log.MustFail(err)

	defer database.Close()

	log.Println("Preparing router...")
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(log.Middleware())

	r.Get("/{feed}", HandleFeed)
	r.Get("/{feed}/{format}", HandleFeed)

	addr := config.Addr()

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var updates sync.WaitGroup

	updates.Go(func() {
		RunFeedUpdates(ctx, config.Feeds)
	})

	go func() {
		log.Printf("Listening at http://localhost%s/\n", addr)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warnln(err)
		}
	}()

	log.WaitForInterrupt()

	log.Warnln("Shutting down...")

	cancel()

	server.Close()

	updates.Wait()
}
