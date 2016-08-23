var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDSP("Cloudflare", "CLOUDFLAREAPI")

var BASE = IP("1.2.3.4")

D("foo.com",REG,CF,
    A("@",BASE),
    A("p1",BASE+1),
    A("p255", BASE+255)
);