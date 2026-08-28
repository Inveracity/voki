package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/inveracity/voki/internal/identity"
	"github.com/inveracity/voki/internal/proto"
)

func Home(home string) string { return filepath.Join(home, "agent") }

func ControllerPinPath(home string) string {
	return filepath.Join(home, "agent", "controller.json")
}

// Run dials the controller, bonds, and serves exec/upload requests until the
// process stops. While pending acceptance it retries until an administrator
// runs 'voki peers accept'.
func Run(home, controller string, retry time.Duration) error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	for {
		if err := once(home, controller, name); err != nil {
			log.Println(err)
		}
		time.Sleep(retry)
	}
}

// once performs one connect/bond/serve cycle. It returns nil when the
// connection ends, so the caller can reconnect.
func once(home, controller, name string) error {
	id, err := identity.Ensure(Home(home))
	if err != nil {
		return err
	}

	pins, err := identity.LoadPins(ControllerPinPath(home))
	if err != nil {
		return err
	}

	var seen string
	cfg := &tls.Config{
		Certificates:       []tls.Certificate{id.Certificate},
		InsecureSkipVerify: true, // pin check below is the real verification
		MinVersion:         tls.VersionTLS13,
		VerifyPeerCertificate: func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
			seen = identity.Fingerprint(rawCerts[0])
			return nil
		},
	}

	conn, err := tls.Dial("tcp", controller, cfg)
	if err != nil {
		return fmt.Errorf("connect to controller %s: %w", controller, err)
	}
	defer func() { _ = conn.Close() }()

	// Trust on first use: pin the controller on the first successful connect.
	if _, ok := pins.Host[controller]; !ok {
		if err := pins.Put(controller, seen); err != nil {
			return err
		}
		log.Printf("pinned controller fingerprint %s", seen[:16])
	} else if err := pins.Check(controller, seen); err != nil {
		return err
	}

	// Bond with the controller.
	body, err := proto.Marshal(proto.BondRequest{Name: name})
	if err != nil {
		return err
	}
	if err := proto.Encode(conn, &proto.Request{Op: proto.OpBond, Body: body}); err != nil {
		return err
	}

	var resp proto.Response
	if err := proto.Decode(conn, &resp); err != nil {
		return fmt.Errorf("bond: %w", err)
	}
	if !resp.OK {
		return fmt.Errorf("bond rejected: %s", resp.Error)
	}

	var bond proto.BondResponse
	if err := proto.Unmarshal(resp.Body, &bond); err != nil {
		return err
	}
	if bond.Status != "accepted" {
		log.Printf("waiting for acceptance on the controller (peer %d, 'voki peers accept %d')", bond.PeerID, bond.PeerID)
		return nil
	}

	log.Printf("connected to controller %s as agent %q", controller, name)
	for {
		var req proto.Request
		if err := proto.Decode(conn, &req); err != nil {
			return fmt.Errorf("controller connection lost: %w", err)
		}

		var out *proto.Response
		switch req.Op {
		case proto.OpExec:
			out = serveExec(req)
		case proto.OpUpload:
			out = serveUpload(req)
		default:
			out = &proto.Response{Error: "unknown op " + req.Op}
		}

		if err := proto.Encode(conn, out); err != nil {
			return fmt.Errorf("respond: %w", err)
		}
	}
}

func serveExec(req proto.Request) *proto.Response {
	var in proto.ExecRequest
	if err := proto.Unmarshal(req.Body, &in); err != nil {
		return &proto.Response{Error: err.Error()}
	}

	cmd := exec.Command("bash", "-c", in.Command)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	body, _ := proto.Marshal(proto.ExecResponse{Stdout: stdout.String(), Stderr: stderr.String()})
	if err != nil {
		return &proto.Response{Error: err.Error(), Body: body}
	}
	return &proto.Response{OK: true, Body: body}
}

func serveUpload(req proto.Request) *proto.Response {
	var in proto.UploadRequest
	if err := proto.Unmarshal(req.Body, &in); err != nil {
		return &proto.Response{Error: err.Error()}
	}

	tmp, err := os.CreateTemp("", "voki-upload-*")
	if err != nil {
		return &proto.Response{Error: err.Error()}
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := tmp.Write(in.Data); err != nil {
		_ = tmp.Close()
		return &proto.Response{Error: err.Error()}
	}
	_ = tmp.Close()

	cmdline := fmt.Sprintf("mkdir -p %q && mv %q %q", filepath.Dir(in.Destination), tmpPath, in.Destination)
	if in.Mode != "" {
		cmdline += fmt.Sprintf(" && chmod %s %q", in.Mode, in.Destination)
	}
	if in.Chown != "" {
		cmdline += fmt.Sprintf(" && chown %s %q", in.Chown, in.Destination)
	}
	if in.Sudo {
		cmdline = "sudo " + cmdline
	}

	if out, err := exec.Command("bash", "-c", cmdline).CombinedOutput(); err != nil {
		return &proto.Response{Error: fmt.Errorf("%s: %s", err, out).Error()}
	}
	return &proto.Response{OK: true}
}
