package config

import (
	"os"
	"strconv"
)

type Config struct {
	AuthToken            string
	CleanupIntervalHours int
	OfflineThresholdSec  int
	GrpcPort             string
	Debug                bool
}

func LoadConfig() *Config {
	cleanupInterval, _ := strconv.Atoi(getEnv("CLEANUP_INTERVAL_HOURS", "1"))
	offlineThreshold, _ := strconv.Atoi(getEnv("OFFLINE_THRESHOLD_SEC", "90"))

	return &Config{
		AuthToken:            getEnv("AUTH_TOKEN", "this_is_bullshit"),
		CleanupIntervalHours: cleanupInterval,
		OfflineThresholdSec:  offlineThreshold,
		GrpcPort:             ":37322",
		Debug:                getEnv("DEBUG", false),
	}
}

func getEnv[T string | int | bool](key string, defaultVal T) T {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultVal
	}
	switch any(defaultVal).(type) {
	case bool:
		return any(value == "true").(T)
	case string:
		return any(value).(T)
	case int:
		if intVal, err := strconv.Atoi(value); err == nil {
			return any(intVal).(T)
		}
	}
	return defaultVal
}
