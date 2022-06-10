module github.com/StackExchange/dnscontrol/v3

go 1.18

require gopkg.in/yaml.v3 v3.0.0 // indirect

require (
	github.com/Azure/azure-sdk-for-go v65.0.0+incompatible
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.11
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/DisposaBoy/JsonConfigReader v0.0.0-20201129172854-99cf318d67e7
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/TomOnTime/utfutil v0.0.0-20210710122150-437f72b26edf
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.1.1
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/aws/aws-sdk-go-v2 v1.16.4
	github.com/aws/aws-sdk-go-v2/config v1.15.9
	github.com/aws/aws-sdk-go-v2/credentials v1.12.4
	github.com/aws/aws-sdk-go-v2/service/route53 v1.20.5
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.12.5
	github.com/babolivier/go-doh-client v0.0.0-20201028162107-a76cff4cb8b6
	github.com/bhendo/go-powershell v0.0.0-20190719160123-219e7fb4e41e
	github.com/billputer/go-namecheap v0.0.0-20210108011502-994a912fb7f9
	github.com/cloudflare/cloudflare-go v0.40.0
	github.com/digitalocean/godo v1.80.0
	github.com/ditashi/jsbeautifier-go v0.0.0-20141206144643-2520a8026a9c
	github.com/dnsimple/dnsimple-go v0.70.1
	github.com/exoscale/egoscale v0.59.0
	github.com/go-acme/lego v2.7.2+incompatible
	github.com/go-gandi/go-gandi v0.4.0
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe
	github.com/google/go-github/v35 v35.3.0
	github.com/gopherjs/jquery v0.0.0-20191017083323-73f4c7416038
	github.com/hashicorp/vault/api v1.6.0
	github.com/hexonet/go-sdk/v3 v3.5.3
	github.com/jarcoal/httpmock v1.0.8 // indirect
	github.com/jinzhu/copier v0.3.5
	github.com/miekg/dns v1.1.49
	github.com/mittwald/go-powerdns v0.5.2
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/goinwx v0.8.1
	github.com/oracle/oci-go-sdk/v32 v32.0.0
	github.com/ovh/go-ovh v1.1.0
	github.com/philhug/opensrs-go v0.0.0-20171126225031-9dfa7433020d
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.3.0
	github.com/qdm12/reprint v0.0.0-20200326205758-722754a53494
	github.com/robertkrimen/otto v0.0.0-20211024170158-b87d35c0b86f
	github.com/softlayer/softlayer-go v1.0.5
	github.com/stretchr/testify v1.7.1
	github.com/tdewolff/minify/v2 v2.11.7
	github.com/transip/gotransip/v6 v6.17.0
	github.com/urfave/cli/v2 v2.8.1
	github.com/vultr/govultr v1.1.1
	github.com/xddxdd/ottoext v0.0.0-20210101073831-439879ee6281
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37
	golang.org/x/oauth2 v0.0.0-20220524215830-622c5d57e401
	google.golang.org/api v0.81.0
	gopkg.in/ns1/ns1-go.v2 v2.5.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/mattn/go-isatty v0.0.14
	golang.org/x/exp v0.0.0-20220602145555-4a0574d9293f
)

require (
	cloud.google.com/go/compute v1.6.1 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.24 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.18 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.5 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/GeertJohan/go.rice v1.0.2 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.11 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.5 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.6 // indirect
	github.com/aws/smithy-go v1.11.2 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/daaku/go.zipexe v1.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/deepmap/oapi-codegen v1.6.1 // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/golang-jwt/jwt/v4 v4.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.5 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault/sdk v0.5.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f // indirect
	github.com/juju/testing v0.0.0-20210324180055-18c50b0c2098 // indirect
	github.com/kolo/xmlrpc v0.0.0-20201022064351-38db28db192b // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/peterhellberg/link v1.1.0 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/tdewolff/parse/v2 v2.5.32 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	moul.io/http2curl v1.0.0 // indirect
)
