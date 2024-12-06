package config

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/parsers/toml"
)

func StoreValueInConfig[T any](name string, value T) {

	// Set a configuration value
	K.Set(name, value)

	writeToConfigFile()

}

func DeleteValueInConfig(name string) {
	K.Delete(name)

	writeToConfigFile()
}

// func GetValueFromConfig(key string, valueType string) (interface{}, error) {
//     if !k.Exists(key) {
//         return nil, fmt.Errorf("key '%s' does not exist", key)
//     }

//     switch valueType {
//     case "string":
//         return k.String(key), nil
//     case "int":
//         return k.Int(key), nil
//     case "float64":
//         return k.Float64(key), nil
//     case "bool":
//         return k.Bool(key), nil
//     case "strings":
//         return k.Strings(key), nil
//     case "ints":
//         return k.Ints(key), nil
//     case "duration":
//         return k.Duration(key), nil
//     default:
//         return nil, fmt.Errorf("unsupported value type '%s'", valueType)
//     }

// }

func writeToConfigFile() {

	// Save the updated configuration back to the file
	configBytes, err := K.Marshal(toml.Parser())
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	if err := os.WriteFile("config.toml", configBytes, 0644); err != nil {
		fmt.Printf("Error writing config file: %v\n", err)
		return
	}

}
