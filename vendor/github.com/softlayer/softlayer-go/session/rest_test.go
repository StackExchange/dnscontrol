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

package session

import (
	"testing"

	"fmt"
	"reflect"

	"github.com/jarcoal/httpmock"
	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"
	"github.com/softlayer/softlayer-go/tests"
)

var s *Session

const restEndpoint = "https://api.softlayer.com/rest/v3"

// structure of a single testcase
type testcase struct {
	description string
	service     string
	method      string
	args        []interface{}
	options     sl.Options
	responder   httpmock.Responder
	expected    interface{}
	expectError bool
}

var testcases = []testcase{
	// Positive tests
	{
		description: "empty array return",
		service:     "SoftLayer_Account",
		method:      "getVirtualGuests",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `[]`),
		expected:    make([]struct{}, 0),
		expectError: false,
	},
	{
		description: "empty struct return",
		service:     "SoftLayer_Account",
		method:      "getCurrentUser",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `{}`),
		expected:    struct{}{},
		expectError: false,
	},
	{
		description: "struct return",
		service:     "SoftLayer_User_Customer",
		method:      "getObject",
		args:        nil,
		options:     sl.Options{Id: sl.Int(12345)},
		responder:   httpmock.NewStringResponder(200, `{"id": 1, "username": "foo"}`),
		expected: struct {
			Id       int    `json:"id,omitempty"`
			Username string `json:"username,omitempty"`
		}{
			Id:       1,
			Username: "foo",
		},
		expectError: false,
	},
	{
		//special case for when SL response omits [] from a list with 1 item
		description: "list of structs return, missing []",
		service:     "SoftLayer_Account",
		method:      "getVirtualGuests",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `{"id": 1, "username": "foo"}`),
		expected: []struct {
			Id       int    `json:"id,omitempty"`
			Username string `json:"username,omitempty"`
		}{{Id: 1, Username: "foo"}},
		expectError: false,
	},
	{
		description: "array of struct return",
		service:     "SoftLayer_Account",
		method:      "getVirtualGuests",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `[{"id": 1, "username": "foo"},{"id": 2, "username": "bar"}]`),
		expected: []struct {
			Id       int    `json:"id,omitempty"`
			Username string `json:"username,omitempty"`
		}{{Id: 1, Username: "foo"}, {Id: 2, Username: "bar"}},
		expectError: false,
	},
	{
		description: "boolean return",
		service:     "SoftLayer_Virtual_Guest",
		method:      "deleteObject",
		args:        nil,
		options:     sl.Options{Id: sl.Int(12345)},
		responder:   httpmock.NewStringResponder(200, `true`),
		expected:    true,
		expectError: false,
	},
	{
		description: "string return",
		service:     "SoftLayer_Account",
		method:      "getAbuseEmail",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `admin@example.com`),
		expected:    "admin@example.com",
		expectError: false,
	},
	{
		description: "string return with quotes",
		service:     "SoftLayer_Account",
		method:      "getAbuseEmail",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `"admin@example.com"`),
		expected:    "admin@example.com",
		expectError: false,
	},
	{
		description: "JSON encoded string with escapes",
		service:     "SoftLayer_Network_Storage",
		method:      "getFileNetworkMountAddress",
		args:        nil,
		options:     sl.Options{Id: sl.Int(123456)},
		responder:   httpmock.NewStringResponder(200, `"fsf-dal0901b-fz.service.softlayer.com:\/IBM02SV123456_1\/data01"`),
		expected:    "fsf-dal0901b-fz.service.softlayer.com:/IBM02SV123456_1/data01",
		expectError: false,
	},
	{
		description: "uint return",
		service:     "SoftLayer_Account",
		method:      "getEvaultCapacityGB",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `256`),
		expected:    uint(256),
		expectError: false,
	},
	{
		description: "int return",
		service:     "SoftLayer_Account",
		method:      "getLargestAllowedSubnetCidr",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `16384`),
		expected:    int(16384),
		expectError: false,
	},
	{
		description: "negative int return",
		service:     "SoftLayer_Scale_Asset_Hardware",
		method:      "getHardwareId",
		args:        nil,
		options:     sl.Options{Id: sl.Int(123456)},
		responder:   httpmock.NewStringResponder(200, `-345`),
		expected:    int(-345),
		expectError: false,
	},
	{
		description: "[]byte return",
		service:     "SoftLayer_Account",
		method:      "getNextInvoicePdf",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewBytesResponder(200, []byte{0xDE, 0xAD, 0xBE, 0xEF}),
		expected:    []uint8{0xDE, 0xAD, 0xBE, 0xEF},
		expectError: false,
	},
	{
		description: "void return",
		service:     "SoftLayer_Virtual_Guest",
		method:      "executeRemoteScript",
		args:        nil,
		options:     sl.Options{Id: sl.Int(12345)},
		responder:   httpmock.NewStringResponder(200, "null"),
		expected:    datatypes.Void(0),
		expectError: false,
	},
	{
		description: "method arguments",
		service:     "SoftLayer_Virtual_Guest",
		method:      "executeRemoteScript",
		args: []interface{}{
			sl.String("https://example.com"),
		},
		options:     sl.Options{Id: sl.Int(12345)},
		responder:   tests.NewEchoResponder(200),
		expected:    `{"parameters":["https://example.com"]}`,
		expectError: false,
	},

	// Negative tests
	{
		description: "Error when boolean expected but null received",
		service:     "SoftLayer_Virtual_Guest",
		method:      "deleteObject",
		args:        nil,
		options:     sl.Options{Id: sl.Int(12345)},
		responder:   httpmock.NewStringResponder(200, `null`),
		expected:    true,
		expectError: true,
	},
	{
		description: "Error when uint expected but null received",
		service:     "SoftLayer_Account",
		method:      "getEvaultCapacityGB",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(200, `null`),
		expected:    uint(0),
		expectError: true,
	},
	{
		description: "HTTP !2xx",
		service:     "SoftLayer_Account",
		method:      "getEvaultCapacityGB",
		args:        nil,
		options:     sl.Options{},
		responder:   httpmock.NewStringResponder(400, `1`),
		expected:    uint(1),
		expectError: true,
	},
}

