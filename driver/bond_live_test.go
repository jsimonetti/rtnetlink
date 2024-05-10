//go:build integration
// +build integration

package driver

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/mdlayher/netlink"
)

func bondT(d rtnetlink.LinkDriver) *Bond {
	b := d.(*Bond)
	return &Bond{
		Mode:         b.Mode,
		Miimon:       b.Miimon,
		ArpInterval:  b.ArpInterval,
		ArpIpTargets: b.ArpIpTargets,
		NsIP6Targets: b.NsIP6Targets,
	}
}

func bondSlaveT(d rtnetlink.LinkDriver) *BondSlave {
	b := d.(*BondSlave)
	return &BondSlave{
		State:     b.State,
		MiiStatus: b.MiiStatus,
		Priority:  b.Priority,
	}
}

func TestBond(t *testing.T) {
	// establish a netlink connection
	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	bns, clean, err := createNS("bns1")
	if err != nil {
		t.Fatal(err)
	}
	defer clean()

	// use ns for testing arp ip targets
	connNS, err := rtnetlink.Dial(&netlink.Config{NetNS: int(bns.Value())})
	if err != nil {
		t.Fatalf("failed to establish netlink socket to ns nkns: %v", err)
	}
	defer connNS.Close()

	var (
		ssa           = BondStateActive
		ssb           = BondStateBackup
		miiup         = BondLinkUp
		u325   uint32 = 5
		u32100 uint32 = 100
		u32    uint32
		i321   int32 = 1
		i32    int32
	)

	tests := []struct {
		name     string
		conn     *rtnetlink.Conn
		driver   *Bond
		bond     *Bond
		setSlave bool
		dummy    []BondSlave
	}{
		{
			name: "with default mode and miion is set",
			conn: conn,
			driver: &Bond{
				Miimon: &u32100,
			},
			bond: &Bond{
				Mode:        BondModeBalanceRR,
				Miimon:      &u32100,
				ArpInterval: &u32,
			},
			dummy: []BondSlave{
				{
					State:     &ssa,
					MiiStatus: &miiup,
					Priority:  &i32,
				},
				{
					State:     &ssa,
					MiiStatus: &miiup,
					Priority:  &i32,
				},
			},
		},
		{
			name: "with active backup, and arp ip targets list",
			conn: connNS,
			driver: &Bond{
				Mode:         BondModeActiveBackup,
				ArpInterval:  &u325,
				ArpIpTargets: []net.IP{{192, 168, 222, 2}, {192, 168, 222, 3}},
			},
			bond: &Bond{
				Mode:         BondModeActiveBackup,
				Miimon:       &u32,
				ArpInterval:  &u325,
				ArpIpTargets: []net.IP{{192, 168, 222, 2}, {192, 168, 222, 3}},
			},
			setSlave: true,
			dummy: []BondSlave{
				{
					State:     &ssb,
					MiiStatus: &miiup,
					Priority:  &i32,
				},
				{
					State:     &ssa,
					MiiStatus: &miiup,
					Priority:  &i321,
				},
			},
		},
		{
			name: "with balanced xor, and arp ns ipv6 list",
			conn: connNS,
			driver: &Bond{
				Mode:        BondModeBalanceXOR,
				ArpInterval: &u325,
				NsIP6Targets: []net.IP{
					{0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
					{0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				},
			},
			bond: &Bond{
				Mode:        BondModeBalanceXOR,
				Miimon:      &u32,
				ArpInterval: &u325,
				NsIP6Targets: []net.IP{
					{0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
					{0xfd, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
				},
			},
			dummy: []BondSlave{
				{
					State:     &ssa,
					MiiStatus: &miiup,
					Priority:  &i32,
				},
				{
					State:     &ssa,
					MiiStatus: &miiup,
					Priority:  &i32,
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bondID := 1100 + uint32(i*10)
			if err := setupInterface(tt.conn, fmt.Sprintf("b%d", bondID), bondID, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup bond interface: %v", err)
			}
			defer tt.conn.Link.Delete(bondID)

			msg, err := getInterface(tt.conn, bondID)
			if err != nil {
				t.Fatalf("failed to get primary netkit interface: %v", err)
			}
			if diff := cmp.Diff(tt.bond, bondT(msg.Attributes.Info.Data)); diff != "" {
				t.Error(diff)
			}

			slave1ID := 1101 + uint32(i*10)
			if err := setupInterface(tt.conn, fmt.Sprintf("d%d", slave1ID), slave1ID, bondID, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
				t.Fatalf("failed to setup d%d interface: %v", slave1ID, err)
			}
			defer tt.conn.Link.Delete(slave1ID)

			slave2ID := 1102 + uint32(i*10)
			if err := setupInterface(tt.conn, fmt.Sprintf("d%d", slave2ID), slave2ID, bondID, &rtnetlink.LinkData{Name: "dummy"}); err != nil {
				t.Fatalf("failed to setup d1%d interface: %v", slave2ID, err)
			}
			defer tt.conn.Link.Delete(slave2ID)

			if tt.setSlave {
				tt.conn.Link.Set(&rtnetlink.LinkMessage{
					Index: slave2ID,
					Attributes: &rtnetlink.LinkAttributes{
						Info: &rtnetlink.LinkInfo{
							SlaveKind: "bond",
							SlaveData: &BondSlave{
								Priority: &i321,
							},
						},
					},
				})
			}

			for i, id := range []uint32{slave1ID, slave2ID} {
				msg, err = getInterface(tt.conn, id)
				if err != nil {
					t.Fatalf("failed to get peer netkit interface: %v", err)
				}
				if diff := cmp.Diff(&tt.dummy[i], bondSlaveT(msg.Attributes.Info.SlaveData)); diff != "" {
					t.Errorf("slave %d %s", i, diff)
				}
			}
		})
	}
}
