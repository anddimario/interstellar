package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	config "github.com/anddimario/interstellar/internal/config"
	deploy "github.com/anddimario/interstellar/internal/deploy"
	"github.com/spf13/viper"
)

func main() {
	// Check the required dependencies
	config.CheckRequirements()

	// Initialize the configuration
	config.InitConfig()

	// Define server
	srv := &http.Server{}
	srv.Addr = viper.GetString("server.address")
	srv.Handler = http.HandlerFunc(balancer.HandleRequest)

	// Define a context to listen for signals
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start the server
	go func() {
		log.Printf("Load balancer starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	// Start healthcheck
	healthCheckInterval := viper.GetDuration("healthcheck.interval")
	go balancer.HealthCheck(healthCheckInterval * time.Second, viper.GetStringSlice("balancer.backends"))

	// Start monitor new releases on github
	checkReleaseInterval := viper.GetDuration("deploy.check_release_interval")
	repo := viper.GetString("deploy.repo")
	go deploy.CheckRelease(checkReleaseInterval, repo)

	<-ctx.Done()

	log.Println("got interruption signal")
	if err := srv.Shutdown(context.TODO()); err != nil {
		log.Printf("server shutdown returned an err: %v\n", err)
	}

	// Stop healthcheck
	balancer.HealthCheckDone <- true
	// Stop monitor new releases on github
	deploy.CheckReleaseDone <- true

	log.Println("Bye!")
}
