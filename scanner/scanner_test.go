package scanner

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

func TestScanner(t *testing.T) {
	ip := netip.MustParseAddr("45.144.178.132")
	out := RunAndFilter(ip, 443, 256, time.Second*3)
	fmt.Println(out)
}
