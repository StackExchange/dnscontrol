// Copyright (c) 2018 Kai Schwarz (HEXONET GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package record provides record functionality to cover API response data
package record

import "errors"

// Record is a struct representing record/row covering API response data.
type Record struct {
	data map[string]string
}

// NewRecord represents the constructor for struct Column.
func NewRecord(data map[string]string) *Record {
	r := &Record{
		data: data,
	}
	return r
}

// GetData method to return the column data
func (c *Record) GetData() map[string]string {
	return c.data
}

// GetDataByKey method to return the column data at the provided index
func (c *Record) GetDataByKey(key string) (string, error) {
	if c.hasData(key) {
		return c.data[key], nil
	}
	return "", errors.New("column name not found in record")
}

// hasData method to check if the given data index exists
func (c *Record) hasData(key string) bool {
	if _, ok := c.data[key]; ok {
		return true
	}
	return false
}
