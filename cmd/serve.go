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
	peer "github.com/anddimario/interstellar/internal/peer"
)

var (
	address      string
	peerAddress  string
	peerName     string
	peerNodeAddr string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the proxy server",
	Long:  `Run the proxy server. This command will start the proxy server, the healthcheck process, and the other used processes.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Check the required dependencies
		config.CheckRequirements()

		// Initialize the configuration
		config.InitConfig()

		// Check pending releases
		go deploy.RecoveryFromCrash(config.PrepareRecoveryConfig())

		// Define server
		srv := &http.Server{}
		srv.Addr = viper.GetString("balancer.address")
		if address != "" {
			srv.Addr = address
		}
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

		// Start peer with gossip protocol
		if peerName != "" {
			newPeer := peer.NewPeer(peerName, peerAddress)
			if peerNodeAddr != "" {
				newPeer.Bootstrap(peerNodeAddr)
			}
			go newPeer.Gossip()
			go newPeer.Listen()
		}

		// Shutdown operations
		<-ctx.Done()

		slog.Info("Got interruption signal")
		if err := srv.Shutdown(context.TODO()); err != nil {
			slog.Error("server shutdown returned an error", "err", err)
		}

		// Stop peer gossip
		peer.PeeringDone <- true
		// close(peer.PeeringDone)
		// Stop healthcheck
		balancer.HealthCheckDone <- true
		// Stop monitor new releases on github
		deploy.CheckReleaseDone <- true

		slog.Info("Bye!")

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&address, "address", "a", "localhost:8080", "Address to listen on")
	serveCmd.Flags().StringVarP(&peerAddress, "peer-address", "e", "localhost:8080", "Address to listen on")
	serveCmd.Flags().StringVarP(&peerName, "peer-name", "n", "", "Peer name")
	serveCmd.Flags().StringVarP(&peerNodeAddr, "peer-bootstrap", "b", "", "Peer node address for bootstrap")
}
