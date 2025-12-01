package rtypeinfo

import "github.com/StackExchange/dnscontrol/v4/pkg/rtypecontrol"

func IsModernType(t string) bool {
	_, ok := rtypecontrol.Func[t]
	return ok
}
