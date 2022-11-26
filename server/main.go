package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-playground/validator/v10"
	"github.com/normanjaeckel/Blitzumfrage/server/public"
	"golang.org/x/sys/unix"
)

const (
	Host     string = "localhost"
	Port     int    = 8000
	DataFile        = "data.jsonl"
)

type Logger interface {
	Printf(format string, v ...any)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := log.Default()

	onSignals(logger, cancel)

	addr := fmt.Sprintf("%s:%d", Host, Port)
	if err := start(ctx, logger, addr); err != nil {
		logger.Fatalf("Error: %v", err)
	}
}

// onSignals starts a goroutine that listens to the operating system signals
// SIGTERM and SIGINT. On incomming signal (SIGTERM or SIGINT), the cancel
// function is called. If SIGINT comes in a second time, os.Exit(1) is called to
// abort the process.
func onSignals(log Logger, cancel context.CancelFunc) {
	go func() {
		msg := "Received operating system signal: %s"

		sigTerm := make(chan os.Signal, 1)
		signal.Notify(sigTerm, unix.SIGTERM)
		sigInt := make(chan os.Signal, 1)
		signal.Notify(sigInt, unix.SIGINT)

		select {
		case s := <-sigInt:
			log.Printf(msg, s.String())
		case s := <-sigTerm:
			log.Printf(msg, s.String())
		}
		cancel()

		s := <-sigInt
		log.Printf(msg, s.String())
		log.Printf("Process aborted")
		os.Exit(1)
	}()
}

// start initiates the HTTP server and lets it listen on the given address.
func start(ctx context.Context, logger Logger, addr string) error {
	s := &http.Server{
		Addr:    addr,
		Handler: handler(logger),
	}

	go func() {
		<-ctx.Done()
		logger.Printf("Server is shuting down")
		if err := s.Shutdown(context.Background()); err != nil {
			logger.Printf("Error: Shutting down server: %v", err)
		}
	}()

	logger.Printf("Server starts and listens on %q", addr)
	err := s.ListenAndServe()
	if err != http.ErrServerClosed {
		return fmt.Errorf("server exited: %w", err)
	}
	logger.Printf("Server is down")

	return nil
}

type payload struct {
	Name   string `json:"name" validate:"required,max=255"`
	Child  string `json:"child" validate:"required,max=255"`
	Amount int    `json:"amount" validate:"required,min=0,max=1000"`
}

func handler(logger Logger) http.Handler {
	mux := http.NewServeMux()

	// Save data
	mux.HandleFunc("/save", saveData(logger))

	// Root
	mux.Handle("/", public.Files())

	return mux
}

func saveData(logger Logger) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: reading request body: %v", err), http.StatusBadRequest)
			return
		}

		p := payload{}

		if err := json.Unmarshal(body, &p); err != nil {
			http.Error(w, fmt.Sprintf("Error: decoding request: %v", err), http.StatusBadRequest)
			return
		}

		v := validator.New()
		if err := v.Struct(p); err != nil {
			http.Error(w, fmt.Sprintf("Error: invalid request: %v", err), http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(p)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: encoding request: %v", err), http.StatusInternalServerError)
		}

		f, err := os.OpenFile(DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write(append(data, "\n"...)); err != nil {
			f.Close() // ignore error; Write error takes precedence
			http.Error(w, fmt.Sprintf("Error: writing to database file: %v", err), http.StatusInternalServerError)
		}
		if err := f.Close(); err != nil {
			http.Error(w, fmt.Sprintf("Error: closing database file: %v", err), http.StatusInternalServerError)
		}
		logger.Printf("Request successfully processed.")
	}
}
