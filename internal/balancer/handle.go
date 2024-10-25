package balancer

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/spf13/viper"
)

var (
	// Index for round-robin selection
	index uint32

	// muCanarytex for canary deployment
	muCanary sync.Mutex

	ResultCanary CanaryInfo
)

type Handler struct {
	backend string
}

type CanaryInfo struct {
	NewReleaseProcessedRequests int
	TotalProcessedRequests      int
	InProgress                  bool
	NewIsLastUsedBacked         bool // used to allow the request to split
	Backends                    []string
}

// HTTP handler to forward requests to backend servers
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	handlerInfo := Handler{}
	err := handlerInfo.getNextBackend()
	if err != nil {
		slog.Error("Get next backend", "err", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	backend, err := url.Parse(handlerInfo.backend)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Create a new request to the backend server
	req, err := http.NewRequest(r.Method, backend.ResolveReference(r.URL).String(), r.Body)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	req.Header = r.Header

	// Forward the request to the backend server
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Forwarding request", "err", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy the response from the backend server to the client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Select a backend server using round-robin algorithm
func (h *Handler) getNextBackend() error {
	healthyBackends, err := GetBackends()

	if err != nil {
		return err
	}

	// Canary deployment is different from normal deployment
	if ResultCanary.InProgress {
		h.backend, err = getCanaryBackend(healthyBackends)
		if err != nil {
			return err
		}
	} else {
		// @todo see if use different algorithms
		i := atomic.AddUint32(&index, 1)
		h.backend = healthyBackends[i%uint32(len(healthyBackends))]
	}
	return nil
}

func ManageCanaryDeployInProgress() {
	slog.Info("Canary deploy in progress\n")
	muCanary.Lock()
	defer muCanary.Unlock()
	ResultCanary = CanaryInfo{
		InProgress:                  true,
		NewReleaseProcessedRequests: 0,
		TotalProcessedRequests:      0,
	}

}

func ManageCanaryDeployCompleted() {
	muCanary.Lock()
	defer muCanary.Unlock()

	// @todo reset the canary counters
	ResultCanary.InProgress = false
	ResultCanary.NewReleaseProcessedRequests = 0
	ResultCanary.TotalProcessedRequests = 0
	ResultCanary.Backends = nil
	ResultCanary.NewIsLastUsedBacked = false 

	slog.Info("Canary deploy completed\n")
}

func AddCanaryBackend(newReleaseBackend string) {
	muCanary.Lock()
	defer muCanary.Unlock()

	ResultCanary.Backends = append(ResultCanary.Backends, newReleaseBackend)
}

func GetCanaryDeployStatus() *CanaryInfo {
	muCanary.Lock()
	defer muCanary.Unlock()
	return &ResultCanary
}

func getCanaryBackend(healthyBackends []string) (string, error) {
	muCanary.Lock()
	defer muCanary.Unlock()

	newReleaseQuota := viper.GetInt("canary.new_release_quota") // @todo inject this value to avoid viper at each request

	ResultCanary.TotalProcessedRequests++
	log.Printf("canaryInfo: %v", ResultCanary)

	// @todo redefine the algorithm?
	// todo if quota <30 get one request on the new for each 2 requests on the old?
	// reset the counter if the quota is reached
	if ResultCanary.TotalProcessedRequests >= 100 {
		ResultCanary.NewReleaseProcessedRequests = 0
		ResultCanary.TotalProcessedRequests = 0
	}

	if ResultCanary.NewReleaseProcessedRequests < newReleaseQuota && len(ResultCanary.Backends) > 0 && !ResultCanary.NewIsLastUsedBacked {
		// Use this to allow the request to split
		ResultCanary.NewIsLastUsedBacked = true
		ResultCanary.NewReleaseProcessedRequests++
		i := atomic.AddUint32(&index, 1)
		return ResultCanary.Backends[i%uint32(len(ResultCanary.Backends))], nil
	} else {
		ResultCanary.NewIsLastUsedBacked = false
		i := atomic.AddUint32(&index, 1)
		return healthyBackends[i%uint32(len(healthyBackends))], nil
	}
}
