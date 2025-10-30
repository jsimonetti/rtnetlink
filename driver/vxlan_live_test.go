//go:build integration
// +build integration

package driver

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func vxlanT(d rtnetlink.LinkDriver) *Vxlan {
	v := d.(*Vxlan)
	return &Vxlan{
		ID:       v.ID,
		Group:    v.Group,
		Group6:   v.Group6,
		Local:    v.Local,
		Local6:   v.Local6,
		Link:     v.Link,
		Port:     v.Port,
		Learning: v.Learning,
		Ageing:   v.Ageing,
		TTL:      v.TTL,
		TOS:      v.TOS,
	}
}

func TestVxlan(t *testing.T) {
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

	// Create a dummy interface to use as the link device for multicast VXLAN
	dummyIndex := uint32(1950)
	if err := setupInterface(connNS, "vxdummy", dummyIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create dummy interface: %v", err)
	}
	defer connNS.Link.Delete(dummyIndex)

	var (
		vni100      uint32 = 100
		vni200      uint32 = 200
		port8472    uint16 = 8472 // Kernel default VXLAN port
		ttl0        uint8  = 0
		ttl64       uint8  = 64
		tos0        uint8  = 0
		ageing300   uint32 = 300
		learningOn         = true
		learningOff        = false
	)

	tests := []struct {
		name     string
		linkName string
		conn     *rtnetlink.Conn
		driver   *Vxlan
		expected *Vxlan
	}{
		{
			name:     "basic vxlan with VNI only",
			linkName: "vxlan0",
			conn:     conn,
			driver: &Vxlan{
				ID: &vni100,
			},
			expected: &Vxlan{
				ID:       &vni100,
				TTL:      &ttl0,     // Kernel sets default TTL to 0
				TOS:      &tos0,     // Kernel sets default TOS to 0
				Port:     &port8472, // Kernel sets default port (IANA assigned)
				Learning: &learningOn,
				Ageing:   &ageing300,
			},
		},
		{
			name:     "vxlan with unicast",
			linkName: "vxlan1",
			conn:     connNS,
			driver: &Vxlan{
				ID:    &vni200,
				Group: net.ParseIP("239.1.1.1"),
				Local: net.ParseIP("192.168.1.1"),
				Link:  &dummyIndex,
			},
			expected: &Vxlan{
				ID:       &vni200,
				Group:    net.ParseIP("239.1.1.1"),
				Local:    net.ParseIP("192.168.1.1"),
				Link:     &dummyIndex,
				TTL:      &ttl0,     // Kernel sets default TTL to 0
				TOS:      &tos0,     // Kernel sets default TOS to 0
				Port:     &port8472, // Kernel sets default port (IANA assigned)
				Learning: &learningOn,
				Ageing:   &ageing300,
			},
		},
		{
			name:     "vxlan with custom port and settings",
			linkName: "vxlan2",
			conn:     connNS,
			driver: &Vxlan{
				ID:       &vni100,
				Port:     &port8472,
				TTL:      &ttl64,
				TOS:      &tos0,
				Learning: &learningOff,
			},
			expected: &Vxlan{
				ID:       &vni100,
				Port:     &port8472,
				TTL:      &ttl64,
				TOS:      &tos0,
				Learning: &learningOff,
				Ageing:   &ageing300,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a unique index for each test
			ifIndex := uint32(2000 + len(tt.name))

			if err := setupInterface(tt.conn, tt.linkName, ifIndex, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup vxlan interface: %v", err)
			}
			defer tt.conn.Link.Delete(ifIndex)

			msg, err := getInterface(tt.conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get vxlan interface: %v", err)
			}

			if msg.Attributes == nil || msg.Attributes.Info == nil || msg.Attributes.Info.Data == nil {
				t.Fatal("interface missing link info data")
			}

			vxlan, ok := msg.Attributes.Info.Data.(*Vxlan)
			if !ok {
				t.Fatalf("expected *Vxlan, got %T", msg.Attributes.Info.Data)
			}

			if diff := cmp.Diff(tt.expected, vxlanT(vxlan)); diff != "" {
				t.Fatalf("unexpected vxlan config (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVxlanIPv6(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	// Create a dummy interface to use as the link device for IPv6 multicast VXLAN
	dummyIndex := uint32(2050)
	if err := setupInterface(conn, "vx6dummy", dummyIndex, 0, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
		t.Fatalf("failed to create dummy interface: %v", err)
	}
	defer conn.Link.Delete(dummyIndex)

	var (
		vni100 uint32 = 100
		ttl64  uint8  = 64
	)

	tests := []struct {
		name     string
		linkName string
		driver   *Vxlan
	}{
		{
			name:     "vxlan with IPv6 multicast group",
			linkName: "vxlan6",
			driver: &Vxlan{
				ID:     &vni100,
				Group6: net.ParseIP("ff05::100"),
				Link:   &dummyIndex,
				TTL:    &ttl64,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifIndex := uint32(2100 + len(tt.name))

			if err := setupInterface(conn, tt.linkName, ifIndex, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup vxlan interface: %v", err)
			}
			defer conn.Link.Delete(ifIndex)

			msg, err := getInterface(conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get vxlan interface: %v", err)
			}

			if msg.Attributes == nil || msg.Attributes.Info == nil || msg.Attributes.Info.Data == nil {
				t.Fatal("interface missing link info data")
			}

			vxlan, ok := msg.Attributes.Info.Data.(*Vxlan)
			if !ok {
				t.Fatalf("expected *Vxlan, got %T", msg.Attributes.Info.Data)
			}

			if vxlan.ID == nil || *vxlan.ID != *tt.driver.ID {
				t.Errorf("expected VNI %d, got %v", *tt.driver.ID, vxlan.ID)
			}

			if vxlan.Group6 == nil || !vxlan.Group6.Equal(tt.driver.Group6) {
				t.Errorf("expected Group6 %v, got %v", tt.driver.Group6, vxlan.Group6)
			}
		})
	}
}

func TestVxlanAdvancedFeatures(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	var (
		vni100     uint32 = 100
		ageing120  uint32 = 120
		limit1000  uint32 = 1000
		proxyOn           = true
		l2missOn          = true
		l3missOn          = true
		udpCsumOn         = true
		learningOn        = true
	)

	tests := []struct {
		name     string
		linkName string
		driver   *Vxlan
	}{
		{
			name:     "vxlan with advanced features",
			linkName: "vxlan-adv",
			driver: &Vxlan{
				ID:       &vni100,
				Ageing:   &ageing120,
				Limit:    &limit1000,
				Proxy:    &proxyOn,
				L2Miss:   &l2missOn,
				L3Miss:   &l3missOn,
				UDPCsum:  &udpCsumOn,
				Learning: &learningOn,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifIndex := uint32(2200 + len(tt.name))

			if err := setupInterface(conn, tt.linkName, ifIndex, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup vxlan interface: %v", err)
			}
			defer conn.Link.Delete(ifIndex)

			msg, err := getInterface(conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get vxlan interface: %v", err)
			}

			if msg.Attributes == nil || msg.Attributes.Info == nil || msg.Attributes.Info.Data == nil {
				t.Fatal("interface missing link info data")
			}

			vxlan, ok := msg.Attributes.Info.Data.(*Vxlan)
			if !ok {
				t.Fatalf("expected *Vxlan, got %T", msg.Attributes.Info.Data)
			}

			if vxlan.Ageing == nil || *vxlan.Ageing != *tt.driver.Ageing {
				t.Errorf("expected Ageing %d, got %v", *tt.driver.Ageing, vxlan.Ageing)
			}

			if vxlan.Limit == nil || *vxlan.Limit != *tt.driver.Limit {
				t.Errorf("expected Limit %d, got %v", *tt.driver.Limit, vxlan.Limit)
			}

			if vxlan.Proxy == nil || *vxlan.Proxy != *tt.driver.Proxy {
				t.Errorf("expected Proxy %v, got %v", *tt.driver.Proxy, vxlan.Proxy)
			}

			if vxlan.L2Miss == nil || *vxlan.L2Miss != *tt.driver.L2Miss {
				t.Errorf("expected L2Miss %v, got %v", *tt.driver.L2Miss, vxlan.L2Miss)
			}

			if vxlan.L3Miss == nil || *vxlan.L3Miss != *tt.driver.L3Miss {
				t.Errorf("expected L3Miss %v, got %v", *tt.driver.L3Miss, vxlan.L3Miss)
			}
		})
	}
}

func TestVxlanPortRange(t *testing.T) {
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	var vni100 uint32 = 100

	tests := []struct {
		name      string
		linkName  string
		portRange *VxlanPortRange
	}{
		{
			name:     "vxlan with custom port range",
			linkName: "vxlan-port",
			portRange: &VxlanPortRange{
				Low:  10000,
				High: 20000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ifIndex := uint32(2300 + len(tt.name))

			driver := &Vxlan{
				ID:        &vni100,
				PortRange: tt.portRange,
			}

			if err := setupInterface(conn, tt.linkName, ifIndex, 0, driver); err != nil {
				t.Fatalf("failed to setup vxlan interface: %v", err)
			}
			defer conn.Link.Delete(ifIndex)

			msg, err := getInterface(conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get vxlan interface: %v", err)
			}

			if msg.Attributes == nil || msg.Attributes.Info == nil || msg.Attributes.Info.Data == nil {
				t.Fatal("interface missing link info data")
			}

			vxlan, ok := msg.Attributes.Info.Data.(*Vxlan)
			if !ok {
				t.Fatalf("expected *Vxlan, got %T", msg.Attributes.Info.Data)
			}

			if vxlan.PortRange == nil {
				t.Fatal("expected PortRange to be set")
			}

			if vxlan.PortRange.Low != tt.portRange.Low {
				t.Errorf("expected PortRange.Low %d, got %d", tt.portRange.Low, vxlan.PortRange.Low)
			}

			if vxlan.PortRange.High != tt.portRange.High {
				t.Errorf("expected PortRange.High %d, got %d", tt.portRange.High, vxlan.PortRange.High)
			}
		})
	}
}
