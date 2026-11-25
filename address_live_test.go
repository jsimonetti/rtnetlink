//go:build integration
// +build integration

package rtnetlink

import (
	"net"
	"testing"

	"github.com/jsimonetti/rtnetlink/v2/internal/testutils"
	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

// lifetimeMatches reports whether the remaining lifetime reported by the kernel
// matches the requested one: finite lifetimes tick down, so allow for some slack.
func lifetimeMatches(got, want uint32) bool {
	const slack = 5

	if want == 0xffffffff || want <= slack {
		return got == want
	}

	return got <= want && got >= want-slack
}

func TestAddressLifetimes(t *testing.T) {
	ns := testutils.NetNS(t)
	conn, err := Dial(&netlink.Config{NetNS: ns})
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	const dummyIndex = 2001

	err = conn.Link.New(&LinkMessage{
		Index: dummyIndex,
		Attributes: &LinkAttributes{
			Name: "testdum0",
			Info: &LinkInfo{
				Kind: "dummy",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to create dummy interface: %v", err)
	}
	defer conn.Link.Delete(dummyIndex)

	const infinity = 0xffffffff

	tests := []struct {
		name      string
		addr      net.IP
		cacheInfo CacheInfo
		wantFlags uint32
		noFlags   uint32
	}{
		{
			name:      "default",
			addr:      net.ParseIP("2001:db8::1"),
			wantFlags: unix.IFA_F_PERMANENT,
			noFlags:   unix.IFA_F_DEPRECATED,
		},
		{
			name: "deprecated",
			addr: net.ParseIP("2001:db8::2"),
			cacheInfo: CacheInfo{
				Preferred: 0,
				Valid:     infinity,
			},
			wantFlags: unix.IFA_F_PERMANENT | unix.IFA_F_DEPRECATED,
		},
		{
			name: "finite",
			addr: net.ParseIP("2001:db8::3"),
			cacheInfo: CacheInfo{
				Preferred: 1800,
				Valid:     3600,
			},
			noFlags: unix.IFA_F_PERMANENT | unix.IFA_F_DEPRECATED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conn.Address.New(&AddressMessage{
				Family:       unix.AF_INET6,
				PrefixLength: 128,
				Index:        dummyIndex,
				Attributes: &AddressAttributes{
					Address:   tt.addr,
					CacheInfo: tt.cacheInfo,
					Flags:     unix.IFA_F_NODAD,
				},
			})
			if err != nil {
				t.Fatalf("failed to add address: %v", err)
			}

			addrs, err := conn.Address.List()
			if err != nil {
				t.Fatalf("failed to list addresses: %v", err)
			}

			var found *AddressMessage

			for i := range addrs {
				if addrs[i].Index == dummyIndex && addrs[i].Attributes.Address.Equal(tt.addr) {
					found = &addrs[i]

					break
				}
			}

			if found == nil {
				t.Fatalf("address %s not found", tt.addr)
			}

			flags := found.Attributes.Flags

			if flags&tt.wantFlags != tt.wantFlags {
				t.Fatalf("expected flags %#x to be set, got %#x", tt.wantFlags, flags)
			}

			if flags&tt.noFlags != 0 {
				t.Fatalf("expected flags %#x to be unset, got %#x", tt.noFlags, flags)
			}

			want := tt.cacheInfo
			if want == (CacheInfo{}) {
				want = CacheInfo{Preferred: infinity, Valid: infinity}
			}

			ci := found.Attributes.CacheInfo

			if !lifetimeMatches(ci.Preferred, want.Preferred) || !lifetimeMatches(ci.Valid, want.Valid) {
				t.Fatalf("unexpected lifetimes: got preferred %d valid %d, want preferred %d valid %d",
					ci.Preferred, ci.Valid, want.Preferred, want.Valid)
			}
		})
	}
}
