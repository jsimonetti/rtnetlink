package rtnetlink

import (
	"errors"
	"net"
	"unsafe"

	"github.com/jsimonetti/rtnetlink/internal/unix"

	"github.com/mdlayher/netlink"
)

var (
	// errInvalidRouteMessage is returned when a RouteMessage is malformed.
	errInvalidRouteMessage = errors.New("rtnetlink RouteMessage is invalid or too short")

	// errInvalidRouteMessageAttr is returned when link attributes are malformed.
	errInvalidRouteMessageAttr = errors.New("rtnetlink RouteMessage has a wrong attribute data length")
)

var _ Message = &RouteMessage{}

type RouteMessage struct {
	Family    uint8 // Address family (current unix.AF_INET or unix.AF_INET6)
	DstLength uint8 // Length of destination prefix
	SrcLength uint8 // Length of source prefix
	Tos       uint8 // TOS filter
	Table     uint8 // Routing table ID
	Protocol  uint8 // Routing protocol
	Scope     uint8 // Distance to the destination
	Type      uint8 // Route type
	Flags     uint32

	Attributes RouteAttributes
}

func (m *RouteMessage) MarshalBinary() ([]byte, error) {
	b := make([]byte, unix.SizeofRtMsg)

	b[0] = m.Family
	b[1] = m.DstLength
	b[2] = m.SrcLength
	b[3] = m.Tos
	b[4] = m.Table
	b[5] = m.Protocol
	b[6] = m.Scope
	b[7] = m.Type
	nativeEndian.PutUint32(b[8:12], m.Flags)

	ae := netlink.NewAttributeEncoder()
	err := m.Attributes.encode(ae)
	if err != nil {
		return nil, err
	}

	a, err := ae.Encode()
	if err != nil {
		return nil, err
	}

	return append(b, a...), nil
}

func (m *RouteMessage) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < unix.SizeofRtMsg {
		return errInvalidRouteMessage
	}

	m.Family = uint8(b[0])
	m.DstLength = uint8(b[1])
	m.SrcLength = uint8(b[2])
	m.Tos = uint8(b[3])
	m.Table = uint8(b[4])
	m.Protocol = uint8(b[5])
	m.Scope = uint8(b[6])
	m.Type = uint8(b[7])
	m.Flags = nativeEndian.Uint32(b[8:12])

	if l > unix.SizeofRtMsg {
		m.Attributes = RouteAttributes{}
		ad, err := netlink.NewAttributeDecoder(b[unix.SizeofRtMsg:])
		if err != nil {
			return err
		}
		err = m.Attributes.decode(ad)
		if err != nil {
			return err
		}
	}

	return nil
}

// rtMessage is an empty method to sattisfy the Message interface.
func (*RouteMessage) rtMessage() {}

type RouteService struct {
	c *Conn
}

func (r *RouteService) execute(m Message, family uint16, flags netlink.HeaderFlags) ([]*RouteMessage, error) {
	msgs, err := r.c.Execute(m, family, flags)

	routes := make([]*RouteMessage, 0, len(msgs))
	for i := range msgs {
		routes[i] = msgs[i].(*RouteMessage)
	}

	return routes, err
}

// Add new route
func (r *RouteService) Add(req *RouteMessage) error {
	flags := netlink.Request | netlink.Create | netlink.Acknowledge | netlink.Excl
	_, err := r.c.Execute(req, unix.RTM_NEWROUTE, flags)

	return err
}

// Replace or add new route
func (r *RouteService) Replace(req *RouteMessage) error {
	flags := netlink.Request | netlink.Create | netlink.Replace | netlink.Acknowledge
	_, err := r.c.Execute(req, unix.RTM_NEWROUTE, flags)

	return err
}

// Delete existing route
func (r *RouteService) Delete(req *RouteMessage) error {
	flags := netlink.Request | netlink.Acknowledge
	_, err := r.c.Execute(req, unix.RTM_DELROUTE, flags)

	return err
}

// Get Route(s)
func (r *RouteService) Get(req *RouteMessage) ([]*RouteMessage, error) {
	flags := netlink.Request | netlink.DumpFiltered
	return r.execute(req, unix.RTM_GETROUTE, flags)
}

// List all routes
func (r *RouteService) List() ([]*RouteMessage, error) {
	flags := netlink.Request | netlink.Dump
	return r.execute(&RouteMessage{}, unix.RTM_GETROUTE, flags)
}

