
var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDSP("Cloudflare", "CLOUDFLAREAPI")
D("foo.com",REG,DSP(CF),
    A("@","1.2.3.4")
);