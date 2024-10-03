package config

import (
	"fmt"

	viper "github.com/spf13/viper"
)

func StoreConfig(name string, value string) {
	
    // Set a configuration value
    viper.Set(name, value)

    // // Write the configuration to the file
    viper.AddConfigPath(".")      // Path to look for the config file in the current directory
	err := viper.WriteConfigAs("config.toml")
    if err != nil {
        fmt.Printf("Error writing config file: %s\n", err)
        return
    }

}