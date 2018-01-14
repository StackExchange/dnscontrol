package namecheap

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestDomainsGetList(t *testing.T) {
	setup()
	defer teardown()

	respXML := `
    <?xml version="1.0" encoding="utf-8"?>
    <ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
      <Errors />
      <Warnings />
      <RequestedCommand>namecheap.domains.getList</RequestedCommand>
      <CommandResponse Type="namecheap.domains.getList">
        <DomainGetListResult>
          <Domain ID="57579" Name="example.com" User="anUser" Created="11/04/2014" Expires="11/04/2015" IsExpired="false" IsLocked="false" AutoRenew="false" WhoisGuard="ENABLED" />
        </DomainGetListResult>
        <Paging>
          <TotalItems>12</TotalItems>
          <CurrentPage>1</CurrentPage>
          <PageSize>20</PageSize>
        </Paging>
      </CommandResponse>
      <Server>WEB1-SANDBOX1</Server>
      <GMTTimeDifference>--5:00</GMTTimeDifference>
      <ExecutionTime>0.009</ExecutionTime>
    </ApiResponse>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// verify that the URL exactly matches...brittle, I know.
		correctURL := "/?ApiKey=anToken&ApiUser=anApiUser&ClientIp=127.0.0.1&Command=namecheap.domains.getList&UserName=anUser"
		if r.URL.String() != correctURL {
			t.Errorf("URL = %v, want %v", r.URL, correctURL)
		}
		testMethod(t, r, "GET")
		fmt.Fprint(w, respXML)
	})

	domains, err := client.DomainsGetList()

	if err != nil {
		t.Errorf("DomainsGetList returned error: %v", err)
	}

	// DomainGetListResult we expect, given the respXML above
	want := []DomainGetListResult{{
		ID:         57579,
		Name:       "example.com",
		User:       "anUser",
		Created:    "11/04/2014",
		Expires:    "11/04/2015",
		IsExpired:  false,
		IsLocked:   false,
		AutoRenew:  false,
		WhoisGuard: "ENABLED",
	}}

	if !reflect.DeepEqual(domains, want) {
		t.Errorf("DomainsGetList returned %+v, want %+v", domains, want)
	}
}

func TestDomainGetInfo(t *testing.T) {
	setup()
	defer teardown()

	respXML := `<?xml version="1.0" encoding="utf-8"?>
<ApiResponse Status="OK" xmlns="http://api.namecheap.com/xml.response">
  <Errors />
  <Warnings />
  <RequestedCommand>namecheap.domains.getInfo</RequestedCommand>
  <CommandResponse Type="namecheap.domains.getInfo">
    <DomainGetInfoResult Status="Ok" ID="57582" DomainName="example.com" OwnerName="anUser" IsOwner="true">
      <DomainDetails>
        <CreatedDate>11/04/2014</CreatedDate>
        <ExpiredDate>11/04/2015</ExpiredDate>
        <NumYears>0</NumYears>
      </DomainDetails>
      <LockDetails />
      <Whoisguard Enabled="True">
        <ID>53536</ID>
        <ExpiredDate>11/04/2015</ExpiredDate>
        <EmailDetails WhoisGuardEmail="08040e11d32d48ebb4346b02b98dda17.protect@whoisguard.com" ForwardedTo="billwiens@gmail.com" LastAutoEmailChangeDate="" AutoEmailChangeFrequencyDays="0" />
      </Whoisguard>
      <DnsDetails ProviderType="FREE" IsUsingOurDNS="true">
        <Nameserver>dns1.registrar-servers.com</Nameserver>
        <Nameserver>dns2.registrar-servers.com</Nameserver>
        <Nameserver>dns3.registrar-servers.com</Nameserver>
        <Nameserver>dns4.registrar-servers.com</Nameserver>
        <Nameserver>dns5.registrar-servers.com</Nameserver>
      </DnsDetails>
      <Modificationrights All="true" />
    </DomainGetInfoResult>
  </CommandResponse>
  <Server>WEB1-SANDBOX1</Server>
  <GMTTimeDifference>--5:00</GMTTimeDifference>
  <ExecutionTime>0.008</ExecutionTime>
</ApiResponse>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// verify that the URL exactly matches...brittle, I know.
		correctURL := "/?ApiKey=anToken&ApiUser=anApiUser&ClientIp=127.0.0.1&Command=namecheap.domains.getInfo&DomainName=example.com&UserName=anUser"
		if r.URL.String() != correctURL {
			t.Errorf("URL = %v, want %v", r.URL, correctURL)
		}
		testMethod(t, r, "GET")
		fmt.Fprint(w, respXML)
	})

	domain, err := client.DomainGetInfo("example.com")

	if err != nil {
		t.Errorf("DomainGetInfo returned error: %v", err)
	}

	// DomainGetListResult we expect, given the respXML above
	want := &DomainInfo{
		ID:        57582,
		Name:      "example.com",
		Owner:     "anUser",
		Created:   "11/04/2014",
		Expires:   "11/04/2015",
		IsExpired: false,
		IsLocked:  false,
		AutoRenew: false,
		DNSDetails: DNSDetails{
			ProviderType:  "FREE",
			IsUsingOurDNS: true,
			Nameservers: []string{
				"dns1.registrar-servers.com",
				"dns2.registrar-servers.com",
				"dns3.registrar-servers.com",
				"dns4.registrar-servers.com",
				"dns5.registrar-servers.com",
			},
		},
	}

	if !reflect.DeepEqual(domain, want) {
		t.Errorf("DomainGetInfo returned %+v, want %+v", domain, want)
	}
}

