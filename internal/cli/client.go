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

func NewInfoClient(socketPath string) (*InfoClient, error) {
    conn, err := net.DialTimeout("unix", socketPath, time.Second*10)
    if err != nil {
        return nil, err
    }
    client := jsonrpc.NewClient(conn)
    return &InfoClient{client: client}, nil
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