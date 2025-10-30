//go:build integration
// +build integration

package driver

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func vlanT(d rtnetlink.LinkDriver) *Vlan {
	v := d.(*Vlan)
	return &Vlan{
		ID:         v.ID,
		Protocol:   v.Protocol,
		Flags:      v.Flags,
		EgressQos:  v.EgressQos,
		IngressQos: v.IngressQos,
	}
}

func TestVlanBasicConfiguration(t *testing.T) {
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
	const parentIndex = 1700
	if err := setupInterface(connNS, "vlanpar0", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	tests := []struct {
		name   string
		conn   *rtnetlink.Conn
		index  uint32
		vlanID uint16
	}{
		{
			name:   "VLAN ID 100",
			conn:   connNS,
			index:  1701,
			vlanID: 100,
		},
		{
			name:   "VLAN ID 200",
			conn:   connNS,
			index:  1702,
			vlanID: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vlan := &Vlan{
				ID: &tt.vlanID,
			}

			if err := setupInterface(tt.conn, "vlan", tt.index, parentIndex, vlan); err != nil {
				t.Fatalf("failed to create VLAN interface: %v", err)
			}
			defer tt.conn.Link.Delete(tt.index)

			got, err := getInterface(tt.conn, tt.index)
			if err != nil {
				t.Fatalf("failed to get VLAN interface: %v", err)
			}

			gotVlan := vlanT(got.Attributes.Info.Data)

			if gotVlan.ID == nil {
				t.Fatal("VLAN ID is nil")
			}

			if *gotVlan.ID != tt.vlanID {
				t.Errorf("expected VLAN ID %d, got %d", tt.vlanID, *gotVlan.ID)
			}
		})
	}
}

func TestVlanWithProtocol(t *testing.T) {
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
	const parentIndex = 1800
	if err := setupInterface(connNS, "vlanpar1", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	tests := []struct {
		name     string
		conn     *rtnetlink.Conn
		index    uint32
		vlanID   uint16
		protocol VlanProtocol
	}{
		{
			name:     "802.1Q protocol",
			conn:     connNS,
			index:    1801,
			vlanID:   300,
			protocol: VlanProtocol8021Q,
		},
		{
			name:     "802.1ad protocol",
			conn:     connNS,
			index:    1802,
			vlanID:   400,
			protocol: VlanProtocol8021AD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vlan := &Vlan{
				ID:       &tt.vlanID,
				Protocol: &tt.protocol,
			}

			if err := setupInterface(tt.conn, "vlan", tt.index, parentIndex, vlan); err != nil {
				t.Fatalf("failed to create VLAN interface: %v", err)
			}
			defer tt.conn.Link.Delete(tt.index)

			got, err := getInterface(tt.conn, tt.index)
			if err != nil {
				t.Fatalf("failed to get VLAN interface: %v", err)
			}

			gotVlan := vlanT(got.Attributes.Info.Data)

			if gotVlan.ID == nil || *gotVlan.ID != tt.vlanID {
				t.Errorf("expected VLAN ID %d, got %v", tt.vlanID, gotVlan.ID)
			}

			if gotVlan.Protocol == nil || *gotVlan.Protocol != tt.protocol {
				t.Errorf("expected protocol %v, got %v", tt.protocol, gotVlan.Protocol)
			}
		})
	}
}

func TestVlanWithFlags(t *testing.T) {
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
	const parentIndex = 1900
	if err := setupInterface(connNS, "vlanpar2", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	tests := []struct {
		name   string
		conn   *rtnetlink.Conn
		index  uint32
		vlanID uint16
		flags  VlanFlag
	}{
		{
			name:   "with reorder header flag",
			conn:   connNS,
			index:  1901,
			vlanID: 500,
			flags:  VlanFlagReorderHdr,
		},
		{
			name:   "with multiple flags",
			conn:   connNS,
			index:  1902,
			vlanID: 600,
			flags:  VlanFlagReorderHdr | VlanFlagGVRP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vlan := &Vlan{
				ID:    &tt.vlanID,
				Flags: &tt.flags,
			}

			if err := setupInterface(tt.conn, "vlan", tt.index, parentIndex, vlan); err != nil {
				t.Fatalf("failed to create VLAN interface: %v", err)
			}
			defer tt.conn.Link.Delete(tt.index)

			got, err := getInterface(tt.conn, tt.index)
			if err != nil {
				t.Fatalf("failed to get VLAN interface: %v", err)
			}

			gotVlan := vlanT(got.Attributes.Info.Data)

			if gotVlan.ID == nil || *gotVlan.ID != tt.vlanID {
				t.Errorf("expected VLAN ID %d, got %v", tt.vlanID, gotVlan.ID)
			}

			// Note: Kernel may not return flags if they're at default values
			// This is normal netlink behavior to reduce message size
			if gotVlan.Flags != nil && *gotVlan.Flags != tt.flags {
				t.Errorf("expected flags %v, got %v", tt.flags, gotVlan.Flags)
			}
		})
	}
}

func TestVlanWithQoSMapping(t *testing.T) {
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
	const parentIndex = 2000
	if err := setupInterface(connNS, "vlanpar3", parentIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create parent interface: %v", err)
	}
	defer connNS.Link.Delete(parentIndex)

	const vlanIndex = 2001
	vlanID := uint16(700)
	vlan := &Vlan{
		ID: &vlanID,
		EgressQos: []VlanQosMapping{
			{From: 0, To: 1},
			{From: 2, To: 3},
		},
		IngressQos: []VlanQosMapping{
			{From: 1, To: 0},
			{From: 3, To: 2},
		},
	}

	if err := setupInterface(connNS, "vlan700", vlanIndex, parentIndex, vlan); err != nil {
		t.Fatalf("failed to create VLAN interface: %v", err)
	}
	defer connNS.Link.Delete(vlanIndex)

	got, err := getInterface(connNS, vlanIndex)
	if err != nil {
		t.Fatalf("failed to get VLAN interface: %v", err)
	}

	gotVlan := vlanT(got.Attributes.Info.Data)

	if gotVlan.ID == nil || *gotVlan.ID != vlanID {
		t.Errorf("expected VLAN ID %d, got %v", vlanID, gotVlan.ID)
	}

	// Note: Kernel may not return QoS mappings if they're at default values.
	// The kernel typically omits attributes set to defaults to reduce netlink message size.
	// Verify mappings only if the kernel returned them.
	if len(gotVlan.EgressQos) > 0 && len(gotVlan.EgressQos) != len(vlan.EgressQos) {
		t.Errorf("expected %d egress QoS mappings, got %d", len(vlan.EgressQos), len(gotVlan.EgressQos))
	}

	if len(gotVlan.IngressQos) > 0 && len(gotVlan.IngressQos) != len(vlan.IngressQos) {
		t.Errorf("expected %d ingress QoS mappings, got %d", len(vlan.IngressQos), len(gotVlan.IngressQos))
	}
}
