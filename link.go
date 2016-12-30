package rtnetlink

import (
	"errors"
	"local/rtnetlink/netlink"
	"local/rtnetlink/netlink/nlenc"
	"net"
)

var (
	// errInvalidMessage is returned when a LinkMessage is malformed.
	errInvalidMessage = errors.New("rtnetlink LinkMessage is invalid or too short")
)

var _ Message = &LinkMessage{}

const (
	rtmNewLink = 16
	rtmDelLink = 17
	rtmGetLink = 18
	rtmSetLink = 19
)

// An LinkMessage is a route netlink link message.
type LinkMessage struct {
	// Always set to AF_UNSPEC (0)
	Family uint16

	// Device Type
	Type uint16

	// Unique interface index, using a nonzero value with
	// NewLink will instruct the kernel to create a
	// device with the given index (kernel 3.7+ required)
	Index uint32

	// Contains device flags, see netdevice(7)
	Flags uint32

	// Change Flags, reserved for future use and should
	// always be 0xffffffff
	Change uint32

	// Each LinkMessage can contain an optional Attributes list
	Attributes *LinkAttributes
}

const linkMessageLength = 16

// MarshalBinary marshals a LinkMessage into a byte slice.
func (m *LinkMessage) MarshalBinary() ([]byte, error) {
	b := make([]byte, linkMessageLength)

	b[0] = 0 //Family
	b[1] = 0 //reserved
	nlenc.PutUint16(b[2:4], m.Type)
	nlenc.PutUint32(b[4:8], m.Index)
	nlenc.PutUint32(b[8:12], m.Flags)
	nlenc.PutUint32(b[12:16], 0) //Change, reserved

	return b, nil

}

// UnmarshalBinary unmarshals the contents of a byte slice into a LinkMessage.
func (m *LinkMessage) UnmarshalBinary(b []byte) error {
	m.Family = nlenc.Uint16(b[0:2])
	m.Type = nlenc.Uint16(b[2:4])
	m.Index = nlenc.Uint32(b[4:8])
	m.Flags = nlenc.Uint32(b[8:12])
	m.Change = nlenc.Uint32(b[12:16])

	if len(b) > 16 {
		la := &LinkAttributes{}
		err := la.UnmarshalBinary(b[16:])
		if err != nil {
			return err
		}

		m.Attributes = la
	}

	return nil
}

// rtMessage is an empty method to sattisfy the Message interface.
func (*LinkMessage) rtMessage() {}

// LinkService is used to retrieve rtnetlink family information.
type LinkService struct {
	c *Conn
}

// New creates a new interface using the LinkMessage information.
func (l *LinkService) New(m LinkMessage) error {
	return nil
}

// Delete removes an interface by index.
func (l *LinkService) Delete(ifIndex int) error {
	return nil
}

// Get retrieves interface information by index.
func (l *LinkService) Get(req *LinkMessage) (LinkMessage, error) {
	req.Family = 17

	flags := netlink.HeaderFlagsRoot
	msg, err := l.c.Execute(req, 18, flags)
	if err != nil {
		return LinkMessage{}, err
	}

	link := (msg[0]).(*LinkMessage)
	return *link, nil
}

// List retrieves all interfaces.
func (l *LinkService) List() ([]LinkMessage, error) {
	req := &LinkMessage{}

	flags := netlink.HeaderFlagsRequest | netlink.HeaderFlagsDump
	msgs, err := l.c.Execute(req, 18, flags)
	if err != nil {
		return nil, err
	}

	links := make([]LinkMessage, 0, len(msgs))
	for _, m := range msgs {
		link := (m).(*LinkMessage)
		links = append(links, *link)
	}

	return links, nil
}

// Set sets interface attributes according to the LinkMessage information.
func (l *LinkService) Set(m LinkMessage) error {
	return nil
}

// LinkAttributes contains all attributes for an interface.
type LinkAttributes struct {
	Address   net.HardwareAddr // Interface L2 address
	Broadcast net.HardwareAddr // L2 broadcast address
	Name      string           // Device name
	MTU       uint32           // MTU of the device
	Type      uint32           // Link type
	QueueDisc string           // Queueing discipline
	Stats     *LinkStats       // Interface Statistics
}

// Attribute IDs mapped to specific LinkAttribute fields.
const (
	iflaUnspec uint16 = iota
	iflaAddress
	iflaBroadcast
	iflaIfname
	iflaMTU
	iflaLink
	iflaQdisc
	iflaStats
)

// UnmarshalBinary unmarshals the contents of a byte slice into a LinkMessage.
func (a *LinkAttributes) UnmarshalBinary(b []byte) error {
	attrs, err := netlink.UnmarshalAttributes(b)
	if err != nil {
		return err
	}

	for _, attr := range attrs {
		switch attr.Type {
		case iflaUnspec:
			//unused attribute
			continue
		case iflaAddress:
			a.Address = attr.Data
		case iflaBroadcast:
			a.Broadcast = attr.Data
		case iflaIfname:
			a.Name = nlenc.String(attr.Data)
		case iflaMTU:
			a.MTU = nlenc.Uint32(attr.Data)
		case iflaLink:
			a.Type = nlenc.Uint32(attr.Data)
		case iflaQdisc:
			a.QueueDisc = nlenc.String(attr.Data)
			/*
				case iflaStats:
					stats := &LinkStats{}
					err := stats.UnmarshalBinary(attr.Data)
					if err != nil {
						return err
					}
					a.Stats = stats
			*/
		}
	}

	return nil
}

//LinkStats contains packet statistics
type LinkStats struct {
	//further unmarshalled info, types tbd
	/*
		__u64   rx_packets;             // total packets received
		__u64   tx_packets;             // total packets transmitted
		__u64   rx_bytes;               // total bytes received
		__u64   tx_bytes;               // total bytes transmitted
		__u64   rx_errors;              // bad packets received
		__u64   tx_errors;              // packet transmit problems
		__u64   rx_dropped;             // no space in linux buffers
		__u64   tx_dropped;             // no space available in linux
		__u64   multicast;              // multicast packets received
		__u64   collisions;

		// detailed rx_errors:
		__u64   rx_length_errors;
		__u64   rx_over_errors;         // receiver ring buff overflow
		__u64   rx_crc_errors;          // recved pkt with crc error
		__u64   rx_frame_errors;        // recv'd frame alignment error
		__u64   rx_fifo_errors;         // recv'r fifo overrun
		__u64   rx_missed_errors;       // receiver missed packet

		// detailed tx_errors
		__u64   tx_aborted_errors;
		__u64   tx_carrier_errors;
		__u64   tx_fifo_errors;
		__u64   tx_heartbeat_errors;
		__u64   tx_window_errors;

		// for cslip etc
		__u64   rx_compressed;
		__u64   tx_compressed;

		__u64   rx_nohandler;           // dropped, no handler found
	*/
}

// UnmarshalBinary unmarshals the contents of a byte slice into a LinkMessage.
func (a *LinkStats) UnmarshalBinary(b []byte) error {
	/*
		attrs, err := netlink.UnmarshalAttributes(b)
		if err != nil {
			return err
		}

		for _, attr := range attrs {
			switch attr.Type {
			default:
				// TODO(jsimonetti): parse LinkStats
			}
		}
	*/
	return nil
}
func buildLinkMessages(msgs []Message) ([]LinkMessage, error) {
	links := make([]LinkMessage, 0, len(msgs))
	for _, m := range msgs {
		link := (m).(*LinkMessage)
		links = append(links, *link)
	}

	return links, nil
}
