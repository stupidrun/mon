package bootstrap

import (
	"context"
	"github.com/stupidrun/mon/api/middlewares"
	"github.com/stupidrun/mon/api/proto"
	"github.com/stupidrun/mon/config"
	"github.com/stupidrun/mon/models"
	"github.com/stupidrun/mon/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

var store *models.MetricsStore

func init() {
	store = models.NewMetricsStore()
	store.AddAllowedIP("127.0.0.1")
}

func Store() *models.MetricsStore {
	return store
}

func Serve(ctx context.Context, c *config.Config) error {
	server := grpc.NewServer(grpc.UnaryInterceptor(middlewares.IPExtractorInterceptor))
	if c.Debug {
		log.Println("Debug mode is enabled")
		log.Printf("Auth Token: %s", c.AuthToken)
		log.Printf("Cleanup Interval: %d hours", c.CleanupIntervalHours)
		log.Printf("Offline Threshold: %d seconds", c.OfflineThresholdSec)
		log.Printf("gRPC Port: %s", c.GrpcPort)
		reflection.Register(server)
	}

	monSrv := services.NewMonitoringService(store, c.Debug)
	proto.RegisterMonitoringServiceServer(server, monSrv)
	lis, err := net.Listen("tcp", c.GrpcPort)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")
		server.GracefulStop()
		log.Println("gRPC server stopped")
	}()

	log.Println("Starting gRPC server on", c.GrpcPort)
	return server.Serve(lis)
}

func WebApi() {
}
