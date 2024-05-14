//go:build integration
// +build integration

package driver

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2"
	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/mdlayher/netlink"
)

func TestVeth(t *testing.T) {
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
		ifIndex     = 1021
		ifPeerIndex = 1022
	)

	tests := []struct {
		name     string
		linkName string
		pconn    *rtnetlink.Conn
		driver   *Veth
	}{
		{
			name:  "with empty names both in default ns",
			pconn: conn,
			driver: &Veth{
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
				},
			},
		},
		{
			name:     "with names both in default ns",
			linkName: "vtp",
			pconn:    conn,
			driver: &Veth{
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
					Attributes: &rtnetlink.LinkAttributes{
						Name: "vte",
					},
				},
			},
		},
		{
			name:     "with names one in other ns",
			linkName: "vtp",
			pconn:    connNS,
			driver: &Veth{
				PeerInfo: &rtnetlink.LinkMessage{
					Index: ifPeerIndex,
					Attributes: &rtnetlink.LinkAttributes{
						Name:  "vte",
						NetNS: rtnetlink.NetNSForFD(uint32(ns)),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setupInterface(conn, tt.linkName, ifIndex, 0, tt.driver); err != nil {
				t.Fatalf("failed to setup veth interface: %v", err)
			}
			defer conn.Link.Delete(ifIndex)

			_, err = getInterface(conn, ifIndex)
			if err != nil {
				t.Fatalf("failed to get primary veth interface: %v", err)
			}

			_, err = getInterface(tt.pconn, ifPeerIndex)
			if err != nil {
				t.Fatalf("failed to get peer veth interface: %v", err)
			}
		})
	}
}
