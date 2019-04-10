package rtnetlink

import (
	"errors"
	"fmt"
	"net"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
	"golang.org/x/sys/unix"
)

var (
	// errInvalidNeighMessage is returned when a LinkMessage is malformed.
	errInvalidNeighMessage = errors.New("rtnetlink NeighMessage is invalid or too short")

	// errInvalidNeighMessageAttr is returned when neigh attributes are malformed.
	errInvalidNeighMessageAttr = errors.New("rtnetlink NeighMessage has a wrong attribute data length")
)

var _ Message = &NeighMessage{}

// A NeighMessage is a route netlink neighbor message.
type NeighMessage struct {
	// Always set to AF_UNSPEC (0)
	Family uint16

	// Unique interface index
	Index uint32

	// Neighbor State is a bitmask of neighbor states (see rtnetlink(7))
	State uint16

	// Neighbor flags
	Flags uint8

	// Neighbor type
	Type uint8

	// Attributes List
	Attributes *NeighAttributes
}

const (
	NTF_USE         = 0x01
	NTF_SELF        = 0x02
	NTF_MASTER      = 0x04
	NTF_PROXY       = 0x08
	NTF_EXT_LEARNED = 0x10
	NTF_OFFLOADED   = 0x20
	NTF_ROUTER      = 0x80
)

const neighMsgLen = 12

// MarshalBinary marshals a NeighMessage into a byte slice.
func (m *NeighMessage) MarshalBinary() ([]byte, error) {
	b := make([]byte, neighMsgLen)

	nlenc.PutUint16(b[0:2], m.Family)
	// bytes 3 and 4 are padding
	nlenc.PutUint32(b[4:8], m.Index)
	nlenc.PutUint16(b[8:10], m.State)
	b[10] = m.Flags
	b[11] = m.Type

	if m.Attributes != nil {
		a, err := m.Attributes.MarshalBinary()
		if err != nil {
			return nil, err
		}

		return append(b, a...), nil
	}
	return b, nil
}

// UnmarshalBinary unmarshals the contents of a byte slice into a NeighMessage.
func (m *NeighMessage) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < neighMsgLen {
		return errInvalidNeighMessage
	}

	m.Family = nlenc.Uint16(b[0:2])
	m.Index = nlenc.Uint32(b[4:8])
	m.State = nlenc.Uint16(b[8:10])
	m.Flags = b[10]
	m.Type = b[11]

	if l > neighMsgLen {
		m.Attributes = &NeighAttributes{}
		err := m.Attributes.UnmarshalBinary(b[neighMsgLen:])
		if err != nil {
			return err
		}
	}

	return nil
}

// rtMessage is an empty method to sattisfy the Message interface.
func (*NeighMessage) rtMessage() {}

// NeighService is used to retrieve rtnetlink family information.
type NeighService struct {
	c *Conn
}

// New creates a new interface using the LinkMessage information.
func (l *NeighService) New(req *NeighMessage) error {
	flags := netlink.Request | netlink.Create | netlink.Acknowledge | netlink.Excl
	_, err := l.c.Execute(req, unix.RTM_NEWNEIGH, flags)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes an neighbor entry by index.
func (l *NeighService) Delete(index uint32) error {
	req := &NeighMessage{}

	flags := netlink.Request | netlink.Acknowledge
	_, err := l.c.Execute(req, unix.RTM_DELNEIGH, flags)
	if err != nil {
		return err
	}

	return nil
}

// List retrieves all neighbors.
func (l *NeighService) List() ([]NeighMessage, error) {
	req := &NeighMessage{}

	flags := netlink.Request | netlink.Dump
	msgs, err := l.c.Execute(req, unix.RTM_GETNEIGH, flags)
	if err != nil {
		return nil, err
	}

	neighs := make([]NeighMessage, 0, len(msgs))
	for _, m := range msgs {
		neigh := (m).(*NeighMessage)
		neighs = append(neighs, *neigh)
	}

	return neighs, nil
}

// NeighCacheInfo contains neigh information
type NeighCacheInfo struct {
	Confirmed uint32
	Used      uint32
	Updated   uint32
	RefCount  uint32
}

// UnmarshalBinary unmarshals the contents of a byte slice into a NeighMessage.
func (n *NeighCacheInfo) UnmarshalBinary(b []byte) error {
	if len(b) != 16 {
		return fmt.Errorf("incorrect size, want: 16, got: %d", len(b))
	}

	n.Confirmed = nlenc.Uint32(b[0:4])
	n.Used = nlenc.Uint32(b[4:8])
	n.Updated = nlenc.Uint32(b[8:12])
	n.RefCount = nlenc.Uint32(b[12:16])

	return nil
}

// NeighAttributes contains all attributes for a neighbor.
type NeighAttributes struct {
	Address   net.IP           // a neighbor cache n/w layer destination address
	LLAddress net.HardwareAddr // a neighbor cache link layer address
	CacheInfo *NeighCacheInfo  // cache statistics
	IfIndex   uint32
}

const (
	NDA_UNSPEC uint16 = iota
	NDA_DST
	NDA_LLADDR
	NDA_CACHEINFO
	NDA_PROBES
	NDA_VLAN
	NDA_PORT
	NDA_VNI
	NDA_IFINDEX
	NDA_MASTER
	NDA_LINK_NETNSID
	NDA_SRC_VNI
)

// NeighAttributes unmarshals the contents of a byte slice into a NeighMessage.
func (a *NeighAttributes) UnmarshalBinary(b []byte) error {
	attrs, err := netlink.UnmarshalAttributes(b)
	if err != nil {
		return err
	}

	for _, attr := range attrs {
		switch attr.Type {
		case NDA_UNSPEC:
			//unused attribute
		case NDA_DST:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidNeighMessageAttr
			}
			a.Address = attr.Data
		case NDA_LLADDR:
			if len(attr.Data) != 6 {
				return errInvalidNeighMessageAttr
			}
			a.LLAddress = attr.Data
		case NDA_CACHEINFO:
			a.CacheInfo = &NeighCacheInfo{}
			err := a.CacheInfo.UnmarshalBinary(attr.Data)
			if err != nil {
				return err
			}
		case NDA_IFINDEX:
			if len(attr.Data) != 4 {
				return errInvalidNeighMessageAttr
			}
			a.IfIndex = nlenc.Uint32(attr.Data)
		}
	}

	return nil
}

// MarshalBinary marshals a NeighAttributes into a byte slice.
func (a *NeighAttributes) MarshalBinary() ([]byte, error) {
	attrs := []netlink.Attribute{
		{
			Type: NDA_UNSPEC,
			Data: nlenc.Uint16Bytes(0),
		},
		{
			Type: NDA_DST,
			Data: a.Address,
		},
		{
			Type: NDA_LLADDR,
			Data: a.LLAddress,
		},
		{
			Type: NDA_IFINDEX,
			Data: nlenc.Uint32Bytes(a.IfIndex),
		},
	}

	return netlink.MarshalAttributes(attrs)
}
