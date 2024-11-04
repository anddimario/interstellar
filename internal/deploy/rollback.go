package deploy

import (
	"fmt"
	"log/slog"

	config "github.com/anddimario/interstellar/internal/config"
	balancer "github.com/anddimario/interstellar/internal/balancer"
)

func Rollback(releaseVersion string) {
	fmt.Println("Rollback")

	deployIsInProgress := CheckIfDeployInProgress()
	
	fmt.Print(deployIsInProgress)
	if deployIsInProgress {
		slog.Error("Deploy in progress, rollback not allowed")
		return
	}

	// todo check if the release exists
	// gh release view <tag> --repo <owner>/<repo>

	deployConfig := config.PrepareDeployConfig()

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
}
