package rtnetlink

import (
	"errors"
	"fmt"
	"net"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
	"syscall"
)

var (
	// errInvalidaddressMessage is returned when a AddressMessage is malformed.
	errInvalidAddressMessage = errors.New("rtnetlink AddressMessage is invalid or too short")

	// errInvalidAddressMessageAttr is returned when link attributes are malformed.
	errInvalidAddressMessageAttr = errors.New("rtnetlink AddressMessage has a wrong attribute data length")
)

var _ Message = &AddressMessage{}

// Address family constants
const (
	AFInet  = 2
	AFInet6 = 10
)

// A AddressMessage is a route netlink address message.
type AddressMessage struct {
	// Address family (current AFInet or AFInet6)
	Family uint8

	// Prefix length
	PrefixLength uint8

	// Contains address flags
	Flags uint8

	// Address Scope
	Scope uint8

	// Interface index
	Index uint32

	// Attributes List
	Attributes AddressAttributes
}

const addressMessageLength = 8

// MarshalBinary marshals a AddressMessage into a byte slice.
func (m *AddressMessage) MarshalBinary() ([]byte, error) {
	b := make([]byte, addressMessageLength)

	b[0] = m.Family
	b[1] = m.PrefixLength
	b[2] = m.Flags
	b[3] = m.Scope
	nlenc.PutUint32(b[4:8], m.Index)

	a, err := m.Attributes.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(b, a...), nil
}

// UnmarshalBinary unmarshals the contents of a byte slice into a AddressMessage.
func (m *AddressMessage) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < addressMessageLength {
		return errInvalidAddressMessage
	}

	m.Family = uint8(b[0])
	m.PrefixLength = uint8(b[1])
	m.Flags = uint8(b[3])
	m.Scope = uint8(b[4])
	m.Index = nlenc.Uint32(b[4:8])

	if l > addressMessageLength {
		m.Attributes = AddressAttributes{}
		err := m.Attributes.UnmarshalBinary(b[addressMessageLength:])
		if err != nil {
			return err
		}
	}

	return nil
}

// rtMessage is an empty method to sattisfy the Message interface.
func (*AddressMessage) rtMessage() {}

// AddressService is used to retrieve rtnetlink family information.
type AddressService struct {
	c *Conn
}

// Constants used to request information from rtnetlink addresses.
const (
	RTM_NEWADDR = syscall.RTM_NEWADDR
	RTM_DELADDR = syscall.RTM_DELADDR
	RTM_GETADDR = syscall.RTM_GETADDR
)

// New creates a new address using the AddressMessage information.
func (a *AddressService) New(req *AddressMessage) error {
	flags := netlink.HeaderFlagsRequest
	_, err := a.c.Send(req, RTM_NEWADDR, flags)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes an address by ip and interface index.
func (a *AddressService) Delete(address net.IP, index uint32) error {
	req := &AddressMessage{
		Index: index,
		Attributes: AddressAttributes{
			Address: address,
		},
	}

	flags := netlink.HeaderFlagsRequest
	_, err := a.c.Send(req, RTM_DELADDR, flags)
	if err != nil {
		return err
	}

	return nil
}

// List retrieves all addresses.
func (a *AddressService) List() ([]AddressMessage, error) {
	req := &AddressMessage{}

	flags := netlink.HeaderFlagsRequest | netlink.HeaderFlagsDump
	msgs, err := a.c.Execute(req, RTM_GETADDR, flags)
	if err != nil {
		return nil, err
	}

	addresses := make([]AddressMessage, 0, len(msgs))
	for _, m := range msgs {
		address := (m).(*AddressMessage)
		addresses = append(addresses, *address)
	}
	return addresses, nil
}

// AddressAttributes contains all attributes for an interface.
type AddressAttributes struct {
	Address   net.IP // Interface Ip address
	Local     net.IP // Local Ip address
	Label     string
	Broadcast net.IP    // Broadcast Ip address
	Anycast   net.IP    // Anycast Ip address
	CacheInfo CacheInfo // Address information
	Multicast net.IP    // Multicast Ip address
	Flags     uint32    // Address flags
}

// Attribute IDs mapped to specific LinkAttribute fields.
const (
	ifaUnspec uint16 = iota
	ifaAddress
	ifaLocal
	ifaLabel
	ifaBroadcast
	ifaAnycast
	ifaCacheInfo
	ifaMulticast
	ifaFlags
)

// UnmarshalBinary unmarshals the contents of a byte slice into a AddressMessage.
func (a *AddressAttributes) UnmarshalBinary(b []byte) error {
	attrs, err := netlink.UnmarshalAttributes(b)
	if err != nil {
		return err
	}
	for _, attr := range attrs {
		switch attr.Type {
		case iflaUnspec:
			//unused attribute
		case ifaAddress:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidAddressMessageAttr
			}
			a.Address = attr.Data
		case ifaLocal:
			if len(attr.Data) != 4 {
				return errInvalidAddressMessageAttr
			}
			a.Local = attr.Data
		case ifaLabel:
			a.Label = nlenc.String(attr.Data)
		case ifaBroadcast:
			if len(attr.Data) != 4 {
				return errInvalidAddressMessageAttr
			}
			a.Broadcast = attr.Data
		case ifaAnycast:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidAddressMessageAttr
			}
			a.Anycast = attr.Data
		case ifaCacheInfo:
			if len(attr.Data) != 16 {
				return errInvalidAddressMessageAttr
			}
			err := a.CacheInfo.UnmarshalBinary(attr.Data)
			if err != nil {
				return err
			}
		case ifaMulticast:
			if len(attr.Data) != 4 && len(attr.Data) != 16 {
				return errInvalidAddressMessageAttr
			}
			a.Multicast = attr.Data
		case ifaFlags:
			if len(attr.Data) != 4 {
				return errInvalidAddressMessageAttr
			}
			a.Flags = nlenc.Uint32(attr.Data)
		}
	}

	return nil
}

// MarshalBinary marshals a AddressAttributes into a byte slice.
func (a *AddressAttributes) MarshalBinary() ([]byte, error) {
	return netlink.MarshalAttributes([]netlink.Attribute{
		{
			Type: ifaUnspec,
			Data: nlenc.Uint16Bytes(0),
		},
		{
			Type: ifaAddress,
			Data: a.Address,
		},
		{
			Type: ifaLocal,
			Data: a.Local,
		},
		{
			Type: ifaBroadcast,
			Data: a.Broadcast,
		},
		{
			Type: ifaAnycast,
			Data: a.Anycast,
		},
		{
			Type: ifaMulticast,
			Data: a.Multicast,
		},
		{
			Type: ifaFlags,
			Data: nlenc.Uint32Bytes(a.Flags),
		},
	})
}

// CacheInfo contains address information
type CacheInfo struct {
	Prefered uint32
	Valid    uint32
	Created  uint32
	Updated  uint32
}

// UnmarshalBinary unmarshals the contents of a byte slice into a LinkMessage.
func (c *CacheInfo) UnmarshalBinary(b []byte) error {
	if len(b) != 16 {
		return fmt.Errorf("incorrect size, want: 16, got: %d", len(b))
	}

	c.Prefered = nlenc.Uint32(b[0:4])
	c.Valid = nlenc.Uint32(b[4:8])
	c.Created = nlenc.Uint32(b[8:12])
	c.Updated = nlenc.Uint32(b[12:16])

	return nil
}
