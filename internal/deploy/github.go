package deploy

import (
	"bytes"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	config "github.com/anddimario/interstellar/internal/config"
	"github.com/spf13/viper"
)

var (
	// CheckReleaseDone is a channel to stop the ticker
	CheckReleaseDone = make(chan bool)
)

func CheckRelease(deployConfig config.DeployConfig) {
	t := time.NewTicker(deployConfig.CheckReleaseInterval * time.Minute)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			getLastRelease(deployConfig)
		case <-CheckReleaseDone:
			return
		}
	}
}

func getLastRelease(deployConfig config.DeployConfig) {
	// Command to execute
	cmd := exec.Command("gh", "release", "list", "--repo", deployConfig.Repo, "--limit", "1")

	// Run the command without capturing its output
	output, err := cmd.Output()
	if err != nil {
		slog.Error("From command output gh", "err", err)
		return
	}

	outputString := string(output)
	releaseInfo := strings.Fields(outputString)
	releaseVersion := releaseInfo[0]

	lastDeployedRelease := viper.GetString(deployConfig.Repo + ".last_release")

	if lastDeployedRelease == releaseVersion {
		fmt.Println("No new release")
		return
	}
	// Could be a new release or a first time release
	// fmt.Println("New release available")
	downloadRelease(deployConfig.Repo, releaseVersion, deployConfig.ReleasePath, deployConfig.AssetName)
	decompressRelease(deployConfig.Repo, releaseVersion, deployConfig.ReleasePath)

	go StartDeploy(deployConfig, releaseVersion)

}

func decompressRelease(repo string, release string, releaseFilePath string) {
	// Decompress the downloaded file
	releaseFileCompletePath := fmt.Sprintf("%s/%s-%s.tar.gz", releaseFilePath, repo, release)
	cmd := exec.Command("tar", "-xvf", releaseFileCompletePath, "-C", releaseFilePath)

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		slog.Error("Error in decompress: %s\n", err)
		slog.Error("Stdout: %s\n", stdout.String())
		slog.Error("Stderr: %s\n", stderr.String())
		return
	}

}

func downloadRelease(repo string, release string, releaseFilePath string, assetName string) {
	// fmt.Printf("Last release: %s\n", release)
	releaseFileCompletePath := fmt.Sprintf("%s/%s-%s.tar.gz", releaseFilePath, repo, release)
	// Command to execute
	cmd := exec.Command("gh", "release", "download", release, "--repo", repo, "--pattern", assetName, "--skip-existing", "--output", releaseFileCompletePath)
	// cmd := exec.Command("gh", "release", "download", release, "--repo", repo, "-A", "tar.gz", "--skip-existing", "--output", releaseFileCompletePath)

	// Create buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		slog.Error("Error in download: %s\n", err)
		slog.Error("Stdout: %s\n", stdout.String())
		slog.Error("Stderr: %s\n", stderr.String())
		return
	}
}
