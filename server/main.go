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
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/normanjaeckel/Blitzumfrage/server/public"
	"golang.org/x/sys/unix"
)

const (
	Host            string = "localhost"
	Port            int    = 8000
	DataFile               = "data.jsonl"
	DataFileMaxSize        = 1000000
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
	mux := sync.Mutex{}

	s := &http.Server{
		Addr:    addr,
		Handler: handler(logger, &mux),
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

func handler(logger Logger, mux *sync.Mutex) http.Handler {
	serveMux := http.NewServeMux()

	// Save data
	serveMux.HandleFunc("/save", saveData(logger, mux))

	// Root
	serveMux.Handle("/", public.Files())

	return serveMux
}

func saveData(logger Logger, mux *sync.Mutex) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Setup mutual exclusion lock
		mux.Lock()
		defer mux.Unlock()

		// Check filesize
		fileInfo, err := os.Stat(DataFile)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: retrieving database file information: %v", err), http.StatusInternalServerError)
			return
		}
		if fileInfo.Size() > DataFileMaxSize {
			http.Error(w, "Error: database file is too large", http.StatusInternalServerError)
			return
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: reading request body: %v", err), http.StatusBadRequest)
			return
		}

		// Decode and validate body
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

		// Encode data
		data, err := json.Marshal(p)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: encoding request: %v", err), http.StatusInternalServerError)
		}

		// Write data to file
		f, err := os.OpenFile(DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: opening database file: %v", err), http.StatusInternalServerError)
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
