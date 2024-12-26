package main

/*
This simple utility reads a TYPES.yml file and outputs the post-processed file. This is a debugging tool.

Run:
  go run *.go ../../../../rtypes/TYPES.yml
*/

import (
	"fmt"
	"os"

	"github.com/StackExchange/dnscontrol/v4/pkg/rtypectl"
	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rtypectl <filename>")
		os.Exit(1)
	}
	arg := os.Args[1]

	db, err := rtypectl.New(arg)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	yamlData, err := yaml.Marshal(db)
	if err != nil {
		fmt.Printf("Error marshaling to YAML: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(yamlData))
}
