JsonConfigReader is a proxy for [golang's io.Reader](http://golang.org/pkg/io/#Reader) that strips line comments and trailing commas, allowing you to use json as a *reasonable* config format.

Comments start with `//` and continue to the end of the line.

If a trailing comma is in front of `]` or `}` it will be stripped as well.


Given `settings.json`

	{
		"key": "value", // k:v
		
		// a list of numbers
		"list": [1, 2, 3],
	}


You can read it in as a *normal* json file:

	package main

	import (
		"encoding/json"
		"fmt"
		"github.com/DisposaBoy/JsonConfigReader"
		"os"
	)

	func main() {
		var v interface{}
		f, _ := os.Open("settings.json")
		// wrap our reader before passing it to the json decoder
		r := JsonConfigReader.New(f)
		json.NewDecoder(r).Decode(&v)
		fmt.Println(v)
	}
