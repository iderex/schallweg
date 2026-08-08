package acoustic

import (
	"net"
	"testing"
	"time"
)

// TestAnOrdinaryTestMayNotReachTheNetwork exists on this branch only, to be
// watched failing. It is the proof that two separate guards refuse a test that
// opens a connection, and that they refuse it for two different reasons.
//
// The reader in cmd/gate refuses this file for importing net, without running
// anything. The test workflow denies outbound network to the user the suite runs
// as, so the dial below is rejected even if the reader were removed.
//
// This branch is never merged. It is kept so the failing run stays readable.
func TestAnOrdinaryTestMayNotReachTheNetwork(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "example.com:80", 10*time.Second)
	if err != nil {
		t.Fatalf("the dial was refused, which is the point: %v", err)
	}
	conn.Close()
}
