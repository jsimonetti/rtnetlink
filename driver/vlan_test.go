package driver

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

func TestVlanEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		vlan *Vlan
	}{
		{
			name: "minimal configuration",
			vlan: &Vlan{
				ID: func() *uint16 { id := uint16(100); return &id }(),
			},
		},
		{
			name: "with 802.1Q protocol",
			vlan: &Vlan{
				ID:       func() *uint16 { id := uint16(200); return &id }(),
				Protocol: func() *VlanProtocol { p := VlanProtocol8021Q; return &p }(),
			},
		},
		{
			name: "with 802.1ad protocol",
			vlan: &Vlan{
				ID:       func() *uint16 { id := uint16(300); return &id }(),
				Protocol: func() *VlanProtocol { p := VlanProtocol8021AD; return &p }(),
			},
		},
		{
			name: "with reorder header flag",
			vlan: &Vlan{
				ID:    func() *uint16 { id := uint16(400); return &id }(),
				Flags: func() *VlanFlag { f := VlanFlagReorderHdr; return &f }(),
			},
		},
		{
			name: "with multiple flags",
			vlan: &Vlan{
				ID:    func() *uint16 { id := uint16(500); return &id }(),
				Flags: func() *VlanFlag { f := VlanFlagReorderHdr | VlanFlagGVRP; return &f }(),
			},
		},
		{
			name: "with egress QoS mapping",
			vlan: &Vlan{
				ID: func() *uint16 { id := uint16(600); return &id }(),
				EgressQos: []VlanQosMapping{
					{From: 0, To: 1},
					{From: 2, To: 3},
				},
			},
		},
		{
			name: "with ingress QoS mapping",
			vlan: &Vlan{
				ID: func() *uint16 { id := uint16(700); return &id }(),
				IngressQos: []VlanQosMapping{
					{From: 1, To: 0},
					{From: 3, To: 2},
				},
			},
		},
		{
			name: "full configuration",
			vlan: &Vlan{
				ID:       func() *uint16 { id := uint16(800); return &id }(),
				Protocol: func() *VlanProtocol { p := VlanProtocol8021Q; return &p }(),
				Flags:    func() *VlanFlag { f := VlanFlagReorderHdr | VlanFlagLooseBinding; return &f }(),
				EgressQos: []VlanQosMapping{
					{From: 0, To: 2},
					{From: 4, To: 6},
				},
				IngressQos: []VlanQosMapping{
					{From: 2, To: 0},
					{From: 6, To: 4},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ae := netlink.NewAttributeEncoder()
			if err := tt.vlan.Encode(ae); err != nil {
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

			decoded := &Vlan{}
			if err := decoded.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			// Verify ID
			if tt.vlan.ID != nil {
				if decoded.ID == nil {
					t.Errorf("expected ID %d, got nil", *tt.vlan.ID)
				} else if *decoded.ID != *tt.vlan.ID {
					t.Errorf("expected ID %d, got %d", *tt.vlan.ID, *decoded.ID)
				}
			}

			// Verify Protocol
			if tt.vlan.Protocol != nil {
				if decoded.Protocol == nil {
					t.Errorf("expected Protocol %v, got nil", *tt.vlan.Protocol)
				} else if *decoded.Protocol != *tt.vlan.Protocol {
					t.Errorf("expected Protocol %v, got %v", *tt.vlan.Protocol, *decoded.Protocol)
				}
			}

			// Verify Flags
			if tt.vlan.Flags != nil {
				if decoded.Flags == nil {
					t.Errorf("expected Flags %v, got nil", *tt.vlan.Flags)
				} else if *decoded.Flags != *tt.vlan.Flags {
					t.Errorf("expected Flags %v, got %v", *tt.vlan.Flags, *decoded.Flags)
				}
			}

			// Verify EgressQos
			if len(tt.vlan.EgressQos) > 0 {
				if len(decoded.EgressQos) != len(tt.vlan.EgressQos) {
					t.Errorf("expected %d egress QoS mappings, got %d", len(tt.vlan.EgressQos), len(decoded.EgressQos))
				} else {
					for i := range tt.vlan.EgressQos {
						if decoded.EgressQos[i].From != tt.vlan.EgressQos[i].From ||
							decoded.EgressQos[i].To != tt.vlan.EgressQos[i].To {
							t.Errorf("egress QoS mapping %d: expected {%d -> %d}, got {%d -> %d}",
								i, tt.vlan.EgressQos[i].From, tt.vlan.EgressQos[i].To,
								decoded.EgressQos[i].From, decoded.EgressQos[i].To)
						}
					}
				}
			}

			// Verify IngressQos
			if len(tt.vlan.IngressQos) > 0 {
				if len(decoded.IngressQos) != len(tt.vlan.IngressQos) {
					t.Errorf("expected %d ingress QoS mappings, got %d", len(tt.vlan.IngressQos), len(decoded.IngressQos))
				} else {
					for i := range tt.vlan.IngressQos {
						if decoded.IngressQos[i].From != tt.vlan.IngressQos[i].From ||
							decoded.IngressQos[i].To != tt.vlan.IngressQos[i].To {
							t.Errorf("ingress QoS mapping %d: expected {%d -> %d}, got {%d -> %d}",
								i, tt.vlan.IngressQos[i].From, tt.vlan.IngressQos[i].To,
								decoded.IngressQos[i].From, decoded.IngressQos[i].To)
						}
					}
				}
			}
		})
	}
}

func TestVlanProtocolString(t *testing.T) {
	tests := []struct {
		name     string
		protocol VlanProtocol
		expected string
	}{
		{
			name:     "802.1Q",
			protocol: VlanProtocol8021Q,
			expected: "802.1Q",
		},
		{
			name:     "802.1ad",
			protocol: VlanProtocol8021AD,
			expected: "802.1ad",
		},
		{
			name:     "unknown",
			protocol: VlanProtocol(0x9999),
			expected: "unknown VlanProtocol value (0x9999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.protocol.String(); got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestVlanKind(t *testing.T) {
	v := &Vlan{}
	if got, want := v.Kind(), "vlan"; got != want {
		t.Errorf("expected kind %q, got %q", want, got)
	}
}

func TestVlanNew(t *testing.T) {
	v := &Vlan{}
	newV := v.New()

	if _, ok := newV.(*Vlan); !ok {
		t.Errorf("expected *Vlan, got %T", newV)
	}
}

func TestVlanDecodeRaw(t *testing.T) {
	// Test decoding raw netlink data
	tests := []struct {
		name     string
		data     []byte
		expected *Vlan
	}{
		{
			name: "ID and protocol",
			data: func() []byte {
				ae := netlink.NewAttributeEncoder()
				ae.Uint16(unix.IFLA_VLAN_ID, 100)
				ae.Uint16(unix.IFLA_VLAN_PROTOCOL, uint16(VlanProtocol8021Q))
				b, _ := ae.Encode()
				return b
			}(),
			expected: &Vlan{
				ID:       func() *uint16 { id := uint16(100); return &id }(),
				Protocol: func() *VlanProtocol { p := VlanProtocol8021Q; return &p }(),
			},
		},
		{
			name: "ID and flags",
			data: func() []byte {
				ae := netlink.NewAttributeEncoder()
				ae.Uint16(unix.IFLA_VLAN_ID, 200)
				ae.Nested(unix.IFLA_VLAN_FLAGS, func(nae *netlink.AttributeEncoder) error {
					nae.Uint32(1, uint32(VlanFlagReorderHdr))
					nae.Uint32(2, uint32(VlanFlagReorderHdr))
					return nil
				})
				b, _ := ae.Encode()
				return b
			}(),
			expected: &Vlan{
				ID:    func() *uint16 { id := uint16(200); return &id }(),
				Flags: func() *VlanFlag { f := VlanFlagReorderHdr; return &f }(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, err := netlink.NewAttributeDecoder(tt.data)
			if err != nil {
				t.Fatalf("failed to create decoder: %v", err)
			}

			v := &Vlan{}
			if err := v.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			if tt.expected.ID != nil {
				if v.ID == nil {
					t.Errorf("expected ID %v, got nil", *tt.expected.ID)
				} else if *v.ID != *tt.expected.ID {
					t.Errorf("expected ID %v, got %v", *tt.expected.ID, *v.ID)
				}
			}

			if tt.expected.Protocol != nil {
				if v.Protocol == nil {
					t.Errorf("expected Protocol %v, got nil", *tt.expected.Protocol)
				} else if *v.Protocol != *tt.expected.Protocol {
					t.Errorf("expected Protocol %v, got %v", *tt.expected.Protocol, *v.Protocol)
				}
			}

			if tt.expected.Flags != nil {
				if v.Flags == nil {
					t.Errorf("expected Flags %v, got nil", *tt.expected.Flags)
				} else if *v.Flags != *tt.expected.Flags {
					t.Errorf("expected Flags %v, got %v", *tt.expected.Flags, *v.Flags)
				}
			}
		})
	}
}
