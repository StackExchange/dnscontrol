// Copyright (c) 2018 Kai Schwarz (1API GmbH). All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// Package listresponse covers all functionality to handle an API response in list format, but as well provides access to the hash format
package listresponse

import (
	"github.com/hexonet/go-sdk/response/hashresponse"
)

// ListResponse class provides extra functionality to work with API responses.
// It provides methods that are useful for data representation in table format.
// In general the apiconnector Client always returns this type of response to be as flexible as possible.
type ListResponse struct {
	*hashresponse.HashResponse
	currentIndex int
	rows         [][]string
}

// NewListResponse represents the constructor for struct ListResponse
func NewListResponse(r string) *ListResponse {
	lr := &ListResponse{
		rows:         [][]string{},
		currentIndex: 0,
	}
	lr.HashResponse = hashresponse.NewHashResponse(r)
	rows := lr.rows
	h := lr.GetHash()
	cols := lr.GetColumnKeys()
	if lr.IsSuccess() && h["PROPERTY"] != nil {
		size := len(cols)
		cc := lr.Count()
		for i := 0; i < cc; i++ { //loop over amount of rows/indexes
			var row []string
			for c := 0; c < size; c++ { //loop over all columns
				colkey := cols[c]
				values := lr.GetColumn(colkey)
				if values != nil && len(values) > i {
					row = append(row, values[i])
				}
			}
			rows = append(rows, row)
		}
	}
	lr.rows = rows
	return lr
}

// GetList method to return the list of available rows
func (lr *ListResponse) GetList() [][]string {
	return lr.rows
}

// HasNext method to check if there's a further row after current row
func (lr *ListResponse) HasNext() bool {
	len := len(lr.rows)
	if len == 0 || lr.currentIndex+1 >= len {
		return false
	}
	return true
}

// Next method to access next row.
// Use HasNext method before.
func (lr *ListResponse) Next() []string {
	lr.currentIndex++
	return lr.rows[lr.currentIndex]
}

// HasPrevious method to check if there is a row available before current row.
func (lr *ListResponse) HasPrevious() bool {
	if lr.currentIndex == 0 {
		return false
	}
	return true
}

// Previous method to access previous row.
// Use HasPrevious method before.
func (lr *ListResponse) Previous() []string {
	lr.currentIndex--
	return lr.rows[lr.currentIndex]
}

// Current method to return current row
func (lr *ListResponse) Current() []string {
	if len(lr.rows) == 0 {
		return nil
	}
	return lr.rows[lr.currentIndex]
}

// Rewind method to reset the iterator index
func (lr *ListResponse) Rewind() {
	lr.currentIndex = 0
}
