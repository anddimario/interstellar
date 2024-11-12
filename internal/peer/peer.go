package peer

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Peer struct {
	ID    string
	Addr  string
	Peers map[string]string // ID -> Addr
	mu    sync.Mutex
}

var PeeringDone = make(chan bool)

func NewPeer(id, addr string) *Peer {
	return &Peer{
		ID:    id,
		Addr:  addr,
		Peers: make(map[string]string),
	}
}
func (p *Peer) Gossip() {
	t := time.NewTicker(5 * time.Second) // todo: make interval configurable?
	defer t.Stop()

	for {
		select {
		case <-t.C:
			fmt.Println("Gossiping...")
			p.mu.Lock()
			peers := p.getRandomPeers(2) // Select 2 random peers to gossip with
			p.mu.Unlock()

			for _, addr := range peers {
				go p.sendPeerList(addr)
			}
		case <-PeeringDone:
			return
		}
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

func (p *Peer) sendPeerList(addr string) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Println("Failed to connect to peer:", err)
		return
	}
	defer conn.Close()

	// Serialize and send peers as JSON (or another format)
	message := fmt.Sprintf("PEERS %s\n", p.Addr)
    conn.Write([]byte(message))
}

func (p *Peer) Listen() {
	udpAddr, _ := net.ResolveUDPAddr("udp", p.Addr)
	conn, _ := net.ListenUDP("udp", udpAddr)
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		fmt.Println("Listening...")
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		message := string(buf[:n])
        fmt.Println("Received message:", message)
		p.handleMessage(message, remoteAddr)
	}

}

func (p *Peer) handleMessage(message string, addr *net.UDPAddr) {
	fmt.Printf("Received message from %s: %s\n", addr.String(), message)

	// Here we can parse message type and content
	if message[:6] == "PEERS " {
		newAddr := message[6:]
		p.mu.Lock()
		p.Peers[addr.String()] = newAddr
		p.mu.Unlock()
		fmt.Println("Added peer:", newAddr)
	}
}

func (p *Peer) Bootstrap(bootstrapAddr string) {
	conn, err := net.Dial("udp", bootstrapAddr)
	if err != nil {
		fmt.Println("Could not connect to bootstrap peer:", err)
		return
	}
	defer conn.Close()
	p.sendPeerList(bootstrapAddr)
}
