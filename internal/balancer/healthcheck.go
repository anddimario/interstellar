package balancer

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/anddimario/interstellar/internal/config"
)

// HealthlyBackends holds the result of the ticker
type HealthlyBackends struct {
	Value []string
}

var (
	Result          HealthlyBackends
	mu              sync.Mutex
	HealthCheckDone = make(chan bool)
)

func InitBackendsFromConfig(backends []string) {
	mu.Lock()
	Result.Value = backends
	mu.Unlock()

	// @todo: start here the healthcheck, or check the status?
}

func HealthCheck(interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			mu.Lock()
			Result.Value = getHealthlyBackends(Result.Value)
			slog.Error("Healthy backends", "list", Result.Value) // @todo: remove
			mu.Unlock()
		case <-HealthCheckDone:
			return
		}
	}
}

// GetResult safely returns the current value of Result
func GetBackends() ([]string, error) {
	mu.Lock()
	defer mu.Unlock()
	if len(Result.Value) == 0 {
		return nil, errors.New("no healthy backends")
	}
	return Result.Value, nil
}

// UpdateBackends safely updates the value of Result when a new backend is added
func UpdateBackends(backends []string) {
	mu.Lock()
	defer mu.Unlock()
	Result.Value = backends
	slog.Info("Updated backends", "list", Result.Value) // @todo: remove

	// Update the config too to keep it in sync
	config.StoreConfig("balancer.backends", backends)
}

func GetHealthlyBackend(backend string) (bool, error) {
	req, err := http.NewRequest("GET", backend, nil)
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

	slog.Error("HealthCheck Status Code for %s: %d", backend, resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}

func getHealthlyBackends(backends []string) []string {
	healthyBackends := make([]string, 0)
	for _, backend := range backends {
		req, err := http.NewRequest("GET", backend, nil)
		if err != nil {
			slog.Error("Error creating request", "err", err)
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Sending request", "err", err)
			continue
		}
		defer resp.Body.Close()

		slog.Error("HealthCheck Status Code for %s: %d", backend, resp.StatusCode)
		if resp.StatusCode == http.StatusOK {
			healthyBackends = append(healthyBackends, backend)
		}

	}

	return healthyBackends
}
