package peers

import (
	"path/filepath"
	"testing"
)

func TestUpsertAndAcceptFlow(t *testing.T) {
	db, err := Load(filepath.Join(t.TempDir(), "peers.json"))
	if err != nil {
		t.Fatal(err)
	}

	id, status := db.UpsertPending("fp-a", "laptop")
	if status != Pending {
		t.Fatalf("new peer should be pending, got %s", status)
	}

	// Bonding twice keeps the same id and does not reset status.
	again, status := db.UpsertPending("fp-a", "laptop")
	if again != id || status != Pending {
		t.Fatalf("re-bond changed peer: id=%d status=%s", again, status)
	}

	if err := db.SetStatus(id, Accepted); err != nil {
		t.Fatal(err)
	}
	if got := db.StatusOf("fp-a"); got != Accepted {
		t.Fatalf("expected accepted, got %s", got)
	}

	if err := db.SetStatus(999, Accepted); err == nil {
		t.Fatal("setting status of unknown id should fail")
	}
}
