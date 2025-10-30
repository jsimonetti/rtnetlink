package driver

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

func TestMacvlanEncodeDecode(t *testing.T) {
	tests := []struct {
		name    string
		macvlan *Macvlan
	}{
		{
			name: "minimal configuration",
			macvlan: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
			},
		},
		{
			name: "bridge mode with flags",
			macvlan: &Macvlan{
				Mode:  func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
				Flags: func() *MacvlanFlag { f := MacvlanFlagNopromisc; return &f }(),
			},
		},
		{
			name: "vepa mode",
			macvlan: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModeVEPA; return &m }(),
			},
		},
		{
			name: "private mode",
			macvlan: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModePrivate; return &m }(),
			},
		},
		{
			name: "passthru mode",
			macvlan: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModePassthru; return &m }(),
			},
		},
		{
			name: "source mode with MAC addresses",
			macvlan: &Macvlan{
				Mode:        func() *MacvlanMode { m := MacvlanModeSource; return &m }(),
				MacaddrMode: func() *MacvlanMacaddrMode { m := MacvlanMacaddrAdd; return &m }(),
				MacaddrData: [][]byte{
					{0x00, 0x11, 0x22, 0x33, 0x44, 0x55},
					{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
				},
				MacaddrCount: func() *uint32 { c := uint32(2); return &c }(),
			},
		},
		{
			name: "full configuration",
			macvlan: &Macvlan{
				Mode:           func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
				Flags:          func() *MacvlanFlag { f := MacvlanFlagNopromisc | MacvlanFlagNodst; return &f }(),
				BcQueueLen:     func() *uint32 { l := uint32(1000); return &l }(),
				BcQueueLenUsed: func() *uint8 { u := uint8(1); return &u }(),
				BcCutoff:       func() *int32 { c := int32(10); return &c }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ae := netlink.NewAttributeEncoder()
			if err := tt.macvlan.Encode(ae); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}

			b, err := ae.Encode()
			if err != nil {
				t.Fatalf("failed to encode attributes: %v", err)
			}

			// Decode
			ad, err := netlink.NewAttributeDecoder(b)
			if err != nil {
				t.Fatalf("failed to create decoder: %v", err)
			}

			decoded := &Macvlan{}
			if err := decoded.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			// Verify
			if tt.macvlan.Mode != nil {
				if decoded.Mode == nil {
					t.Errorf("expected Mode %v, got nil", *tt.macvlan.Mode)
				} else if *decoded.Mode != *tt.macvlan.Mode {
					t.Errorf("expected Mode %v, got %v", *tt.macvlan.Mode, *decoded.Mode)
				}
			}

			if tt.macvlan.Flags != nil {
				if decoded.Flags == nil {
					t.Errorf("expected Flags %v, got nil", *tt.macvlan.Flags)
				} else if *decoded.Flags != *tt.macvlan.Flags {
					t.Errorf("expected Flags %v, got %v", *tt.macvlan.Flags, *decoded.Flags)
				}
			}

			if tt.macvlan.MacaddrMode != nil {
				if decoded.MacaddrMode == nil {
					t.Errorf("expected MacaddrMode %v, got nil", *tt.macvlan.MacaddrMode)
				} else if *decoded.MacaddrMode != *tt.macvlan.MacaddrMode {
					t.Errorf("expected MacaddrMode %v, got %v", *tt.macvlan.MacaddrMode, *decoded.MacaddrMode)
				}
			}

			if len(tt.macvlan.MacaddrData) > 0 {
				if len(decoded.MacaddrData) != len(tt.macvlan.MacaddrData) {
					t.Errorf("expected %d MAC addresses, got %d", len(tt.macvlan.MacaddrData), len(decoded.MacaddrData))
				}
			}

			if tt.macvlan.MacaddrCount != nil {
				if decoded.MacaddrCount == nil {
					t.Errorf("expected MacaddrCount %v, got nil", *tt.macvlan.MacaddrCount)
				} else if *decoded.MacaddrCount != *tt.macvlan.MacaddrCount {
					t.Errorf("expected MacaddrCount %v, got %v", *tt.macvlan.MacaddrCount, *decoded.MacaddrCount)
				}
			}

			if tt.macvlan.BcQueueLen != nil {
				if decoded.BcQueueLen == nil {
					t.Errorf("expected BcQueueLen %v, got nil", *tt.macvlan.BcQueueLen)
				} else if *decoded.BcQueueLen != *tt.macvlan.BcQueueLen {
					t.Errorf("expected BcQueueLen %v, got %v", *tt.macvlan.BcQueueLen, *decoded.BcQueueLen)
				}
			}

			if tt.macvlan.BcQueueLenUsed != nil {
				if decoded.BcQueueLenUsed == nil {
					t.Errorf("expected BcQueueLenUsed %v, got nil", *tt.macvlan.BcQueueLenUsed)
				} else if *decoded.BcQueueLenUsed != *tt.macvlan.BcQueueLenUsed {
					t.Errorf("expected BcQueueLenUsed %v, got %v", *tt.macvlan.BcQueueLenUsed, *decoded.BcQueueLenUsed)
				}
			}

			if tt.macvlan.BcCutoff != nil {
				if decoded.BcCutoff == nil {
					t.Errorf("expected BcCutoff %v, got nil", *tt.macvlan.BcCutoff)
				} else if *decoded.BcCutoff != *tt.macvlan.BcCutoff {
					t.Errorf("expected BcCutoff %v, got %v", *tt.macvlan.BcCutoff, *decoded.BcCutoff)
				}
			}
		})
	}
}

