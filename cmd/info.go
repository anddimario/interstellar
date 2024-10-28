/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cli "github.com/anddimario/interstellar/internal/cli"
)

var (
    showVersion   bool
    showDeploy bool
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about the application",
	Long: `Information about the application`,
	Run: func(cmd *cobra.Command, args []string) {
		socketPath := viper.GetString("cli.socket_path") // todo see if injectable

		infoClient, err := cli.NewInfoClient(socketPath)
        if err != nil {
            slog.Error("Failed to connect to server", "err", err)
        }

		query := "version"
		// if showVersion {
        //     fmt.Printf("Version: %s\n", version)
		// }
        if showDeploy {
			query = "deploy"
		}

        info, err := infoClient.GetInfo(query)
        if err != nil {
            slog.Error("Failed to get info", "err", err)
        }

        fmt.Println(info)

        // if !showVersion && !showBuildDate && !showAuthor {
        //     fmt.Printf("Application Information:\n")
        //     fmt.Printf("Version: %s\n", version)
        //     fmt.Printf("Build Date: %s\n", buildDate)
        //     fmt.Printf("Author: %s\n", author)
        // }
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Version of the actual deployed application")
	infoCmd.Flags().BoolVarP(&showDeploy, "deploy", "d", false, "Deploy informations")
}
