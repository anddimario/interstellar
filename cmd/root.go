/*
Copyright © 2024 Andrea Di Mario
*/
package cmd

import (
	"os"

	"github.com/anddimario/interstellar/internal/config"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "interstellar",
	Short: "Application Deployer",
	Long:  `Watch repository release and deploy binary using canary or blue green deployment.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
    cobra.OnInitialize(config.InitConfig)

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
