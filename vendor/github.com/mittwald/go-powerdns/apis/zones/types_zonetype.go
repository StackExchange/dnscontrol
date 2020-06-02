package zones

import "fmt"

type ZoneType int

const (
	ZoneTypeZone ZoneType = iota
)

func (k ZoneType) MarshalJSON() ([]byte, error) {
	switch k {
	case ZoneTypeZone:
		return []byte(`"Zone"`), nil
	default:
		return nil, fmt.Errorf("unsupported zone type: %d", k)
	}
}

func (k *ZoneType) UnmarshalJSON(input []byte) error {
	switch string(input) {
	case `"Zone"`:
		*k = ZoneTypeZone
	default:
		return fmt.Errorf("unsupported zone type: %s", string(input))
	}

	return nil
}
