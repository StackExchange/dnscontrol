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

package tests

import (
	"testing"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/services"
	"reflect"
)

// Tests for each service/method that follows special case logic during code
// generation

func TestVirtualDiskImageType(t *testing.T) {
	expected := "*datatypes.Virtual_Disk_Image_Type"
	actual := reflect.TypeOf(datatypes.Virtual_Guest_Block_Device_Template_Group{}.ImageType).String()

	if actual != expected {
		t.Errorf("Expect type of Virtual_Guest_Block_Device_Template_Group.ImageType to be %s, but was %s", expected, actual)
	}
}

func TestDomainResourceRecordPropertiesCopiedToBaseType(t *testing.T) {
	fields := []string{"IsGatewayAddress", "Port", "Priority", "Protocol", "Service", "Weight"}
	for _, field := range fields {
		if _, ok := reflect.TypeOf(datatypes.Dns_Domain_ResourceRecord{}).FieldByName(field); !ok {
			t.Errorf("Expect property %s not found for datatypes.Dns_Domain_ResourceRecord", field)
		}
	}
}

func TestVoidPatchedReturnTypes(t *testing.T) {
	tests := map[interface{}]string{
		services.Network_Application_Delivery_Controller_LoadBalancer_Service{}:       "DeleteObject",
		services.Network_Application_Delivery_Controller_LoadBalancer_VirtualServer{}: "DeleteObject",
		services.Network_Application_Delivery_Controller{}:                            "DeleteLiveLoadBalancerService",
	}

	for service, method := range tests {
		reflectedService := reflect.TypeOf(service)
		reflectedMethod, _ := reflectedService.MethodByName(method)
		if reflectedMethod.Type.NumOut() > 1 {
			t.Errorf("Expect %s() to have only one (error) return value, but multiple values found", reflectedService.String()+method)
		}
	}
}

func TestPlaceOrder(t *testing.T) {
	services := []interface{}{
		services.Billing_Order_Quote{},
		services.Product_Order{},
	}

	for _, service := range services {
		serviceType := reflect.TypeOf(service)
		method, _ := serviceType.MethodByName("PlaceOrder")
		argType := method.Type.In(1).String()

		if argType != "interface {}" {
			t.Errorf(
				"Expect %s.PlaceOrder() to accept interface {} as parameter, but %s found instead",
				serviceType.String(),
				argType,
			)
		}
	}
}
