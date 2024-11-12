package deploy

import (
	"log/slog"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	config "github.com/anddimario/interstellar/internal/config"
)

func RecoveryFromCrash(recoveryConfig config.RecoveryConfig) {
	if recoveryConfig.DeployInProgress {
		slog.Warn("Recovery from crash a deploy in progress...\n")
		// remove the process, and the healthcheck will remove backend from the list
		balancer.RemoveProcesses([]int{recoveryConfig.NewProcessPID})

		config.DeleteValueInConfig("deploy.in_progress")
		config.DeleteValueInConfig("deploy.new_process_pid")
		config.DeleteValueInConfig("deploy.new_process_port")

	}
}
