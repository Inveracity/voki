package controller

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/inveracity/voki/internal/identity"
	"github.com/inveracity/voki/internal/peers"
	"github.com/inveracity/voki/internal/proto"
)

func Home(home string) string       { return filepath.Join(home, "controller") }
func PeersPath(home string) string  { return filepath.Join(home, "controller", "peers.json") }
func SocketPath(home string) string { return filepath.Join(home, "controller", "voki.sock") }

// agentConn is one live agent connection. The mutex serializes dispatches so
// two concurrent runs against the same agent can not interleave frames.
type agentConn struct {
	mu   sync.Mutex
	conn net.Conn
	name string
}

// Hub holds accepted agent connections and serves CLI dispatch over a unix socket.
type Hub struct {
	mu     sync.Mutex
	agents map[string]*agentConn

	dbPath     string
	socketPath string
}

func (h *Hub) get(name string) *agentConn {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.agents[name]
}

func (h *Hub) put(name string, c *agentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, exists := h.agents[name]; exists {
		_ = old.conn.Close()
		log.Printf("agent %s reconnected, replacing old connection", name)
	}
	h.agents[name] = c
}

func (h *Hub) remove(name string, c *agentConn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, exists := h.agents[name]; exists && cur == c {
		delete(h.agents, name)
	}
}

// Run starts the controller daemon: a TLS listener for agents and a unix
// socket for the local CLI. Agents are trusted only after an administrator
// accepted them (trust on first use).
func Run(home, addr string) error {
	id, err := identity.Ensure(Home(home))
	if err != nil {
		return err
	}

	hub := &Hub{
		agents:     map[string]*agentConn{},
		dbPath:     PeersPath(home),
		socketPath: SocketPath(home),
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{id.Certificate},
		ClientAuth:   tls.RequireAnyClientCert,
		MinVersion:   tls.VersionTLS13,
		// Every connection must present the agent certificate for verification,
		// so resumption is disabled.
		SessionTicketsDisabled: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			db, err := peers.Load(hub.dbPath)
			if err != nil {
				return err
			}
			if db.StatusOf(identity.Fingerprint(rawCerts[0])) == peers.Rejected {
				return fmt.Errorf("peer rejected")
			}
			// Unknown and pending agents pass the handshake but may only bond.
			return nil
		},
	}

	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return err
	}
	log.Printf("voki controller listening on %s, fingerprint %s", addr, id.Fingerprint[:16])

	if err := os.Remove(hub.socketPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	sockListener, err := net.Listen("unix", hub.socketPath)
	if err != nil {
		return err
	}
	if err := os.Chmod(hub.socketPath, 0600); err != nil {
		return err
	}
	log.Printf("cli socket at %s", hub.socketPath)

	go func() {
		for {
			conn, err := sockListener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go hub.serveCLI(conn)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go hub.handleAgent(conn)
	}
}

// handleAgent performs the bond handshake and, once accepted, parks the
// connection in the pool for dispatch. A pooled connection is owned by the pool.
func (h *Hub) handleAgent(conn net.Conn) {
	tlsConn := conn.(*tls.Conn)
	// tls.Listen defers the handshake to first use, so force it here.
	if err := tlsConn.Handshake(); err != nil {
		_ = conn.Close()
		return
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		_ = conn.Close()
		return
	}
	fp := identity.Fingerprint(state.PeerCertificates[0].Raw)

	db, err := peers.Load(h.dbPath)
	if err != nil {
		log.Println("load peers:", err)
		_ = conn.Close()
		return
	}
	// ponytail: handshake hook rejects too, but TLS 1.3 resumption can skip it,
	// so this connection-level check is the real gate.
	if db.StatusOf(fp) == peers.Rejected {
		_ = conn.Close()
		return
	}

	var req proto.Request
	if err := proto.Decode(conn, &req); err != nil || req.Op != proto.OpBond {
		_ = conn.Close()
		return
	}
	var bond proto.BondRequest
	if err := proto.Unmarshal(req.Body, &bond); err != nil {
		_ = conn.Close()
		return
	}

	id, status := db.UpsertPending(fp, bond.Name)
	if err := db.Save(); err != nil {
		log.Println("save peers:", err)
		_ = conn.Close()
		return
	}

	body, err := proto.Marshal(proto.BondResponse{Status: string(status), PeerID: id})
	if err != nil {
		_ = conn.Close()
		return
	}
	if err := proto.Encode(conn, &proto.Response{OK: true, Body: body}); err != nil {
		_ = conn.Close()
		return
	}

	if status != peers.Accepted {
		_ = conn.Close()
		return
	}

	// ponytail: no keepalive read here — it would race with dispatch reads on
	// the same connection. Dead connections surface when a dispatch fails.
	// Add agent pings if stale connection detection matters.
	h.put(bond.Name, &agentConn{conn: conn, name: bond.Name})
	log.Printf("agent %s connected (peer %d)", bond.Name, id)
}

// serveCLI routes CLI requests to a connected agent.
func (h *Hub) serveCLI(conn net.Conn) {
	defer func() { _ = conn.Close() }()
	for {
		var req proto.Request
		if err := proto.Decode(conn, &req); err != nil {
			return
		}

		resp := h.dispatch(&req)
		if err := proto.Encode(conn, resp); err != nil {
			return
		}
	}
}

func (h *Hub) dispatch(req *proto.Request) *proto.Response {
	agent := h.get(req.Agent)
	if agent == nil {
		return &proto.Response{Error: fmt.Sprintf("agent %q not connected", req.Agent)}
	}

	agent.mu.Lock()
	defer agent.mu.Unlock()

	if err := proto.Encode(agent.conn, &proto.Request{Op: req.Op, Body: req.Body}); err != nil {
		h.remove(agent.name, agent)
		return &proto.Response{Error: fmt.Sprintf("agent %s connection lost: %s", agent.name, err)}
	}

	var resp proto.Response
	if err := proto.Decode(agent.conn, &resp); err != nil {
		h.remove(agent.name, agent)
		return &proto.Response{Error: fmt.Sprintf("agent %s connection lost: %s", agent.name, err)}
	}
	return &resp
}
