package models

import (
	// "crypto/sha256"
	"encoding/base64"
	// "encoding/hex"
	"errors"
	"regexp"
)

const (
	opgpktail = "._openpgpkey"
)

func (rc *RecordConfig) GetOpenPGPKeyField() string {
	// nameLen := len(rc.Name)
	// opgpktail := "._openpgpkey"

	// // fmt.Println("GetOpenPGPKeyField()")
	// if rc.Name[nameLen-1:] == "@" {
	// 	rc.Name = buildSha256Prefix(rc.Name[:nameLen-1])
	// }

	// tailLen := len(opgpktail)
	// nameLen = len(rc.Name)
	// if (nameLen - tailLen >= 0) {
	// 	if rc.Name[nameLen-tailLen:] != opgpktail {
	// 		rc.Name += opgpktail
	// 	}
	// }

	keyField := rc.OpenPgpKeyPublicKey
	// get the record text, minus any enclosing ()
	var re = regexp.MustCompile(`\(?\s{1,}?([^\)]+)\s{1,}?\)?`)
	str := re.ReplaceAllString(keyField, "$1")

	// verify base64
	if _, err := base64.StdEncoding.DecodeString(str); err != nil {
		panic("SetTargetOpenPGPKey: Error decoding base64: " + err.Error())
	}

	// this is subtle: return rc.target and you can create multiple sub records
	// return rc.OpenPgpKeyPublicKey and you can only have one sub record for a hash
	return rc.target
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetTargetOpenPGPKey sets the OPENPGPKEY fields.
func (rc *RecordConfig) SetTargetOpenPGPKey(target string) error {
	// nameLen := len(rc.Name)
	// opgpktail := "._openpgpkey"

	// if rc.Name[nameLen-1:] == "@" {
	// 	rc.Name = buildSha256Prefix(rc.Name[:nameLen-1])
	// }

	// tailLen := len(opgpktail)
	// nameLen = len(rc.Name)

	// if (nameLen - tailLen >= 0) {
	// 	if rc.Name[nameLen-tailLen:] != opgpktail {
	// 		rc.Name += opgpktail
	// 	}
	// }

	// if len(target) > 0 {
	// 	if target[0:1] == "(" && target[len(target)-1:] == ")" {
	// 		target = strings.TrimSpace(target[1:len(target)-1])
	// 	}
	// }

	// get the record text, minus any enclosing ()
	var re = regexp.MustCompile(`\(?\s{1,}?([^\)]+)\s{1,}?\)?`)
	str := re.ReplaceAllString(target, "$1")

	// verify base64
	if _, err := base64.StdEncoding.DecodeString(str); err != nil {
		return errors.New("SetTargetOpenPGPKey: Error decoding base64: " + err.Error())
	}

	// if the key base64 length is longer than 256, enclose it in ()
	var tgtstr string
	if len(target) > 256 {
		tgtstr = "( "
		for i := 0; i < len(target); i += 64 {
			// Get the next 64-byte chunk of the string suffixed with newline
			tgtstr += target[i:min(i+64, len(target))] + "\n"
		}
		tgtstr += " )"
	}

	rc.SetTarget(tgtstr)

	rc.OpenPgpKeyPublicKey = str

	if rc.Type == "" {
		rc.Type = "OPENPGPKEY"
	}
	if rc.Type != "OPENPGPKEY" {
		panic("assertion failed: SetTargetOpenPGPKey called when .Type is not OPENPGPKEY")
	}

	return nil
}

// // build sha256
// func buildSha256Prefix(input string) string {
// 	// create a new SHA256 hash object
// 	hasher := sha256.New()

// 	// write the input string to the hash object
// 	hasher.Write([]byte(input))

// 	// get the resulting hash as a byte slice
// 	hashBytes := hasher.Sum(nil)

// 	// convert the hash to a hex-encoded string
// 	hashString := hex.EncodeToString(hashBytes)

// 	// fmt.Println(hashString)

// 	return hashString[:56]
// }
