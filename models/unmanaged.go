package models

import (
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

// func NewUnmanagedConfig(label, rtype, target string) (*UnmanagedConfig, error) {
// 	var err error

// 	result := &UnmanagedConfig{
// 		LabelPattern:  label,
// 		RTypePattern:  rtype,
// 		TargetPattern: target,
// 	}

// 	result.LabelGlob, err = glob.Compile(label)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if rtype != "*" && rtype != "" {
// 		for _, part := range strings.Split(rtype, ",") {
// 			part = strings.TrimSpace(part)
// 			result.RTypeMap[part] = struct{}{}
// 		}
// 	}

// 	result.TargetGlob, err = glob.Compile(target)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil
// }
