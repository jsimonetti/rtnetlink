//go:build integration
// +build integration

package rtnetlink

import (
	"testing"

	"github.com/jsimonetti/rtnetlink/v2/internal/unix"
	"github.com/mdlayher/netlink"
)

func TestListMatch(t *testing.T) {
	c, err := Dial(&netlink.Config{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	routes, err := c.Route.ListMatch(&RouteMessage{
		Family: unix.AF_INET,
		Attributes: RouteAttributes{
			Table: unix.RT_TABLE_MAIN,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(routes) <= 1 {
		t.Fatal("expected multiple routes to be returned")
	}

	for _, rx := range routes {
		if rx.Attributes.Table != unix.RT_TABLE_MAIN {
			t.Fatalf("unepxected route from table %d", rx.Attributes.Table)
		}
	}
}
