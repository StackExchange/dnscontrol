D('dnscontrol.zone',
  PROVIDER_NONE,
  DnsProvider(PROVIDER_TRANSIP),
  DefaultTTL(3600),
  TXT('@', 'google-site-verification=nJ_ftpMt6KigyTtD-JBZTp9yd0-EfK5nknDXI2ZRG1k'),
  TXT('@', 'v=spf1 -all'),
  TXT('_dmarc', 'v=DMARC1; p=none;'),
  CNAME('docs', 'db3053e25d-hosting.gitbook.io.')
)
