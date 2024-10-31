package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"strconv"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	"github.com/anddimario/interstellar/internal/deploy"
	"github.com/spf13/viper"
)

type CliConfig struct {
	SocketPath string
}

type InfoService struct{}
type DeployService struct{}

type CommandRequest struct {
	Command string
	Param   string
}

type CommandResponse struct {
	Result string
}

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
	switch req.Query {
	case "version":
		repo := viper.GetString("deploy.repo")
		versionConfigPath := fmt.Sprintf("%s.%s", repo, "last_release")
		res.Info = viper.GetString(versionConfigPath)
	case "deploy":
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
	default:
		return errors.New("invalid query")
	}
	return nil
}

func (s *DeployService) Canary(req CommandRequest, res *CommandResponse) error {
	switch req.Command {
	case "canary-update-quota":
		deployIsInProgress := deploy.CheckIfDeployInProgress()
		if !deployIsInProgress {
			return errors.New("no deploy in progress")
		}
		// Convert the string to an integer
		quota, err := strconv.Atoi(req.Param)
		if err != nil {
			return errors.New("invalid quota")
		}
		balancer.UpdateCanaryNewReleaseQuota(int(quota))

		payload := ResponsePayload{
			Message: "Quota updated",
			Status:  "ok",
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return errors.New("error marshaling JSON " + err.Error())
		}
		res.Result = string(payloadJSON)
	default:
		return errors.New("invalid command")
	}
	return nil
}

func (config CliConfig) StartCliServer() {
	os.Remove(config.SocketPath) // Remove the socket file if it already exists

	infoService := new(InfoService)
	rpc.Register(infoService)

	deployService := new(DeployService)
	rpc.Register(deployService)

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
