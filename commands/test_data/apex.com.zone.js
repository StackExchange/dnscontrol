var DSP_BIND = NewDnsProvider("bind", "BIND");
var REG_CHANGEME = NewRegistrar("none");

D("apex.com", REG_CHANGEME,
	DnsProvider(DSP_BIND),
	//SOA("@", "ns3.serverfault.com.", "sysadmin.stackoverflow.com.", 2020022300, 3600, 600, 604800, 1440),
	//NAMESERVER("ns-1313.awsdns-36.org."),
	//NAMESERVER("ns-736.awsdns-28.net."),
	//NAMESERVER("ns-cloud-c1.googledomains.com."),
	//NAMESERVER("ns-cloud-c2.googledomains.com."),
	// NOTE: CNAME at apex may require manual editing.
	CNAME("@", "cnametest1.example.com."),
	CNAME("www", "cnametest2.example.com."),
END);

