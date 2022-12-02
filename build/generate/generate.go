package main

import "log"

func main() {
	if err := generateFeatureMatrix(); err != nil {
		log.Fatal(err)
	}
	if err := generateFunctionTypes(); err != nil {
		log.Fatal(err)
	}
	if err := combineTypes(); err != nil {
		log.Fatal(err)
	}
}
