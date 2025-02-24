package models

import "fmt"

func BuilderCAA(rawfields []string, meta map[string]string, origin string) ([]string, map[string]string, error) {
	// To be compatible with existing configurations, this takes CAA(label, tag, value, { flags }) and returns CAA(label, tag, flag, value).

	if len(rawfields) < 3 {
		return rawfields, meta, fmt.Errorf("CAA record must have at least 3 fields")
	}

	fmt.Printf("DEBUG: BuilderCAA: rawfields=%v meta=%+v\n", rawfields, meta)

	flag := "0"
	if meta["caa_critical"] != "" {
		flag = "128"
		delete(meta, "caa_critical")
	}

	rawfields = []string{
		rawfields[0], // Label
		flag,         // Flag
		rawfields[1], // Tag
		rawfields[2], // Value
	}

	return rawfields, meta, nil
}
