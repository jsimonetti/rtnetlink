//go:build integration
// +build integration

package rtnl

import (
	"net"
	"testing"

	"golang.org/x/sys/unix"
)

func TestLiveNeighbours(t *testing.T) {
	c, err := Dial(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Trigger a DNS lookup, only for a side effect of pushing our gateway or NS onto the neighbour table
	net.LookupHost("github.com")

	neigtab, err := c.Neighbours(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(neigtab) == 0 {
		t.Skip("no neighbours")
	}
	for i, e := range neigtab {
		t.Logf("* neighbour table entry [%d]: %v", i, e)

		// Ignore neighbor entries in internal/pseudo state.
		if e.State == unix.NUD_NONE {
			continue
		}
		// Loopback and p2p interfaces can have neigh entries with a zero IP address.
		if e.IP.IsUnspecified() {
			continue
		}

		if e.Interface == nil {
			t.Error("nil e.Interface, expected non-nil")
		}
		if len(e.Interface.Name) == 0 {
			t.Error("zero-length e.Interface.Name")
		}

		// Don't (always) expect hardware address info on entries marked noarp,
		// as they include link-local multicast and loopback addresses that are
		// valid for all interfaces on the host.
		if e.State == unix.NUD_NOARP {
			continue
		}

		if hardwareAddrIsUnspecified(e.HwAddr) {
			t.Error("zero e.HwAddr, expected non-zero")
		}
		if hardwareAddrIsUnspecified(e.Interface.HardwareAddr) {
			t.Error("zero e.Interface.HardwareAddr, expected non-zero")
		}
	}
}
