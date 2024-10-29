package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	"github.com/anddimario/interstellar/internal/deploy"
	"github.com/spf13/viper"
)

type InfoService struct{}

type InfoRequest struct {
	Query string
}

type InfoResponse struct {
	Info string
}

type ResponsePayload struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (s *InfoService) GetInfo(req InfoRequest, res *InfoResponse) error {
	if req.Query == "" {
		return errors.New("query cannot be empty")
	}
	if req.Query == "version" {
		repo := viper.GetString("deploy.repo")
		versionConfigPath := fmt.Sprintf("%s.%s", repo, "last_release")
		res.Info = viper.GetString(versionConfigPath)
		return nil
	}
	// todo: format the output in json?
	if req.Query == "deploy" {
		deployIsInProgress := deploy.CheckIfDeployInProgress()
		if deployIsInProgress {
			canaryInfo := balancer.GetCanaryDeployStatus()
			if canaryInfo.InProgress {
				// Marshal the struct to JSON
				canaryInfoJSON, err := json.Marshal(balancer.GetCanaryDeployStatus())
				if err != nil {
					fmt.Println("Error marshaling JSON:", err)
					return err
				}
				res.Info = string(canaryInfoJSON)
			} else {
				payload := ResponsePayload{
					Message: "Blue-green deploy in progress",
					Status:  "ok",
				}
				payloadJSON, err := json.Marshal(payload)
				if err != nil {
					fmt.Println("Error marshaling JSON:", err)
					return err
				}
				res.Info = string(payloadJSON)
			}
		} else {
			payload := ResponsePayload{
				Message: "No deploy in progress",
				Status:  "ok",
			}
			payloadJSON, err := json.Marshal(payload)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return err
			}
			res.Info = string(payloadJSON)
		}
	}
	return nil
}

type CliConfig struct {
	SocketPath string
}

func (config CliConfig) StartCliServer() {
	os.Remove(config.SocketPath) // Remove the socket file if it already exists

	infoService := new(InfoService)
	rpc.Register(infoService)

	listener, err := net.Listen("unix", config.SocketPath)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
