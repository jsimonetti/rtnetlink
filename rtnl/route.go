package rtnl

import (
	"net"

	"golang.org/x/sys/unix"

	"github.com/jsimonetti/rtnetlink"
)

// RouteAdd add infomation about a network route.
func (c *Conn) RouteAdd(ifc *net.Interface, dst net.IPNet, gw net.IP) (err error) {
	af, err := addrFamily(dst.IP)
	if err != nil {
		return err
	}
	prefixlen, _ := dst.Mask.Size()
	scope := addrScope(dst.IP)
	if len(dst.IP) == net.IPv6len && dst.IP.To4() == nil {
		scope = unix.RT_SCOPE_UNIVERSE
	}
	attr := rtnetlink.RouteAttributes{
		Dst:      dst.IP,
		OutIface: uint32(ifc.Index),
	}
	if gw != nil {
		attr.Gateway = gw
	}
	tx := &rtnetlink.RouteMessage{
		Family:     uint8(af),
		Table:      unix.RT_TABLE_MAIN,
		Protocol:   unix.RTPROT_BOOT,
		Type:       unix.RTN_UNICAST,
		Scope:      uint8(scope),
		DstLength:  uint8(prefixlen),
		Attributes: attr,
	}
	return c.Conn.Route.Add(tx)
}
