package peer

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewPeer(t *testing.T) {
    secret := "shared_secret"
    peer := NewPeer("localhost:12345", secret)

    if peer.ID == "" {
        t.Error("Expected peer ID to be set")
    }
    if peer.Addr != "localhost:12345" {
        t.Errorf("Expected peer address to be 'localhost:12345', got '%s'", peer.Addr)
    }
    if peer.Secret != secret {
        t.Errorf("Expected peer secret to be '%s', got '%s'", secret, peer.Secret)
    }
}

func TestSendPeerList(t *testing.T) {
    secret := "shared_secret"
    peer := NewPeer("localhost:12345", secret)

    // Start a dummy UDP server to receive the peer list
    addr := "localhost:12346"
    udpAddr, _ := net.ResolveUDPAddr("udp", addr)
    conn, _ := net.ListenUDP("udp", udpAddr)
    defer conn.Close()

    go func() {
        buf := make([]byte, 1024)
        n, _, err := conn.ReadFromUDP(buf)
        if err != nil {
            t.Errorf("Error receiving data: %v", err)
        }
        message := string(buf[:n])
        expectedPrefix := "PEERS"
        if message[:len(expectedPrefix)] != expectedPrefix {
            t.Errorf("Expected message to start with '%s', got '%s'", expectedPrefix, message)
        }
    }()

    peer.sendPeerList(addr)
    time.Sleep(1 * time.Second) // Give some time for the message to be received
}

func TestHandleMessage(t *testing.T) {
    secret := "shared_secret"
    peer := NewPeer("localhost:12345", secret)

    // Simulate receiving a PEERS message
    message := fmt.Sprintf("PEERS peer1 localhost:12346 peer2-localhost:12347,peer3-localhost:12348 %s", secret)
    addr, _ := net.ResolveUDPAddr("udp", "localhost:12346")
    peer.handleMessage(message, addr)

    if len(peer.Peers) != 3 {
        t.Errorf("Expected 3 peers, got %d", len(peer.Peers))
    }
    if peer.Peers["peer1"] != "localhost:12346" {
        t.Errorf("Expected peer1 address to be 'localhost:12346', got '%s'", peer.Peers["peer1"])
    }
    if peer.Peers["peer2"] != "localhost:12347" {
        t.Errorf("Expected peer2 address to be 'localhost:12347', got '%s'", peer.Peers["peer2"])
    }
    if peer.Peers["peer3"] != "localhost:12348" {
        t.Errorf("Expected peer3 address to be 'localhost:12348', got '%s'", peer.Peers["peer3"])
    }
}

func TestBootstrap(t *testing.T) {
    secret := "shared_secret"
    peer := NewPeer("localhost:12345", secret)

    // Start a dummy UDP server to act as the bootstrap peer
    addr := "localhost:12346"
    udpAddr, _ := net.ResolveUDPAddr("udp", addr)
    conn, _ := net.ListenUDP("udp", udpAddr)
    defer conn.Close()

    go func() {
        buf := make([]byte, 1024)
        n, _, err := conn.ReadFromUDP(buf)
        if err != nil {
            t.Errorf("Error receiving data: %v", err)
        }
        message := string(buf[:n])
        expectedPrefix := "PEERS"
        if message[:len(expectedPrefix)] != expectedPrefix {
            t.Errorf("Expected message to start with '%s', got '%s'", expectedPrefix, message)
        }
    }()

    peer.Bootstrap(addr)
    time.Sleep(1 * time.Second) // Give some time for the message to be received
}

func TestCleanupPeerList(t *testing.T) {
    secret := "shared_secret"
    peer := NewPeer("localhost:12345", secret)

    // Add some peers with different last seen times
    peer.Peers["peer1"] = "localhost:12346"
    peer.LastSeen["peer1"] = time.Now().Add(-40 * time.Second) // Should be removed
    peer.Peers["peer2"] = "localhost:12347"
    peer.LastSeen["peer2"] = time.Now().Add(-20 * time.Second) // Should be kept

    peer.cleanupPeerList()

    if len(peer.Peers) != 1 {
        t.Errorf("Expected 1 peer, got %d", len(peer.Peers))
    }
    if _, exists := peer.Peers["peer1"]; exists {
        t.Errorf("Expected peer1 to be removed")
    }
    if _, exists := peer.Peers["peer2"]; !exists {
        t.Errorf("Expected peer2 to be kept")
    }
}
