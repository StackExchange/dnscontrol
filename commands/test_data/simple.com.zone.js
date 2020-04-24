var bind = NewDnsProvider("bind", "BIND");
var REG_CHANGEME = NewRegistrar("ThirdParty", "NONE");
D("simple.com", REG_CHANGEME,
	DnsProvider(bind),
	//SOA('@', 'ns3.serverfault.com.', 'sysadmin.stackoverflow.com.', 2020022300, 3600, 600, 604800, 1440),
	NAMESERVER('ns-1313.awsdns-36.org.'),
	NAMESERVER('ns-736.awsdns-28.net.'),
	NAMESERVER('ns-cloud-c1.googledomains.com.'),
	NAMESERVER('ns-cloud-c2.googledomains.com.'),
	MX('@', 1, 'aspmx.l.google.com.'),
	MX('@', 5, 'alt1.aspmx.l.google.com.'),
	MX('@', 5, 'alt2.aspmx.l.google.com.'),
	MX('@', 10, 'alt3.aspmx.l.google.com.'),
	MX('@', 10, 'alt4.aspmx.l.google.com.'),
	TXT('@', 'google-site-verification=O54a_pYHGr4EB8iLoGFgX8OTZ1DkP1KWnOLpx0YCazI'),
	TXT('@', 'v=spf1 mx include:mktomail.com ~all'),
	TXT('m1._domainkey', 'v=DKIM1;k=rsa;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCZfEV2C82eJ4OA3Mslz4C6msjYYalg1eUcHeJQ//QM1hOZSvn4qz+hSKGi7jwNDqsZNzM8vCt2+XzdDYL3JddwUEhoDsIsZsJW0qzIVVLLWCg6TLNS3FpVyjc171o94dpoHFekfswWDoEwFQ03Woq2jchYWBrbUf7MMcdEj/EQqwIDAQAB'),
	CNAME('dev', 'stackoverflowsandbox2.mktoweb.com.'),
	CNAME('dev-email', 'mkto-sj310056.com.'),
	TXT('m1._domainkey.dev-email', 'v=DKIM1;k=rsa;p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCIBezZ2Gc+/3PghWk+YOE6T9HdwgUTMTR0Fne2i51MNN9Qs7AqDitVdG/949iDbI2fPNZSnKtOcnlLYwvve9MhMAMI1nZ26ILhgaBJi2BMZQpGFlO4ucuo/Uj4DPZ5Ge/NZHCX0CRhAhR5sRmL2OffNcFXFrymzUuz4KzI/NyUiwIDAQAB'),
	CNAME('email', 'mkto-sj280138.com.'),
	CNAME('info', 'stackoverflow.mktoweb.com.'),
	SRV('_sip._tcp', 10, 60, 5060, 'bigbox.example.com.')
)
