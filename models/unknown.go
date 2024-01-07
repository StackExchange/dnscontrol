package models

func MakeUnknown(rc *RecordConfig, rtype string, contents string, origin string) error {
	rc.Type = "UNKNOWN"
	rc.UnknownTypeName = rtype
	rc.target = contents

	return nil
	//return fmt.Errorf("unknown rtype (%s) when parsing (%s) domain=(%s)", rtype, contents, origin)
}
