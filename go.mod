module github.com/StackExchange/dnscontrol/v4

go 1.25

retract v4.8.0

require google.golang.org/protobuf v1.36.10 // indirect

require golang.org/x/net v0.47.0

require (
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0
	github.com/Azure/go-autorest/autorest/to v0.4.1
	github.com/DisposaBoy/JsonConfigReader v0.0.0-20201129172854-99cf318d67e7
	github.com/PuerkitoBio/goquery v1.11.0
	github.com/TomOnTime/utfutil v1.0.0
	github.com/akamai/AkamaiOPEN-edgegrid-golang v1.2.2
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/aws/aws-sdk-go-v2 v1.40.0
	github.com/aws/aws-sdk-go-v2/config v1.32.2
	github.com/aws/aws-sdk-go-v2/credentials v1.19.2
	github.com/aws/aws-sdk-go-v2/service/route53 v1.61.0
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.34.12
	github.com/babolivier/go-doh-client v0.0.0-20201028162107-a76cff4cb8b6
	github.com/billputer/go-namecheap v0.0.0-20210108011502-994a912fb7f9
	github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v4 v4.0.7
	github.com/cloudflare/cloudflare-go v0.116.0
	github.com/digitalocean/godo v1.169.0
	github.com/ditashi/jsbeautifier-go v0.0.0-20141206144643-2520a8026a9c
	github.com/dnsimple/dnsimple-go v1.7.0
	github.com/exoscale/egoscale v0.102.4
	github.com/go-gandi/go-gandi v0.7.0
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe
	github.com/gopherjs/jquery v0.0.0-20191017083323-73f4c7416038
	github.com/jinzhu/copier v0.4.0
	github.com/miekg/dns v1.1.68
	github.com/mittwald/go-powerdns v0.6.7
	github.com/namedotcom/go v0.0.0-20180403034216-08470befbe04
	github.com/nrdcg/goinwx v0.12.0
	github.com/ovh/go-ovh v1.9.0
	github.com/philhug/opensrs-go v0.0.0-20171126225031-9dfa7433020d
	github.com/pkg/errors v0.9.1
	github.com/pquerna/otp v1.5.0
	github.com/qdm12/reprint v0.0.0-20200326205758-722754a53494
	github.com/robertkrimen/otto v0.5.1
	github.com/softlayer/softlayer-go v1.2.1
	github.com/stretchr/testify v1.11.1
	github.com/transip/gotransip/v6 v6.26.1
	github.com/urfave/cli/v2 v2.27.7
	github.com/xddxdd/ottoext v0.0.0-20221109171055-210517fa4419
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/oauth2 v0.33.0
	google.golang.org/api v0.256.0
	gopkg.in/ns1/ns1-go.v2 v2.16.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0
	github.com/G-Core/gcore-dns-sdk-go v0.3.3
	github.com/aliyun/alibaba-cloud-sdk-go v1.63.107
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.2
	github.com/centralnicgroup-opensource/rtldev-middleware-go-sdk/v5 v5.0.18
	github.com/failsafe-go/failsafe-go v0.9.2
	github.com/fatih/color v1.18.0
	github.com/fbiville/markdown-table-formatter v0.3.0
	github.com/google/go-cmp v0.7.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hetznercloud/hcloud-go/v2 v2.32.0
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.178
	github.com/kylelemons/godebug v1.1.0
	github.com/luadns/luadns-go v0.3.0
	github.com/mattn/go-isatty v0.0.20
	github.com/nicholas-fedor/shoutrrr v0.12.1
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/oracle/oci-go-sdk/v65 v65.105.0
	github.com/vercel/terraform-provider-vercel v1.14.1
	github.com/vultr/govultr/v2 v2.17.2
	golang.org/x/exp v0.0.0-20251125195548-87e1e737ad39
	golang.org/x/text v0.31.0
	golang.org/x/time v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.10 // indirect
	github.com/aws/smithy-go v1.23.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.24.0 // indirect
	github.com/boombuler/barcode v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.9.8 // indirect
	github.com/gofrs/flock v0.12.1 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/terraform-plugin-log v0.9.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opentracing/opentracing-go v1.2.1-0.20220228012449-10b1cf09e00b // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/peterhellberg/link v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.13.1 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251103181224-f26f9409b101 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	moul.io/http2curl v1.0.0 // indirect
)

tool golang.org/x/tools/cmd/stringer
