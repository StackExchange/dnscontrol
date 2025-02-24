package rfc1035

import (
	"encoding/csv"
	"fmt"
	"strings"
)

func Fields(s string) ([]string, error) {

	r := csv.NewReader(strings.NewReader(s))
	r.Comma = ' '

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) != 1 {
		return nil, fmt.Errorf("string contains more than one record: %q", s)
	}

	for j, field := range records[0] {
		if len(field) >= 2 && field[0] == '"' && field[len(field)-1] == '"' {
			records[0][j] = field[1 : len(field)-1]
		}
	}

	return records[0], nil
}
