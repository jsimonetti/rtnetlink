//go:build integration
// +build integration

package driver

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func macvlanT(d rtnetlink.LinkDriver) *Macvlan {
	m := d.(*Macvlan)
	return &Macvlan{
		Mode:       m.Mode,
		Flags:      m.Flags,
		BcQueueLen: m.BcQueueLen,
		BcCutoff:   m.BcCutoff,
	}
}

func TestMacvlanBridgeMode(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	connNS, err := rtnetlink.Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket to netns: %v", err)
	}
	defer connNS.Close()

	// Create parent interface in netns
	const parentIndex = 1400
	if err := setupInterface(connNS, "macvpar0", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	tests := []struct {
		name   string
		conn   *rtnetlink.Conn
		index  uint32
		driver *Macvlan
		want   *Macvlan
	}{
		{
			name:  "bridge mode",
			conn:  connNS,
			index: 1401,
			driver: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
			},
			want: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
			},
		},
		{
			name:  "bridge mode with nopromisc flag",
			conn:  connNS,
			index: 1402,
			driver: &Macvlan{
				Mode:  func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
				Flags: func() *MacvlanFlag { f := MacvlanFlagNopromisc; return &f }(),
			},
			want: &Macvlan{
				Mode:  func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
				Flags: func() *MacvlanFlag { f := MacvlanFlagNopromisc; return &f }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setupInterface(tt.conn, "macvln", tt.index, parentIndex, tt.driver); err != nil {
				t.Fatalf("failed to create macvlan interface: %v", err)
			}
			defer tt.conn.Link.Delete(tt.index)

			got, err := getInterface(tt.conn, tt.index)
			if err != nil {
				t.Fatalf("failed to get interface: %v", err)
			}

			gotMacvlan := macvlanT(got.Attributes.Info.Data)

			if gotMacvlan.Mode == nil {
				t.Fatal("expected Mode, got nil")
			}

			if tt.want.Mode != nil && *gotMacvlan.Mode != *tt.want.Mode {
				t.Errorf("expected mode %v, got %v", *tt.want.Mode, *gotMacvlan.Mode)
			}

			if tt.want.Flags != nil {
				if gotMacvlan.Flags == nil {
					t.Errorf("expected Flags %v, got nil", *tt.want.Flags)
				} else if *gotMacvlan.Flags != *tt.want.Flags {
					t.Errorf("expected Flags %v, got %v", *tt.want.Flags, *gotMacvlan.Flags)
				}
			}
		})
	}
}

func TestMacvlanDifferentModes(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	connNS, err := rtnetlink.Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket to netns: %v", err)
	}
	defer connNS.Close()

	// Create parent interface in netns
	const parentIndex = 1500
	if err := setupInterface(connNS, "macvpar1", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	tests := []struct {
		name  string
		conn  *rtnetlink.Conn
		index uint32
		mode  MacvlanMode
	}{
		{
			name:  "private mode",
			conn:  connNS,
			index: 1501,
			mode:  MacvlanModePrivate,
		},
		{
			name:  "vepa mode",
			conn:  connNS,
			index: 1502,
			mode:  MacvlanModeVEPA,
		},
		{
			name:  "passthru mode",
			conn:  connNS,
			index: 1503,
			mode:  MacvlanModePassthru,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := &Macvlan{
				Mode: &tt.mode,
			}

			if err := setupInterface(tt.conn, "macvln", tt.index, parentIndex, driver); err != nil {
				t.Fatalf("failed to create macvlan interface: %v", err)
			}
			defer tt.conn.Link.Delete(tt.index)

			got, err := getInterface(tt.conn, tt.index)
			if err != nil {
				t.Fatalf("failed to get interface: %v", err)
			}

			gotMacvlan := macvlanT(got.Attributes.Info.Data)

			if gotMacvlan.Mode == nil {
				t.Fatal("expected Mode, got nil")
			}

			if *gotMacvlan.Mode != tt.mode {
				t.Errorf("expected mode %v (%s), got %v (%s)", tt.mode, tt.mode.String(), *gotMacvlan.Mode, gotMacvlan.Mode.String())
			}
		})
	}
}

func TestMacvlanBroadcastQueueConfig(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	connNS, err := rtnetlink.Dial(&netlink.Config{NetNS: testutils.NetNS(t)})
	if err != nil {
		t.Fatalf("failed to establish netlink socket to netns: %v", err)
	}
	defer connNS.Close()

	// Create parent interface in netns
	const parentIndex = 1600
	if err := setupInterface(connNS, "macvpar2", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	const macvlanIndex = 1601
	mode := MacvlanModeBridge
	qlen := uint32(500)
	cutoff := int32(5)

	driver := &Macvlan{
		Mode:       &mode,
		BcQueueLen: &qlen,
		BcCutoff:   &cutoff,
	}

	if err := setupInterface(connNS, "macvlnbcq", macvlanIndex, parentIndex, driver); err != nil {
		t.Fatalf("failed to create macvlan interface: %v", err)
	}
	defer connNS.Link.Delete(macvlanIndex)

	got, err := getInterface(connNS, macvlanIndex)
	if err != nil {
		t.Fatalf("failed to get interface: %v", err)
	}

	gotMacvlan := macvlanT(got.Attributes.Info.Data)

	if gotMacvlan.BcQueueLen != nil {
		if *gotMacvlan.BcQueueLen != qlen {
			t.Errorf("expected BcQueueLen %v, got %v", qlen, *gotMacvlan.BcQueueLen)
		}
	}

	if gotMacvlan.BcCutoff != nil {
		if *gotMacvlan.BcCutoff != cutoff {
			t.Errorf("expected BcCutoff %v, got %v", cutoff, *gotMacvlan.BcCutoff)
		}
	}
}
