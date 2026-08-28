package controller_test

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/inveracity/voki/internal/agent"
	"github.com/inveracity/voki/internal/controller"
	"github.com/inveracity/voki/internal/peers"
	"github.com/inveracity/voki/internal/remote"
	"github.com/inveracity/voki/internal/targets"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

// TestAgentBondAcceptDispatch runs the full hub lifecycle: agent bonds as
// pending, the administrator accepts, dispatch of exec and upload succeeds.
func TestAgentBondAcceptDispatch(t *testing.T) {
	home := t.TempDir()
	agentHome := t.TempDir()
	addr := freePort(t)

	go func() { _ = controller.Run(home, addr) }()

	go func() { _ = agent.Run(agentHome, addr, 50*time.Millisecond) }()

	dbPath := controller.PeersPath(home)
	waitFor(t, "pending agent", func() bool {
		db, err := peers.Load(dbPath)
		return err == nil && len(db.Peers) == 1
	})

	db, err := peers.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Peers) != 1 || db.Peers[0].Status != peers.Pending {
		t.Fatalf("expected 1 pending peer, got %+v", db.Peers)
	}
	agentName := db.Peers[0].Name

	if err := db.SetStatus(db.Peers[0].ID, peers.Accepted); err != nil {
		t.Fatal(err)
	}

	conn, err := remote.Dial(controller.SocketPath(home))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	ex := conn.Agent(agentName)

	waitFor(t, "agent in pool", func() bool {
		_, _, err := ex.Exec("true")
		return err == nil
	})

	stdout, stderr, err := ex.Exec("echo hub-ok; echo err-out >&2")
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "hub-ok\n" || stderr != "err-out\n" {
		t.Fatalf("unexpected output: %q / %q", stdout, stderr)
	}

	if _, _, err := ex.Exec("exit 7"); err == nil {
		t.Fatal("expected failing command to return an error")
	}

	dst := filepath.Join(t.TempDir(), "uploaded.txt")
	src := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(src, []byte("upload-data"), 0644); err != nil {
		t.Fatal(err)
	}

	err = ex.Upload(targetsStep(src, dst))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "upload-data" {
		t.Fatalf("uploaded content = %q", got)
	}
}

// TestRejectedAgentCannotRun dispatches to an agent the administrator rejected.
func TestRejectedAgentCannotRun(t *testing.T) {
	home := t.TempDir()
	agentHome := t.TempDir()
	addr := freePort(t)

	go func() { _ = controller.Run(home, addr) }()
	go func() { _ = agent.Run(agentHome, addr, 50*time.Millisecond) }()

	dbPath := controller.PeersPath(home)
	waitFor(t, "pending agent", func() bool {
		db, err := peers.Load(dbPath)
		return err == nil && len(db.Peers) == 1
	})

	db, _ := peers.Load(dbPath)
	if err := db.SetStatus(db.Peers[0].ID, peers.Rejected); err != nil {
		t.Fatal(err)
	}

	conn, err := remote.Dial(controller.SocketPath(home))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	_, _, err = conn.Agent(db.Peers[0].Name).Exec("true")
	if err == nil {
		t.Fatal("expected rejected agent dispatch to fail")
	}
}

func targetsStep(src, dst string) targets.Step {
	return targets.Step{Action: "file", Source: src, Destination: dst}
}
