package services

import (
	"context"
	"errors"
	"github.com/stupidrun/mon/api/middlewares"
	"github.com/stupidrun/mon/api/proto"
	"github.com/stupidrun/mon/models"
	"log"
)

type MonitoringService struct {
	proto.UnimplementedMonitoringServiceServer
	store *models.MetricsStore
	debug bool
}

func NewMonitoringService(store *models.MetricsStore, debug bool) *MonitoringService {
	return &MonitoringService{
		store: store,
		debug: debug,
	}
}

func (s *MonitoringService) PushMetrics(ctx context.Context, req *proto.PushMetricsRequest) (*proto.PushMetricsResponse, error) {
	clientIP, ok := middlewares.GetClientIP(ctx)
	if !ok {
		return nil, errors.New("could not extract client IP from context")
	}
	for _, metric := range req.Metrics {
		if metric.Name == "" {
			return nil, errors.New("metric name cannot be empty")
		}
		if !s.store.IsAllowed(metric.Name) {
			continue
		}
		s.store.AddMetric(metric.Name, models.Metric{
			Name:        metric.Name,
			IP:          clientIP,
			CPUUsage:    metric.CpuUsage,
			MemoryUsage: metric.MemoryUsage,
			NetworkIn:   metric.NetworkIn,
			NetworkOut:  metric.NetworkOut,
			Timestamp:   metric.Timestamp,
		})
	}
	if s.debug {
		log.Printf("Received metrics from %s: %v", clientIP, req.Metrics)
		log.Printf("Current metrics for %s: %v", req.Metrics[0].Name, s.store.GetMetrics(req.Metrics[0].Name))
	}
	return &proto.PushMetricsResponse{
		Success: true,
		Message: "Metrics pushed successfully",
	}, nil
}
