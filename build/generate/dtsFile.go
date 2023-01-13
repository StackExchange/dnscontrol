package main

import (
	"os"
	"strings"
)

func generateDTSFile(funcs string) error {
	names := []string{
		"base-types",
		"fetch",
		"others",
	}

	combined := []string{
		"// WARNING: These type definitions are experimental and subject to change in future releases.",
	}
	for _, name := range names {
		content, err := os.ReadFile(join("types", name+".d.ts"))
		if err != nil {
			return err
		}
		combined = append(combined, string(content))
	}
	combined = append(combined, funcs)
	os.WriteFile(join("types", "dnscontrol.d.ts"), []byte(strings.Join(combined, "\n\n")), 0644)
	return nil
}
