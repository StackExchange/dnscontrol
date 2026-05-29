package txtutil

import (
	"strings"

	"codeberg.org/miekg/dns/pkg/pool"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil/ddd"
)

var builderPool = pool.NewBuilder()

func Zoneify(txt []string) string {
	sb := builderPool.Get()
	defer builderPool.Put(sb)

	for i, s := range txt {
		sb.Grow(3 + len(s))
		if i > 0 {
			sb.WriteString(` "`)
		} else {
			sb.WriteByte('"')
		}
		for j := 0; j < len(s); {
			b, n := ddd.Next(s, j)
			if n == 0 {
				break
			}
			writeTxtByte(&sb, b)
			j += n
		}
		sb.WriteByte('"')
	}
	return sb.String()
}

func writeTxtByte(sb *strings.Builder, b byte) {
	switch {
	case b == '"' || b == '\\':
		sb.WriteByte('\\')
		sb.WriteByte(b)
	case b < ' ' || b > '~':
		sb.WriteString(ddd.Escape(b))
	default:
		sb.WriteByte(b)
	}
}
