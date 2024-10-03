package balancer

// import (
// 	"log"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/spf13/viper"
// )

// func TestHandleRequest(t *testing.T) {

// 	url := viper.GetString("server.address")
// 	log.Println(url)
// 	// Create a mock request
// 	mockRequest, err := http.NewRequest(http.MethodGet, url, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Create a mock response writer
// 	mockResponseWriter := httptest.NewRecorder()

// 	// Call the function being tested
// 	HandleRequest(mockResponseWriter, mockRequest)

// 	// Verify the response
// 	response := mockResponseWriter.Result()
// 	if response.StatusCode != http.StatusOK {
// 		t.Errorf("Expected status code %d, got %d", http.StatusOK, response.StatusCode)
// 	}

// 	// Verify the response body
// 	expectedBody := "Hello, World!"
// 	actualBody := mockResponseWriter.Body.String()
// 	if actualBody != expectedBody {
// 		t.Errorf("Expected response body %q, got %q", expectedBody, actualBody)
// 	}
// }
