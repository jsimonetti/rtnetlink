package rtnetlink

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jsimonetti/rtnetlink/internal/unix"
)

// Tests will only pass on little endian machines

func TestRouteMessageMarshalUnmarshalBinary(t *testing.T) {
	skipBigEndian(t)

	var (
		timeout = uint32(255)
		pref    = uint8(1)
	)

	tests := []struct {
		name string
		m    *RouteMessage
		b    []byte
	}{
		{
			name: "empty",
			m:    &RouteMessage{},
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "no attributes",
			m: &RouteMessage{
				Family:    unix.AF_INET,
				DstLength: 8,
				Type:      unix.RTN_UNICAST,
			},
			b: []byte{
				0x02, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
				0x00, 0x00, 0x00, 0x00,
			},
		},
		{
			name: "full",
			m: &RouteMessage{
				Family:    2,
				DstLength: 8,
				Table:     unix.RT_TABLE_MAIN,
				Protocol:  unix.RTPROT_STATIC,
				Scope:     unix.RT_SCOPE_UNIVERSE,
				Type:      unix.RTN_UNICAST,
				Attributes: RouteAttributes{
					Dst:      net.IPv4(10, 0, 0, 0),
					Src:      net.IPv4(10, 100, 10, 1),
					Gateway:  net.IPv4(10, 0, 0, 1),
					OutIface: 5,
					Priority: 1,
					Table:    2,
					Mark:     3,
					Pref:     &pref,
					Expires:  &timeout,
					Metrics: &RouteMetrics{
						AdvMSS:   1,
						Features: 0xffffffff,
						InitCwnd: 2,
						InitRwnd: 3,
						MTU:      1500,
					},
					Multipath: []NextHop{
						{
							Hop: RTNextHop{
								Length:  16,
								IfIndex: 1,
							},
							Gateway: net.IPv4(10, 0, 0, 2),
						},
						{
							Hop: RTNextHop{
								Length:  16,
								IfIndex: 2,
							},
							Gateway: net.IPv4(10, 0, 0, 3),
						},
					},
				},
			},
			b: []byte{
				// RouteMessage struct literal
				//
				// Family
				0x02,
				// DstLength
				0x08,
				// SrcLength
				0x00,
				// Tos
				0x00,
				// Table
				0xfe,
				// Protocol
				0x04,
				// Scope
				0x00,
				// Type
				0x01,
				// Flags
				0x00, 0x00, 0x00, 0x00,
				// RouteAttributes
				// 2 bytes length, 2 bytes type, N bytes value
				//
				// Dst
				0x08, 0x00, 0x01, 0x00,
				0x0a, 0x00, 0x00, 0x00,
				// Src
				0x08, 0x00, 0x07, 0x00,
				0x0a, 0x64, 0x0a, 0x01,
				// Gateway
				0x08, 0x00, 0x05, 0x00,
				0x0a, 0x00, 0x00, 0x01,
				// OutIface
				0x08, 0x00, 0x04, 0x00,
				0x05, 0x00, 0x00, 0x00,
				// Priority
				0x08, 0x00, 0x06, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// Table
				0x08, 0x00, 0x0f, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// Mark
				0x08, 0x00, 0x10, 0x00,
				0x03, 0x00, 0x00, 0x00,
				// Pref
				0x05, 0x00, 0x14, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// Expires
				0x08, 0x00, 0x17, 0x00,
				0xff, 0x00, 0x00, 0x00,
				// RouteMetrics
				// Length must be manually adjusted as more fields are added.
				0x2c, 0x00, 0x08, 0x80,
				// AdvMSS
				0x08, 0x00, 0x08, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// Features
				0x08, 0x00, 0x0c, 0x00,
				0xff, 0xff, 0xff, 0xff,
				// InitCwnd
				0x08, 0x00, 0x0b, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// InitRwnd
				0x08, 0x00, 0x0e, 0x00,
				0x03, 0x00, 0x00, 0x00,
				// MTU
				0x08, 0x00, 0x02, 0x00,
				0xdc, 0x05, 0x00, 0x00,
				// Multipath
				//
				// 2 bytes length, 2 bytes type, then repeated 8 byte rtnexthop
				// structures followed by their nested netlink attributes.
				0x24, 0x00, 0x09, 0x00,
				// rtnexthop
				0x10, 0x00, 0x00, 0x00,
				0x01, 0x00, 0x00, 0x00,
				// rtnexthop attributes
				0x08, 0x00, 0x05, 0x00,
				// Gateway
				10, 0, 0, 2,
				// rtnexthop
				0x10, 0x00, 0x00, 0x00,
				0x02, 0x00, 0x00, 0x00,
				// rtnexthop attributes
				0x08, 0x00, 0x05, 0x00,
				// Gateway
				10, 0, 0, 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// It's important to be able to parse raw bytes into valid
			// structures so we start with that step first. After, we'll do a
			// marshaling round-trip to ensure that the structure's byte output
			// and parsed form match what is expected, while also comparing
			// against the expected fixtures throughout.
			var m1 RouteMessage
			if err := m1.UnmarshalBinary(tt.b); err != nil {
				t.Fatalf("failed to unmarshal first message from binary: %v", err)
			}

			if diff := cmp.Diff(tt.m, &m1); diff != "" {
				t.Fatalf("unexpected first message (-want +got):\n%s", diff)
			}

			b, err := m1.MarshalBinary()
			if err != nil {
				t.Fatalf("failed to marshal first message binary: %v", err)
			}

			if diff := cmp.Diff(tt.b, b); diff != "" {
				t.Fatalf("unexpected first message bytes (-want +got):\n%s", diff)
			}

			var m2 RouteMessage
			if err := m2.UnmarshalBinary(b); err != nil {
				t.Fatalf("failed to unmarshal second message from binary: %v", err)
			}

			if diff := cmp.Diff(&m1, &m2); diff != "" {
				t.Fatalf("unexpected parsed messages (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRouteMessageMarshalRoundTrip(t *testing.T) {
	skipBigEndian(t)

	// The above tests begin with unmarshaling raw bytes and are more
	// comprehensive, but due to the complexity of nested route message
	// attributes and structures, it has become rather difficult to maintain
	// over time. These tests will focus on a subset of that functionality to
	// ensure that marshaling and unmarshaling perform symmetrical operations
	// given a proper Go type as input, rather than raw bytes.

	tests := []struct {
		name string
		m    *RouteMessage
	}{
		{
			name: "multipath IPv4 MPLS",
			m: &RouteMessage{
				Attributes: RouteAttributes{
					Multipath: []NextHop{
						{
							Hop: RTNextHop{
								Length:  36,
								IfIndex: 1,
							},
							Gateway: net.IPv4(10, 0, 0, 2),
							MPLS: []MPLSNextHop{{
								Label:         1,
								TrafficClass:  1,
								BottomOfStack: true,
								TTL:           1,
							}},
						},
						{
							Hop: RTNextHop{
								Length:  40,
								IfIndex: 2,
							},
							Gateway: net.IPv4(10, 0, 0, 3),
							MPLS: []MPLSNextHop{
								{
									Label:        1,
									TrafficClass: 1,
									TTL:          1,
								},
								{
									Label:         2,
									TrafficClass:  2,
									BottomOfStack: true,
									TTL:           2,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multipath IPv6 MPLS",
			m: &RouteMessage{
				Attributes: RouteAttributes{
					Multipath: []NextHop{
						{
							Hop: RTNextHop{
								Length:  48,
								IfIndex: 1,
							},
							Gateway: net.ParseIP("2001:db8::1"),
							MPLS: []MPLSNextHop{{
								Label:         1,
								TrafficClass:  1,
								BottomOfStack: true,
								TTL:           1,
							}},
						},
						{
							Hop: RTNextHop{
								Length:  52,
								IfIndex: 2,
							},
							Gateway: net.ParseIP("2001:db8::2"),
							MPLS: []MPLSNextHop{
								{
									Label:        1,
									TrafficClass: 1,
									TTL:          1,
								},
								{
									Label:         2,
									TrafficClass:  2,
									BottomOfStack: true,
									TTL:           2,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First do a marshaling and unmarshaling round trip to ensure
			// the inputs and outputs are identical.
			b1, err := tt.m.MarshalBinary()
			if err != nil {
				t.Fatalf("failed to marshal test message: %v", err)
			}

			var m RouteMessage
			if err := m.UnmarshalBinary(b1); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if diff := cmp.Diff(tt.m, &m); diff != "" {
				t.Fatalf("unexpected RouteMessage after round-trip (-want +got):\n%s", diff)
			}

			// Then compare the results of the first marshaled bytes against
			// the newly marshaled bytes.
			b2, err := m.MarshalBinary()
			if err != nil {
				t.Fatalf("failed to marshal parsed message: %v", err)
			}

			if diff := cmp.Diff(b1, b2); diff != "" {
				t.Fatalf("unexpected final raw byte output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRouteMessageUnmarshalBinaryErrors(t *testing.T) {
	skipBigEndian(t)

	tests := []struct {
		name string
		b    []byte
		m    Message
	}{
		{
			name: "empty",
		},
		{
			name: "short",
			b:    make([]byte, 3),
		},
		{
			name: "invalid attr",
			b: []byte{
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x04, 0x00, 0x01, 0x00, 0x04, 0x00, 0x02, 0x00,
				0x05, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x08, 0x00, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x05, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m RouteMessage
			err := m.UnmarshalBinary(tt.b)
			if err == nil {
				t.Fatal("expected an error, but none occurred")
			}

			t.Logf("err: %v", err)
		})
	}
}

func TestRouteMessageFuzz(t *testing.T) {
	skipBigEndian(t)

	tests := []struct {
		name string
		s    string
	}{
		// Strings in this test table are copied from go-fuzz crashers.
		{
			name: "short rtnexthop",
			s: "\xef\xbf\xea\x00\a\x00\xd1\xea\xf9A\b\xf9\b\x00\t\x00\xbfA\b\xf9" +
				"\b\x00\a\x00\xf9A\b\xf9\b\x00\a\x00\xbfA\b\xf9\b\x00\a\x00" +
				"\xd3\xea\xf9A\b\x00\a\u007f\xff\xff\xffA\b\x00\a\x00\xd3\xea\xf9A" +
				"\b\x00\a\x00\xbfA\b\xf9\b\x00\a\x00\xd3\xea\xf9A\b\x00\a\x00" +
				"\xd3\xea\xf9A\b\x00\a\x00\xbfA\b\xf9\b\x00\a\x00\xd3-\xbf\xbd",
		},
		{
			name: "out of bounds attributes length",
			s: "000000000000\x14\x00\t\x000\xea00" +
				"000000000000",
		},
		{
			name: "bad rtnexthop length",
			s: "000000000000!\x00\t\x00\b\x0000" +
				"0000\b\x00000000\x06\x00000000" +
				"00000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m RouteMessage
			if err := m.UnmarshalBinary([]byte(tt.s)); err == nil {
				t.Fatal("expected an error, but none occurred")
			}
		})
	}
}
