package models

import (
	"strconv"

	"github.com/pkg/errors"
)

func atou32(s string) uint32 {
	i64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(errors.Errorf("atou32 failed (%v) (err=%v", s, err))
	}
	return uint32(i64)
}
