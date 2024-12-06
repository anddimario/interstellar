/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	cli "github.com/anddimario/interstellar/internal/cli"
	config "github.com/anddimario/interstellar/internal/config"
)

var (
	canaryQuota     string
	rollbackVersion string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Manage application deployment",
	Long:  `Manage deploy and rollback`,
	Run: func(cmd *cobra.Command, args []string) {

		socketPath := config.K.String("cli.socket_path") // todo see if injectable

		deployCliClient, err := cli.NewDeployClient(socketPath)
		if err != nil {
			slog.Error("Failed to connect to server", "err", err)
		}

		if canaryQuota != "0" {
			response, err := deployCliClient.ExecuteAction("DeployService.ExecuteAction", "canary-update-quota", canaryQuota)
			if err != nil {
				slog.Error("Response from server", "err", err)
				return
			}

			fmt.Println(response)
			return
		}

		if rollbackVersion != "0" {

			response, err := deployCliClient.ExecuteAction("DeployService.ExecuteAction", "rollback", rollbackVersion)
			if err != nil {
				slog.Error("Response from server", "err", err)
				return
			}

			fmt.Println(response)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&canaryQuota, "canary-quota", "q", "0", "Canary new release quota")
	deployCmd.Flags().StringVarP(&rollbackVersion, "rollback", "r", "0", "Rollback to previous release")
}
