//go:build integration
// +build integration

package driver

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func bridgeT(d rtnetlink.LinkDriver) *Bridge {
	b := d.(*Bridge)
	return &Bridge{
		StpState:      b.StpState,
		ForwardDelay:  b.ForwardDelay,
		HelloTime:     b.HelloTime,
		MaxAge:        b.MaxAge,
		VlanFiltering: b.VlanFiltering,
		VlanProtocol:  b.VlanProtocol,
		McastSnooping: b.McastSnooping,
		Priority:      b.Priority,
	}
}

func bridgePortT(d rtnetlink.LinkDriver) *BridgePort {
	b := d.(*BridgePort)
	return &BridgePort{
		Priority:     b.Priority,
		Cost:         b.Cost,
		Mode:         b.Mode,
		Learning:     b.Learning,
		UnicastFlood: b.UnicastFlood,
		BcastFlood:   b.BcastFlood,
	}
}

func TestBridge(t *testing.T) {
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

	var (
		stpEnabled   = BridgeStpStateEnabled
		stpDisabled  = BridgeStpStateDisabled
		vlanEnabled  = BridgeEnableEnabled
		vlanDisabled = BridgeEnableDisabled
		mcastEnabled = BridgeEnableEnabled
		proto8021Q   = VlanProtocol8021Q
		u321500      = uint32(1500)
		u32200       = uint32(200)
		u322000      = uint32(2000)
		u1632768     = uint16(32768) // Default bridge priority
	)

	tests := []struct {
		name   string
		conn   *rtnetlink.Conn
		driver *Bridge
		bridge *Bridge
	}{
		{
			name: "bridge with STP enabled",
			conn: conn,
			driver: &Bridge{
				StpState:     &stpEnabled,
				ForwardDelay: &u321500,
				HelloTime:    &u32200,
				MaxAge:       &u322000,
			},
			bridge: &Bridge{
				StpState:      &stpEnabled,
				ForwardDelay:  &u321500,
				HelloTime:     &u32200,
				MaxAge:        &u322000,
				Priority:      &u1632768,
				VlanFiltering: &vlanDisabled,
				VlanProtocol:  &proto8021Q,
				McastSnooping: &mcastEnabled,
			},
		},
		{
			name: "bridge with STP disabled",
			conn: connNS,
			driver: &Bridge{
				StpState: &stpDisabled,
			},
			bridge: &Bridge{
				StpState:      &stpDisabled,
				ForwardDelay:  &u321500,
				HelloTime:     &u32200,
				MaxAge:        &u322000,
				Priority:      &u1632768,
				VlanFiltering: &vlanDisabled,
				VlanProtocol:  &proto8021Q,
				McastSnooping: &mcastEnabled,
			},
		},
		{
			name: "VLAN filtering enabled with 802.1Q",
			conn: connNS,
			driver: &Bridge{
				VlanFiltering: &vlanEnabled,
				VlanProtocol:  &proto8021Q,
			},
			bridge: &Bridge{
				StpState:      &stpDisabled,
				ForwardDelay:  &u321500,
				HelloTime:     &u32200,
				MaxAge:        &u322000,
				Priority:      &u1632768,
				VlanFiltering: &vlanEnabled,
				VlanProtocol:  &proto8021Q,
				McastSnooping: &mcastEnabled,
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridgeID := 1200 + uint32(i*10)
			if err := setupInterface(tt.conn, fmt.Sprintf("br%d", bridgeID), bridgeID, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup bridge interface: %v", err)
			}
			defer tt.conn.Link.Delete(bridgeID)

			msg, err := getInterface(tt.conn, bridgeID)
			if err != nil {
				t.Fatalf("failed to get bridge interface: %v", err)
			}
			if diff := cmp.Diff(tt.bridge, bridgeT(msg.Attributes.Info.Data)); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestBridgePort(t *testing.T) {
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

	var (
		enabled  = BridgeEnableEnabled
		disabled = BridgeEnableDisabled
		u1632    = uint16(32)
		u32100   = uint32(100)
	)

	tests := []struct {
		name     string
		conn     *rtnetlink.Conn
		driver   *Bridge
		portCfg  *BridgePort
		expected *BridgePort
	}{
		{
			name: "port with cost and priority",
			conn: conn,
			driver: &Bridge{
				StpState: ptrBridgeStpState(BridgeStpStateEnabled),
			},
			portCfg: &BridgePort{
				Priority: &u1632,
				Cost:     &u32100,
			},
			expected: &BridgePort{
				Priority:     &u1632,
				Cost:         &u32100,
				Mode:         &disabled,
				Learning:     &enabled,
				UnicastFlood: &enabled,
				BcastFlood:   &enabled,
			},
		},
		{
			name: "port with hairpin mode",
			conn: connNS,
			driver: &Bridge{
				StpState: ptrBridgeStpState(BridgeStpStateEnabled),
			},
			portCfg: &BridgePort{
				Mode: &enabled,
			},
			expected: &BridgePort{
				Mode:         &enabled,
				Cost:         &u32100,
				Priority:     &u1632,
				Learning:     &enabled,
				UnicastFlood: &enabled,
				BcastFlood:   &enabled,
			},
		},
		{
			name: "port with learning and flooding",
			conn: connNS,
			driver: &Bridge{
				StpState: ptrBridgeStpState(BridgeStpStateEnabled),
			},
			portCfg: &BridgePort{
				Learning:     &enabled,
				UnicastFlood: &enabled,
				BcastFlood:   &enabled,
			},
			expected: &BridgePort{
				Learning:     &enabled,
				UnicastFlood: &enabled,
				BcastFlood:   &enabled,
				Cost:         &u32100,
				Priority:     &u1632,
				Mode:         &disabled,
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bridgeID := 1300 + uint32(i*10)
			if err := setupInterface(tt.conn, fmt.Sprintf("br%d", bridgeID), bridgeID, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup bridge interface: %v", err)
			}
			defer tt.conn.Link.Delete(bridgeID)

			// Create dummy interface as bridge port
			dummyID := bridgeID + 1
			if err := setupInterface(tt.conn, fmt.Sprintf("d%d", dummyID), dummyID, bridgeID, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
				t.Fatalf("failed to setup dummy interface: %v", err)
			}
			defer tt.conn.Link.Delete(dummyID)

			// Configure the bridge port
			if tt.portCfg != nil {
				err := tt.conn.Link.Set(&rtnetlink.LinkMessage{
					Index: dummyID,
					Attributes: &rtnetlink.LinkAttributes{
						Info: &rtnetlink.LinkInfo{
							SlaveKind: "bridge",
							SlaveData: tt.portCfg,
						},
					},
				})
				if err != nil {
					t.Fatalf("failed to configure bridge port: %v", err)
				}
			}

			// Verify bridge port configuration
			msg, err := getInterface(tt.conn, dummyID)
			if err != nil {
				t.Fatalf("failed to get dummy interface: %v", err)
			}
			if diff := cmp.Diff(tt.expected, bridgePortT(msg.Attributes.Info.SlaveData)); diff != "" {
				t.Error(diff)
			}
		})
	}
}
