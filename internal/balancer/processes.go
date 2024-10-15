package balancer

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/viper"
)

func GetProcessesPID() ([]int, error) {
	// Get the list of processes from the balancer
	backends := viper.GetStringSlice("balancer.backends") // @todo: see if inject

	// Create a slice to store the PIDs of the processes
	pids := make([]int, 0)

	// Iterate over the backends to extract the PIDs
	for _, backend := range backends {
		pid, err := getPID(backend)
		if err != nil {
			slog.Error("Parsing PID from backend URL", "err", err)
			return nil, err
		}

		// Append the PID to the slice
		pids = append(pids, pid)
	}

	return pids, nil
}

func RemoveProcesses(pids []int) {
	for _, pid := range pids {
		// Find the process by its PID
		process, err := os.FindProcess(pid)
		if err != nil {
			slog.Error("Process not found", "pid", pid, "err", err)
			return
		}

		// Send a termination signal to the process
		if err := process.Signal(syscall.SIGTERM); err != nil {
			slog.Error("Sending termination signal for process", "pid", pid, "err", err)
			return
		}

		// Print a message indicating that the process was terminated
		slog.Info("Process terminated", "pid", pid)
	}
}

func getPID(backend string) (int, error) {

	parts := strings.Split(backend, ":")

	if len(parts) < 2 {
		return 0, errors.New("invalid backend URL")
	}

	port := parts[len(parts)-1]

	cmd := exec.Command("ss", "-tulnp")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		slog.Error("Executing ss to get PID", "err", err)
		return 0, err
	}

	// Parse the command output to find the port
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, string(port)) {
			// Use a regular expression to extract the PID
			re := regexp.MustCompile(`pid=(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return strconv.Atoi(matches[1])
			}
		}
	}

	return 0, nil
}
