package deploy

import (
	"fmt"
	"log/slog"
	"os/exec"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	config "github.com/anddimario/interstellar/internal/config"
)

func Rollback(releaseVersion string) {
	slog.Info("Rollback to release", "version", releaseVersion)

	deployIsInProgress := CheckIfDeployInProgress()

	if deployIsInProgress {
		slog.Error("Deploy in progress, rollback not allowed")
		return
	}

	deployConfig := config.PrepareDeployConfig()
	// Get last release info for the old version
	lastReleaseConfig := config.PrepareReleaseConfig(deployConfig.Repo)

	releaseExists := checkIfReleaseExists(deployConfig.Repo, releaseVersion)
	if !releaseExists {
		slog.Error("Release does not exist", "releaseVersion", releaseVersion)
		return
	}

	DownloadRelease(deployConfig.Repo, releaseVersion, deployConfig.ReleasePath, deployConfig.AssetName)
	DecompressRelease(deployConfig.Repo, releaseVersion, deployConfig.ReleasePath)

	// get old version processes
	oldVersionProcessesPID, err := balancer.GetProcessesPID()
	if err != nil {
		slog.Error("Getting old version processes", "err", err)
		return
	}

	newReleasProcess, err := LaunchNewVersion(deployConfig, releaseVersion)
	if err != nil {
		slog.Error("Launching new version", "err", err)
		return
	}

	newBackend := fmt.Sprintf("http://localhost:%d", newReleasProcess.Port)

	// replace the backends with the new version
	replaceBackendInConfig([]string{newBackend})

	// kill old version
	balancer.RemoveProcesses(oldVersionProcessesPID)

	// update the release in the config
	config.StoreValueInConfig(deployConfig.Repo+".last_release", releaseVersion)

	// set the older version in the ignore to avoid the next deploy with the same version
	config.StoreValueInConfig(deployConfig.Repo+".ignore", lastReleaseConfig.LastRelease)
}

func checkIfReleaseExists(repo string, releaseVersion string) bool {
	// gh release view <tag> --repo <owner>/<repo>
	// Command to execute
	cmd := exec.Command("gh", "release", "view", releaseVersion, "--repo", repo)

	// Run the command without capturing its output
	_, err := cmd.Output() // The output is non necessary
	// NOTE: if release not exists gh will return a status code 1 and enter the err
	if err != nil {
		slog.Error("From command output gh", "err", err)
		return false
	}

	//	outputString := string(output)
	//	releaseInfo := strings.Fields(outputString)

	return true
}
