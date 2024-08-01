package models

// Eventually we'll autogenerate this file.

func NewMX(short, origin string, priority uint16, mx string) *RecordConfig {
	rc := &RecordConfig{}
	rc.SetLabel(short, origin)
	rc.SetTargetMX(priority, mx)
	return rc
}
