package rtnetlink

import "local/rtnetlink/netlink"
import "encoding"

// Protocol is the netlink protocol constant used to specify rtnetlink.
const Protocol = 0x0

// A Conn is a route netlink connection.  A Conn can be used to send and
// receive route netlink messages to and from netlink.
type Conn struct {
	c    conn
	Link *LinkService
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

	return rtc
}

// Close closes the connection.
func (c *Conn) Close() error {
	return c.c.Close()
}

// Execute sends a single Message to netlink using Conn.Send, receives one or
// more replies using Conn.Receive, and then checks the validity of the replies
// against the request using netlink.Validate.
//
// See the documentation of Conn.Send, Conn.Receive, and netlink.Validate for
// details about each function.
func (c *Conn) Execute(m Message, family uint16, flags netlink.HeaderFlags) ([]Message, error) {
	return nil, nil
}

//Message is the interface used for passing around different kinds of rtnetlink messages
type Message interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	rtMessage()
}
