package main

import (
	"log"
	"os"

	"github.com/mjibson/esc/embed"
)

func main() {
	conf := &embed.Config{
		ModTime:    "0",
		OutputFile: "pkg/js/static.go",
		Package:    "js",
		Prefix:     "pkg/js",
		Private:    true,
		Files:      []string{`pkg/js/helpers.js`},
	}

	var err error
	out := os.Stdout
	if conf.OutputFile != "" {
		if out, err = os.Create(conf.OutputFile); err != nil {
			log.Fatal(err)
		}
		defer out.Close()
	}

	embed.Run(conf, out)

	if err := generateFeatureMatrix(); err != nil {
		log.Fatal(err)
	}
}
