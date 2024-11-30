package peer

import (
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type PeerConfig struct {
	secret   string
	memThreshold float64
	cpuThreshold float64
}

type Peer struct {
	ID       string
	Addr     string
	Peers    map[string]string // ID -> Addr
	LastSeen map[string]time.Time
	mu       sync.Mutex
	Config   *PeerConfig
}

var PeeringDone = make(chan bool)

func NewPeer(addr string, secret string, memThreshold float64, cpuThreshold float64) *Peer {
	id := generateRandomString(8)
	peerConfig := &PeerConfig{
		secret: secret,
		memThreshold: memThreshold,
		cpuThreshold: cpuThreshold,
	}
	return &Peer{
		ID:       id,
		Addr:     addr,
		Peers:    make(map[string]string),
		LastSeen: make(map[string]time.Time),
		Config:   peerConfig,
	}
}

func (p *Peer) Gossip() {
	t := time.NewTicker(10 * time.Second) // todo: make interval configurable?
	defer t.Stop()

	for {
		select {
		case <-t.C:
			// slog.Info("Gossiping", "addr", p.Addr)
			p.mu.Lock()
			p.cleanupPeerList()
			peers := p.getRandomPeers(2) // Select 2 random peers to gossip with
			p.mu.Unlock()

			for _, addr := range peers {
				// add info if it's available to serve requests from other peer based on the load
				go p.sendPeerAvailability(addr)
				// send the peer list to the selected peer
				go p.sendPeerList(addr)
			}
		case <-PeeringDone:
			return
		}
	}
}

func (p *Peer) Listen() {
	slog.Info("Listening for peers", "addr", p.Addr)
	udpAddr, _ := net.ResolveUDPAddr("udp", p.Addr)
	conn, _ := net.ListenUDP("udp", udpAddr)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		message := string(buf[:n])
		p.handleMessage(message, remoteAddr)
	}

}

func (p *Peer) sendPeerList(addr string) {
	peerList := ""
	for id, peerAddr := range p.Peers {
		peerList += fmt.Sprintf("%s-%s,", id, peerAddr)
	}
	peerList = strings.TrimRight(peerList, ",")

	// Serialize and send peers as JSON (or another format)
	// The last part of the message should be the secret
	message := fmt.Sprintf("PEERS %s %s %s %s", p.ID, p.Addr, peerList, p.Config.secret)

	p.sendMessage(addr, message)
}

func (p *Peer) sendMessage(addr string, message string) {
	conn, err := net.Dial("udp4", addr)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(message))
}

func (p *Peer) sendPeerAvailability(addr string) {
	canServeExternalRequests := IsAvailableToServe(p.Config.memThreshold, p.Config.cpuThreshold)
	if canServeExternalRequests {
		message := fmt.Sprintf("AVAILABLE %s %s %s", p.ID, p.Addr, p.Config.secret)
		p.sendMessage(addr, message)
	} else {
		message := fmt.Sprintf("UNAVAILABLE %s %s %s", p.ID, p.Addr, p.Config.secret)
		p.sendMessage(addr, message)
	}
}

func (p *Peer) getRandomPeers(n int) []string {

	fmt.Println("Peers:", p.Peers)
	peerAddrs := make([]string, 0, len(p.Peers))
	for _, addr := range p.Peers {
		peerAddrs = append(peerAddrs, addr)
	}

	// Shuffle and select n peers
	rand.Shuffle(len(peerAddrs), func(i, j int) {
		peerAddrs[i], peerAddrs[j] = peerAddrs[j], peerAddrs[i]
	})

	if len(peerAddrs) < n {
		return peerAddrs
	}
	return peerAddrs[:n]
}

func (p *Peer) handleMessage(message string, addr *net.UDPAddr) {
	slog.Info("Received message", "addr", addr.String(), "message", message)

	if len(message) < 6 {
		slog.Error("Message too short", "message", message)
		return
	}

	// Split message into parts, the first part is the message type (command)
	parts := strings.Split(message, " ")

	// Check if the message is from a valid peer
	// The last part of the message should be the secret
	if parts[len(parts)-1] != p.Config.secret {
		slog.Error("Invalid secret", "addr", addr.String())
		return
	}

	// Here we can parse message type and content
	switch parts[0] {
	case "PEERS":
		// Get the peer name and address from the message
		p.mu.Lock()
		p.Peers[parts[1]] = parts[2]
		p.mu.Unlock()
		// slog.Info("Added peer", "peer", parts[1], "addr", parts[2])
		// Update the peer list with received peers
		receivedPeers := strings.Split(parts[3], ",")
		p.LastSeen[parts[1]] = time.Now()
		p.mu.Lock()
		for _, peer := range receivedPeers {
			peerParts := strings.Split(peer, "-")
			if len(peerParts) == 2 && peerParts[0] != p.ID { // Avoid to add the same peer to its own list
				p.Peers[peerParts[0]] = peerParts[1]
				p.LastSeen[peerParts[0]] = time.Now()
			}
		}
		fmt.Printf("Updated peer list: %v\n", p.Peers)
		p.mu.Unlock()
	case "AVAILABLE":
	// todo
	case "UNAVAILABLE":
	// todo
	default:
		slog.Error("Unknown message", "type", parts[0])
	}
}

func (p *Peer) Bootstrap(bootstrapAddr string) {
	conn, err := net.Dial("udp", bootstrapAddr)
	if err != nil {
		slog.Error("Could not connect to bootstrap peer", "addr", bootstrapAddr)
		return
	}
	defer conn.Close()

	p.sendPeerList(bootstrapAddr)
}

func (p *Peer) cleanupPeerList() {
	// get the last seen variable for each peer and remove the ones that haven't been seen for a while
	for id, lastSeen := range p.LastSeen {
		if time.Since(lastSeen) > 30*time.Second { // todo: make timeout configurable?
			delete(p.Peers, id)
			delete(p.LastSeen, id)
		}
	}
	// todo: if there are peers with the same address, but different IDs, remove the oldest one?
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