type RouteAttributes struct {
	Dst       net.IP
	Src       net.IP
	Gateway   net.IP
	OutIface  uint32
	Priority  uint32
	Table     uint32
	Mark      uint32
	Expires   *uint32
	Metrics   *RouteMetrics
	Multipath []NextHop
}

func (a *RouteAttributes) decode(ad *netlink.AttributeDecoder) error {

	for ad.Next() {
		switch ad.Type() {
		case unix.RTA_UNSPEC:
			//unused attribute
		case unix.RTA_DST:
			l := len(ad.Bytes())
			if l != 4 && l != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Dst = ad.Bytes()
		case unix.RTA_PREFSRC:
			l := len(ad.Bytes())
			if l != 4 && l != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Src = ad.Bytes()
		case unix.RTA_GATEWAY:
			l := len(ad.Bytes())
			if l != 4 && l != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Gateway = ad.Bytes()
		case unix.RTA_OIF:
			a.OutIface = ad.Uint32()
		case unix.RTA_PRIORITY:
			a.Priority = ad.Uint32()
		case unix.RTA_TABLE:
			a.Table = ad.Uint32()
		case unix.RTA_MARK:
			a.Mark = ad.Uint32()
		case unix.RTA_EXPIRES:
			timeout := ad.Uint32()
			a.Expires = &timeout
		case unix.RTA_METRICS:
			a.Metrics = &RouteMetrics{}
			ad.Nested(a.Metrics.decode)
		case unix.RTA_MULTIPATH:
			ad.Do(a.parseMultipath)
		}
	}

	return ad.Err()
}

func (a *RouteAttributes) encode(ae *netlink.AttributeEncoder) error {
	if a.Dst != nil {
		if ipv4 := a.Dst.To4(); ipv4 == nil {
			// Dst Addr is IPv6
			ae.Bytes(unix.RTA_DST, a.Dst)
		} else {
			// Dst Addr is IPv4
			ae.Bytes(unix.RTA_DST, ipv4)
		}
	}

	if a.Src != nil {
		if ipv4 := a.Src.To4(); ipv4 == nil {
			// Src Addr is IPv6
			ae.Bytes(unix.RTA_PREFSRC, a.Src)
		} else {
			// Src Addr is IPv4
			ae.Bytes(unix.RTA_PREFSRC, ipv4)
		}
	}

	if a.Gateway != nil {
		if ipv4 := a.Gateway.To4(); ipv4 == nil {
			// Gateway Addr is IPv6
			ae.Bytes(unix.RTA_GATEWAY, a.Gateway)
		} else {
			// Gateway Addr is IPv4
			ae.Bytes(unix.RTA_GATEWAY, ipv4)
		}
	}

	if a.OutIface != 0 {
		ae.Uint32(unix.RTA_OIF, a.OutIface)
	}

	if a.Priority != 0 {
		ae.Uint32(unix.RTA_PRIORITY, a.Priority)
	}

	if a.Table != 0 {
		ae.Uint32(unix.RTA_TABLE, a.Table)
	}

	if a.Mark != 0 {
		ae.Uint32(unix.RTA_MARK, a.Mark)
	}

	if a.Expires != nil {
		ae.Uint32(unix.RTA_EXPIRES, *a.Expires)
	}

	if a.Metrics != nil {
		ae.Nested(unix.RTA_METRICS, a.Metrics.encode)
	}

	if len(a.Multipath) > 0 {
		ae.Do(unix.RTA_MULTIPATH, a.encodeMultipath)
	}

	return nil
}

// RouteMetrics holds some advanced metrics for a route
type RouteMetrics struct {
	AdvMSS   uint32
	Features uint32
	InitCwnd uint32
	MTU      uint32
}

func (rm *RouteMetrics) decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.RTAX_ADVMSS:
			rm.AdvMSS = ad.Uint32()
		case unix.RTAX_FEATURES:
			rm.Features = ad.Uint32()
		case unix.RTAX_INITCWND:
			rm.InitCwnd = ad.Uint32()
		case unix.RTAX_MTU:
			rm.MTU = ad.Uint32()
		}
	}

	// ad.Err call handled by Nested method in calling attribute decoder.
	return nil
}

