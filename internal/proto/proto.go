package proto

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

const (
	OpBond   = "bond"
	OpExec   = "exec"
	OpUpload = "upload"
)

// MaxFrameSize caps a single message to stop a peer from exhausting memory.
const MaxFrameSize = 1 << 30

type Request struct {
	// Agent routes a request to a connected agent. Empty for agent-to-controller messages.
	Agent string `msgpack:"agent,omitempty"`
	Op    string `msgpack:"op"`
	Body  []byte `msgpack:"body,omitempty"`
}

type Response struct {
	OK    bool   `msgpack:"ok"`
	Error string `msgpack:"error"`
	Body  []byte `msgpack:"body"`
}

type BondRequest struct {
	Name string `msgpack:"name"`
}

type BondResponse struct {
	Status string `msgpack:"status"` // pending | accepted
	PeerID int    `msgpack:"peer_id"`
}

type ExecRequest struct {
	Command string `msgpack:"command"`
}

type ExecResponse struct {
	Stdout string `msgpack:"stdout"`
	Stderr string `msgpack:"stderr"`
}

type UploadRequest struct {
	Destination string `msgpack:"destination"`
	Mode        string `msgpack:"mode"`
	Chown       string `msgpack:"chown"`
	Sudo        bool   `msgpack:"sudo"`
	Data        []byte `msgpack:"data"`
}

func Encode(w io.Writer, v any) error {
	payload, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	if len(payload) > MaxFrameSize {
		return fmt.Errorf("message too large: %d", len(payload))
	}
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(payload)))
	if _, err := w.Write(header[:]); err != nil {
		return err
	}
	_, err = w.Write(payload)
	return err
}

func Decode(r io.Reader, v any) error {
	var header [4]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return err
	}
	size := binary.BigEndian.Uint32(header[:])
	if size > MaxFrameSize {
		return fmt.Errorf("frame too large: %d", size)
	}
	payload := make([]byte, size)
	if _, err := io.ReadFull(r, payload); err != nil {
		return err
	}
	return msgpack.Unmarshal(payload, v)
}

func Marshal(v any) ([]byte, error)   { return msgpack.Marshal(v) }
func Unmarshal(b []byte, v any) error { return msgpack.Unmarshal(b, v) }
