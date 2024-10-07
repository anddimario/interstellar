package config

import "time"

type DeployConfig struct {
	CheckReleaseInterval time.Duration
	Repo                 string
	ReleasePath          string
	AssetName            string
	ExecutableCommand    string
	ExecutableEnv        []string
	ExecutableArgs       []string
}
