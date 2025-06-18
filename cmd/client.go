package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/stupidrun/mon/bootstrap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	clientName := flag.String("n", "", "Name of the client")
	interval := flag.Int("i", 10, "Metrics collection interval in seconds")
	flag.Parse()

	if *clientName == "" {
		log.Fatal("Client name must be provided using -n flag")
	}

	serverHost := os.Getenv("MONITORING_SERVER_HOST")
	client, err := bootstrap.NewClient(fmt.Sprintf("%s:37322", serverHost), *clientName)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bootstrap.StartPeriodicMetricsCollection(ctx, client, *interval)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")
	<-time.After(5 * time.Second)
	log.Println("shutting down complete")
}
