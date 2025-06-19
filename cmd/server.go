package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stupidrun/mon/bootstrap"
	"github.com/stupidrun/mon/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func main() {
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

	go func() {
		gin.SetMode(gin.ReleaseMode)
		e := gin.Default()
		e.Use(corsMiddleware())
		bootstrap.WebApi(e, cfg)
		srv := http.Server{
			Addr:    ":7920",
			Handler: e,
		}

		log.Println("Starting HTTP server on :7920")
		go func() {
			<-ctx.Done()
			if err := srv.Shutdown(context.Background()); err != nil {
				log.Printf("HTTP server shutdown failed: %v", err)
			} else {
				log.Println("HTTP server shutdown gracefully")
			}
		}()

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	log.Println("Received shutdown signal, shutting down gracefully...")
	<-time.After(5 * time.Second)
	log.Println("shutting down complete")
}
