package balancer

import (
	"bytes"
	"errors"
	"log"
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
			log.Printf("Error parsing PID from backend URL: %s\n", err)
			return nil, err
		}

		log.Printf("Found process with pid %d\n", pid) // @todo: remove
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
			log.Printf("Error finding process with pid %d: %s\n", pid, err)
			return
		}

		// Send a termination signal to the process
		if err := process.Signal(syscall.SIGTERM); err != nil {
			log.Printf("Error sending termination signal for process with pid %d: %s\n", pid, err)
			return
		}

		// Print a message indicating that the process was terminated
		log.Printf("Process with pid %d terminated", pid)
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
		log.Printf("Error executing ss: %s\n", err)
		return 0, err
	}

	// Parse the command output to find the port
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, string(port)) {
			log.Printf("Found line: %s\n", line) // @todo: remove
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
