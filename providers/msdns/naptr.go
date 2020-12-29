package msdns

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

type nabuf struct {
	bytes.Buffer
}

// Pack

func naptrToHex(rc *models.RecordConfig) string {
	if rc.Type != "NAPTR" {
		panic("naptrToHex: That's not a NAPTR!")
	}

	var b nabuf
	b.writeU16(rc.NaptrOrder)
	b.writeU16(rc.NaptrPreference)
	b.writeString(rc.NaptrFlags)
	b.writeString(rc.NaptrService)
	b.writeString(rc.NaptrRegexp)
	b.WriteByte(0)
	return strings.ToUpper(hex.EncodeToString(b.Bytes()))
}

func (b *nabuf) writeU16(u uint16) {
	b.WriteByte(byte((u) & 0xff))
	b.WriteByte(byte((u >> 8) & 0xff))
}

func (b *nabuf) writeString(s string) {
	b.WriteByte(byte(len(s)))
	b.Write([]byte(s))
}

// Unpack

func populateFromHex(rc *models.RecordConfig, s string) error {
	msg, err := hex.DecodeString(s)
	if err != nil {
		return err
	}

	off := 0

	rc.NaptrOrder, off, err = unpackUint16(msg, off)
	if err != nil {
		return err
	}
	rc.NaptrPreference, off, err = unpackUint16(msg, off)
	if err != nil {
		return err
	}
	rc.NaptrFlags, off, err = unpackString(msg, off)
	if err != nil {
		return err
	}
	rc.NaptrService, off, err = unpackString(msg, off)
	if err != nil {
		return err
	}
	rc.NaptrRegexp, off, err = unpackString(msg, off)
	if err != nil {
		return err
	}
	// should find a 0-byte

	return nil
}

func unpackUint16(msg []byte, off int) (i uint16, off1 int, err error) {
	// Based on unpackUint16() in msg_helpers.go
	// (Changed BigEndian to LittleEndian)
	if off+2 > len(msg) {
		return 0, len(msg), fmt.Errorf("overflow unpacking uint16")
	}
	return binary.LittleEndian.Uint16(msg[off:]), off + 2, nil
}

//var order uint16
//var preference uint16
//var flags string
//var service string
//var regexp string

func unpackString(msg []byte, off int) (string, int, error) {
	// Based on unpackString() in msg_helpers.go
	if off+1 > len(msg) {
		return "", off, fmt.Errorf("overflow unpacking txt")
	}
	l := int(msg[off])
	off++
	if off+l > len(msg) {
		return "", off, fmt.Errorf("overflow unpacking txt")
	}
	var s strings.Builder
	consumed := 0
	for i, b := range msg[off : off+l] {
		switch {
		case b == '"' || b == '\\':
			if consumed == 0 {
				s.Grow(l * 2)
			}
			s.Write(msg[off+consumed : off+i])
			s.WriteByte('\\')
			s.WriteByte(b)
			consumed = i + 1
		case b < ' ' || b > '~': // unprintable
			if consumed == 0 {
				s.Grow(l * 2)
			}
			s.Write(msg[off+consumed : off+i])
			s.WriteString(escapeByte(b))
			consumed = i + 1
		}
	}
	if consumed == 0 { // no escaping needed
		return string(msg[off : off+l]), off + l, nil
	}
	s.Write(msg[off+consumed : off+l])
	return s.String(), off + l, nil
}

const (
	escapedByteSmall = "" +
		`\000\001\002\003\004\005\006\007\008\009` +
		`\010\011\012\013\014\015\016\017\018\019` +
		`\020\021\022\023\024\025\026\027\028\029` +
		`\030\031`
	escapedByteLarge = `\127\128\129` +
		`\130\131\132\133\134\135\136\137\138\139` +
		`\140\141\142\143\144\145\146\147\148\149` +
		`\150\151\152\153\154\155\156\157\158\159` +
		`\160\161\162\163\164\165\166\167\168\169` +
		`\170\171\172\173\174\175\176\177\178\179` +
		`\180\181\182\183\184\185\186\187\188\189` +
		`\190\191\192\193\194\195\196\197\198\199` +
		`\200\201\202\203\204\205\206\207\208\209` +
		`\210\211\212\213\214\215\216\217\218\219` +
		`\220\221\222\223\224\225\226\227\228\229` +
		`\230\231\232\233\234\235\236\237\238\239` +
		`\240\241\242\243\244\245\246\247\248\249` +
		`\250\251\252\253\254\255`
)

// escapeByte returns the \DDD escaping of b which must
// satisfy b < ' ' || b > '~'.
func escapeByte(b byte) string {
	// Based on escapeByte() in types.go
	if b < ' ' {
		return escapedByteSmall[b*4 : b*4+4]
	}

	b -= '~' + 1
	// The cast here is needed as b*4 may overflow byte.
	return escapedByteLarge[int(b)*4 : int(b)*4+4]
}
