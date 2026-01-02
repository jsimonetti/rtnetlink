package driver

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

func TestBridgeEncodeDecode(t *testing.T) {
	tests := []struct {
		name   string
		bridge *Bridge
	}{
		{
			name:   "minimal configuration",
			bridge: &Bridge{},
		},
		{
			name: "basic STP configuration",
			bridge: &Bridge{
				StpState:     ptr(BridgeStpStateEnabled),
				Priority:     ptr(uint16(32768)),
				ForwardDelay: ptr(uint32(1500)),
				HelloTime:    ptr(uint32(200)),
				MaxAge:       ptr(uint32(2000)),
			},
		},
		{
			name: "with VLAN filtering",
			bridge: &Bridge{
				VlanFiltering:   ptr(BridgeEnableEnabled),
				VlanProtocol:    ptr(VlanProtocol8021Q),
				VlanDefaultPvid: ptr(uint16(1)),
			},
		},
		{
			name: "with multicast settings",
			bridge: &Bridge{
				McastSnooping:       ptr(BridgeEnableEnabled),
				McastQuerier:        ptr(BridgeEnableEnabled),
				McastHashElasticity: ptr(uint32(4)),
				McastHashMax:        ptr(uint32(512)),
				McastIgmpVersion:    ptr(uint8(3)),
				McastMldVersion:     ptr(uint8(2)),
			},
		},
		{
			name: "with netfilter settings",
			bridge: &Bridge{
				NfCallIptables:  ptr(BridgeEnableEnabled),
				NfCallIp6tables: ptr(BridgeEnableEnabled),
				NfCallArptables: ptr(BridgeEnableEnabled),
			},
		},
		{
			name: "full configuration",
			bridge: &Bridge{
				ForwardDelay:            ptr(uint32(1500)),
				HelloTime:               ptr(uint32(200)),
				MaxAge:                  ptr(uint32(2000)),
				AgeingTime:              ptr(uint32(30000)),
				StpState:                ptr(BridgeStpStateEnabled),
				Priority:                ptr(uint16(32768)),
				VlanFiltering:           ptr(BridgeEnableEnabled),
				VlanProtocol:            ptr(VlanProtocol8021Q),
				GroupFwdMask:            ptr(uint16(0)),
				GroupAddr:               net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x00},
				McastRouter:             ptr(uint8(1)),
				McastSnooping:           ptr(BridgeEnableEnabled),
				McastQueryUseIfaddr:     ptr(BridgeEnableDisabled),
				McastQuerier:            ptr(BridgeEnableEnabled),
				McastHashElasticity:     ptr(uint32(4)),
				McastHashMax:            ptr(uint32(512)),
				McastLastMemberCnt:      ptr(uint32(2)),
				McastStartupQueryCnt:    ptr(uint32(2)),
				McastLastMemberIntvl:    ptr(uint64(100)),
				McastMembershipIntvl:    ptr(uint64(26000)),
				McastQuerierIntvl:       ptr(uint64(25500)),
				McastQueryIntvl:         ptr(uint64(12500)),
				McastQueryResponseIntvl: ptr(uint64(1000)),
				McastStartupQueryIntvl:  ptr(uint64(3125)),
				NfCallIptables:          ptr(BridgeEnableDisabled),
				NfCallIp6tables:         ptr(BridgeEnableDisabled),
				NfCallArptables:         ptr(BridgeEnableDisabled),
				VlanDefaultPvid:         ptr(uint16(1)),
				VlanStatsEnabled:        ptr(BridgeEnableDisabled),
				McastStatsEnabled:       ptr(BridgeEnableDisabled),
				McastIgmpVersion:        ptr(uint8(2)),
				McastMldVersion:         ptr(uint8(1)),
				VlanStatsPerPort:        ptr(BridgeEnableDisabled),
				FdbMaxLearned:           ptr(uint32(0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ae := netlink.NewAttributeEncoder()
			if err := tt.bridge.Encode(ae); err != nil {
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

			decoded := &Bridge{}
			if err := decoded.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			// Compare
			if diff := cmp.Diff(tt.bridge, decoded); diff != "" {
				t.Fatalf("unexpected bridge (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridgeKind(t *testing.T) {
	b := &Bridge{}
	if got, want := b.Kind(), "bridge"; got != want {
		t.Fatalf("unexpected Kind:\n got: %q\nwant: %q", got, want)
	}
}

func TestBridgeNew(t *testing.T) {
	b := &Bridge{}
	newBridge := b.New()
	if _, ok := newBridge.(*Bridge); !ok {
		t.Fatalf("New() returned wrong type: %T", newBridge)
	}
}

func TestBridgeStpStateString(t *testing.T) {
	tests := []struct {
		state BridgeStpState
		want  string
	}{
		{BridgeStpStateDisabled, "disabled"},
		{BridgeStpStateEnabled, "enabled"},
		{BridgeStpState(99), "unknown BridgeStpState value (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Fatalf("unexpected string:\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestBridgeEnableString(t *testing.T) {
	tests := []struct {
		enable BridgeEnable
		want   string
	}{
		{BridgeEnableDisabled, "disabled"},
		{BridgeEnableEnabled, "enabled"},
		{BridgeEnable(99), "unknown BridgeEnable value (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.enable.String(); got != tt.want {
				t.Fatalf("unexpected string:\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestBridgeDecodeRaw(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want *Bridge
	}{
		{
			name: "STP state",
			b: []byte{
				0x08, 0x00, // Length: 8
				unix.IFLA_BR_STP_STATE, 0x00, // Type: IFLA_BR_STP_STATE
				0x01, 0x00, 0x00, 0x00, // Value: 1
			},
			want: &Bridge{
				StpState: ptr(BridgeStpStateEnabled),
			},
		},
		{
			name: "VLAN filtering",
			b: []byte{
				0x05, 0x00, // Length: 5
				unix.IFLA_BR_VLAN_FILTERING, 0x00, // Type: IFLA_BR_VLAN_FILTERING
				0x01,             // Value: 1
				0x00, 0x00, 0x00, // Padding
			},
			want: &Bridge{
				VlanFiltering: ptr(BridgeEnableEnabled),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, err := netlink.NewAttributeDecoder(tt.b)
			if err != nil {
				t.Fatalf("failed to create decoder: %v", err)
			}

			got := &Bridge{}
			if err := got.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("unexpected bridge (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridgePortStateString(t *testing.T) {
	tests := []struct {
		state BridgePortState
		want  string
	}{
		{BridgePortStateDisabled, "disabled"},
		{BridgePortStateListening, "listening"},
		{BridgePortStateLearning, "learning"},
		{BridgePortStateForwarding, "forwarding"},
		{BridgePortStateBlocking, "blocking"},
		{BridgePortState(99), "unknown BridgePortState value (99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Fatalf("unexpected string:\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestBridgePortEncodeDecode(t *testing.T) {
	tests := []struct {
		name string
		port *BridgePort
	}{
		{
			name: "minimal configuration",
			port: &BridgePort{},
		},
		{
			name: "basic port configuration",
			port: &BridgePort{
				State:    ptr(BridgePortStateForwarding),
				Priority: ptr(uint16(32)),
				Cost:     ptr(uint32(100)),
			},
		},
		{
			name: "with hairpin and guard",
			port: &BridgePort{
				Mode:  ptr(BridgeEnableEnabled), // Hairpin mode
				Guard: ptr(BridgeEnableEnabled), // BPDU guard
			},
		},
		{
			name: "with flood control",
			port: &BridgePort{
				Learning:     ptr(BridgeEnableEnabled),
				UnicastFlood: ptr(BridgeEnableEnabled),
				McastFlood:   ptr(BridgeEnableEnabled),
				BcastFlood:   ptr(BridgeEnableEnabled),
			},
		},
		{
			name: "with proxy ARP and isolation",
			port: &BridgePort{
				ProxyArp:     ptr(BridgeEnableEnabled),
				ProxyArpWifi: ptr(BridgeEnableEnabled),
				Isolated:     ptr(BridgeEnableEnabled),
			},
		},
		{
			name: "full configuration",
			port: &BridgePort{
				State:             ptr(BridgePortStateForwarding),
				Priority:          ptr(uint16(32)),
				Cost:              ptr(uint32(100)),
				Mode:              ptr(BridgeEnableDisabled),
				Guard:             ptr(BridgeEnableDisabled),
				Protect:           ptr(BridgeEnableDisabled),
				FastLeave:         ptr(BridgeEnableDisabled),
				Learning:          ptr(BridgeEnableEnabled),
				UnicastFlood:      ptr(BridgeEnableEnabled),
				ProxyArp:          ptr(BridgeEnableDisabled),
				LearningSync:      ptr(BridgeEnableDisabled),
				ProxyArpWifi:      ptr(BridgeEnableDisabled),
				MulticastRouter:   ptr(uint8(1)),
				McastFlood:        ptr(BridgeEnableEnabled),
				McastToUcast:      ptr(BridgeEnableDisabled),
				VlanTunnel:        ptr(BridgeEnableDisabled),
				BcastFlood:        ptr(BridgeEnableEnabled),
				GroupFwdMask:      ptr(uint16(0)),
				NeighSuppress:     ptr(BridgeEnableDisabled),
				Isolated:          ptr(BridgeEnableDisabled),
				Locked:            ptr(BridgeEnableDisabled),
				Mab:               ptr(BridgeEnableDisabled),
				NeighVlanSuppress: ptr(BridgeEnableDisabled),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encode
			ae := netlink.NewAttributeEncoder()
			if err := tt.port.Encode(ae); err != nil {
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

			decoded := &BridgePort{}
			if err := decoded.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			// Compare
			if diff := cmp.Diff(tt.port, decoded); diff != "" {
				t.Fatalf("unexpected bridge port (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBridgePortKind(t *testing.T) {
	bp := &BridgePort{}
	if got, want := bp.Kind(), "bridge"; got != want {
		t.Fatalf("unexpected Kind:\n got: %q\nwant: %q", got, want)
	}
}

func TestBridgePortNew(t *testing.T) {
	bp := &BridgePort{}
	newPort := bp.New()
	if _, ok := newPort.(*BridgePort); !ok {
		t.Fatalf("New() returned wrong type: %T", newPort)
	}
}

func TestBridgePortSlave(t *testing.T) {
	bp := &BridgePort{}
	// Should not panic
	bp.Slave()
}

func TestBridgePortDecodeRaw(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want *BridgePort
	}{
		{
			name: "port state",
			b: []byte{
				0x05, 0x00, // Length: 5
				unix.IFLA_BRPORT_STATE, 0x00, // Type: IFLA_BRPORT_STATE
				0x03,             // Value: 3 (forwarding)
				0x00, 0x00, 0x00, // Padding
			},
			want: &BridgePort{
				State: ptr(BridgePortStateForwarding),
			},
		},
		{
			name: "port cost",
			b: []byte{
				0x08, 0x00, // Length: 8
				unix.IFLA_BRPORT_COST, 0x00, // Type: IFLA_BRPORT_COST
				0x64, 0x00, 0x00, 0x00, // Value: 100
			},
			want: &BridgePort{
				Cost: ptr(uint32(100)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ad, err := netlink.NewAttributeDecoder(tt.b)
			if err != nil {
				t.Fatalf("failed to create decoder: %v", err)
			}

			got := &BridgePort{}
			if err := got.Decode(ad); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("unexpected bridge port (-want +got):\n%s", diff)
			}
		})
	}
}

// ptr is a generic helper function for creating a pointer to an arbitrary type.
func ptr[T any](t T) *T {
	return &t
}
