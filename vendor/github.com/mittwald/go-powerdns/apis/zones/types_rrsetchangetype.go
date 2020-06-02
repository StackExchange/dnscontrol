package zones

import "fmt"

type RecordSetChangeType int

const (
	_                                    = iota
	ChangeTypeDelete RecordSetChangeType = iota
	ChangeTypeReplace
)

func (k RecordSetChangeType) MarshalJSON() ([]byte, error) {
	switch k {
	case ChangeTypeDelete:
		return []byte(`"DELETE"`), nil
	case ChangeTypeReplace:
		return []byte(`"REPLACE"`), nil
	default:
		return nil, fmt.Errorf("unsupported change type: %d", k)
	}
}

func (k *RecordSetChangeType) UnmarshalJSON(input []byte) error {
	switch string(input) {
	case `"DELETE"`:
		*k = ChangeTypeDelete
	case `"REPLACE"`:
		*k = ChangeTypeReplace
	default:
		return fmt.Errorf("unsupported change type: %s", string(input))
	}

	return nil
}
