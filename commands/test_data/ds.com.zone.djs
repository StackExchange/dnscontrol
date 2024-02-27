var DSP_BIND = NewDnsProvider("bind", "BIND");
var REG_CHANGEME = NewRegistrar("none");

D("ds.com", REG_CHANGEME
	, DnsProvider(DSP_BIND)
	//, SOA("@", "ns3.serverfault.com.", "sysadmin.stackoverflow.com.", 2020022300, 3600, 600, 604800, 1440)
	, DS("geo", 14480, 13, 2, "BB1C4B615CDED2B34347CF23710471934D972F1E34F53B54ED8D5F786202C73B")
)

