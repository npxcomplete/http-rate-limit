package ratelimit

import (
	"testing"
	"time"
)

func TestPeerGroup_SetGetDelete(t *testing.T) {
	pg := NewPeerGroup()
	if _, ok := pg.Get("alice"); ok {
		t.Fatalf("expected no entry")
	}

	c := &Capacity{Burst: 10, Rate: 5, Available: 5, LastUpdated: time.Now()}
	pg.Set("alice", c)

	rc, ok := pg.Get("alice")
	if !ok {
		t.Fatalf("expected entry to exist")
	}
	if rc != c {
		t.Fatalf("returned capacity mismatch")
	}

	pg.Delete("alice")
	if _, ok := pg.Get("alice"); ok {
		t.Fatalf("expected entry removed")
	}
}
