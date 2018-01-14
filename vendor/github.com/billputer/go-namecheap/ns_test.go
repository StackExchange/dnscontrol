package namecheap

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestNSGetInfo(t *testing.T) {
	setup()
	defer teardown()

	respXML := `
<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse xmlns="http://api.namecheap.com/xml.response" Status="OK">
  <Errors />
  <RequestedCommand>namecheap.domains.ns.getInfo</RequestedCommand>
  <CommandResponse Type="namecheap.domains.ns.getInfo">
    <DomainNSInfoResult Domain="domain.com" Nameserver="ns1.domain.com" IP="12.23.23.23">
      <NameserverStatuses>
        <Status>OK</Status>
        <Status>Linked</Status>
      </NameserverStatuses>
    </DomainNSInfoResult>
  </CommandResponse>
  <Server>SERVER-NAME</Server>
  <GMTTimeDifference>+5</GMTTimeDifference>
  <ExecutionTime>32.76</ExecutionTime>
</ApiResponse>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		correctURL := "/?ApiKey=anToken&ApiUser=anApiUser&ClientIp=127.0.0.1&Command=namecheap.domains.ns.getInfo&Nameserver=ns1.domain.com&SLD=domain&TLD=com&UserName=anUser"
		if r.URL.String() != correctURL {
			t.Errorf("URL = %v, want %v", r.URL, correctURL)
		}
		testMethod(t, r, "GET")
		fmt.Fprint(w, respXML)
	})

	ns, err := client.NSGetInfo("domain", "com", "ns1.domain.com")
	if err != nil {
		t.Errorf("NSGetInfo returned error: %v", err)
	}
	want := &DomainNSInfoResult{
		Domain:     "domain.com",
		Nameserver: "ns1.domain.com",
		IP:         "12.23.23.23",
		Statuses:   []string{"OK", "Linked"},
	}
	if !reflect.DeepEqual(ns, want) {
		t.Errorf("NSGetInfo returned %+v, want %+v", ns, want)
	}
}
