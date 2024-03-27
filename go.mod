module github.com/StackExchange/dnscontrol/v4

go 1.22.1

retract v4.8.0

require github.com/go-jose/go-jose/v3 v3.0.3 // indirect

require google.golang.org/protobuf v1.33.0 // indirect

require (
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.2.0
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/DisposaBoy/JsonConfigReader v0.0.0-20201129172854-99cf318d67e7
	github.com/PuerkitoBio/goquery v1.9.1
	github.com/TomOnTime/utfutil v0.0.0-20230223141146-125e65197b36
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.2.2
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/aws/aws-sdk-go-v2 v1.26.0
	github.com/aws/aws-sdk-go-v2/config v1.27.9
	github.com/aws/aws-sdk-go-v2/credentials v1.17.9
	github.com/aws/aws-sdk-go-v2/service/route53 v1.40.3
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.23.3
	github.com/babolivier/go-doh-client v0.0.0-20201028162107-a76cff4cb8b6
	github.com/billputer/go-namecheap v0.0.0-20210108011502-994a912fb7f9
	github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v3 v3.5.6
	github.com/cloudflare/cloudflare-go v0.91.0
	github.com/digitalocean/godo v1.110.0
	github.com/ditashi/jsbeautifier-go v0.0.0-20141206144643-2520a8026a9c
	github.com/dnsimple/dnsimple-go v1.5.1
	github.com/exoscale/egoscale v0.90.2
	github.com/go-acme/lego v2.7.2+incompatible
	github.com/go-gandi/go-gandi v0.7.0
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe
	github.com/gopherjs/jquery v0.0.0-20191017083323-73f4c7416038
	github.com/hashicorp/vault/api v1.12.2
	github.com/jinzhu/copier v0.4.0
	github.com/miekg/dns v1.1.58
	github.com/mittwald/go-powerdns v0.6.2
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/goinwx v0.10.0
	github.com/oracle/oci-go-sdk/v32 v32.0.0
	github.com/ovh/go-ovh v1.4.3
	github.com/philhug/opensrs-go v0.0.0-20171126225031-9dfa7433020d
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.4.0
	github.com/qdm12/reprint v0.0.0-20200326205758-722754a53494
	github.com/robertkrimen/otto v0.3.0
	github.com/softlayer/softlayer-go v1.1.3
	github.com/stretchr/testify v1.9.0
	github.com/transip/gotransip/v6 v6.23.0
	github.com/urfave/cli/v2 v2.27.1
	github.com/xddxdd/ottoext v0.0.0-20221109171055-210517fa4419
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/net v0.22.0
	golang.org/x/oauth2 v0.18.0
	google.golang.org/api v0.171.0
	gopkg.in/ns1/ns1-go.v2 v2.7.11
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.10.0
	github.com/G-Core/gcore-dns-sdk-go v0.2.8
	github.com/fatih/color v1.16.0
	github.com/fbiville/markdown-table-formatter v0.3.0
	github.com/google/go-cmp v0.6.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/juju/errors v1.0.0
	github.com/kylelemons/godebug v1.1.0
	github.com/mattn/go-isatty v0.0.20
	github.com/vultr/govultr/v2 v2.17.2
	golang.org/x/exp v0.0.0-20240318143956-a85f2c67cd81
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute v1.23.4 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.5 // indirect
	github.com/aws/smithy-go v1.20.1 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-test/deep v1.0.3 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.3 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.6 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/peterhellberg/link v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240314234333-6e1732d8331c // indirect
	google.golang.org/grpc v1.62.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
