package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boofexxx/todolist/internal/handlers"
)

type config struct {
	user     string
	password string
	dbname   string
	addr     string
}

func main() {
	cfg := config{}
	cfg.user = os.Getenv("DBUSER")
	if cfg.user == "" {
		log.Fatal("DBUSER not provided")
	}
	cfg.password = os.Getenv("DBPASSWORD")
	if cfg.password == "" {
		log.Fatal("DBPASSWORD not provided")
	}
	cfg.dbname = os.Getenv("DBNAME")
	if cfg.dbname == "" {
		log.Fatal("DBNAME not provided")
	}
	cfg.addr = os.Getenv("ADDR")
	if cfg.addr == "" {
		cfg.addr = ":8080"
	}

	mux, err := handlers.NewServerMux(
		http.NewServeMux(),
		log.New(os.Stdout, "todolist: ", log.Default().Flags()),
		fmt.Sprintf("user=%s password=%s dbname=%s", cfg.user, cfg.password, cfg.dbname))
	if err != nil {
		log.Fatal(err)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World")
	})
	mux.HandleFunc("/tasks/", mux.TaskHandler)

	wrappedMux := mux.LoggerMiddleware(mux)

	s := http.Server{
		Addr:              cfg.addr,
		Handler:           wrappedMux,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 0,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       120 * time.Second,
		ErrorLog:          mux.Logger,
	}

	go func() {
		mux.Logger.Printf("starting to listen %s\n", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err = s.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
