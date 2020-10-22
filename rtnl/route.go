package rtnl

import (
	"net"

	"github.com/jsimonetti/rtnetlink/internal/unix"

	"github.com/jsimonetti/rtnetlink"
)

// generating route message
func genRouteMessage(ifc *net.Interface, dst net.IPNet, gw net.IP, options ...RouteOption) (rm *rtnetlink.RouteMessage, err error) {

	opts := DefaultRouteOptions(ifc, dst, gw)

	for _, option := range options {
		option(opts)
	}

	af, err := addrFamily(dst.IP)
	if err != nil {
		return nil, err
	}

	// Determine scope
	var scope uint8
	switch {
	case gw != nil:
		scope = unix.RT_SCOPE_UNIVERSE
	case len(dst.IP) == net.IPv6len && dst.IP.To4() == nil:
		scope = unix.RT_SCOPE_UNIVERSE
	default:
		// Set default scope to LINK
		scope = unix.RT_SCOPE_LINK
	}

	var srclen int
	if opts.Src != nil {
		srclen, _ = opts.Src.Mask.Size()
		opts.Attrs.Src = opts.Src.IP
	}

	dstlen, _ := dst.Mask.Size()

	tx := &rtnetlink.RouteMessage{
		Family:     uint8(af),
		Table:      unix.RT_TABLE_MAIN,
		Protocol:   unix.RTPROT_BOOT,
		Type:       unix.RTN_UNICAST,
		Scope:      scope,
		DstLength:  uint8(dstlen),
		SrcLength:  uint8(srclen),
		Attributes: opts.Attrs,
	}
	return tx, nil
}

// RouteAdd adds infomation about a network route.
func (c *Conn) RouteAdd(ifc *net.Interface, dst net.IPNet, gw net.IP, options ...RouteOption) (err error) {
	rm, err := genRouteMessage(ifc, dst, gw, options...)
	if err != nil {
		return err
	}

	return c.Conn.Route.Add(rm)
}

// RouteReplace adds or replace information about a network route.
func (c *Conn) RouteReplace(ifc *net.Interface, dst net.IPNet, gw net.IP, options ...RouteOption) (err error) {
	rm, err := genRouteMessage(ifc, dst, gw, options...)
	if err != nil {
		return err
	}
	return c.Conn.Route.Replace(rm)
}

// RouteDel deletes the route to the given destination.
func (c *Conn) RouteDel(ifc *net.Interface, dst net.IPNet) error {
	af, err := addrFamily(dst.IP)
	if err != nil {
		return err
	}
	prefixlen, _ := dst.Mask.Size()
	attr := rtnetlink.RouteAttributes{
		Dst:      dst.IP,
		OutIface: uint32(ifc.Index),
	}
	tx := &rtnetlink.RouteMessage{
		Family:     uint8(af),
		Table:      unix.RT_TABLE_MAIN,
		DstLength:  uint8(prefixlen),
		Attributes: attr,
	}
	return c.Conn.Route.Delete(tx)
}
