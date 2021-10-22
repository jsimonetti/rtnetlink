//go:build integration
// +build integration

package rtnl

import (
	"net"
	"testing"
)

func TestLiveRoute(t *testing.T) {
	c, err := Dial(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	route, err := c.RouteGet(net.ParseIP("8.8.8.8"))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("got route: %v", route)
	if route.Gateway.IsUnspecified() {
		t.Error("zero route.Gateway, expected non-zero")
	}
	if route.Gateway.IsLoopback() {
		t.Error("lo route.Gatway, expected non lo")
	}
	if route.Interface == nil {
		t.Error("nil route.Interface, expected non-nil")
	}
	if len(route.Interface.Name) == 0 {
		t.Error("zero-length route.Interface.Name")
	}
	if hardwareAddrIsUnspecified(route.Interface.HardwareAddr) {
		t.Error("zero route.Interface.HardwareAddr, expected non-zero")
	}
}
