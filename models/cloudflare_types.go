package models

import (
	"strconv"
)

func init() {
	MustRegisterType("CF_SINGLE_REDIRECT", RegisterOpts{PopulateFromRaw: PopulateFromRawCFSINGLEREDIRECT})
}

// CFSINGLEREDIRECT

func NewFromRawCFSINGLEREDIRECT(rawfields []string, meta map[string]string, origin string, ttl uint32) (*RecordConfig, error) {
	rc := &RecordConfig{TTL: ttl}
	if err := PopulateFromRawCFSINGLEREDIRECT(rc, rawfields, meta, origin); err != nil {
		return nil, err
	}
	return rc, nil
}

// GetCFSINGLEREDIRECTFields returns rc.Fields as individual typed values.
func (rc *RecordConfig) GetCFSINGLEREDIRECTFields() (string, uint16, string, string) {
	n := rc.AsCFSINGLEREDIRECT()
	return n.SRName, n.Code, (n.SRWhen), (n.SRThen)
}

// GetCFSINGLEREDIRECTStrings returns rc.Fields as individual strings.
func (rc *RecordConfig) GetCFSINGLEREDIRECTStrings() [4]string {
	n := rc.AsCFSINGLEREDIRECT()
	return [4]string{n.SRName, strconv.Itoa(int(n.Code)), n.SRWhen, n.SRThen}
}
