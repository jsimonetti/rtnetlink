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
		attrs.Master = &master
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
