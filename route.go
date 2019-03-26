package rtnetlink

import (
	"errors"
	"net"

	"github.com/mdlayher/netlink"

	"github.com/mdlayher/netlink/nlenc"
	"golang.org/x/sys/unix"
)

var (
	// errInvalidRouteMessage is returned when a RouteMessage is malformed.
	errInvalidRouteMessage = errors.New("rtnetlink RouteMessage is invalid or too short")

	// errInvalidRouteMessageAttr is returned when link attributes are malformed.
	errInvalidRouteMessageAttr = errors.New("rtnetlink RouteMessage has a wrong attribute data length")
)

// var _ Message = &RouteMessage{}

type RouteMessage struct {
	Family    uint8 // Address family (current AFInet or AFInet6)
	DstLength uint8 // Length of destination
	SrcLength uint8 // Length of source
	Tos       uint8 // TOS filter
	Table     uint8 // Routing table ID
	Protocol  uint8 // Routing protocol
	Scope     uint8 // Distance to the destination
	Type      uint8 // Route type
	Flags     uint32

	Attributes RouteAttributes
}

const routeMessageLength = 12

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
	nlenc.PutUint32(b[8:12], m.Flags)

	a, err := m.Attributes.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(b, a...), nil
}

func (m *RouteMessage) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < routeMessageLength {
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
	m.Flags = nlenc.Uint32(b[8:12])

	if l > routeMessageLength {
		m.Attributes = RouteAttributes{}
		err := m.Attributes.UnmarshalBinary(b[routeMessageLength:])
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

// Constants used to request information from rtnetlink.
const (
	RTM_NEWROUTE = 0x18
	RTM_DELROUTE = 0x19
	RTM_GETROUTE = 0x1a
)

// Add new route
func (r *RouteService) Add(req *RouteMessage) error {
	flags := netlink.Request | netlink.Create | netlink.Acknowledge | netlink.Excl
	_, err := r.c.Execute(req, RTM_NEWROUTE, flags)
	if err != nil {
		return err
	}

	return nil
}

// Delete existing route
func (r *RouteService) Delete(req *RouteMessage) error {
	flags := netlink.Request
	_, err := r.c.Send(req, RTM_DELROUTE, flags)
	if err != nil {
		return err
	}

	return nil
}

// List all routes
func (r *RouteService) List() ([]RouteMessage, error) {
	req := &RouteMessage{}

	flags := netlink.Request | netlink.Dump
	msgs, err := r.c.Execute(req, RTM_GETROUTE, flags)
	if err != nil {
		return nil, err
	}

	routes := make([]RouteMessage, 0, len(msgs))
	for _, m := range msgs {
		route := (m).(*RouteMessage)
		routes = append(routes, *route)
	}

	return routes, nil
}

type RouteAttributes struct {
	Dst      net.IP
	Src      net.IP
	Gateway  net.IP
	OutIface uint32
	Priority uint32
	Table    uint32
}

const (
	RTA_UNSPEC uint16 = iota
	RTA_DST
	RTA_SRC
	RTA_IIF
	RTA_OIF
	RTA_GATEWAY
	RTA_PRIORITY
	RTA_PREFSRC
	RTA_METRICS
	RTA_MULTIPATH
	RTA_PROTOINFO // no longer used
	RTA_FLOW
	RTA_CACHEINFO
	RTA_SESSION // no longer used
	RTA_MP_ALGO // no longer used
	RTA_TABLE
	RTA_MARK
	RTA_MFC_STATS
	RTA_VIA
	RTA_NEWDST
	RTA_PREF
	RTA_ENCAP_TYPE
	RTA_ENCAP
	RTA_EXPIRES
	RTA_PAD
	RTA_UID
	RTA_TTL_PROPAGATE
	RTA_IP_PROTO
	RTA_SPORT
	RTA_DPORT
)

func (a *RouteAttributes) UnmarshalBinary(b []byte) error {
	attrs, err := netlink.UnmarshalAttributes(b)
	if err != nil {
		return err
	}
	for _, attr := range attrs {
		switch attr.Type {
		case RTA_UNSPEC:
		case RTA_DST:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Dst = attr.Data
		case RTA_PREFSRC:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Src = attr.Data
		case RTA_GATEWAY:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidRouteMessageAttr
			}
			a.Gateway = attr.Data
		case RTA_OIF:
			if len(attr.Data) != 4 {
				return errInvalidRouteMessageAttr
			}
			a.OutIface = nlenc.Uint32(attr.Data)
		case RTA_PRIORITY:
			if len(attr.Data) != 4 {
				return errInvalidRouteMessageAttr
			}
			a.Priority = nlenc.Uint32(attr.Data)
		case RTA_TABLE:
			if len(attr.Data) != 4 {
				return errInvalidRouteMessageAttr
			}
			a.Table = nlenc.Uint32(attr.Data)
		}
	}

	return nil
}

func (a *RouteAttributes) MarshalBinary() ([]byte, error) {
	attrs := make([]netlink.Attribute, 0)

	if a.Dst != nil {
		if ipv4 := a.Dst.To4(); ipv4 == nil {
			// Dst Addr is IPv6
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_DST,
				Data: a.Dst,
			})
		} else {
			// Dst Addr is IPv4
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_DST,
				Data: ipv4,
			})
		}
	}

	if a.Src != nil {
		if ipv4 := a.Src.To4(); ipv4 == nil {
			// Src Addr is IPv6
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_PREFSRC,
				Data: a.Src,
			})
		} else {
			// Src Addr is IPv4
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_PREFSRC,
				Data: ipv4,
			})
		}
	}

	if a.Gateway != nil {
		if ipv4 := a.Gateway.To4(); ipv4 == nil {
			// Gateway Addr is IPv6
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_GATEWAY,
				Data: a.Gateway,
			})
		} else {
			// Gateway Addr is IPv4
			attrs = append(attrs, netlink.Attribute{
				Type: RTA_GATEWAY,
				Data: ipv4,
			})
		}
	}

	if a.OutIface != 0 {
		attrs = append(attrs, netlink.Attribute{
			Type: RTA_OIF,
			Data: nlenc.Uint32Bytes(a.OutIface),
		})
	}

	if a.Priority != 0 {
		attrs = append(attrs, netlink.Attribute{
			Type: RTA_PRIORITY,
			Data: nlenc.Uint32Bytes(a.Priority),
		})
	}

	if a.Table != 0 {
		attrs = append(attrs, netlink.Attribute{
			Type: RTA_TABLE,
			Data: nlenc.Uint32Bytes(a.Table),
		})
	}

	return netlink.MarshalAttributes(attrs)
}

