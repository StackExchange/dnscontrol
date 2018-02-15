package models

import (
	"bytes"
	"encoding/gob"
	"strconv"

	"github.com/pkg/errors"
)

func copyObj(input interface{}, output interface{}) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)
	if err := enc.Encode(input); err != nil {
		return err
	}
	return dec.Decode(output)
}

// atou32 converts a string  to uint32 or panics.
// DEPRECATED: This will go away when SOA record handling is rewritten.
func atou32(s string) uint32 {
	i64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		panic(errors.Errorf("atou32 failed (%v) (err=%v", s, err))
	}
	return uint32(i64)
}
