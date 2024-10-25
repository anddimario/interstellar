package config

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config") // Name of the config file (without extension)
	viper.SetConfigType("toml")   // Config file type
	viper.AddConfigPath(".")      // Path to look for the config file in the current directory

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	validateConfig()
}

func validateConfig() {
	if !viper.IsSet("cli.socket_path") {
		log.Fatal("cli.socket_path is not set in the config file")
	}
	if !viper.IsSet("balancer.address") {
		log.Fatal("balancer.address is not set in the config file")
	}

	if !viper.IsSet("healthcheck.interval") {
		log.Fatal("healthcheck.interval is not set in the config file")
	}
}
