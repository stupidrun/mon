package models

import (
	"slices"
	"sync"
	"time"
)

type Metric struct {
	IP          string
	Name        string
	CPUUsage    float64
	MemoryUsage float64
	NetworkIn   float64
	NetworkOut  float64
	Timestamp   int64
}

type MetricsStore struct {
	mu           sync.RWMutex
	metrics      map[string][]Metric
	allowedNames []string
}

func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		metrics:      make(map[string][]Metric),
		allowedNames: make([]string, 0, 20),
	}
}

func (ms *MetricsStore) AddAllowedName(name string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.IsAllowed(name) {
		ms.allowedNames = append(ms.allowedNames, name)
	}
}

func (ms *MetricsStore) IsAllowed(name string) bool {
	return slices.Contains(ms.allowedNames, name)
}

func (ms *MetricsStore) AddMetric(name string, metric Metric) {
	if !ms.IsAllowed(name) {
		return
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.metrics[name]; !exists {
		ms.metrics[name] = make([]Metric, 0, 100)
	}

	ms.metrics[name] = append(ms.metrics[name], metric)
}

func (ms *MetricsStore) GetMetrics(name string) []Metric {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if !ms.IsAllowed(name) {
		return nil
	}

	return ms.metrics[name]
}

func (ms *MetricsStore) Cleanup() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for name, metrics := range ms.metrics {
		if len(metrics) == 0 {
			delete(ms.metrics, name)
			continue
		}

		// Keep only the last 100 metrics for each name
		if len(metrics) > 100 {
			ms.metrics[name] = metrics[len(metrics)-100:]
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

func (ms *MetricsStore) AliveStatus(threshold int) map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	status := make(map[string]interface{})
	for _, name := range ms.allowedNames {
		metrics, exists := ms.metrics[name]
		if !exists || len(metrics) == 0 {
			status[name] = map[string]interface{}{
				"alive": false,
				"ip":    "",
			}
			continue
		}

		// Check if the last metric is within the offline threshold
		lastMetric := metrics[len(metrics)-1]
		t := map[string]interface{}{
			"ip": lastMetric.IP,
		}
		if time.Now().UTC().Unix()-lastMetric.Timestamp <= int64(threshold) {
			t["alive"] = true
		} else {
			t["alive"] = false
		}
		status[name] = t
	}
	return status
}
