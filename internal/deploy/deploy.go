package deploy

import (
	"bytes"
	"fmt"
	"log/slog"
	"math/rand"
	"os/exec"
	"strings"
	"sync"
	"time"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	config "github.com/anddimario/interstellar/internal/config"
	"github.com/spf13/viper"
)

type DeployStatus struct {
	Progress bool
	mu       sync.Mutex
}

var (
	Status DeployStatus
)

func StartDeploy(deployConfig config.DeployConfig, releaseVersion string) {

	Status.mu.Lock()
	defer Status.mu.Unlock()
	if Status.Progress {
		slog.Warn("Deploy in progress, skipping...\n")
		return
	}
	Status.Progress = true
	config.StoreValueInConfig("deploy.in_progress", true)

	newReleasProcess, err := LaunchNewVersion(deployConfig, releaseVersion)
	if err != nil {
		slog.Error("Launching new version", "err", err)
		return
	}

	if deployConfig.Type == "canary" {
		go canaryDeploy(newReleasProcess.Port, newReleasProcess.PID, deployConfig.Repo, releaseVersion)
	} else {
		go blueGreenDeploy(newReleasProcess.Port, newReleasProcess.PID, deployConfig.Repo, releaseVersion)
	}

}

func canaryDeploy(processPort int, newProcessPID int, repo string, releaseVersion string) {
	balancer.ManageCanaryDeployInProgress()

	newBackend := fmt.Sprintf("http://localhost:%d", processPort)

	// wait for the new version to start
	waitStartup := viper.GetInt("canary.wait_startup_in_sec") // @todo: see if inject
	time.Sleep(time.Duration(waitStartup) * time.Second)

	// see if the new version it's healthy (if not close the new version and go to notify)
	newVersionIsHealthy, err := balancer.GetHealthyBackend(newBackend)
	if err != nil {
		slog.Error("Checking health of new version", "err", err)
		return
	}

	if !newVersionIsHealthy {
		slog.Error("New version is not healthy, removing...")
		// remove new version with problems
		balancer.RemoveProcesses([]int{newProcessPID})

		// @todo: notify, where?
		return
	}

	// set the backends for the canary deploy
	balancer.AddCanaryBackend(newBackend)

	// get old version processes
	oldVersionProcessesPID, err := balancer.GetProcessesPID()
	if err != nil {
		slog.Error("Getting old version processes", "err", err)
		return
	}

	addBackendToConfig(processPort)

	// use a timer timeout that wait until the window time is over and compleate the deploy
	canaryWaitWindow := viper.GetDuration("canary.wait_window_in_min") // @todo: see if inject
	t := time.NewTimer(canaryWaitWindow * time.Minute)

	<-t.C
	// remove the old version
	balancer.RemoveProcesses(oldVersionProcessesPID)

	balancer.ManageCanaryDeployCompleted()

	go postDeploy(repo, releaseVersion)

}

func blueGreenDeploy(processPort int, newProcessPID int, repo string, releaseVersion string) {

	// wait for positive health check
	positiveHealthCheck := viper.GetInt("bluegreen.positive_healthchecks") // @todo: see if inject
	healthCheckInterval := viper.GetDuration("healthcheck.interval")       // @todo: see if inject
	time.Sleep(time.Duration(positiveHealthCheck) * time.Duration(healthCheckInterval) * time.Second)

	newBackend := fmt.Sprintf("http://localhost:%d", processPort)

	// see if the new version it's healthy (if not close the new version and go to notify)
	newVersionIsHealthy, err := balancer.GetHealthyBackend(newBackend)
	if err != nil {
		slog.Error("Checking health of new version", "err", err)
		return
	}

	if !newVersionIsHealthy {
		slog.Warn("New version is not healthy, removing...")
		// remove new version with problems
		balancer.RemoveProcesses([]int{newProcessPID})

		// @todo: notify, where?
		return
	}

	// get old version processes
	oldVersionProcessesPID, err := balancer.GetProcessesPID()
	if err != nil {
		slog.Error("Getting old version processes", "err", err)
		return
	}

	// replace the backends with the new version
	replaceBackendInConfig([]string{newBackend})

	// kill old version
	balancer.RemoveProcesses(oldVersionProcessesPID)

	// @todo: notify, where?

	go postDeploy(repo, releaseVersion)

}

func postDeploy(repo string, release string) {

	config.StoreValueInConfig(repo+".last_release", release)
	// reset the ignore release, and other temp values
	config.DeleteValueInConfig(repo + ".ignore")
	config.DeleteValueInConfig("deploy.in_progress")
	config.DeleteValueInConfig("deploy.new_process_pid")
	config.DeleteValueInConfig("deploy.new_process_port")

	Status.mu.Lock()
	Status.Progress = false
	Status.mu.Unlock()
}

func chooseNextReleasePort() (int, error) {
	// Choose a random port number between 1024 and 65535
	min := 1024
	max := 65535

	// Generate a random number within the range
	randomNumber := rand.Intn(max-min+1) + min

	cmd := exec.Command("ss", "-tuln")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		slog.Error("Executing ss to choose release port", "err", err)
		return 0, err
	}
	// Parse the command output to find the port
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, string(randomNumber)) {
			return chooseNextReleasePort()
		}
	}

	return randomNumber, nil
}

func addBackendToConfig(port int) {
	backendUrl := fmt.Sprintf("http://localhost:%d", port)

	backends := viper.GetStringSlice("balancer.backends")
	backends = append(backends, backendUrl)

	// add backend to balancer and config
	balancer.UpdateBackends(backends)
}

func replaceBackendInConfig(newBackends []string) {
	// add backend to balancer and config
	balancer.UpdateBackends(newBackends)
}

func CheckIfDeployInProgress() bool {
	Status.mu.Lock()
	defer Status.mu.Unlock()
	return Status.Progress
}
