package config

import (
	"log"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var K = koanf.New(".")

func InitConfig() {
	if err := K.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	validateConfig()

}


func validateConfig() {
	if !K.Exists("cli.socket_path") {
		log.Fatal("cli.socket_path is not set in the config file")
	}
	if !K.Exists("balancer.address") {
		log.Fatal("balancer.address is not set in the config file")
	}

	if !K.Exists("healthcheck.interval") {
		log.Fatal("healthcheck.interval is not set in the config file")
	}
}
