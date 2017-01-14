package rtnetlink

import (
	"errors"
	"fmt"
	"net"
	"unsafe"

	"github.com/mdlayher/netlink"
	"github.com/mdlayher/netlink/nlenc"
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

	PutUint8(b[0], m.Family)
	PutUint8(b[1], m.PrefixLength)
	PutUint8(b[2], m.Flags)
	PutUint8(b[3], m.Scope)
	nlenc.PutUint32(b[4:8], m.Index)

	a, err := m.Attributes.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return append(b, a...), nil
}

// PutUint8 encodes a uint8 into b using the host machine's native endianness.
func PutUint8(b byte, v uint8) {
	*(*uint8)(unsafe.Pointer(&b)) = v
}

// Uint8 decodes a uint8 from b using the host machine's native endianness.
func Uint8(b byte) uint8 {
	return *(*uint8)(unsafe.Pointer(&b))
}

// UnmarshalBinary unmarshals the contents of a byte slice into a AddressMessage.
func (m *AddressMessage) UnmarshalBinary(b []byte) error {
	l := len(b)
	if l < addressMessageLength {
		return errInvalidAddressMessage
	}

	m.Family = Uint8(b[0])
	m.PrefixLength = Uint8(b[1])
	m.Flags = Uint8(b[3])
	m.Scope = Uint8(b[4])
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
	rtmNewAddress = 20
	rtmDelAddress = 21
	rtmGetAddress = 22
)

// New creates a new address using the AddressMessage information.
func (a *AddressService) New(req *AddressMessage) error {
	flags := netlink.HeaderFlagsRequest
	_, err := a.c.Send(req, rtmNewAddress, flags)
	if err != nil {
		return err
	}

	return nil
}

// Delete removes an address by index.
//TODO: dont use index for deletion
func (a *AddressService) Delete(index uint32) error {
	req := &AddressMessage{
		Index: index,
	}

	flags := netlink.HeaderFlagsRequest
	_, err := a.c.Send(req, rtmDelAddress, flags)
	if err != nil {
		return err
	}

	return nil
}

// Get retrieves address information by index.
func (a *AddressService) Get(index uint32) (AddressMessage, error) {
	req := &AddressMessage{
		Index: index,
	}

	flags := netlink.HeaderFlagsRequest | netlink.HeaderFlagsDumpFiltered
	msg, err := a.c.Execute(req, rtmGetAddress, flags)
	if err != nil {
		return AddressMessage{}, err
	}

	if len(msg) != 1 {
		return AddressMessage{}, fmt.Errorf("too many/little matches, expected 1")
	}

	address := (msg[0]).(*AddressMessage)
	return *address, nil
}

// List retrieves all addresses.
func (a *AddressService) List() ([]AddressMessage, error) {
	req := &AddressMessage{}

	flags := netlink.HeaderFlagsRequest | netlink.HeaderFlagsDump
	msgs, err := a.c.Execute(req, rtmGetAddress, flags)
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
	Broadcast net.IP // Broadcast Ip address
	Anycast   net.IP // Anycast Ip address
	CacheInfo []byte //Cache Info
	Multicast net.IP // Multicast Ip address
	Flags     uint32 // Address flags
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
		case iflaAddress:
			//if len(attr.Data) != 6 {
			//	return errInvalidAddressMessageAttr
			//}
			a.Address = attr.Data
		case iflaBroadcast:
			//if len(attr.Data) != 6 {
			//	return errInvalidAddressMessageAttr
			//}
			a.Broadcast = attr.Data
		}
	}

	return nil
}

// MarshalBinary marshals a AddressAttributes into a byte slice.
func (a *AddressAttributes) MarshalBinary() ([]byte, error) {
	return netlink.MarshalAttributes([]netlink.Attribute{
		{
			Type: iflaUnspec,
			Data: nlenc.Uint16Bytes(0),
		},
		{
			Type: iflaAddress,
			Data: a.Address,
		},
		{
			Type: iflaBroadcast,
			Data: a.Broadcast,
		},
	})
}
