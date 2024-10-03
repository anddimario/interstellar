package config

import (
	"fmt"
	"log"
	"os/exec"
)

func CheckRequirements() error {
	// Program to check
	programs := []string{"gh", "podman"}

	for _, program := range programs {

		path, err := exec.LookPath(program)
		if err != nil {
			fmt.Printf("%s is not installed\n", program)
			log.Panic("Please install the required dependencies")
		} else {
			fmt.Printf("%s is installed at %s\n", program, path)
		}

	}
	return nil
}
