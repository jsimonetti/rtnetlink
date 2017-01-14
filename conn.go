package rtnetlink

import (
	"encoding"

	"github.com/mdlayher/netlink"
)

// Protocol is the netlink protocol constant used to specify rtnetlink.
const Protocol = 0x0

// A Conn is a route netlink connection. A Conn can be used to send and
// receive route netlink messages to and from netlink.
type Conn struct {
	c       conn
	Link    *LinkService
	Address *AddressService
}

var _ conn = &netlink.Conn{}

// A conn is a netlink connection, which can be swapped for tests.
type conn interface {
	Close() error
	Send(m netlink.Message) (netlink.Message, error)
	Receive() ([]netlink.Message, error)
}

// Dial dials a route netlink connection.  Config specifies optional
// configuration for the underlying netlink connection.  If config is
// nil, a default configuration will be used.
func Dial(config *netlink.Config) (*Conn, error) {
	c, err := netlink.Dial(Protocol, config)
	if err != nil {
		return nil, err
	}

	return newConn(c), nil
}

// newConn is the internal constructor for Conn, used in tests.
func newConn(c conn) *Conn {
	rtc := &Conn{
		c: c,
	}

	rtc.Link = &LinkService{c: rtc}
	rtc.Address = &AddressService{c: rtc}

	return rtc
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.c.Close()
}

// Send sends a single Message to netlink, wrapping it in a netlink.Message
// using the specified generic netlink family and flags.  On success, Send
// returns a copy of the netlink.Message with all parameters populated, for
// later validation.
func (c *Conn) Send(m Message, family uint16, flags netlink.HeaderFlags) (netlink.Message, error) {
	nm := netlink.Message{
		Header: netlink.Header{
			Type:  netlink.HeaderType(family),
			Flags: flags,
		},
	}

	mb, err := m.MarshalBinary()
	if err != nil {
		return netlink.Message{}, err
	}
	nm.Data = mb
	reqnm, err := c.c.Send(nm)
	if err != nil {
		return netlink.Message{}, err
	}

	return reqnm, nil
}

// Receive receives one or more Messages from netlink.  The netlink.Messages
// used to wrap each Message are available for later validation.
func (c *Conn) Receive() ([]Message, []netlink.Message, error) {
	msgs, err := c.c.Receive()
	if err != nil {
		return nil, nil, err
	}

	return messageUnmarshall(msgs)
}

// messageUnmarshall will unmarshal the message based on its type
func messageUnmarshall(msgs []netlink.Message) ([]Message, []netlink.Message, error) {
	lmsgs := make([]Message, 0, len(msgs))

	for _, nm := range msgs {
		switch nm.Header.Type {
		case rtmNewAddress:
			m := &AddressMessage{}
			if err := (m).UnmarshalBinary(nm.Data); err != nil {
				return nil, nil, err
			}
			lmsgs = append(lmsgs, m)
		case rtmNewLink:
			m := &LinkMessage{}
			if err := (m).UnmarshalBinary(nm.Data); err != nil {
				return nil, nil, err
			}
			lmsgs = append(lmsgs, m)

		}

	}

	return lmsgs, msgs, nil
}

// Execute sends a single Message to netlink using Conn.Send, receives one or
// more replies using Conn.Receive, and then checks the validity of the replies
// against the request using netlink.Validate.
//
// See the documentation of Conn.Send, Conn.Receive, and netlink.Validate for
// details about each function.
func (c *Conn) Execute(m Message, family uint16, flags netlink.HeaderFlags) ([]Message, error) {
	req, err := c.Send(m, family, flags)
	if err != nil {
		return nil, err
	}

	msgs, replies, err := c.Receive()
	if err != nil {
		return nil, err
	}

	if err := netlink.Validate(req, replies); err != nil {
		return nil, err
	}

	return msgs, nil
}

//Message is the interface used for passing around different kinds of rtnetlink messages
type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	rtMessage()
}
