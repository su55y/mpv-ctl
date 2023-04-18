package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/blang/mpv"
	mpvctl "github.com/su55y/mpv-ctl/internal/mpv-ctl"
)

var (
	logsPath   string
	port       string
	socketPath string
	healthy    int32
)

const (
	DEFAULT_SOCKET_PATH = "/tmp/mpv.sock"
	DEFAULT_LOG_FILE    = "/tmp/mpv-ctl.log"
	DEFAULT_PORT        = "5000"
)

func init() {
	flag.StringVar(&socketPath, "s", DEFAULT_SOCKET_PATH, "socket path")
	flag.StringVar(&logsPath, "l", DEFAULT_LOG_FILE, "log file path")
	flag.StringVar(&port, "p", DEFAULT_PORT, "port")

	flag.Parse()

}

func main() {
	logFile, err := os.OpenFile(logsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %s", err)
	} else {
		log.SetOutput(logFile)
	}
	defer logFile.Close()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mpvctl.GetNewHandler(&mpv.Client{LLClient: mpv.NewIPCClient(socketPath)}),
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
