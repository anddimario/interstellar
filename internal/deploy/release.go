package deploy

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	config "github.com/anddimario/interstellar/internal/config"
)

type ProcessInfo struct {
	PID  int
	Port int
}

func LaunchNewVersion(deployConfig config.DeployConfig, releaseVersion string) (*ProcessInfo, error) {

	processPort, err := chooseNextReleasePort()
	if err != nil {
		slog.Error("Choosing port", "err", err)
		return nil, err
	}

	executablePath := deployConfig.ReleasePath + "/" + deployConfig.ExecutableCommand

	cmd := exec.Command(executablePath, deployConfig.ExecutableArgs...)

	processEnvVariable := append(deployConfig.ExecutableEnv, fmt.Sprintf("PORT=%d", processPort))

	// Set the environment variables
	cmd.Env = append(os.Environ(), processEnvVariable...)

	slog.Info("Starting release", "command", executablePath)

	// Configure the command to detach from the parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Redirect stdout and stderr to files
	stdoutFile, err := os.Create("stdout.log") // @todo: use a log package, or define dir in config
	if err != nil {
		slog.Error("Creating stdout file", "err", err)
		return nil, err
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create("stderr.log") // @todo: use a log package, or define dir in config
	if err != nil {
		slog.Error("Creating stderr file", "err", err)
		return nil, err
	}
	defer stderrFile.Close()

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the command
	err = cmd.Start()
	if err != nil {
		slog.Error("Starting command", "err", err)
		return nil, err
	}

	newProcessPID := cmd.Process.Pid

	// Store the process info in the config, useful for rollback and recovery from crash
	config.StoreValueInConfig("deploy.new_process_pid", newProcessPID)
	config.StoreValueInConfig("deploy.new_process_port", processPort)

	// Print the PID of the detached process
	slog.Info("Detached process started with PID", "pid", newProcessPID)

	return &ProcessInfo{
		PID:  newProcessPID,
		Port: processPort,
	}, nil
}