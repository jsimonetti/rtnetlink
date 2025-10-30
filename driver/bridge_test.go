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
				StpState:     ptrBridgeStpState(BridgeStpStateEnabled),
				Priority:     ptrUint16(32768),
				ForwardDelay: ptrUint32(1500),
				HelloTime:    ptrUint32(200),
				MaxAge:       ptrUint32(2000),
			},
		},
		{
			name: "with VLAN filtering",
			bridge: &Bridge{
				VlanFiltering:   ptrBridgeEnable(BridgeEnableEnabled),
				VlanProtocol:    ptrVlanProtocol(VlanProtocol8021Q),
				VlanDefaultPvid: ptrUint16(1),
			},
		},
		{
			name: "with multicast settings",
			bridge: &Bridge{
				McastSnooping:       ptrBridgeEnable(BridgeEnableEnabled),
				McastQuerier:        ptrBridgeEnable(BridgeEnableEnabled),
				McastHashElasticity: ptrUint32(4),
				McastHashMax:        ptrUint32(512),
				McastIgmpVersion:    ptrUint8(3),
				McastMldVersion:     ptrUint8(2),
			},
		},
		{
			name: "with netfilter settings",
			bridge: &Bridge{
				NfCallIptables:  ptrBridgeEnable(BridgeEnableEnabled),
				NfCallIp6tables: ptrBridgeEnable(BridgeEnableEnabled),
				NfCallArptables: ptrBridgeEnable(BridgeEnableEnabled),
			},
		},
		{
			name: "full configuration",
			bridge: &Bridge{
				ForwardDelay:            ptrUint32(1500),
				HelloTime:               ptrUint32(200),
				MaxAge:                  ptrUint32(2000),
				AgeingTime:              ptrUint32(30000),
				StpState:                ptrBridgeStpState(BridgeStpStateEnabled),
				Priority:                ptrUint16(32768),
				VlanFiltering:           ptrBridgeEnable(BridgeEnableEnabled),
				VlanProtocol:            ptrVlanProtocol(VlanProtocol8021Q),
				GroupFwdMask:            ptrUint16(0),
				GroupAddr:               net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x00},
				McastRouter:             ptrUint8(1),
				McastSnooping:           ptrBridgeEnable(BridgeEnableEnabled),
				McastQueryUseIfaddr:     ptrBridgeEnable(BridgeEnableDisabled),
				McastQuerier:            ptrBridgeEnable(BridgeEnableEnabled),
				McastHashElasticity:     ptrUint32(4),
				McastHashMax:            ptrUint32(512),
				McastLastMemberCnt:      ptrUint32(2),
				McastStartupQueryCnt:    ptrUint32(2),
				McastLastMemberIntvl:    ptrUint64(100),
				McastMembershipIntvl:    ptrUint64(26000),
				McastQuerierIntvl:       ptrUint64(25500),
				McastQueryIntvl:         ptrUint64(12500),
				McastQueryResponseIntvl: ptrUint64(1000),
				McastStartupQueryIntvl:  ptrUint64(3125),
				NfCallIptables:          ptrBridgeEnable(BridgeEnableDisabled),
				NfCallIp6tables:         ptrBridgeEnable(BridgeEnableDisabled),
				NfCallArptables:         ptrBridgeEnable(BridgeEnableDisabled),
				VlanDefaultPvid:         ptrUint16(1),
				VlanStatsEnabled:        ptrBridgeEnable(BridgeEnableDisabled),
				McastStatsEnabled:       ptrBridgeEnable(BridgeEnableDisabled),
				McastIgmpVersion:        ptrUint8(2),
				McastMldVersion:         ptrUint8(1),
				VlanStatsPerPort:        ptrBridgeEnable(BridgeEnableDisabled),
				FdbMaxLearned:           ptrUint32(0),
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
				StpState: ptrBridgeStpState(BridgeStpStateEnabled),
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
				VlanFiltering: ptrBridgeEnable(BridgeEnableEnabled),
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
				State:    ptrBridgePortState(BridgePortStateForwarding),
				Priority: ptrUint16(32),
				Cost:     ptrUint32(100),
			},
		},
		{
			name: "with hairpin and guard",
			port: &BridgePort{
				Mode:  ptrBridgeEnable(BridgeEnableEnabled), // Hairpin mode
				Guard: ptrBridgeEnable(BridgeEnableEnabled), // BPDU guard
			},
		},
		{
			name: "with flood control",
			port: &BridgePort{
				Learning:     ptrBridgeEnable(BridgeEnableEnabled),
				UnicastFlood: ptrBridgeEnable(BridgeEnableEnabled),
				McastFlood:   ptrBridgeEnable(BridgeEnableEnabled),
				BcastFlood:   ptrBridgeEnable(BridgeEnableEnabled),
			},
		},
		{
			name: "with proxy ARP and isolation",
			port: &BridgePort{
				ProxyArp:     ptrBridgeEnable(BridgeEnableEnabled),
				ProxyArpWifi: ptrBridgeEnable(BridgeEnableEnabled),
				Isolated:     ptrBridgeEnable(BridgeEnableEnabled),
			},
		},
		{
			name: "full configuration",
			port: &BridgePort{
				State:             ptrBridgePortState(BridgePortStateForwarding),
				Priority:          ptrUint16(32),
				Cost:              ptrUint32(100),
				Mode:              ptrBridgeEnable(BridgeEnableDisabled),
				Guard:             ptrBridgeEnable(BridgeEnableDisabled),
				Protect:           ptrBridgeEnable(BridgeEnableDisabled),
				FastLeave:         ptrBridgeEnable(BridgeEnableDisabled),
				Learning:          ptrBridgeEnable(BridgeEnableEnabled),
				UnicastFlood:      ptrBridgeEnable(BridgeEnableEnabled),
				ProxyArp:          ptrBridgeEnable(BridgeEnableDisabled),
				LearningSync:      ptrBridgeEnable(BridgeEnableDisabled),
				ProxyArpWifi:      ptrBridgeEnable(BridgeEnableDisabled),
				MulticastRouter:   ptrUint8(1),
				McastFlood:        ptrBridgeEnable(BridgeEnableEnabled),
				McastToUcast:      ptrBridgeEnable(BridgeEnableDisabled),
				VlanTunnel:        ptrBridgeEnable(BridgeEnableDisabled),
				BcastFlood:        ptrBridgeEnable(BridgeEnableEnabled),
				GroupFwdMask:      ptrUint16(0),
				NeighSuppress:     ptrBridgeEnable(BridgeEnableDisabled),
				Isolated:          ptrBridgeEnable(BridgeEnableDisabled),
				Locked:            ptrBridgeEnable(BridgeEnableDisabled),
				Mab:               ptrBridgeEnable(BridgeEnableDisabled),
				NeighVlanSuppress: ptrBridgeEnable(BridgeEnableDisabled),
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
				State: ptrBridgePortState(BridgePortStateForwarding),
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
				Cost: ptrUint32(100),
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

// Helper functions for creating pointers
func ptrUint8(v uint8) *uint8 {
	return &v
}

func ptrUint16(v uint16) *uint16 {
	return &v
}

func ptrUint32(v uint32) *uint32 {
	return &v
}

func ptrUint64(v uint64) *uint64 {
	return &v
}

func ptrBridgeStpState(s BridgeStpState) *BridgeStpState {
	return &s
}

func ptrBridgeEnable(e BridgeEnable) *BridgeEnable {
	return &e
}

func ptrVlanProtocol(p VlanProtocol) *VlanProtocol {
	return &p
}

func ptrBridgePortState(s BridgePortState) *BridgePortState {
	return &s
}
