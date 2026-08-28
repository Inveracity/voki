package identity

import (
	"path/filepath"
	"testing"
)

func TestEnsureIsStable(t *testing.T) {
	dir := t.TempDir()

	first, err := Ensure(dir)
	if err != nil {
		t.Fatal(err)
	}
	again, err := Ensure(dir)
	if err != nil {
		t.Fatal(err)
	}

	if first.Fingerprint != again.Fingerprint {
		t.Fatalf("fingerprint changed across loads: %s != %s", first.Fingerprint, again.Fingerprint)
	}
	if len(first.Fingerprint) != 64 {
		t.Fatalf("expected sha256 hex fingerprint, got %q", first.Fingerprint)
	}

	// A different directory must give a different identity.
	other, err := Ensure(filepath.Join(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if other.Fingerprint == first.Fingerprint {
		t.Fatal("two generated identities share a fingerprint")
	}
}

func TestPinsCheck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pins.json")
	pins, err := LoadPins(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := pins.Check("srv:7331", "aaa"); err == nil {
		t.Fatal("expected error for unpinned host")
	}

	if err := pins.Put("srv:7331", "aaa"); err != nil {
		t.Fatal(err)
	}

	if err := pins.Check("srv:7331", "aaa"); err != nil {
		t.Fatalf("pinned fingerprint rejected: %s", err)
	}
	if err := pins.Check("srv:7331", "bbb"); err == nil {
		t.Fatal("changed fingerprint accepted")
	}

	reloaded, err := LoadPins(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := reloaded.Check("srv:7331", "aaa"); err != nil {
		t.Fatalf("pin not persisted: %s", err)
	}
}
