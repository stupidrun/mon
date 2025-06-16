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
	// server side example
	// ---------------------
	//cfg := config.LoadConfig()
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	//go func() {
	//	err := bootstrap.Serve(ctx, cfg)
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
	//
	//go func() {
	//	store := bootstrap.Store()
	//	for {
	//		select {
	//		case <-time.After(time.Hour * time.Duration(cfg.CleanupIntervalHours)):
	//			log.Println("cleanup metrics store")
	//			store.Cleanup()
	//		case <-ctx.Done():
	//			return
	//		}
	//	}
	//}()
	//
	//go func() {
	//	gin.SetMode(gin.ReleaseMode)
	//	e := gin.Default()
	//	bootstrap.WebApi(e, cfg)
	//	srv := http.Server{
	//		Addr:    ":7920",
	//		Handler: e,
	//	}
	//
	//	log.Println("Starting HTTP server on :7920")
	//	go func() {
	//		<-ctx.Done()
	//		if err := srv.Shutdown(context.Background()); err != nil {
	//			log.Printf("HTTP server shutdown failed: %v", err)
	//		} else {
	//			log.Println("HTTP server shutdown gracefully")
	//		}
	//	}()
	//
	//	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	//		log.Fatalf("listen: %s\n", err)
	//	}
	//}()

	// client side example
	// ---------------------
	clientName := flag.String("n", "", "Name of the client")
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
	go bootstrap.StartPeriodicMetricsCollection(ctx, client, 10*time.Second)

	// common shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")
	<-time.After(5 * time.Second)
	log.Println("shutting down complete")
}
