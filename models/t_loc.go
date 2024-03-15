package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miekg/dns"
)

// SetTargetLOC sets the LOC fields from the rr.LOC type properties.
func (rc *RecordConfig) SetTargetLOC(ver uint8, lat uint32, lon uint32, alt uint32, siz uint8, hzp uint8, vtp uint8) error {
	rc.LocVersion = ver
	rc.LocLatitude = lat
	rc.LocLongitude = lon
	rc.LocAltitude = alt
	rc.LocSize = siz
	rc.LocHorizPre = hzp
	rc.LocVertPre = vtp

	if rc.Type == "" {
		rc.Type = "LOC"
	}
	if rc.Type != "LOC" {
		panic("assertion failed: SetTargetLOC called when .Type is not LOC")
	}
	return nil
}

// SetLOCParams is an intermediate function which passes the 12 input parameters
// for further processing to the LOC native 7 input binary format:
// LocVersion (0), LocLatitude, LocLongitude, LocAltitude, LocSize, LocVertPre, LocHorizPre
func (rc *RecordConfig) SetLOCParams(d1 uint8, m1 uint8, s1 float32, ns string,
	d2 uint8, m2 uint8, s2 float32, ew string, al int32, sz float32, hp float32, vp float32) error {

	err := rc.calculateLOCFields(d1, m1, s1, ns, d2, m2, s2, ew, al, sz, hp, vp)

	return err
}

// SetTargetLOCString is like SetTargetLOC but accepts one big string and origin
// Normally this is used when we receive a record string from provider records
// because e.g. the provider API passed rc.PopulateFromString()
func (rc *RecordConfig) SetTargetLOCString(origin string, contents string) error {
	// This is where text from provider records ingresses into the target field.
	// Fill the other fields derived from the TEXT here. LOC is special, and
	// needs more math.
	// We have to re-invent the wheel because the miekg dns library gives no
	// access to the objects properties, and internally the object is represented
	// by the dns.LOC format 💩

	// Build a string with which to init the rr.LOC object:
	str := fmt.Sprintf("%s. LOC %s\n", origin, contents)
	loc, err := dns.NewRR(str)
	if err != nil {
		return fmt.Errorf("can't parse LOC data: %w", err)
	}
	// We 'normalize' the record thru rr.LOC, to get defaults for absent properties.

	loctext := loc.String()
	loctext = strings.TrimSpace(strings.Split(loctext, "LOC")[1])

	err = rc.extractLOCFieldsFromStringInput(loctext)
	if err != nil {
		return fmt.Errorf("can't extractLOCFieldsFromStringInput from LOC data: %w", err)
	}
	rc.target = loctext
	if rc.Type == "" {
		rc.Type = "LOC"
	}
	if rc.Type != "LOC" {
		panic("assertion failed: SetTargetLOC called when .Type is not LOC")
	}
	return nil
}

// extractLOCFieldsFromStringInput is a helper to split an input string to
// the 12 variable inputs of integers and strings.
func (rc *RecordConfig) extractLOCFieldsFromStringInput(input string) error {
	var d1, m1, d2, m2 uint8
	var al int32
	var s1, s2 float32
	var ns, ew string
	var sz, hp, vp float32

	var err error
	_, err = fmt.Sscanf(input+"~", "%d %d %f %s %d %d %f %s %dm %fm %fm %fm~",
		&d1, &m1, &s1, &ns, &d2, &m2, &s2, &ew, &al, &sz, &hp, &vp)
	if err != nil {
		return fmt.Errorf("extractLOCFieldsFromStringInput: can't unpack LOC tex input data: %w", err)
	}
	// fmt.Printf("\ngot: %d %d %g %s %d %d %g %s %dm %0.2fm %0.2fm %0.2fm \n", d1, m1, s1, ns, d2, m2, s2, ew, al, sz, hp, vp)

	rc.calculateLOCFields(d1, m1, s1, ns, d2, m2, s2, ew, al, sz, hp, vp)

	return nil
}

