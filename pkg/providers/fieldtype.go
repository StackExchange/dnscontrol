package providers

//go:generate stringer -type FieldType -trimprefix FieldType

type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeBool
)
