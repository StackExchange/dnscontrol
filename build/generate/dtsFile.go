package main

import (
	"fmt"
	"os"
	"path"
	"strings"
)

var commentStart = "@dnscontrol-auto-doc-comment "

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
		content, err := os.ReadFile(join("commands", "types", name+".d.ts"))
		if err != nil {
			return err
		}
		// Find all instances of `/** @dnscontrol-auto-doc-comment <path> */`
		// and replace them with the contents of the file at <path>.
		// This allows us to keep the documentation in the same file as the code.
		for {
			start := strings.Index(string(content), commentStart)
			if start == -1 {
				break
			}
			end := strings.Index(string(content[start:]), "\n")
			if end == -1 {
				return fmt.Errorf("unterminated @dnscontrol-auto-doc-comment in '%s'", name)
			}

			docPath := string(content[start+len(commentStart) : start+end])
			println("Replacing", docPath)

			if strings.Contains(docPath, "..") {
				return fmt.Errorf("invalid path '%s' in '%s'", docPath, name)
			}

			newPath := path.Clean(join("documentation", docPath))
			if !strings.HasPrefix(newPath, "documentation") {
				return fmt.Errorf("invalid path '%s' in '%s'", docPath, name)
			}
			_, body, err := readDocFile(newPath)
			if err != nil {
				return err
			}

			body = strings.ReplaceAll(strings.Trim(body, "\n"), "\n", "\n * ")

			content = append(content[:start], append([]byte(body), content[start+end:]...)...)
		}

		combined = append(combined, string(content))
	}
	combined = append(combined, funcs)
	fileContent := strings.Join(combined, "\n\n")
	lines := strings.Split(fileContent, "\n")
	fileContent = ""
	for _, line := range lines {
		fileContent += strings.TrimRight(line, " \t") + "\n"
	}
	fileContent = strings.TrimRight(fileContent, "\n")
	os.WriteFile(join("commands", "types", "dnscontrol.d.ts"), []byte(fileContent+"\n"), 0644)
	return nil
}
