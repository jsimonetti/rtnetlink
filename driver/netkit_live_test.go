//go:build integration
// +build integration

package driver

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func TestNetkit(t *testing.T) {
	kernelMinReq(t, 6, 7)

	conn, err := rtnetlink.Dial(nil)
	if err != nil {
		t.Fatalf("failed to establish netlink socket: %v", err)
	}
	defer conn.Close()

	ns := testutils.NetNS(t)
	connNS, err := rtnetlink.Dial(&netlink.Config{NetNS: ns})
	if err != nil {
		t.Fatalf("failed to establish netlink socket to netns: %v", err)
	}
	defer connNS.Close()

	const (
		ifIndex     = 1011
		ifPeerIndex = 1012
	)

	modeL2 := NetkitModeL2
	modeL3 := NetkitModeL3
	polPass := NetkitPolicyPass
	polDrop := NetkitPolicyDrop

	tests := []struct {
		name     string
		linkName string
		pconn    *rtnetlink.Conn
		driver   *Netkit
		primary  *Netkit
		peer     *Netkit
	}{
		{
			name:     "with empty link names both in default ns",
			linkName: "",
			pconn:    conn,
			driver: &Netkit{
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
				},
			},
			primary: &Netkit{
				Mode:       &modeL3,
				Policy:     &polPass,
				PeerPolicy: &polPass,
				Primary:    true,
			},
			peer: &Netkit{
				Mode:       &modeL3,
				Policy:     &polPass,
				PeerPolicy: &polPass,
			},
		},
		{
			name:     "with names both in default ns",
			linkName: "nkp",
			pconn:    conn,
			driver: &Netkit{
				Mode: &modeL2,
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
					Attributes: &rtnetlink.LinkAttributes{
						Name: "nke",
					},
				},
			},
			primary: &Netkit{
				Mode:       &modeL2,
				Policy:     &polPass,
				PeerPolicy: &polPass,
				Primary:    true,
			},
			peer: &Netkit{
				Mode:       &modeL2,
				Policy:     &polPass,
				PeerPolicy: &polPass,
			},
		},
		{
			name:     "with one is in other ns",
			linkName: "nkp",
			pconn:    connNS,
			driver: &Netkit{
				Policy:     &polPass,
				PeerPolicy: &polDrop,
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
					Attributes: &rtnetlink.LinkAttributes{
						Name:  "nke",
						NetNS: rtnetlink.NetNSForFD(uint32(ns)),
					},
				},
			},
			primary: &Netkit{
				Mode:       &modeL3,
				Policy:     &polPass,
				PeerPolicy: &polDrop,
				Primary:    true,
			},
			peer: &Netkit{
				Mode:       &modeL3,
				Policy:     &polDrop,
				PeerPolicy: &polPass,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setupInterface(conn, tt.linkName, ifIndex, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup netkit interface: %v", err)
			}
			defer conn.Link.Delete(ifIndex)

			msg, err := getInterface(conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get primary netkit interface: %v", err)
			}
			if diff := cmp.Diff(tt.primary, msg.Attributes.Info.Data); diff != "" {
				t.Error(diff)
			}

			msg, err = getInterface(tt.pconn, ifPeerIndex)
			if err != nil {
				t.Fatalf("failed to get peer netkit interface: %v", err)
			}
			if diff := cmp.Diff(tt.peer, msg.Attributes.Info.Data); diff != "" {
				t.Error(diff)
			}
		})
	}
}