func (rm *RouteMetrics) encode(ae *netlink.AttributeEncoder) error {
	if rm.AdvMSS != 0 {
		ae.Uint32(unix.RTAX_ADVMSS, rm.AdvMSS)
	}

	if rm.Features != 0 {
		ae.Uint32(unix.RTAX_FEATURES, rm.Features)
	}

	if rm.InitCwnd != 0 {
		ae.Uint32(unix.RTAX_INITCWND, rm.InitCwnd)
	}

	if rm.MTU != 0 {
		ae.Uint32(unix.RTAX_MTU, rm.MTU)
	}

	return nil
}

// TODO(mdlayher): probably eliminate Length field from the API to avoid the
// caller possibly tampering with it since we can compute it.

// RTNextHop represents the netlink rtnexthop struct (not an attribute)
type RTNextHop struct {
	Length  uint16 // length of this hop including nested values
	Flags   uint8  // flags defined in rtnetlink.h line 311
	Hops    uint8
	IfIndex uint32 // the interface index number
}

// NextHop wraps struct rtnexthop to provide access to nested attributes
type NextHop struct {
	Hop     RTNextHop // a rtnexthop struct
	Gateway net.IP    // that struct's nested Gateway attribute
}

func (a *RouteAttributes) encodeMultipath() ([]byte, error) {
	var b []byte
	for _, nh := range a.Multipath {
		// Encode the attributes first so their total length can be used to
		// compute the length of each (rtnexthop, attributes) pair.
		ae := netlink.NewAttributeEncoder()

		if a.Gateway != nil {
			// TODO(mdlayher): more validation.
			ae.Bytes(unix.RTA_GATEWAY, nh.Gateway)
		}

		ab, err := ae.Encode()
		if err != nil {
			return nil, err
		}

		// Assume the caller wants the length updated so they don't have to
		// keep track of it themselves when encoding attributes.
		nh.Hop.Length = unix.SizeofRtNexthop + uint16(len(ab))
		var nhb [unix.SizeofRtNexthop]byte

		copy(
			nhb[:],
			(*(*[unix.SizeofRtNexthop]byte)(unsafe.Pointer(&nh.Hop)))[:],
		)

		// rtnexthop first, then attributes.
		b = append(b, nhb[:]...)
		b = append(b, ab...)
	}

	return b, nil
}

func (a *RouteAttributes) parseMultipath(b []byte) error {
	// check for truncated message
	if len(b) <= unix.SizeofRtNexthop {
		return errInvalidRouteMessageAttr
	}

	// Iterate through the nested array of rtnexthop, unpacking each and appending them to mp
	for i := 0; i <= len(b); {
		// check for end of message
		if len(b)-i < unix.SizeofRtNexthop {
			return nil
		}

		// Copy over the struct portion
		var nh NextHop
		var nhb [unix.SizeofRtNexthop]byte
		copy(nhb[:], b[i:i+unix.SizeofRtNexthop])

		copy(
			(*(*[unix.SizeofRtNexthop]byte)(unsafe.Pointer(&nh.Hop)))[:],
			(*(*[unix.SizeofRtNexthop]byte)(unsafe.Pointer(&nhb[0])))[:],
		)

		// check again for a truncated message
		if int(nh.Hop.Length) > len(b) {
			return errInvalidRouteMessageAttr
		}

		// grab a new attributedecoder for the nested attributes
		start := (i + unix.SizeofRtNexthop)
		end := (i + int(nh.Hop.Length))

		ad, err := netlink.NewAttributeDecoder(b[start:end])
		if err != nil {
			return err
		}

		// read in the nested attributes
		if err := nh.decode(ad); err != nil {
			return err
		}

		// append this hop to the parent Multipath struct
		a.Multipath = append(a.Multipath, nh)

		// move forward to the next element in multipath.[]nexthop
		i += int(nh.Hop.Length)
	}

	return nil
}

// TODO: Implement func (mp *RTMultiPath) encode()

// rtnexthop payload is at least one nested attribute RTA_GATEWAY
// possibly others?
func (nh *NextHop) decode(ad *netlink.AttributeDecoder) error {
	for ad.Next() {
		switch ad.Type() {
		case unix.RTA_GATEWAY:
			l := len(ad.Bytes())
			if l != 4 && l != 16 {
				return errInvalidRouteMessageAttr
			}

			nh.Gateway = ad.Bytes()
		}
	}

	return ad.Err()
}
