package config

import (
	"time"

)

type DeployConfig struct {
	CheckReleaseInterval time.Duration
	Repo                 string
	ReleasePath          string
	AssetName            string
	ExecutableCommand    string
	ExecutableEnv        []string
	ExecutableArgs       []string
	Type                 string
}

type ReleaseConfig struct {
	LastRelease string
	Ignore      string
}

type RecoveryConfig struct {
	DeployInProgress bool
	NewProcessPID    int
	NewProcessPort   int
}

func PrepareDeployConfig() DeployConfig {

	// Start monitor new releases on github
	checkReleaseInterval := K.Duration("deploy.check_release_interval")
	repo := K.String("deploy.repo")
	releasePath := K.String("deploy.release_path")
	assetName := K.String("deploy.asset_name")
	executableCommand := K.String("deploy.executable_command")
	executableEnv := K.Strings("deploy.executable_env")
	executableArgs := K.Strings("deploy.executable_args")
	deployType := K.String("deploy.type")

	return DeployConfig{
		CheckReleaseInterval: checkReleaseInterval,
		Repo:                 repo,
		ReleasePath:          releasePath,
		AssetName:            assetName,
		ExecutableCommand:    executableCommand,
		ExecutableEnv:        executableEnv,
		ExecutableArgs:       executableArgs,
		Type:                 deployType,
	}
}

func PrepareReleaseConfig(repo string) ReleaseConfig {

	lastRelease := K.String(repo + ".last_release")
	ignore := K.String(repo + ".ignore")

	return ReleaseConfig{
		LastRelease: lastRelease,
		Ignore:      ignore,
	}
}

func PrepareRecoveryConfig() RecoveryConfig {

	inProgress := K.Bool("deploy.in_progress")
	newProcessPID := K.Int("deploy.new_process_pid")
	newProcessPort := K.Int("deploy.new_process_port")

	return RecoveryConfig{
		DeployInProgress: inProgress,
		NewProcessPID:    newProcessPID,
		NewProcessPort:   newProcessPort,
	}
}
