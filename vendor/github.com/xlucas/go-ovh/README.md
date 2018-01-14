# go-ovh
A simple helper library around the OVH API for golang developers.

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/xlucas/go-ovh/ovh)


## Requirements
Firsteval, you will need to generate the application credentials in order to use the API. Check [the official guide here](https://api.ovh.com/g934.test) for more details.

## Usage example

```go

import "github.com/xlucas/go-ovh/ovh"

type RequestStruct struct {
    MyRequestField     string
}

type ResponseStruct struct {
    MyResponseField    uint
}

c := ovh.Client(ovh.ENDPOINT_EU_OVHCOM, "MyAppKey", "MyAppSecretKey", "MyConsumerKey")

// Check for time lag
if err := c.PollTimeshift(); err != nil {
    log.Fatal("Failed to retrieve timeshift, reason : ", err)
}

var response ResponseStruct

request = RequestStruct {
    MyRequestField: "foo",
}


// Send our request
if err := c.Call("POST", "/cloud/createProject", request, &response); err != nil {
    log.Fatal("Failed to call the API, reason : ", err)
}

// Access the response struct
fmt.Printf("My response field is %d", response.MyResponseField)
...
```
