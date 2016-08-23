package transform

import (
	"fmt"
	"net"
	"strings"
)

type IpConversion struct {
	Low, High, NewBase net.IP
	NewIPs             []net.IP
}

func ipToUint(i net.IP) (uint32, error) {
	parts := i.To4()
	if parts == nil || len(parts) != 4 {
		return 0, fmt.Errorf("%s is not an ipv4 address", parts.String())
	}
	r := uint32(parts[0])<<24 | uint32(parts[1])<<16 | uint32(parts[2])<<8 | uint32(parts[3])
	return r, nil
}

func UintToIP(u uint32) net.IP {
	return net.IPv4(
		byte((u>>24)&255),
		byte((u>>16)&255),
		byte((u>>8)&255),
		byte((u)&255))
}

// DecodeTransformTable turns a string-encoded table into a list of conversions.
func DecodeTransformTable(transforms string) ([]IpConversion, error) {
	result := []IpConversion{}
	rows := strings.Split(transforms, ";")
	for ri, row := range rows {
		items := strings.Split(row, "~")
		if len(items) != 4 {
			return nil, fmt.Errorf("transform_table rows should have 4 elements. (%v) found in row (%v) of %#v\n", len(items), ri, transforms)
		}
		for i, item := range items {
			items[i] = strings.TrimSpace(item)
		}

		con := IpConversion{
			Low:     net.ParseIP(items[0]),
			High:    net.ParseIP(items[1]),
			NewBase: net.ParseIP(items[2]),
		}
		for _, ip := range strings.Split(items[3], ",") {
			if ip == "" {
				continue
			}
			addr := net.ParseIP(ip)
			if addr == nil {
				return nil, fmt.Errorf("%s is not a valid ip address", ip)
			}
			con.NewIPs = append(con.NewIPs, addr)
		}

		low, _ := ipToUint(con.Low)
		high, _ := ipToUint(con.High)
		if low > high {
			return nil, fmt.Errorf("transform_table Low should be less than High. row (%v) %v>%v (%v)\n", ri, con.Low, con.High, transforms)
		}
		if con.NewBase != nil && con.NewIPs != nil {
			return nil, fmt.Errorf("transform_table_rows should only specify one of NewBase or NewIP. Not both.")
		}
		result = append(result, con)
	}

	return result, nil
}

// TransformIP transforms a single ip address. If the transform results in multiple new targets, an error will be returned.
func TransformIP(address net.IP, transforms []IpConversion) (net.IP, error) {
	ips, err := TransformIPToList(address, transforms)
	if err != nil {
		return nil, err
	}
	if len(ips) != 1 {
		return nil, fmt.Errorf("Expect exactly one ip for TransformIP result. Got: %s", ips)
	}
	return ips[0], err
}

// TransformIPToList manipulates an net.IP based on a list of IpConversions. It can potentially expand one ip address into multiple addresses.
func TransformIPToList(address net.IP, transforms []IpConversion) ([]net.IP, error) {
	thisIP, err := ipToUint(address)
	if err != nil {
		return nil, err
	}
	for _, conv := range transforms {
		min, err := ipToUint(conv.Low)
		if err != nil {
			return nil, err
		}
		max, err := ipToUint(conv.High)
		if err != nil {
			return nil, err
		}
		if (thisIP >= min) && (thisIP <= max) {
			if conv.NewIPs != nil {
				return conv.NewIPs, nil
			}
			newbase, err := ipToUint(conv.NewBase)
			if err != nil {
				return nil, err
			}
			return []net.IP{UintToIP(newbase + (thisIP - min))}, nil
		}
	}
	return []net.IP{address}, nil
}
