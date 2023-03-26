package main

import "log"

const (
	langRefStartMarker  = "<!-- LANG_REF start -->\n"
	langRefEndMarker    = "<!-- LANG_REF end -->"
	providerStartMarker = "<!-- PROVIDER start -->\n"
	providerEndMarker   = "<!-- PROVIDER end -->"
)

func main() {
	if err := generateFeatureMatrix(); err != nil {
		log.Fatal(err)
	}
	funcs, err := generateFunctionTypes()
	if err != nil {
		log.Fatal(err)
	}
	if err := generateDTSFile(funcs); err != nil {
		log.Fatal(err)
	}
	if err := generateDocuTOC("documentation", "SUMMARY.md", "language_reference", langRefStartMarker, langRefEndMarker); err != nil {
		log.Print(err)
	}
	if err := generateDocuTOC("documentation", "SUMMARY.md", "service_providers", providerStartMarker, providerEndMarker); err != nil {
		log.Print(err)
	}
}
