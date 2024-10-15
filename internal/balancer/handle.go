package balancer

import (
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
)

var (
	// Index for round-robin selection
	index uint32

	Canary bool
)

type Handler struct {
	backend string
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
	healthlyBackends, err := GetBackends()

	if err != nil {
		return err
	}

	// Canary deployment is different from normal deployment
	if Canary {
		// @todo
		slog.Info("Canary deployment in progress") // @todo remove
		// @todo see if there's a storage that initializes the counters
		// @todo get the counters
		// @todo see the new release quotas and check who process the last request
		// @todo if the last request was processed by the new release, send the request to the old release
		// @todo if the last request was processed by the old release, check if the new one must process other requests and send the request to the new release, otherwise send to the old release
		// @todo check if the old release must process other requests, if not, reset the counters

		i := atomic.AddUint32(&index, 1)
		h.backend = healthlyBackends[i%uint32(len(healthlyBackends))]
		return nil
	} else {
		// @todo see if use different algorithms
		i := atomic.AddUint32(&index, 1)
		h.backend = healthlyBackends[i%uint32(len(healthlyBackends))]
		return nil
	}
}

func ManageCanaryDeployInProgress() {

	// @todo: mutex?
	Canary = true

}

func ManageCanaryDeployCompleted() {
	// @todo reset the canary counters and release the mutex (use mutex)

	Canary = false

}