// calculateLOCFields converts from 12 user inputs to the LOC 7 binary fields
func (rc *RecordConfig) calculateLOCFields(d1 uint8, m1 uint8, s1 float32, ns string,
	d2 uint8, m2 uint8, s2 float32, ew string, al int32, sz float32, hp float32, vp float32) error {
	// Crazy hairy shit happens here.
	// We already got the useful "string" version earlier. ¯\_(ツ)_/¯ code golf...
	const LOCEquator uint64 = 0x80000000       // 1 << 31 // RFC 1876, Section 2.
	const LOCPrimeMeridian uint64 = 0x80000000 // 1 << 31 // RFC 1876, Section 2.
	const LOCHours uint32 = 60 * 1000
	const LOCDegrees = 60 * LOCHours
	const LOCAltitudeBase int32 = 100000

	// Some providers want the original values, so we should keep them around
	rc.LocLatDegrees = d1
	rc.LocLatMinutes = m1
	rc.LocLatSeconds = s1
	rc.LocLatDirection = ns
	
	rc.LocLongDegrees = d2
	rc.LocLongMinutes = m2
	rc.LocLongSeconds = s2
	rc.LocLongDirection = ew

	rc.LocOrigAltitude = al
	rc.LocOrigSize = sz
	rc.LocOrigHorizPre = hp
	rc.LocOrigVertPre = vp

	lat := uint64((uint32(d1) * LOCDegrees) + (uint32(m1) * LOCHours) + uint32(s1*1000))
	lon := uint64((uint32(d2) * LOCDegrees) + (uint32(m2) * LOCHours) + uint32(s2*1000))
	if strings.ToUpper(ns) == "N" {
		rc.LocLatitude = uint32(LOCEquator + lat)
	} else { // "S"
		rc.LocLatitude = uint32(LOCEquator - lat)
	}
	if strings.ToUpper(ew) == "E" {
		rc.LocLongitude = uint32(LOCPrimeMeridian + lon)
	} else { // "W"
		rc.LocLongitude = uint32(LOCPrimeMeridian - lon)
	}
	// Altitude
	rc.LocAltitude = uint32(al+LOCAltitudeBase) * 100
	var err error
	// Size
	rc.LocSize, err = getENotationInt(sz)
	if err != nil {
		return err
	}
	// Horizontal Precision
	rc.LocHorizPre, err = getENotationInt(hp)
	if err != nil {
		return err
	}
	// Vertical Precision
	rc.LocVertPre, err = getENotationInt(vp)
	if err != nil {
		return err
	}
	// if hp != 0 {
	// } else {
	// 	rc.LocHorizPre = 22 // 1e6 10,000m default
	// }
	// if vp != 0 {
	// } else {
	// 	rc.LocVertPre = 19 // 1e3 10m default
	// }

	return nil
}

// getENotationInt produces a mantissa_exponent 4bits:4bits into a uint8
func getENotationInt(x float32) (uint8, error) {
	/*
	   9000000000cm = 9e9 == 153 (9^4 + 9) or 9<<4 + 9
	   800000000cm = 8e8 == 136 (8^4 + 8) or 8<<4 + 8
	   70000000cm = 7e7 == 119 (7^4 + 7) or 7<<4 + 7
	   6000000cm = 6e6 == 102 (6^4 + 6) or 6<<4 + 6
	   1000000cm = 1e6 == 22 (1^4 + 6) or 1<<4 + 6
	   500000cm = 5e5 == 85 (5^4 + 5) or 5<<4 + 5
	   40000cm = 4e4 == 68 (4^4 + 4) or 4<<4 + 4
	   3000cm = 3e3 == 51 (3^4 + 3) or 3<<4 + 3
	   1000cm = 1e3 == 19 (1^4 + 3) or 1<<4 + 1
	   200cm = 2e2 == 34 (2^4 + 2) or 2<<4 + 2
	   100cm = 1e2 == 18 (1^4 + 2) or 1<<4 + 2
	   10cm = 1e1 == 17 (1^4 + 1) or 1<<4 + 1
	   1cm = 1e0 == 16 (1^4 + 0) or 0<<4 + 0
	   0cm = 0e0 == 0
	*/
	// get int from cm value:
	num := strconv.Itoa(int(x * 100))
	// fmt.Printf("num: %s\n", num)
	// split string on zeroes to count zeroes:
	arr := strings.Split(num, "0")
	// fmt.Printf("arr: %s\n", arr)
	// get the leading digit:
	prefix, err := strconv.Atoi(arr[0])
	if err != nil {
		return 0, fmt.Errorf("can't unpack LOC base/mantissa: %w", err)
	}
	// fmt.Printf("prefix: %d\n", prefix)
	// fmt.Printf("lenArr-1: %d\n", len(arr)-1)
	// construct our x^e uint8
	value := uint8((prefix << 4) | (len(arr) - 1))
	// fmt.Printf("m_e: %d\n", value)
	return value, err
}
