/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	"github.com/anddimario/interstellar/internal/cli"
	config "github.com/anddimario/interstellar/internal/config"
	deploy "github.com/anddimario/interstellar/internal/deploy"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Check the required dependencies
		config.CheckRequirements()

		// Initialize the configuration
		config.InitConfig()

		// Define server
		srv := &http.Server{}
		srv.Addr = viper.GetString("balancer.address")
		srv.Handler = http.HandlerFunc(balancer.HandleRequest)

		// Define a context to listen for signals
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// Start the server
		go func() {
			slog.Info("Load balancer starting", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}()

		// Start healthcheck
		healthCheckConfig := balancer.HealthCheckConfig{
			Interval: viper.GetDuration("healthcheck.interval"),
			Path:     viper.GetString("healthcheck.path"),
		}
		go healthCheckConfig.InitBackendsFromConfig(viper.GetStringSlice("balancer.backends"))

		// Start Deploy process
		deployConfig := config.PrepareDeployConfig()
		go deploy.CheckRelease(deployConfig)

		// Cli server
		cliServerConfig := cli.CliConfig{
			SocketPath: viper.GetString("cli.socket_path"),
		}
		go cliServerConfig.StartCliServer()

		<-ctx.Done()

		slog.Info("Got interruption signal")
		if err := srv.Shutdown(context.TODO()); err != nil {
			slog.Error("server shutdown returned an error", "err", err)
		}

		// Stop healthcheck
		balancer.HealthCheckDone <- true
		// Stop monitor new releases on github
		deploy.CheckReleaseDone <- true

		slog.Info("Bye!")

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
