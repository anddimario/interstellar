package balancer

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/anddimario/interstellar/internal/config"
)

// healthyBackends holds the result of the ticker
type healthyBackends struct {
	Value []string
	mu    sync.Mutex
}

type HealthCheckConfig struct {
	Interval time.Duration
	Path     string
}

var (
	healthCheckConfig HealthCheckConfig
	Result          healthyBackends
	HealthCheckDone = make(chan bool)
)

func (c HealthCheckConfig) InitBackendsFromConfig(backends []string) {
	Result.mu.Lock()
	defer Result.mu.Unlock()
	Result.Value = backends
	healthCheckConfig = c
	go healthCheck(c.Interval * time.Second)
}

func healthCheck(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			Result.mu.Lock()
			Result.Value = GetHealthyBackends(Result.Value)
			// slog.Info("Healthy backends", "list", Result.Value) // @todo: remove
			Result.mu.Unlock()
		case <-HealthCheckDone:
			return
		}
	}
}

// GetResult safely returns the current value of Result
func GetBackends() ([]string, error) {
	Result.mu.Lock()
	defer Result.mu.Unlock()
	if len(Result.Value) == 0 {
		return nil, errors.New("no healthy backends")
	}
	return Result.Value, nil
}

// UpdateBackends safely updates the value of Result when a new backend is added
func UpdateBackends(backends []string) {
	Result.mu.Lock()
	defer Result.mu.Unlock()
	Result.Value = backends
	// slog.Info("Updated backends", "list", Result.Value) // @todo: remove

	// Update the config too to keep it in sync
	config.StoreConfig("balancer.backends", backends)
}

func GetHealthyBackend(backend string) (bool, error) {
	completeBackendPath := backend + healthCheckConfig.Path

	req, err := http.NewRequest("GET", completeBackendPath, nil)
	if err != nil {
		slog.Error("Error creating request", "err", err)
		return false, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Sending request", "err", err)
		return false, err
	}
	defer resp.Body.Close()

	// slog.Info("HealthCheck Status Code", "backend", completeBackendPath, "status", resp.StatusCode) // @todo: remove
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func GetHealthyBackends(backends []string) []string {

	healthyBackends := make([]string, 0)
	needReplaceInConfig := false

	for _, backend := range backends {
		if ok, _ := GetHealthyBackend(backend); ok {
			healthyBackends = append(healthyBackends, backend)
		} else {
			// check if the process is still running, if not remove from the list
			processPID, err := GetProcessPID(backend)
			if err != nil {
				slog.Error("Parsing PID from backend URL", "err", err)
			}
			if processPID == 0 {
				// Remove the backend from the list
				slog.Info("Removing unhealthy backend", "backend", backend)
				needReplaceInConfig = true
			}
		}
	}

	if needReplaceInConfig {
		config.StoreConfig("balancer.backends", healthyBackends)
	}

	return healthyBackends
}
