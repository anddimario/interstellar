package cli

import (
    "net"
    "net/rpc"
    "net/rpc/jsonrpc"
    "testing"
    "time"
)

type DeployServiceTest struct{}

func (s *DeployServiceTest) ExecuteAction(req CommandRequest, res *CommandResponse) error {
    res.Result = "Action executed: " + req.Command + " with param: " + req.Param
    return nil
}

type InfoServiceTest struct{}

func (s *InfoServiceTest) GetInfo(req InfoRequest, res *InfoResponse) error {
    res.Info = "Info for query: " + req.Query
    return nil
}

func startServer(t *testing.T, socketPath string) net.Listener {
    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }

    deployService := new(DeployServiceTest)
    infoServiceTest := new(InfoServiceTest)
    rpc.Register(deployService)
    rpc.Register(infoServiceTest)

    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                return
            }
            go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
        }
    }()

    return listener
}

func TestDeployService_ExecuteAction(t *testing.T) {
    socketPath := "/tmp/test_deploy_service.sock"
    listener := startServer(t, socketPath)
    defer listener.Close()

    client, err := net.DialTimeout("unix", socketPath, time.Second*10)
    if err != nil {
        t.Fatalf("Failed to connect to server: %v", err)
    }
    defer client.Close()

    rpcClient := jsonrpc.NewClient(client)
    req := CommandRequest{Command: "deploy", Param: "v1.0.0"}
    var res CommandResponse
    err = rpcClient.Call("DeployServiceTest.ExecuteAction", req, &res)
    if err != nil {
        t.Fatalf("ExecuteAction failed: %v", err)
    }

    expected := "Action executed: deploy with param: v1.0.0"
    if res.Result != expected {
        t.Errorf("Expected '%s', got '%s'", expected, res.Result)
    }
}

func TestInfoServiceTest_GetInfo(t *testing.T) {
    socketPath := "/tmp/test_info_service.sock" // todo: use config?
    listener := startServer(t, socketPath)
    defer listener.Close()

    client, err := net.DialTimeout("unix", socketPath, time.Second*10)
    if err != nil {
        t.Fatalf("Failed to connect to server: %v", err)
    }
    defer client.Close()

    rpcClient := jsonrpc.NewClient(client)
    req := InfoRequest{Query: "test query"}
    var res InfoResponse
    err = rpcClient.Call("InfoServiceTest.GetInfo", req, &res)
    if err != nil {
        t.Fatalf("GetInfo failed: %v", err)
    }

    expected := "Info for query: test query"
    if res.Info != expected {
        t.Errorf("Expected '%s', got '%s'", expected, res.Info)
    }
}
