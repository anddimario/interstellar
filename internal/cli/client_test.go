package cli

import (
    "net"
    "net/rpc"
    "net/rpc/jsonrpc"
    "testing"
)


func startMockServer(t *testing.T, socketPath string) net.Listener {
    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        t.Fatalf("Failed to start mock server: %v", err)
    }

    deployService := new(DeployService)
    infoService := new(InfoService)
    rpc.Register(deployService)
    rpc.Register(infoService)

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

func TestDeployClient_ExecuteAction(t *testing.T) {
    socketPath := "/tmp/test_deploy.sock"
    listener := startMockServer(t, socketPath)
    defer listener.Close()

    client, err := NewDeployClient(socketPath)
    if err != nil {
        t.Fatalf("Failed to create DeployClient: %v", err)
    }

    // todo: use another action?
    result, err := client.ExecuteAction("DeployService.ExecuteAction", "rollback", "v1.0.0")
    if err != nil {
        t.Fatalf("ExecuteAction failed: %v", err)
    }
    
    expected := "{\"message\":\"Started rollback to version v1.0.0\",\"status\":\"ok\"}"
    if result != expected {
        t.Errorf("Expected '%s', got '%s'", expected, result)
    }
}

func TestInfoClient_GetInfo(t *testing.T) {
    socketPath := "/tmp/test_info.sock" // todo: use config?
    listener := startMockServer(t, socketPath)
    defer listener.Close()

    client, err := NewInfoClient(socketPath)
    if err != nil {
        t.Fatalf("Failed to create InfoClient: %v", err)
    }

    result, err := client.GetInfo("version")
    if err != nil {
        t.Fatalf("GetInfo failed: %v", err)
    }

    expected := ""
    if result != expected {
        t.Errorf("Expected '%s', got '%s'", expected, result)
    }
}
