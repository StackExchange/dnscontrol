/**
 * Copyright 2016 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package sl

import (
	"testing"

	"github.com/softlayer/softlayer-go/datatypes"
	"reflect"
)

func TestGet(t *testing.T) {
	// primitive test values
	testString := "Test"
	testInt := 123
	testFloat := 3.14159
	testBool := true

	// Pass by value
	if Get(testString) != "Test" {
		t.Errorf("Expected %s, got %s", testString, Get(testString))
	}

	if Get(testInt) != 123 {
		t.Errorf("Expected %d, got %d", testInt, Get(testInt))
	}

	if Get(testFloat) != 3.14159 {
		t.Errorf("Expected %f, got %f", testFloat, Get(testFloat))
	}

	if Get(testBool) != true {
		t.Errorf("Expected %t, got %t", testBool, Get(testBool))
	}

	// Pass by reference
	if Get(&testString) != "Test" {
		t.Errorf("Expected %s, got %s", testString, Get(&testString))
	}

	if Get(&testInt) != 123 {
		t.Errorf("Expected %d, got %d", testInt, Get(&testInt))
	}

	if Get(&testFloat) != 3.14159 {
		t.Errorf("Expected %f, got %f", testFloat, Get(&testFloat))
	}

	if Get(&testBool) != true {
		t.Errorf("Expected %t, got %t", testBool, Get(&testBool))
	}

	// zero value primitive
	var testZero int

	if Get(testZero) != 0 {
		t.Errorf("Expected %d, got %d", testZero, Get(testZero))
	}

	if Get(&testZero) != 0 {
		t.Errorf("Expected %d, got %d", testZero, Get(&testZero))
	}

	// nil pointer to primitive
	var testNil *int

	if Get(testNil) != 0 {
		t.Errorf("Expected %d, got %d", testNil, Get(testNil))
	}
}

func TestGetStruct(t *testing.T) {
	// Complex datatype test with
	// 1. non-nil primitive
	// 2. nil primitive (anything not set)
	// 3. non-nil ptr-to-struct type (Account)
	// 4. nil ptr-to-struct type (any unspecified ptr-to-struct field, e.g., Location)
	// 5. non-nil slice type (ActiveTickets)
	// 6. zero-val slice type (unspecified, e.g., ActiveTransactions)

	account := datatypes.Account{
		Id: Int(456),
	}

	tickets := []datatypes.Ticket{
		datatypes.Ticket{
			Id: Int(789),
		},
	}

	location := datatypes.Location{}

	var transactions []datatypes.Provisioning_Version1_Transaction

	d := datatypes.Virtual_Guest{
		Id:            Int(123),
		Account:       &account,
		ActiveTickets: tickets,
	}

	if Get(d.Id) != 123 {
		t.Errorf("Expected %d, got %d", *d.Id, Get(d.Id))
	}

	if !reflect.DeepEqual(Get(d.Account), account) {
		t.Errorf("Expected %v, got %v", account, Get(d.Account))
	}

	if !reflect.DeepEqual(Get(d.Location), location) {
		t.Errorf("Expected %v, got %v", location, Get(d.Location))
	}

	if !reflect.DeepEqual(Get(d.ActiveTickets), tickets) {
		t.Errorf("Expected %v, got %v", tickets, Get(d.ActiveTickets))
	}

	if !reflect.DeepEqual(Get(d.ActiveTransactions), transactions) {
		t.Errorf("Expected %#v, got %#v", transactions, Get(d.ActiveTransactions))
	}
}

func TestGetDefaults(t *testing.T) {
	var testptr *int

	if Get(testptr, 123) != 123 {
		t.Errorf("Expected 123, got %d", Get(testptr, 123))
	}
}
