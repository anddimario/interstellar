package deploy

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	// config "github.com/anddimario/interstellar/internal/config"
	"github.com/spf13/viper"
)

var (
	// CheckReleaseDone is a channel to stop the ticker
	CheckReleaseDone = make(chan bool)
)

func downloadRelease(repo string, release string) {
	fmt.Printf("Last release: %s\n", release)
	// Command to execute
	cmd := exec.Command("gh", "release", "download", release, "--repo", repo, "-A", "tar.gz", "--skip-existing", "--dir", "/tmp") // @todo: use config for tmp dir
	
    // Create buffers to capture stdout and stderr
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

	// Run the command without capturing its output
	err := cmd.Run()
	if err != nil {
		log.Printf("Error in download: %s\n", err)
		log.Printf("Stdout: %s\n", stdout.String())
        log.Printf("Stderr: %s\n", stderr.String())
		return
	}
	// config.StoreConfig(repo + ".last_release", release)

	go StartDeploy()
}

func getLastRelease(repo string) {
	// Command to execute
	cmd := exec.Command("gh", "release", "list", "--repo", repo, "--limit", "1")

	// Run the command without capturing its output
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	outputString := string(output)
	releaseInfo := strings.Fields(outputString)

	lastDeployedRelease := viper.GetString(repo + ".last_release")

	switch lastDeployedRelease {
	case releaseInfo[0]:
		fmt.Println("No new release")
		return
	case "":
		fmt.Println("First time checking for release")
		downloadRelease(repo, releaseInfo[0])
		return
	default:
		fmt.Println("New release available")
		downloadRelease(repo, releaseInfo[0])
		return
	}
}

func CheckRelease(checkReleaseInterval time.Duration, repo string) {
	t := time.NewTicker(checkReleaseInterval * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			getLastRelease(repo)
		case <-CheckReleaseDone:
			return
		}
	}
}
