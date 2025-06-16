package bootstrap

import (
	"context"
	"github.com/gin-gonic/gin"
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

func WebApi(engine *gin.Engine, cfg *config.Config) {
	g := engine.Group("/api")
	g.Use(authMiddleware(config.LoadConfig().AuthToken))
	g.POST("/add-allowed-name", func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		store.AddAllowedName(req.Name)
		c.JSON(200, gin.H{
			"success": true,
			"message": "name added to allowed list",
		})
	})

	g.GET("/all-metrics", func(c *gin.Context) {
		metrics := store.GetAllMetrics()
		c.JSON(200, gin.H{
			"success": true,
			"metrics": metrics,
		})
	})

	g.GET("/alive", func(c *gin.Context) {
		result := store.AliveStatus(cfg.OfflineThresholdSec)
		c.JSON(200, gin.H{
			"state": result,
		})
	})
}

func authMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"success": false,
				"message": "unauthorized",
			})
			return
		}
		if authHeader != token {
			c.AbortWithStatusJSON(403, gin.H{
				"success": false,
				"message": "forbidden",
			})
			return
		}
		c.Next()
	}
}
