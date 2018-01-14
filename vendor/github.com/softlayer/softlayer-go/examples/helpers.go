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

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/helpers/location"
	"github.com/softlayer/softlayer-go/helpers/order"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

func testHelpers() {
	sess := session.New() // default endpoint

	sess.Debug = true

	// Demonstrate order status helpers using a simulated product order receipt

	// First, get any valid order item ID
	items, err := services.GetAccountService(sess).
		Mask("orderItemId").
		GetNextInvoiceTopLevelBillingItems()

	// Create a receipt object to pass to the method
	receipt := datatypes.Container_Product_Order_Receipt{
		PlacedOrder: &datatypes.Billing_Order{
			Items: []datatypes.Billing_Order_Item{
				datatypes.Billing_Order_Item{
					Id: items[0].OrderItemId,
				},
			},
		},
	}

	complete, _, err := order.CheckBillingOrderStatus(sess, &receipt, []string{"COMPLETE", "PENDING"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Order in COMPLETE or PENDING status: %t\n", complete)
	}

	complete, _, err = order.CheckBillingOrderStatus(sess, &receipt, []string{"PENDING", "CANCELLED"})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Order in CANCELLED or PENDING status: %t\n", complete)
	}

	complete, _, err = order.CheckBillingOrderComplete(sess, &receipt)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Order is Complete: %t\n", complete)
	}

	// Demonstrate GetDataCenterByName

	l, err := location.GetDatacenterByName(sess, "ams01")
	fmt.Printf("Found Datacenter: %d\n", *l.Id)
}
