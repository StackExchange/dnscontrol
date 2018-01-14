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
	"github.com/softlayer/softlayer-go/filter"
)

func testFilters() {
	fmt.Println(filter.Path("virtualGuests.hostname").Eq("example.com").Build())

	fmt.Println(
		filter.New(
			filter.Path("id").Eq("134"),
			filter.Path("datacenter.locationName").Eq("Dallas"),
			filter.Path("something.creationDate").Date("01/01/01"),
		).Build(),
	)

	fmt.Println(
		filter.Build(
			filter.Path("virtualGuests.domain").Eq("example.com"),
			filter.Path("virtualGuests.id").NotEq(12345),
		),
	)

	filters := filter.New(
		filter.Path("virtualGuests.hostname").StartsWith("KM078"),
		filter.Path("virtualGuests.id").NotEq(12345),
	)

	filters = append(filters, filter.Path("virtualGuests.domain").Eq("example.com"))

	fmt.Println(filters.Build())
}