func TestMacvlanModeString(t *testing.T) {
	tests := []struct {
		name     string
		mode     MacvlanMode
		expected string
	}{
		{
			name:     "private",
			mode:     MacvlanModePrivate,
			expected: "private",
		},
		{
			name:     "vepa",
			mode:     MacvlanModeVEPA,
			expected: "vepa",
		},
		{
			name:     "bridge",
			mode:     MacvlanModeBridge,
			expected: "bridge",
		},
		{
			name:     "passthru",
			mode:     MacvlanModePassthru,
			expected: "passthru",
		},
		{
			name:     "source",
			mode:     MacvlanModeSource,
			expected: "source",
		},
		{
			name:     "unknown",
			mode:     MacvlanMode(99),
			expected: "unknown MacvlanMode value (99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestMacvlanKind(t *testing.T) {
	m := &Macvlan{}
	if got := m.Kind(); got != "macvlan" {
		t.Errorf("expected %q, got %q", "macvlan", got)
	}
}

func TestMacvlanNew(t *testing.T) {
	m := &Macvlan{}
	n := m.New()
	if _, ok := n.(*Macvlan); !ok {
		t.Errorf("expected *Macvlan, got %T", n)
	}
}

func TestMacvlanDecodeRaw(t *testing.T) {
	// Test decoding raw netlink data
	tests := []struct {
		name     string
		data     []byte
		expected *Macvlan
	}{
		{
			name: "mode only",
			data: func() []byte {
				ae := netlink.NewAttributeEncoder()
				ae.Uint32(unix.IFLA_MACVLAN_MODE, uint32(MacvlanModeBridge))
				b, _ := ae.Encode()
				return b
			}(),
			expected: &Macvlan{
				Mode: func() *MacvlanMode { m := MacvlanModeBridge; return &m }(),
			},
		},
		{
			name: "mode and flags",
			data: func() []byte {
				ae := netlink.NewAttributeEncoder()
				ae.Uint32(unix.IFLA_MACVLAN_MODE, uint32(MacvlanModeVEPA))
				ae.Uint16(unix.IFLA_MACVLAN_FLAGS, uint16(MacvlanFlagNopromisc))
				b, _ := ae.Encode()
				return b
			}(),
			expected: &Macvlan{
				Mode:  func() *MacvlanMode { m := MacvlanModeVEPA; return &m }(),
				Flags: func() *MacvlanFlag { f := MacvlanFlagNopromisc; return &f }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, err := netlink.NewAttributeDecoder(tt.data)
			if err != nil {
				t.Fatalf("failed to create decoder: %v", err)
			}

			m := &Macvlan{}
			if err := m.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			if tt.expected.Mode != nil {
				if m.Mode == nil {
					t.Errorf("expected Mode %v, got nil", *tt.expected.Mode)
				} else if *m.Mode != *tt.expected.Mode {
					t.Errorf("expected Mode %v, got %v", *tt.expected.Mode, *m.Mode)
				}
			}

			if tt.expected.Flags != nil {
				if m.Flags == nil {
					t.Errorf("expected Flags %v, got nil", *tt.expected.Flags)
				} else if *m.Flags != *tt.expected.Flags {
					t.Errorf("expected Flags %v, got %v", *tt.expected.Flags, *m.Flags)
				}
			}
		})
	}
}
