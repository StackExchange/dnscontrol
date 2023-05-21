package models

import (
	"bytes"
	"fmt"

	"github.com/gobwas/glob"
)

// UnmanagedConfig describes an UNMANAGED() rule.
type UnmanagedConfig struct {
	// Glob pattern for matching labels.
	LabelPattern string    `json:"label_pattern,omitempty"`
	LabelGlob    glob.Glob `json:"-"` // Compiled version

	// Comma-separated list of DNS Resource Types.
	RTypePattern string              `json:"rType_pattern,omitempty"`
	RTypeMap     map[string]struct{} `json:"-"` // map of RTypes or len()=0 for all

	// Glob pattern for matching targets.
	TargetPattern string    `json:"target_pattern,omitempty"`
	TargetGlob    glob.Glob `json:"-"` // Compiled version
}

func DebugUnmanagedConfig(uc []*UnmanagedConfig) string {
	if len(uc) == 0 {
		return "UnmanagedConfig{}"
	}

	var buf bytes.Buffer
	b := &buf

	fmt.Fprint(b, "UnmanagedConfig{\n")
	for i, c := range uc {
		fmt.Fprintf(b, "%00d: (%v, %+v, %v)\n",
			i,
			c.LabelGlob,
			c.RTypeMap,
			c.TargetGlob,
		)
	}
	fmt.Fprint(b, "}")

	return b.String()
}
