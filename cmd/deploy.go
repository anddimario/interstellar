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
	canaryQuota     string
	rollbackVersion string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Manage application deployment",
	Long:  `Manage deploy and rollback`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("cmd: %v\n", cmd)
		fmt.Printf("args: %v\n", args)
		fmt.Printf("canaryQuota: %v\n", canaryQuota)
		fmt.Printf("rollbackVersion: %v\n", rollbackVersion)
		socketPath := viper.GetString("cli.socket_path") // todo see if injectable

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
