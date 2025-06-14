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
		debug: true,
	}
}

func (s *MonitoringService) PushMetrics(ctx context.Context, req *proto.PushMetricsRequest) (*proto.PushMetricsResponse, error) {
	clientIP, ok := middlewares.GetClientIP(ctx)
	if !ok {
		return nil, errors.New("could not extract client IP from context")
	}
	isAllowed := s.store.IsIPAllowed(clientIP)
	if !isAllowed {
		log.Printf("IP %s is not allowed to push metrics", clientIP)
		return &proto.PushMetricsResponse{
			Success: false,
			Message: "IP not allowed",
		}, nil
	}
	for _, metric := range req.Metrics {
		s.store.AddMetric(clientIP, models.Metric{
			CPUUsage:    metric.CpuUsage,
			MemoryUsage: metric.MemoryUsage,
			NetworkIn:   metric.NetworkIn,
			NetworkOut:  metric.NetworkOut,
			Timestamp:   metric.Timestamp,
		})
	}
	if s.debug {
		log.Printf("Received metrics from %s: %v", clientIP, req.Metrics)
		log.Printf("Current metrics for %s: %v", clientIP, s.store.GetMetrics(clientIP))
	}
	return &proto.PushMetricsResponse{
		Success: true,
		Message: "Metrics pushed successfully",
	}, nil
}
