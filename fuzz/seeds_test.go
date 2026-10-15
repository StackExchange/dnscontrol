package fuzz

// Seed corpora for FuzzConfigJSON.  Each seed is a JSON document in
// the shape that JSON.stringify(conf) produces after dnsconfig.js runs
// (see pkg/js.ExecuteJavascriptString).  Only fields consumed by the
// Go pipeline are included; helpers.js omits the rest.

const seedBasic = `{
  "registrars": [{"name": "r", "type": "NONE"}],
  "dns_providers": [{"name": "d", "type": "BIND"}],
  "domains": [{
    "name": "example.com",
    "uniquename": "example.com",
    "registrar": "r",
    "dnsProviders": {"d": -1},
    "nameservers": [{"name": "ns1.example.com"}],
    "rawrecords": [
      {"type": "A",    "args": ["@", "1.2.3.4"], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "A",    "args": ["www", "1.2.3.5"], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "AAAA", "args": ["www", "2001:db8::1"], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "CNAME", "args": ["alias", "www.example.com."], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "TXT",  "args": ["txt", "hello world"], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "MX",   "args": ["@", 10, "mail.example.com."], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "NS",   "args": ["sub", "ns1.example.com."], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "CAA",  "args": ["@", 0, "issue", "letsencrypt.org"], "ttl": 300, "filepos": "[line:1:1]"}
    ]
  }]
}`

const seedRecordTypes = `{
  "registrars": [{"name": "r", "type": "NONE"}],
  "domains": [{
    "name": "example.com",
    "uniquename": "example.com",
    "registrar": "r",
    "rawrecords": [
      {"type": "SRV", "args": ["_ntp._udp", 1, 100, 123, "one.example.com."], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "TLSA", "args": ["_443._tcp", 3, 1, 1, "MDFiYTQ3MTljODBiNmZlOTExYjA5MWE3YzA1MTI0YjY0ZWVlY2U5NjRlMDljMDU4ZWY4Zjk4MDVkYWNhNTQ2YiAgLQo="], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "TXT", "args": ["long", "This TXT record is longer than 255 octets so that it is stored as multiple segments. This TXT record is longer than 255 octets so that it is stored as multiple segments. This TXT record is longer than 255 octets."], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "LOC", "args": ["@", 42, 21, 54, "N", 71, 6, 18, "W", -24, 30, 0, 0], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "SOA", "args": ["@", "ns1.example.com.", "hostmaster.example.com.", 3600, 600, 3600, 1209600, 300], "ttl": 300, "filepos": "[line:1:1]"}
    ]
  }]
}`

const seedSubdomainAndMetas = `{
  "registrars": [{"name": "r", "type": "NONE"}],
  "domains": [{
    "name": "example.com",
    "uniquename": "example.com",
    "registrar": "r",
    "rawrecords": [
      {"type": "A", "args": ["@", "1.2.3.4"], "metas": [{"cloudflare_proxy": "ON"}], "ttl": 300, "filepos": "[line:1:1]"},
      {"type": "A", "args": ["deep", "10.9.9.9"], "subdomain": "sub", "ttl": 300, "filepos": "[line:1:1]"}
    ]
  }]
}`

const seedSplitHorizon = `{
  "registrars": [{"name": "r", "type": "NONE"}],
  "domains": [
    {
      "name": "example.com!tag",
      "uniquename": "example.com!tag",
      "registrar": "r",
      "rawrecords": [
        {"type": "A", "args": ["@", "1.2.3.4"], "ttl": 300, "filepos": "[line:1:1]"}
      ]
    }
  ]
}`
