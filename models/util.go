package models

import (
	"strconv"

	"github.com/pkg/errors"
)

// func atou8(s string) uint8 {
// 	i64, err := strconv.ParseUint(s, 10, 8)
// 	if err != nil {
// 		panic(errors.Errorf("atou8 failed (%v) (err=%v", s, err))
// 	}
// 	return uint8(i64)
// }

// func atou16(s string) uint16 {
// 	i64, err := strconv.ParseUint(s, 10, 16)
// 	if err != nil {
// 		panic(errors.Errorf("atou16 failed (%v) (err=%v", s, err))
// 	}
// 	return uint16(i64)
// }

func atou32(s string) uint32 {
	i64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(errors.Errorf("atou32 failed (%v) (err=%v", s, err))
	}
	return uint32(i64)
}
