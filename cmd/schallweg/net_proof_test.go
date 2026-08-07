package main

import (
	"net"
	"testing"
	"time"
)

// TestProofOutboundNetworkIsDenied is a deliberate violation of the testability
// rule, added to show that the unit test check fails on it rather than skipping
// it. It dials an address outside the machine, which the ordinary suite may
// never do.
//
// This file is not merged. It exists on a proof branch so the failing run can be
// pointed at from the pull request that adds the check.
func TestProofOutboundNetworkIsDenied(t *testing.T) {
	conn, err := net.DialTimeout("tcp", "93.184.216.34:443", 10*time.Second)
	if err != nil {
		t.Fatalf("outbound connection refused, which is what the job is supposed to cause: %v", err)
	}
	conn.Close()
	t.Log("outbound connection succeeded, which means the job did not deny the network")
}
