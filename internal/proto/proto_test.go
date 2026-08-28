package proto

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoundtrip(t *testing.T) {
	var buf bytes.Buffer

	req := Request{Op: OpExec, Body: []byte("payload")}
	if err := Encode(&buf, &req); err != nil {
		t.Fatal(err)
	}

	var got Request
	if err := Decode(&buf, &got); err != nil {
		t.Fatal(err)
	}
	if got.Op != req.Op || string(got.Body) != "payload" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestDecodeRejectsOversizedFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // claims ~4GB

	var got Request
	if err := Decode(&buf, &got); err == nil {
		t.Fatal("oversized frame accepted")
	}
}
