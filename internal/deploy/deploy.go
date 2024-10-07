package deploy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	// config "github.com/anddimario/interstellar/internal/config"
)

func StartDeploy(releaseFilePath string, executableCommand string, executableEnv []string, executableArgs []string) {
	executablePath := releaseFilePath + "/" + executableCommand

	cmd := exec.Command(executablePath, executableArgs...)

	// Set the environment variables
	cmd.Env = append(os.Environ(), executableEnv...)

	log.Printf("Starting release: %s\n", executablePath)

	// Configure the command to detach from the parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Redirect stdout and stderr to files
	stdoutFile, err := os.Create("stdout.log") // @todo: use a log package, or define dir in config
	if err != nil {
		fmt.Printf("Error creating stdout file: %s\n", err)
		return
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create("stderr.log") // @todo: use a log package, or define dir in config
	if err != nil {
		fmt.Printf("Error creating stderr file: %s\n", err)
		return
	}
	defer stderrFile.Close()

	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the command
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %s\n", err)
		return
	}

	// Print the PID of the detached process
	fmt.Printf("Detached process started with PID %d\n", cmd.Process.Pid)

	postDeploy()
}

func postDeploy() {

	// config.StoreConfig(repo + ".last_release", release)

}