// Type (rtm_type)
const (
	RTN_UNSPEC      = iota
	RTN_UNICAST     // Gateway or direct route
	RTN_LOCAL       // Accept locally
	RTN_BROADCAST   // Accept locally as broadcast, send as broadcast
	RTN_ANYCAST     // Accept locally as broadcast, but send as unicast
	RTN_MULTICAST   // Multicast route
	RTN_BLACKHOLE   // Drop
	RTN_UNREACHABLE // Destination is unreachable
	RTN_PROHIBIT    // Administratively prohibited
	RTN_THROW       // Not in this table
	RTN_NAT         // Translate this address
	RTN_XRESOLVE    // Use external resolver
)

// Protocol (rtm_protocol)
const (
	RTPROT_UNSPEC   = 0
	RTPROT_REDIRECT = 1   // Route installed by ICMP redirects; not used by current IPv4
	RTPROT_KERNEL   = 2   // Route installed by kernel
	RTPROT_BOOT     = 3   // Route installed during boot
	RTPROT_STATIC   = 4   // Route installed by administrator
	RTPROT_GATED    = 8   // Apparently, GateD
	RTPROT_RA       = 9   // RDISC/ND router advertisements
	RTPROT_MRT      = 10  // Merit MRT
	RTPROT_ZEBRA    = 11  // Zebra
	RTPROT_BIRD     = 12  // BIRD
	RTPROT_DNROUTED = 13  // DECnet routing daemon
	RTPROT_XORP     = 14  // XORP
	RTPROT_NTK      = 15  // Netsukuku
	RTPROT_DHCP     = 16  // DHCP client
	RTPROT_MROUTED  = 17  // Multicast daemon
	RTPROT_BABEL    = 42  // Babel daemon
	RTPROT_BGP      = 186 // BGP Routes
	RTPROT_ISIS     = 187 // ISIS Routes
	RTPROT_OSPF     = 188 // OSPF Routes
	RTPROT_RIP      = 189 // RIP Routes
	RTPROT_EIGRP    = 192 // EIGRP Routes

)

// Scope
const (
	RT_SCOPE_UNIVERSE = 0
	RT_SCOPE_SITE     = 200
	RT_SCOPE_LINK     = 253
	RT_SCOPE_HOST     = 254
	RT_SCOPE_NOWHERE  = 255
)

// Table
const (
	RT_TABLE_UNSPEC  = 0
	RT_TABLE_COMPAT  = 252
	RT_TABLE_DEFAULT = 253
	RT_TABLE_MAIN    = 254
	RT_TABLE_LOCAL   = 255
	RT_TABLE_MAX     = 0xFFFFFFFF
)

// Flags
const (
	RTM_F_NOTIFY       = 0x100  // Notify user of route change
	RTM_F_CLONED       = 0x200  // This route is cloned
	RTM_F_EQUALIZE     = 0x400  // Multipath equalizer: NI
	RTM_F_PREFIX       = 0x800  // Prefix addresses
	RTM_F_LOOKUP_TABLE = 0x1000 // set rtm_table to FIB lookup result
	RTM_F_FIB_MATCH    = 0x2000 // return full fib lookup match
)
