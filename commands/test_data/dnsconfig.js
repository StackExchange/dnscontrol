var REGISTRAR1 = NewRegistrar("registrar1");
var REGISTRAR2 = NewRegistrar("registrar2");

var DSP1 = NewDnsProvider("dsp1");
var DSP2 = NewDnsProvider("dsp2");

D("example.org", REGISTRAR1, DnsProvider(DSP1))

D("example.com", REGISTRAR2, DnsProvider(DSP2))

