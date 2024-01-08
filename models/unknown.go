package models

func MakeUnknown(rc *RecordConfig, rtype string, contents string, origin string) error {
	rc.Type = "UNKNOWN"
	rc.UnknownTypeName = rtype
	rc.target = contents

	return nil
}
