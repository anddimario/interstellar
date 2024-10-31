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
	canaryQuota string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Manage deploy",
	Long: `Manage deploy`,
	Run: func(cmd *cobra.Command, args []string) {

		socketPath := viper.GetString("cli.socket_path") // todo see if injectable

		deployCliClient, err := cli.NewDeployClient(socketPath)
		if err != nil {
			slog.Error("Failed to connect to server", "err", err)
		}
		
		if canaryQuota != "" {
			response, err := deployCliClient.Canary("canary-update-quota", canaryQuota)
			if err != nil {
				slog.Error("Response from server", "err", err)
				return
			}
			
			fmt.Println(response)			

		}

	},
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVarP(&canaryQuota, "canary-quota", "q", "0", "Canary new release quota")

}
