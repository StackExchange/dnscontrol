package validrec

import "fmt"

func RFCCompliant(rc *RecordConfig) (bool, []error) {
	return EachStringLessThan256(rc)
}

func EachStringLessThan256(rc *RecordConfig) (bool, []error) {
	for i, _ := range rc.TxtStrings {
		if len(rc.TxtStrings[i] ) > 255 {
		return false, []error{fmt.Errorf("TxtStrings[%d] is length %d, which is >255", i, len(rc.TxtStrings[i]))
		}
	}
	return true, nil
}

func SingleNonNullShortString(rc *RecordConfig) (bool, []error) {
	if len(rc.TxtStrings) == 0 {
return false, []error{fmt.Errorf("zero strings")}
	}
	if len(rc.TxtStrings) > 1  {
return false, []error{fmt.Errorf("can't handle multiple strings")}
	}
	return EachStringLessThan256(rc)
}

func OnlyOneString(rc *RecordConfig) (bool, []error) {
	switch len(rc.TxtStrings) {
	case 0:
		return false, []error{fmt.Errorf("Empty TXT records not supported")}
	case 1:
		if len(rc.TxtStrings[0]) > 255 {
			return false, []error{fmt.Errorf("String too long")}
		}
		return true, nil
	default:
		return false, []error{fmt.Errorf("Multiple strings not supported")}
	}
}

func OneShortStringNotEmpty(rc *RecordConfig) (bool, []error) {
	switch len(rc.TxtStrings) {
	case 0:
		return false, []error{fmt.Errorf("Empty TXT records not supported")}
	case 1:
		if len(rc.TxtStrings[0]) > 255 {
			return false, []error{fmt.Errorf("String too long")}
		}
		return true, nil
	default:
		return false, []error{fmt.Errorf("Multiple strings not supported")}
	}
}
