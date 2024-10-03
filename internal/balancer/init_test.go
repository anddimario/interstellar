package balancer

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/spf13/viper"
// )

// func TestMain(m *testing.M) {
// 	// beforeAll Logic here, example
// 	// Setup DB Connection
// 	// Seed data, etc
// 	// Create a mock backend server
// 	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("Hello, World!"))
// 	}))
// 	defer mockBackend.Close()

// 	// Set up the backends with the mock server URL
// 	backends := []string{mockBackend.URL}

// 	viper.Set("balancer.backends", backends)

// 	go HealthCheck(10 * time.Second, backends)

// 	time.Sleep(5 * time.Second)

// 	m.Run() // run all the test function

// 	// afterAll logic here, example
// 	// Close DB connection
// 	// Cleanup DB, etc
// 	HealthCheckDone <- true
// }
