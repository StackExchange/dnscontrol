package models

import (
	"bytes"
	"encoding/gob"
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
