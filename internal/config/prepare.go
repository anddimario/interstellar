package config

import (
	"time"

	"github.com/spf13/viper"
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
	checkReleaseInterval := viper.GetDuration("deploy.check_release_interval")
	repo := GetValueFromConfig("deploy.repo")
	releasePath := GetValueFromConfig("deploy.release_path")
	assetName := GetValueFromConfig("deploy.asset_name")
	executableCommand := GetValueFromConfig("deploy.executable_command")
	executableEnv := viper.GetStringSlice("deploy.executable_env")
	executableArgs := viper.GetStringSlice("deploy.executable_args")
	deployType := GetValueFromConfig("deploy.type")

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

	lastRelease := GetValueFromConfig(repo + ".last_release")
	ignore := GetValueFromConfig(repo + ".ignore")

	return ReleaseConfig{
		LastRelease: lastRelease,
		Ignore:      ignore,
	}
}

func PrepareRecoveryConfig() RecoveryConfig {

	inProgress := viper.GetBool("deploy.in_progress")
	newProcessPID := viper.GetInt("deploy.new_process_pid")
	newProcessPort := viper.GetInt("deploy.new_process_port")

	return RecoveryConfig{
		DeployInProgress: inProgress,
		NewProcessPID:    newProcessPID,
		NewProcessPort:   newProcessPort,
	}
}
