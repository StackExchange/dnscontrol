// FYI: go-yaml writes an empty string as ""; python's yaml writes an empty string as "".
// For that reason:
// 006-apex:       tests YAML -> JSON.
// 007-apex-quote: tests JS -> YAML -> JSON
// It would be nice if go-yaml has an option to write '' instead of "".

var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("bind", "BIND")
D("example.tld",REG,DnsProvider(CF),
    DefaultTTL(307),
    A("@","1.2.3.4")
);
