package main

import (
	"context"
	"database/sql"
		"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/httpapi"
	"github.com/ViViD420/RSGalaxies-/server/internal/storage/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ViViD420/RSGalaxies-/server/internal/storage/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	address := os.Getenv("RS_GALAXIES_ADDR")
	if address == "" {
		address = ":8080"
	}

	var players game.PlayerRepository = game.NewMemoryPlayerRepository()
	var database *sql.DB
	if databaseURL := os.Getenv("RS_GALAXIES_DATABASE_URL"); databaseURL != "" {
		var err error
		database, err = sql.Open("pgx", databaseURL)
		if err != nil { log.Fatalf("open PostgreSQL: %v", err) }
		database.SetMaxOpenConns(20)
		database.SetMaxIdleConns(5)
		database.SetConnMaxLifetime(30 * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil { log.Fatalf("connect PostgreSQL: %v", err) }
		if err := postgres.Migrate(ctx, database); err != nil { log.Fatalf("migrate PostgreSQL: %v", err) }
		players = postgres.NewPlayerRepository(database)
		defer database.Close()
		log.Printf("using PostgreSQL player/world persistence")
	} else {
		log.Printf("using in-memory development storage; set RS_GALAXIES_DATABASE_URL for PostgreSQL")
	}

	repository := game.PlayerRepository(game.NewMemoryPlayerRepository())
	var database *sql.DB
	if databaseURL := os.Getenv("RS_GALAXIES_DATABASE_URL"); databaseURL != "" {
		var err error
		database, err = sql.Open("pgx", databaseURL)
		if err != nil { log.Fatalf("open PostgreSQL: %v", err) }
		defer database.Close()
		database.SetMaxOpenConns(20)
		database.SetMaxIdleConns(5)
		database.SetConnMaxLifetime(30 * time.Minute)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := database.PingContext(ctx); err != nil { log.Fatalf("connect PostgreSQL: %v", err) }
		if err := postgres.Migrate(ctx, database); err != nil { log.Fatalf("migrate PostgreSQL: %v", err) }
		repository = postgres.NewPlayerRepository(database)
		log.Print("using PostgreSQL player and world persistence")
	} else {
		log.Print("using in-memory development storage; set RS_GALAXIES_DATABASE_URL for PostgreSQL")
	}

	server := &http.Server{
		Addr:              address,
		Handler:           httpapi.New(players).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("RS Galaxies world development service listening on %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
