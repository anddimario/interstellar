package deploy

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	config "github.com/anddimario/interstellar/internal/config"
	"github.com/spf13/viper"
)

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
		log.Printf("Error executing ss: %s\n", err)
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

func StartDeploy(deployConfig config.DeployConfig, releaseVersion string) {

	processPort, err := chooseNextReleasePort()
	if err != nil {
		log.Printf("Error choosing port: %s\n", err)
		return
	}

	executablePath := deployConfig.ReleasePath + "/" + deployConfig.ExecutableCommand

	cmd := exec.Command(executablePath, deployConfig.ExecutableArgs...)

	processEnvVariable := append(deployConfig.ExecutableEnv, fmt.Sprintf("PORT=%d", processPort))

	// Set the environment variables
	cmd.Env = append(os.Environ(), processEnvVariable...)

	log.Printf("Starting release: %s\n", executablePath)

	// Configure the command to detach from the parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Redirect stdout and stderr to files
	stdoutFile, err := os.Create("stdout.log") // @todo: use a log package, or define dir in config
	if err != nil {
		log.Printf("Error creating stdout file: %s\n", err)
		return
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create("stderr.log") // @todo: use a log package, or define dir in config
	if err != nil {
		log.Printf("Error creating stderr file: %s\n", err)
		return
	}
	defer stderrFile.Close()

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the command
	err = cmd.Start()
	if err != nil {
		log.Printf("Error starting command: %s\n", err)
		return
	}

	// Print the PID of the detached process
	log.Printf("Detached process started with PID %d\n", cmd.Process.Pid)

	if deployConfig.Type == "canary" {
		canaryDeploy(processPort)
	} else {
		blueGreenDeploy(processPort)
	}

	postDeploy(deployConfig.Repo, releaseVersion)

}

func canaryDeploy(processPort int) {

	addBackendToConfig(processPort)
}

func blueGreenDeploy(processPort int) {
	// wait for positive health check
	positiveHealthCheck := viper.GetInt("bluegreen.positive_healthchecks") // @todo: see if inject
	healthCheckInterval := viper.GetDuration("healthcheck.interval") // @todo: see if inject
	time.Sleep(time.Duration(positiveHealthCheck) * time.Duration(healthCheckInterval) * time.Second)

	newBackend := fmt.Sprintf("http://localhost:%d", processPort)

	// see if the new version it's healthy (if not close the new version and go to notify)
	newVersionIsHealthy, err := balancer.GetHealthlyBackend(newBackend)
	if err != nil {
		log.Printf("Error checking health of new version: %s\n", err)
		return
	}

	if !newVersionIsHealthy {
		log.Printf("New version is not healthy, removing...\n")
		// @todo: remove new version

		// @ todo: notify, where?
		return
	}

	// replace the backends with the new version
	replaceBackendInConfig([]string{newBackend})

	// @todo: remove old version

	// @todo: notify, where?

}

func postDeploy(repo string, release string) {

	config.StoreConfig(repo + ".last_release", release)

}
