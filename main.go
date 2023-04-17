package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/blang/mpv"
)

var (
	logsPath   string
	port       string
	socketPath string
	healthy    int32
	mpvc       mpv.Client
)

const (
	DEFAULT_SOCKET_PATH = "/tmp/mpv.sock"
	DEFAULT_LOG_FILE    = "/tmp/mpv-ctl.log"
	DEFAULT_PORT        = "5000"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}

type ResponseModel struct {
	Ok bool `json:"ok"`
}

func init() {
	flag.StringVar(&socketPath, "s", DEFAULT_SOCKET_PATH, "socket path")
	flag.StringVar(&logsPath, "l", DEFAULT_LOG_FILE, "log file path")
	flag.StringVar(&port, "p", DEFAULT_PORT, "port")

	flag.Parse()
	mpvc = mpv.Client{LLClient: mpv.NewIPCClient(socketPath)}
}

func main() {
	logFile, err := os.OpenFile(logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %s", err)
	} else {
		log.SetOutput(logFile)
	}
	defer logFile.Close()

	router := http.NewServeMux()
	router.HandleFunc("/append", appendHandler)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      middleHandler()(router),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		log.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	log.Printf("Server is ready to handle requests at :%s", port)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", port, err)
	}

	<-done
	log.Println("Server stopped")
}

func writeDefaultResponse(w *http.ResponseWriter) {
	if err := json.NewEncoder(*w).Encode(ResponseModel{true}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}
func writeError(w *http.ResponseWriter, msg string) {
	if err := json.NewEncoder(*w).Encode(ErrorResponse{msg, false}); err != nil {
		http.Error(*w, err.Error(), http.StatusInternalServerError)
	}
}

func middleHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			log.Printf("%s %s\n", r.Method, r.RequestURI)
			if r.Method != http.MethodGet {
				writeError(&w, "Method Not Allowed")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func appendHandler(w http.ResponseWriter, r *http.Request) {
	if !r.URL.Query().Has("path") {
		writeError(&w, "'path' param should be present")
		return
	}
	if err := mpvc.Loadfile(r.URL.Query().Get("path"), mpv.LoadFileModeAppendPlay); err != nil {
		errorMsg := fmt.Sprintf("append error: %s", err.Error())
		log.Println(errorMsg)
		writeError(&w, errorMsg)
	} else {
		writeDefaultResponse(&w)
	}
}
