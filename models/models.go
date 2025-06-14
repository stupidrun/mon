package models

import (
	"slices"
	"sync"
)

type Metric struct {
	IP          string
	CPUUsage    float64
	MemoryUsage float64
	NetworkIn   float64
	NetworkOut  float64
	Timestamp   int64
}

type MetricsStore struct {
	mu         sync.RWMutex
	metrics    map[string][]Metric
	allowedIPs []string
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		metrics:    make(map[string][]Metric),
		allowedIPs: make([]string, 0, 20),
	}
}

func (ms *MetricsStore) AddAllowedIP(ip string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.IsIPAllowed(ip) {
		ms.allowedIPs = append(ms.allowedIPs, ip)
	}
}

func (ms *MetricsStore) IsIPAllowed(ip string) bool {
	return slices.Contains(ms.allowedIPs, ip)
}

func (ms *MetricsStore) AddMetric(ip string, metric Metric) {
	if !ms.IsIPAllowed(ip) {
		return
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.metrics[ip]; !exists {
		ms.metrics[ip] = make([]Metric, 0, 100)
	}

	ms.metrics[ip] = append(ms.metrics[ip], metric)
}

func (ms *MetricsStore) GetMetrics(ip string) []Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.IsIPAllowed(ip) {
		return nil
	}

	return ms.metrics[ip]
}

func (ms *MetricsStore) Cleanup() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for ip, metrics := range ms.metrics {
		if len(metrics) == 0 {
			delete(ms.metrics, ip)
			continue
		}

		// Keep only the last 100 metrics for each IP
		if len(metrics) > 100 {
			ms.metrics[ip] = metrics[len(metrics)-100:]
		}
	}
}

func (ms *MetricsStore) GetAllMetrics() map[string][]Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Return a copy of the metrics map to avoid external modifications
	metricsCopy := make(map[string][]Metric, len(ms.metrics))
	for ip, metrics := range ms.metrics {
		metricsCopy[ip] = append([]Metric(nil), metrics...)
	}
	return metricsCopy
}
