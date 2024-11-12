package config

import (
	"fmt"

	viper "github.com/spf13/viper"
)

func StoreValueInConfig[T any](name string, value T) {
	
    // Set a configuration value
    viper.Set(name, value)

    writeToConfigFile()

}

func DeleteValueInConfig(name string) {
	// IMP: there's a workaround here, because it does not seem possible to delete a key in viper 

    // Set a configuration value
    viper.Set(name, false)

    writeToConfigFile()
}

func GetValueFromConfig(name string) string {
    
    // Get a configuration value
    value := viper.GetString(name)
    return value

}

func writeToConfigFile() {
    
    // Write the configuration to the file
    viper.AddConfigPath(".")      // Path to look for the config file in the current directory
    err := viper.WriteConfigAs("config.toml")
    if err != nil {
        fmt.Printf("Error writing config file: %s\n", err)
        return
    }

}   