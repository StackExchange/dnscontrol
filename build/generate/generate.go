package main

import (
	"github.com/mjibson/esc/embed"
)

func main() {
	//go:generate esc -modtime 0 -o js/static.go -pkg js -include helpers\.js -ignore go -prefix js js
	conf := &embed.Config{
		ModTime:    "0",
		OutputFile: "js/static.go",
		Package:    "js",
		Prefix:     "js",
		Private:    true,
		Files:      []string{`js/helpers.js`},
	}
	embed.Run(conf)
}
