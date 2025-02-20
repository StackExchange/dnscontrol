package main

import (
	"log"

	"github.com/fatih/structtag"
)

// MustParseTags parses a tag string. If the parsing fails, an empty Tags
// struct is returned and a warning is printed to stderr.
// This should only be used with static data that is known to be good.
func MustParseTags(s string) *structtag.Tags {
	tags, err := structtag.Parse(s)
	if err != nil {
		log.Printf("skipping unparsable tag string: %q %v", s, err)
		t, _ := structtag.Parse("")
		return t
	}

	return tags
}

// HasTagOption returns a bool indicating if a key+option exists.
// In the string `json:"name,foo"` "json" is the key, "foo"  is the option.
func HasTagOption(tags *structtag.Tags, key string, option string) bool {

	k, err := tags.Get(key)
	if err != nil {
		return false
	}

	if key == "dns" {
		// dns: has no name, just options. We check the name as if it is an option.
		if k.Name == option {
			return true
		}
	}

	return k.HasOption(option)
}
