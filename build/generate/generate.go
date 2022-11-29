package main

import "log"

func main() {
	// if err := generateFeatureMatrix(); err != nil {
	// 	log.Fatal(err)
	// }
	if err := generateTypes(); err != nil {
		log.Fatal(err)
	}
}
