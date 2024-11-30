package peer

import (
	"testing"
)

// TestIsAvailableToServe tests the IsAvailableToServe function.
func TestIsAvailableToServe(t *testing.T) {

	// Test when CPU load is below the allowed threshold
	if !IsAvailableToServe(50, 50) {
		t.Errorf("Expected peer to be available to serve")
	}

	// Test when CPU load is above the allowed threshold
	if IsAvailableToServe(0.5, 0.5) {
		t.Errorf("Expected peer to not be available to serve")
	}
}

// TestGetCPULoad tests the getCPULoad function.
func TestGetCPULoad(t *testing.T) {

	// Test the getCPULoad function
	load, err := getCPULoad()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if load <= 0 {
		t.Errorf("Expected positive CPU load, got %f", load)
	}
}

// TestCalculateCPUUtilization tests the calculateCPUUtilization function.
func TestCalculateCPUUtilization(t *testing.T) {
	initialStats := []uint64{100, 200, 300, 400, 500, 600, 700, 800}
	finalStats := []uint64{200, 300, 400, 500, 600, 700, 800, 900}

	utilization := calculateCPUUtilization(initialStats, finalStats)
	if utilization <= 0 {
		t.Errorf("Expected positive CPU utilization, got %f", utilization)
	}
}

// TestSum tests the sum function.
func TestSum(t *testing.T) {
	stats := []uint64{100, 200, 300, 400, 500, 600, 700, 800}
	expected := uint64(3600)

	total := sum(stats)
	if total != expected {
		t.Errorf("Expected %d, got %d", expected, total)
	}
}
