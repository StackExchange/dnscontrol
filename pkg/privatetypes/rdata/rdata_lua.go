package privatetypesrdata

import (
	"fmt"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v4/pkg/mustbe"
	"github.com/DNSControl/dnscontrol/v4/pkg/txtutil"
)

type LUA struct {
	LuaType    string `json:"lua_type"`
	LuaPayload string `json:"lua_payload"`
}

func (rd LUA) Len() int {
	return 0
}

func (rd LUA) String() string {
	return txtutil.Zoneify([]string{rd.LuaType, rd.LuaPayload})
}

func MakeLUA(origin string, args ...any) (dnsv2.RDATA, error) {
	if len(args) != 2 {
		return LUA{}, fmt.Errorf("LUA requires exactly 2 arguments, got %d: %+v", len(args), args)
	}
	return LUA{mustbe.RawString(args[0]), mustbe.RawString(args[1])}, nil
}
