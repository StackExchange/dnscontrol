package main

import (
	"os"
	"strings"
)

func combineTypes() error {
	names := []string{
		"base-types",
		"fetch",
		"functions",
		"others",
	}

	combined := []string{}
	for _, name := range names {
		content, err := os.ReadFile(join("types", "src", name+".d.ts"))
		if err != nil {
			return err
		}
		combined = append(combined, string(content))
	}
	os.WriteFile(join("types", "dnscontrol.d.ts"), []byte(strings.Join(combined, "\n\n")), 0644)
	return nil
}
