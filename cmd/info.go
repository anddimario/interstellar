
package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	cli "github.com/anddimario/interstellar/internal/cli"
	config "github.com/anddimario/interstellar/internal/config"
)

var (
	showVersion bool
	showDeploy  bool
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information about the application",
	Long:  `Information about the application`,
	Run: func(cmd *cobra.Command, args []string) {
		socketPath := config.GetValueFromConfig("cli.socket_path") // todo see if injectable

		infoCliClient, err := cli.NewInfoClient(socketPath)
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

		info, err := infoCliClient.GetInfo(query)
		if err != nil {
			slog.Error("Failed to get info", "err", err)
		}

		fmt.Println(info)

	},
}

func init() {
	rootCmd.AddCommand(infoCmd)

	infoCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Version of the actual deployed application")
	infoCmd.Flags().BoolVarP(&showDeploy, "deploy", "d", false, "Deploy informations")
}
