package cli

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"
)

type InfoClient struct {
    client *rpc.Client
}

type DeployClient struct {
    client *rpc.Client
}

func NewInfoClient(socketPath string) (*InfoClient, error) {
    conn, err := net.DialTimeout("unix", socketPath, time.Second*10)
    if err != nil {
        return nil, err
    }
    client := jsonrpc.NewClient(conn)
    return &InfoClient{client: client}, nil
}

func NewDeployClient(socketPath string) (*DeployClient, error) {
    conn, err := net.DialTimeout("unix", socketPath, time.Second*10)
    if err != nil {
        return nil, err
    }
    client := jsonrpc.NewClient(conn)
    return &DeployClient{client: client}, nil
}

func (c *DeployClient) ExecuteAction(service string, action string, param string) (string, error) {
    req := CommandRequest{Command: action, Param: param}
    var res CommandResponse
    err := c.client.Call(service, req, &res)
    if err != nil {
        return "", err
    }
    return res.Result, nil
}

func (c *InfoClient) GetInfo(query string) (string, error) {
    req := InfoRequest{Query: query}
    var res InfoResponse
    err := c.client.Call("InfoService.GetInfo", req, &res)
    if err != nil {
        return "", err
    }
    return res.Info, nil
}