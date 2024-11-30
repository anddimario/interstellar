package peer

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

func IsAvailableToServe(allowedMem float64, allowedCPU float64) bool {
	// logic to check if the peer is available to serve requests based on mem and cpu usage
    cpuLoad, err := getCPULoad()
    if err != nil {
        slog.Error("Error getting CPU load:", "err", err)
        return false
    }

    memUsage, err := getMemoryUsage()
    if err != nil {
        slog.Error("Error getting memory usage:", "err", err)
        return false
    }

    if cpuLoad < allowedCPU && memUsage < allowedMem {
        return true
    }
	return false
}

func getCPULoad() (float64, error) {
    // Get initial CPU stats
    initialStats, err := getCPUStats()
    if err != nil {
        // fmt.Println("Error getting initial CPU stats:", err)
        return 0.0, err
    }

    // Wait for a second
    time.Sleep(1 * time.Second)

    // Get CPU stats after a second
    finalStats, err := getCPUStats()
    if err != nil {
        // fmt.Println("Error getting final CPU stats:", err)
        return 0.0, err
    }

    // Calculate CPU utilization
    utilization := calculateCPUUtilization(initialStats, finalStats)
	return utilization, nil
}

func getMemoryUsage() (float64, error) {
    data, err := os.ReadFile("/proc/meminfo")
    if err != nil {
        return 0.0, err
    }

    // Parse the memory usage information
    lines := strings.Split(string(data), "\n")
    var totalMem, freeMem, buffers, cached uint64
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) < 2 {
            continue
        }
        value, err := strconv.ParseUint(fields[1], 10, 64)
        if err != nil {
            return 0.0, err
        }
        switch fields[0] {
        case "MemTotal:":
            totalMem = value
        case "MemFree:":
            freeMem = value
        case "Buffers:":
            buffers = value
        case "Cached:":
            cached = value
        }
    }

    // Calculate used memory
    usedMem := totalMem - freeMem - buffers - cached
    usedMemPercentage := (float64(usedMem) / float64(totalMem)) * 100.0
    return usedMemPercentage, nil
}

func getCPUStats() ([]uint64, error) {
    // Read the /proc/stat file
    data, err := os.ReadFile("/proc/stat")
    if err != nil {
        return nil, err
    }

    // Find the line starting with "cpu "
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        if strings.HasPrefix(line, "cpu ") {
            // Parse the CPU stats
            fields := strings.Fields(line)
            if len(fields) < 8 {
                return nil, fmt.Errorf("unexpected format in /proc/stat")
            }

            // Convert the fields to uint64
            var stats []uint64
            for _, field := range fields[1:8] {
                var value uint64
                fmt.Sscanf(field, "%d", &value)
                stats = append(stats, value)
            }
            return stats, nil
        }
    }

    return nil, fmt.Errorf("cpu stats not found in /proc/stat")
}

func calculateCPUUtilization(initialStats, finalStats []uint64) float64 {
    // Calculate the total and idle time
    initialTotal := sum(initialStats)
    finalTotal := sum(finalStats)
    initialIdle := initialStats[3] + initialStats[4]
    finalIdle := finalStats[3] + finalStats[4]

    // Calculate the delta values
    totalDelta := finalTotal - initialTotal
    idleDelta := finalIdle - initialIdle

    // Calculate the CPU utilization
    utilization := (1.0 - float64(idleDelta)/float64(totalDelta)) * 100.0
    return utilization
}

func sum(stats []uint64) uint64 {
    var total uint64
    for _, value := range stats {
        total += value
    }
    return total
}