func TestRest(t *testing.T) {
	// setup session and mock environment
	s = New()
	s.Endpoint = restEndpoint

	//s.Debug = true
	httpmock.Activate()
	defer httpmock.Deactivate()

	// For each test case:
	// 1. Setup the mock environment
	// 2. Allocate an empty variable for the response, based on the 'expected' response
	// 3. Call DoRequest
	// 4. Check result matches expected
	// 5. Reset the mock
	for _, tc := range testcases {
		setup(tc)

		pResult := reflect.New(reflect.TypeOf(tc.expected)).Interface()

		fmt.Printf("Test [%s]: ", tc.description)
		err := s.DoRequest(tc.service, tc.method, tc.args, &tc.options, pResult)

		// Report results
		switch {

		// Positive tests - no error expected
		case !tc.expectError && err != nil:
			fmt.Println("Unexpected error:", err.Error())
			t.Fail()
		case !tc.expectError && err == nil:
			result := reflect.Indirect(reflect.ValueOf(pResult)).Interface()
			if !reflect.DeepEqual(tc.expected, result) {
				fmt.Println("FAIL")
				t.Errorf("Expected %#v, got %#v", tc.expected, result)
			} else {
				fmt.Println("OK")
			}

		// Negative tests - error expected
		case tc.expectError && err == nil:
			fmt.Println("FAIL")
			t.Errorf("Expected error not received")
		case tc.expectError && err != nil:
			fmt.Println("OK")
		}

		teardown()
	}
}

func setup(tc testcase) {
	httpmock.RegisterResponder(
		httpMethod(tc.method, tc.args),
		fmt.Sprintf("%s/%s", restEndpoint, buildPath(tc.service, tc.method, &tc.options)),
		tc.responder)
}

// remove any existig mocks (e.g., between tests)
func teardown() {
	httpmock.Reset()
}
