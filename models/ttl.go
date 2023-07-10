package models

// TTL implements an optional TTL field, with a default of DefaultTTL.
type TTL struct {
	value *uint32
}

func NewTTL(ttl uint32) TTL {
	value := new(uint32)
	*value = ttl
	return TTL{
		value,
	}
}

// EmptyTTL returns a new TTL without an explicit value.
func EmptyTTL() TTL {
	return TTL{
		value: nil,
	}
}

func (ttl TTL) IsSet() bool {
	return ttl.value != nil
}

func (ttl TTL) Value() uint32 {
	if ttl.IsSet() {
		return *ttl.value
	} else {
		return DefaultTTL
	}
}

func (ttl *TTL) ValueRef() *uint32 {
	if ttl.IsSet() {
		return ttl.value
	} else {
		defaultTTL := new(uint32)
		*defaultTTL = DefaultTTL
		return defaultTTL
	}
}
