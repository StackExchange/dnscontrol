package main

// Demo the utfutil ReadFile function.

import (
	"fmt"
	"log"
	"strings"

	"github.com/TomOnTime/utfutil"
)

func main() {
	data, err := utfutil.ReadFile("inputfile.txt", utfutil.HTML5)
	if err != nil {
		log.Fatal(err)
	}
	final := strings.Replace(string(data), "\r\n", "\n", -1)
	fmt.Println(final)
}
