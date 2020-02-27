// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package column provides column functionality to cover API response data
package column

import "errors"

// Column is a struct representing column covering API response data.
type Column struct {
	Length int
	key    string
	data   []string
}

// NewColumn represents the constructor for struct Column.
func NewColumn(key string, data []string) *Column {
	sc := &Column{
		Length: len(data),
		key:    key,
		data:   data,
	}
	return sc
}

// GetKey method to return the column name
func (c *Column) GetKey() string {
	return c.key
}

// GetData method to return the column data
func (c *Column) GetData() []string {
	return c.data
}

// GetDataByIndex method to return the column data at the provided index
func (c *Column) GetDataByIndex(idx int) (string, error) {
	if c.hasDataIndex(idx) {
		return c.data[idx], nil
	}
	return "", errors.New("Index not found")
}

// hasDataIndex method to check if the given data index exists
func (c *Column) hasDataIndex(idx int) bool {
	return (idx >= 0 && idx < c.Length)
}
