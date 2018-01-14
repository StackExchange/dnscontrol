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

package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
	"github.com/softlayer/softlayer-go/sl"
)

func main() {
	sess := session.New() // default endpoint

	sess.Debug = true

	// List all Virtual Guests for an account
	//doListAccountVMsTest(sess)

	// Execute a remote script on a Virtual Guest
	//doExecuteRemoteScriptTest(sess)

	// Example: Provision and destroy a Virtual Guest
	//doCreateVMTest(sess)

	// Example: Get disk usage metrics by date
	//doGetDiskUsageMetricsTest(sess)

	// Example: Get the last bill date
	//doGetLatestBillDate(sess)

	// Demonstrate API Error
	//doError(sess)
}

func doListAccountVMsTest(sess *session.Session) {
	// Get the Account service
	service := services.GetAccountService(sess)

	// List VMs
	vms, err := service.Mask("id;hostname;domain").Limit(10).GetVirtualGuests()
	if err != nil {
		fmt.Printf("Error retrieving Virtual Guests from Account: %s\n", err)
		return
	} else {
		fmt.Println("VMs under Account:")
	}

	for _, vm := range vms {
		fmt.Printf("\t[%d]%s.%s\n", *vm.Id, *vm.Hostname, *vm.Domain)
	}
}

func doExecuteRemoteScriptTest(sess *session.Session) {
	// Get the VirtualGuest service
	service := services.GetVirtualGuestService(sess)

	// Execute the remote script
	err := service.Id(22870595).ExecuteRemoteScript(sl.String("http://example.com"))
	if err != nil {
		fmt.Println("Error executing remote script on VM:", err)
	} else {
		fmt.Println("Remote script sent for execution on VM")
	}
}

func doCreateVMTest(sess *session.Session) {
	service := services.GetVirtualGuestService(sess)

	// Create a Virtual_Guest instance as a template
	vGuestTemplate := datatypes.Virtual_Guest{
		Hostname:                     sl.String("sample"),
		Domain:                       sl.String("example.com"),
		MaxMemory:                    sl.Int(4096),
		StartCpus:                    sl.Int(1),
		Datacenter:                   &datatypes.Location{Name: sl.String("wdc01")},
		OperatingSystemReferenceCode: sl.String("UBUNTU_LATEST"),
		LocalDiskFlag:                sl.Bool(true),
		HourlyBillingFlag:            sl.Bool(true),
	}

	vGuest, err := service.Mask("id;domain").CreateObject(&vGuestTemplate)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	} else {
		fmt.Printf("\nNew Virtual Guest created with ID %d\n", *vGuest.Id)
		fmt.Printf("Domain: %s\n", *vGuest.Domain)
	}

	// Wait for transactions to finish
	fmt.Printf("Waiting for transactions to complete before destroying.")
	service = service.Id(*vGuest.Id)

	// Delay to allow transactions to be registered
	time.Sleep(10 * time.Second)

	for transactions, _ := service.GetActiveTransactions(); len(transactions) > 0; {
		fmt.Print(".")
		time.Sleep(10 * time.Second)
		transactions, err = service.GetActiveTransactions()
	}

	fmt.Println("Deleting virtual guest")

	success, err := service.DeleteObject()
	if err != nil {
		fmt.Printf("Error deleting virtual guest: %s", err)
	} else if success == false {
		fmt.Printf("Error deleting virtual guest")
	} else {
		fmt.Printf("Virtual Guest deleted successfully")
	}
}

func doGetDiskUsageMetricsTest(sess *session.Session) {
	service := services.GetAccountService(sess)

	tEnd := sl.Time(time.Now())
	tStart := sl.Time(tEnd.AddDate(0, -6, 0))

	data, err := service.GetDiskUsageMetricDataByDate(tStart, tEnd)
	if err != nil {
		fmt.Println("Error calling GetDiskUsageMetricDataByDate: ", err)
		return
	}

	fmt.Printf("Number of elements returned: %d\n", len(data))

	// Retrieve and print a DateTime (sl.Time) value
	if len(data) > 0 {
		fmt.Printf("item.DateTime = %s\n", data[0].DateTime)
	}
}

func doGetLatestBillDate(sess *session.Session) {
	service := services.GetAccountService(sess)

	d, _ := service.GetLatestBillDate()

	fmt.Printf("date of last bill: %s\n", d)
	fmt.Printf("type of date field: %s\n", reflect.TypeOf(d))
}

func handleError(err error) {
	apiErr := err.(sl.Error)
	fmt.Printf(
		"Exception: %s\nMessage: %s\nHTTP Status Code: %d\n",
		apiErr.Exception,
		apiErr.Message,
		apiErr.StatusCode)

	// Note that we could instead just dump the error, if we are not interested
	// in the individual fields
	//fmt.Println("Error:", err)
}

func doError(sess *session.Session) {
	service := services.GetVirtualGuestService(sess)

	// Example of an API error
	_, err := service.Id(0).GetObject() // invalid object ID
	if err != nil {
		handleError(err)
	}

	// Example of an HTTP, but non-API error
	sess.Endpoint = "http://example.com" // invalid endpoint
	_, err = service.GetObject()
	if err != nil {
		handleError(err)
	}

	// Example of a non-HTTP, non-API error
	sess.Endpoint = session.DefaultEndpoint
	var result struct {
		Id string `json:"id"` // type mismatch (unmarshal an integer value into a string)
	}
	err = sess.DoRequest("SoftLayer_Account", "getObject", nil, &sl.Options{}, &result)
	if err != nil {
		handleError(err)
	}
}
