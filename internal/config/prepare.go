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

func PrepareDeployConfig() DeployConfig {

	// Start monitor new releases on github
	checkReleaseInterval := viper.GetDuration("deploy.check_release_interval")
	repo := viper.GetString("deploy.repo")
	releasePath := viper.GetString("deploy.release_path")
	assetName := viper.GetString("deploy.asset_name")
	executableCommand := viper.GetString("deploy.executable_command")
	executableEnv := viper.GetStringSlice("deploy.executable_env")
	executableArgs := viper.GetStringSlice("deploy.executable_args")
	deployType := viper.GetString("deploy.type")

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