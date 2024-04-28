package models

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"strconv"
	"strings"
)

// SetTargetSVCB sets the SVCB fields.
func (rc *RecordConfig) SetTargetSVCB(priority uint16, target string, params []dns.SVCBKeyValue) error {
	rc.SvcPriority = priority
	rc.SetTarget(target)
	paramsStr := []string{}
	for _, kv := range params {
		paramsStr = append(paramsStr, fmt.Sprintf("%s=%s", kv.Key(), kv.String()))
	}
	rc.SvcParams = strings.Join(paramsStr, " ")
	if rc.Type == "" {
		rc.Type = "SVCB"
	}
	if rc.Type != "SVCB" && rc.Type != "HTTPS" {
		panic("assertion failed: SetTargetSVCB called when .Type is not SVCB or HTTPS")
	}
	return nil
}

// SetTargetSVCBStrings is like SetTargetSVCB but accepts strings.
func (rc *RecordConfig) SetTargetSVCBStrings(priority, target, params string) error {
	u64prio, err := strconv.ParseUint(priority, 10, 16)
	if err != nil {
		return fmt.Errorf("can't parse SVCB data: %w", err)
	}

	paramsData := []dns.SVCBKeyValue{}

	for _, kv := range strings.Split(params, " ") {
		kv = strings.TrimSpace(kv)
		args := strings.Split(kv, "=")
		if len(args) != 2 {
			return fmt.Errorf("can't parse SVCB data as key=value pair: %s", kv)
		}
		key := strings.TrimSpace(args[0])
		value := strings.TrimSpace(args[1])
		value = strings.Trim(value, `"`)
		switch key {
		case "alpn":
			alpn := new(dns.SVCBAlpn)
			alpn.Alpn = strings.Split(value, ",")
			paramsData = append(paramsData, alpn)
		case "ipv4hint":
			alpn := new(dns.SVCBIPv4Hint)
			data := strings.Split(value, ",")
			for _, ip := range data {
				alpn.Hint = append(alpn.Hint, net.ParseIP(strings.TrimSpace(ip)))
			}
			paramsData = append(paramsData, alpn)
		case "ipv6hint":
			alpn := new(dns.SVCBIPv6Hint)
			data := strings.Split(value, ",")
			for _, ip := range data {
				alpn.Hint = append(alpn.Hint, net.ParseIP(strings.TrimSpace(ip)))
			}
			paramsData = append(paramsData, alpn)
		case "port":
			port := new(dns.SVCBPort)
			uport, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				return fmt.Errorf("can't parse port: %s", value)
			}
			port.Port = uint16(uport)
			paramsData = append(paramsData, port)
		}
	}
	return rc.SetTargetSVCB(uint16(u64prio), target, paramsData)
}

// SetTargetSVCBString is like SetTargetSVCB but accepts one big string.
func (rc *RecordConfig) SetTargetSVCBString(s string) error {
	part := strings.Fields(s)
	if len(part) != 3 {
		return fmt.Errorf("SVCB value does not contain 3 fields: (%#v)", s)
	}
	return rc.SetTargetSVCBStrings(part[0], part[1], part[2])
}
