package bind

import (
	"log"
	"strconv"
	"strings"
	"time"
)

var nowFunc func() time.Time = time.Now

// generate_serial takes an old SOA serial number and increments it.
func generate_serial(old_serial uint32) uint32 {
	// Serial numbers are in the format yyyymmddvv
	// where vv is a version count that starts at 01 each day.
	// Multiple serial numbers generated on the same day increase vv.
	// If the old serial number is not in this format, it gets replaced
	// with the new format. However if that would mean a new serial number
	// that is smaller than the old one, we punt and increment the old number.
	// At no time will a serial number == 0 be returned.

	original := old_serial
	old_serialStr := strconv.FormatUint(uint64(old_serial), 10)
	var new_serial uint32

	// Make draft new serial number:
	today := nowFunc().UTC()
	todayStr := today.Format("20060102")
	version := uint32(1)
	todayNum, err := strconv.ParseUint(todayStr, 10, 32)
	if err != nil {
		log.Fatalf("new serial won't fit in 32 bits: %v", err)
	}
	draft := uint32(todayNum)*100 + version

	method := "none" // Used only in debugging.
	if old_serial > draft {
		// If old_serial was really slow, upgrade to new yyyymmddvv standard:
		method = "o>d"
		new_serial = old_serial + 1
		new_serial = old_serial + 1
	} else if old_serial == draft {
		// Edge case: increment old serial:
		method = "o=d"
		new_serial = draft + 1
	} else if len(old_serialStr) != 10 {
		// If old_serial is wrong number of digits, upgrade to yyyymmddvv standard:
		method = "len!=10"
		new_serial = draft
	} else if strings.HasPrefix(old_serialStr, todayStr) {
		// If old_serial just needs to be incremented:
		method = "prefix"
		new_serial = old_serial + 1
	} else {
		// First serial number to be requested today:
		method = "default"
		new_serial = draft
	}

	if new_serial == 0 {
		// We never return 0 as the serial number.
		new_serial = 1
	}
	if old_serial == new_serial {
		log.Fatalf("%v: old_serial == new_serial (%v == %v) draft=%v method=%v", original, old_serial, new_serial, draft, method)
	}
	if old_serial > new_serial {
		log.Fatalf("%v: old_serial > new_serial (%v > %v) draft=%v method=%v", original, old_serial, new_serial, draft, method)
	}
	return new_serial
}
