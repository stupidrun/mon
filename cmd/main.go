package main

import (
	"context"
	"github.com/stupidrun/mon/bootstrap"
	"github.com/stupidrun/mon/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// server side example
	// ---------------------
	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := bootstrap.Serve(ctx, cfg)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		store := bootstrap.Store()
		for {
			select {
			case <-time.After(time.Hour * time.Duration(cfg.CleanupIntervalHours)):
				log.Println("cleanup metrics store")
				store.Cleanup()
			case <-ctx.Done():
				return
			}
		}
	}()

	// client side example
	// ---------------------
	//client, err := bootstrap.NewClient(":37322")
	//if err != nil {
	//	log.Fatalf("Failed to create client: %v", err)
	//}
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//go bootstrap.StartPeriodicMetricsCollection(ctx, client, 10*time.Second)

	// common shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")
	<-time.After(5 * time.Second)
	log.Println("shutting down complete")
}
