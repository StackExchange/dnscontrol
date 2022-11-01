D('dnscontrol.zone',
  PROVIDER_NONE,
  DnsProvider(PROVIDER_TRANSIP),
  DefaultTTL(3600),
  TXT('@', 'v=spf1 -all'),
  TXT('_dmarc', 'v=DMARC1; p=none;')
)
