package balancer

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"sync/atomic"
)

// Index for round-robin selection
var index uint32

type Handler struct {
	backend string
}

// Select a backend server using round-robin algorithm
func (h *Handler) getNextBackend() (error) {
    healthlyBackends, err := GetBackends()

	if err != nil {
		return err
	}

	// @todo use round-robin algorithm to select the next backend server, or use a different algorithm
	// @todo for canary deployment, use a different algorithm to select the backend server, maybe we need to store some data
	i := atomic.AddUint32(&index, 1)
	h.backend = healthlyBackends[i%uint32(len(healthlyBackends))]
	return nil
}

// HTTP handler to forward requests to backend servers
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	handlerInfo := Handler{}
	err := handlerInfo.getNextBackend()
	if err != nil {
		log.Println("Error getting next backend:", err)
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
		log.Println("Error forwarding request:", err)
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
