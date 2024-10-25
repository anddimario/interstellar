package cli

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"

	balancer "github.com/anddimario/interstellar/internal/balancer"
	"github.com/anddimario/interstellar/internal/deploy"
)

type InfoService struct{}

type InfoRequest struct {
	Query string
}

type InfoResponse struct {
	Info string
}

func (s *InfoService) GetInfo(req InfoRequest, res *InfoResponse) error {
	if req.Query == "" {
		return errors.New("query cannot be empty")
	}
	if req.Query == "version" {
		res.Info = "1.0.0"
		return nil
	}
	// todo: format the output in json?
	if req.Query == "deploy" {
		deployIsInProgress := deploy.CheckIfDeployInProgress()
		fmt.Println(deployIsInProgress)
		if deployIsInProgress {
			canaryInfo := balancer.GetCanaryDeployStatus()
			if canaryInfo.InProgress {
				res.Info = fmt.Sprintf("Deploy status: %v", balancer.GetCanaryDeployStatus())
			} else {
				res.Info = "Blue-green deploy in progress"
			}
		} else {
			res.Info = "No deploy in progress"
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