func TestDomainsCheck(t *testing.T) {
	setup()
	defer teardown()

	respXML := `<?xml version="1.0" encoding="UTF-8"?>
<ApiResponse xmlns="http://api.namecheap.com/xml.response" Status="OK">
  <Errors />
  <RequestedCommand>namecheap.domains.check</RequestedCommand>
  <CommandResponse Type="namecheap.domains.check">
    <DomainCheckResult Domain="domain1.com" Available="true" />
    <DomainCheckResult Domain="availabledomain.com" Available="false" />
  </CommandResponse>
  <Server>SERVER-NAME</Server>
  <GMTTimeDifference>+5</GMTTimeDifference>
  <ExecutionTime>32.76</ExecutionTime>
</ApiResponse>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// verify that the URL exactly matches...brittle, I know.
		correctURL := "/?ApiKey=anToken&ApiUser=anApiUser&ClientIp=127.0.0.1&Command=namecheap.domains.check&DomainList=domain1.com,availabledomain.com&UserName=anUser"
		if r.URL.String() != correctURL {
			t.Errorf("URL = %v, want %v", r.URL, correctURL)
		}
		testMethod(t, r, "GET")
		fmt.Fprint(w, respXML)
	})

	domain, err := client.DomainsCheck("domain1.com", "availabledomain.com")
	if err != nil {
		t.Errorf("DomainsCheck returned error: %v", err)
	}

	// DomainGetListResult we expect, given the respXML above
	want := []DomainCheckResult{
		DomainCheckResult{
			Domain:    "domain1.com",
			Available: true,
		},
		DomainCheckResult{
			Domain:    "availabledomain.com",
			Available: false,
		},
	}

	if !reflect.DeepEqual(domain, want) {
		t.Errorf("DomainsCheck returned %+v, want %+v", domain, want)
	}
}

func TestDomainCreate(t *testing.T) {
	setup()
	defer teardown()

	respXML := `<?xml version="1.0" encoding="UTF-8"?>
	<ApiResponse xmlns="http://api.namecheap.com/xml.response" Status="OK">
	  <Errors />
	  <RequestedCommand>namecheap.domains.create</RequestedCommand>
	  <CommandResponse Type="namecheap.domains.create">
	    <DomainCreateResult Domain="domain1.com" Registered="true" ChargedAmount="20.3600" DomainID="9007" OrderID="196074" TransactionID="380716" WhoisguardEnable="false" NonRealTimeDomain="false" />
	  </CommandResponse>
	  <Server>SERVER-NAME</Server>
	  <GMTTimeDifference>+5</GMTTimeDifference>
	  <ExecutionTime>0.078</ExecutionTime>
	</ApiResponse>`

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// verify that the URL exactly matches...brittle, I know.
		correctURL := "/?AdminAddress1=8939%20S.cross%20Blvd&ApiUser=anApiUser&ApiKey=anToken&UserName=anUser&Command=namecheap.domains.create&ClientIp=127.0.0.1&DomainName=domain1.com&Years=2&AuxBillingFirstName=John&AuxBillingLastName=Smith&AuxBillingAddress1=8939%20S.cross%20Blvd&AuxBillingStateProvince=CA&AuxBillingPostalCode=90045&AuxBillingCountry=US&AuxBillingPhone=+1.6613102107&AuxBillingEmailAddress=john@gmail.com&AuxBillingCity=CA&TechFirstName=John&TechLastName=Smith&TechAddress1=8939%20S.cross%20Blvd&TechStateProvince=CA&TechPostalCode=90045&TechCountry=US&TechPhone=+1.6613102107&TechEmailAddress=john@gmail.com&TechCity=CA&AdminFirstName=John&AdminLastName=Smith&AdminStateProvince=CA&AdminPostalCode=90045&AdminCountry=US&AdminPhone=+1.6613102107&AdminEmailAddress=john@gmail.com&AdminCity=CA&RegistrantFirstName=John&RegistrantLastName=Smith&RegistrantAddress1=8939%20S.cross%20Blvd&RegistrantStateProvince=CA&RegistrantPostalCode=90045&RegistrantCountry=US&RegistrantPhone=+1.6613102107&RegistrantEmailAddress=john@gmail.com&RegistrantCity=CA"
		correctValues, err := url.ParseQuery(correctURL)
		if err != nil {
			t.Fatal(err)
		}
		values, err := url.ParseQuery(r.URL.String())
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(values, correctValues) {
			t.Fatalf("URL = \n%v,\nwant \n%v", values, correctValues)
		}
		testMethod(t, r, "POST")

		fmt.Fprint(w, respXML)
	})

	client.NewRegistrant(
		"John", "Smith",
		"8939 S.cross Blvd", "",
		"CA", "CA", "90045", "US",
		" 1.6613102107", "john@gmail.com",
	)

	result, err := client.DomainCreate("domain1.com", 2)
	if err != nil {
		t.Fatalf("DomainCreate returned error: %v", nil)
	}

	// DomainGetListResult we expect, given the respXML above
	want := &DomainCreateResult{
		"domain1.com", true, 20.36, 9007, 196074, 380716, false, false,
	}

	if !reflect.DeepEqual(result, want) {
		t.Fatalf("DomainCreate returned\n%+v,\nwant\n%+v", result, want)
	}
}
