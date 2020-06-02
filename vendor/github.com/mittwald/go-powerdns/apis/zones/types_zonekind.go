package zones

import "fmt"

type ZoneKind int

const (
	_                       = iota
	ZoneKindNative ZoneKind = iota
	ZoneKindMaster
	ZoneKindSlave
)

func (k ZoneKind) MarshalJSON() ([]byte, error) {
	switch k {
	case ZoneKindNative:
		return []byte(`"Native"`), nil
	case ZoneKindMaster:
		return []byte(`"Master"`), nil
	case ZoneKindSlave:
		return []byte(`"Slave"`), nil
	default:
		return nil, fmt.Errorf("unsupported zone kind: %d", k)
	}
}

func (k *ZoneKind) UnmarshalJSON(input []byte) error {
	switch string(input) {
	case `"Native"`:
		*k = ZoneKindNative
	case `"Master"`:
		*k = ZoneKindMaster
	case `"Slave"`:
		*k = ZoneKindSlave
	default:
		return fmt.Errorf("unsupported zone kind: %s", string(input))
	}

	return nil
}
