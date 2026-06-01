package mustbe

import (
	"net/netip"
)

func IPv4(a any) netip.Addr {
	switch v := a.(type) {
	case string:
		a, err := netip.ParseAddr(v)
		if err != nil || !a.Is4() {
			return netip.Addr{}
		}
		return a
	case netip.Addr:
		if !v.Is4() {
			return netip.Addr{}
		}
		return v
	}
	panic("mustbe.IPv4: unhandled type")
}

func IPv6(a any) netip.Addr {
	switch v := a.(type) {
	case string:
		a, err := netip.ParseAddr(v)
		if err != nil || !a.Is6() {
			return netip.Addr{}
		}
		return a
	case netip.Addr:
		if !v.Is6() {
			return netip.Addr{}
		}
		return v
	}
	panic("mustbe.IPv6: unhandled type")
}
