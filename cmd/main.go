package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jalilbengoufa/ptp/internal"
	"github.com/jalilbengoufa/ptp/internal/database/postgres"
	"github.com/jalilbengoufa/ptp/internal/database/sqlite"
	"github.com/jalilbengoufa/ptp/internal/messaging"
	"github.com/jalilbengoufa/ptp/internal/server"
)

var (
	addr = flag.String("addr", ":8080", "http service address")
)

func main() {
	filename := "file.txt"
	internal.Filename = filename

	go closeOnSignal()

	_, err := sqlite.InitSqliteDbInstance()
	if err != nil {
		fmt.Printf("Failed to connect to the database sqlite: %v", err)
	}

	_, errPostgres := postgres.InitPostgresDbInstance()
	if errPostgres != nil {
		fmt.Printf("Failed to connect to the database postgres: %v", err)
	}

	go messaging.InitKafkaProducer()
	go messaging.InitKafkaConsumer()
	go sqlite.PullDataFromSqlite()

	http.HandleFunc("/", server.ServeHome)
	http.HandleFunc("/ws", server.ServeWs)
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func closeOnSignal() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	os.Exit(0)
}
