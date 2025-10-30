//go:build integration
// +build integration

package driver

import (
	"github.com/jsimonetti/rtnetlink/v2"
	"golang.org/x/sys/unix"
)

// setupInterface create a interface for testing
func setupInterface(conn *rtnetlink.Conn, name string, index, master uint32, driver rtnetlink.LinkDriver) error {
	attrs := &rtnetlink.LinkAttributes{
		Name: name,
		Info: &rtnetlink.LinkInfo{Kind: driver.Kind(), Data: driver},
	}
	flag := uint32(unix.IFF_UP)
	if master > 0 {
		// Check if this is a VLAN, VXLAN, or MACVLAN interface
		// These types need the parent interface specified via Type/IFLA_LINK
		kind := driver.Kind()
		if kind == "vlan" || kind == "vxlan" || kind == "macvlan" {
			// For VLAN/VXLAN/MACVLAN, the master parameter is actually the parent link index
			attrs.Type = master
		} else {
			// For other types (like dummy being added to bridge), master is for enslaving
			attrs.Master = &master
		}
		flag = 0
	}
	// construct an interface to test drivers
	err := conn.Link.New(&rtnetlink.LinkMessage{
		Family:     unix.AF_UNSPEC,
		Index:      index,
		Flags:      flag,
		Change:     flag,
		Attributes: attrs,
	})
	if err != nil {
		conn.Link.Delete(index)
	}
	return err
}

func getInterface(conn *rtnetlink.Conn, index uint32) (*rtnetlink.LinkMessage, error) {
	interf, err := conn.Link.Get(index)
	if err != nil {
		conn.Link.Delete(interf.Index)
		return nil, err
	}
	return &interf, err
}
