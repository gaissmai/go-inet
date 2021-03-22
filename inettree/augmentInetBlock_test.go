package inettree

import (
	"net"
	"testing"
)

// augment inet.ParseBlock or inet.FromStdIPNet
func TestNewItem(t *testing.T) {
	s := "::1"
	_, err := NewItem(s, "")
	if err != nil {
		t.Errorf("NewItem(string), got error: %v", err)
	}

	_, cidr, _ := net.ParseCIDR("::/0")
	_, err = NewItem(*cidr, "")
	if err != nil {
		t.Errorf("NewItem(net.IPNet), got error: %v", err)
	}

	_, err = NewItem(5, "")
	if err == nil {
		t.Errorf("NewItem(int), expected error, got: %v", err)
	}
}
