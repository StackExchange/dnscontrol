package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/providers"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
)

func generateFeatureMatrix() error {
	allNames := map[string]bool{}
	for n := range providers.RegistrarTypes {
		allNames[n] = true
	}
	for n := range providers.DNSProviderTypes {
		allNames[n] = true
	}
	providerTypes := []string{}
	for n := range allNames {
		providerTypes = append(providerTypes, n)
	}
	sort.Strings(providerTypes)
	matrix := &FeatureMatrix{
		Providers: map[string]FeatureMap{},
		Features: []FeatureDef{
			{"Official Support", "This means the provider is actively used at Stack Exchange, bugs are more likely to be fixed, and failing integration tests will block a release. See below for details"},
			{"DNS Provider", "Can manage and serve DNS zones"},
			{"Registrar", "The provider has registrar capabilities to set nameservers for zones"},
			{"ALIAS", "Provider supports some kind of ALIAS, ANAME or flattened CNAME record type"},
			{"AUTODNSSEC", "Provider can automatically manage DNSSEC"},
			{"CAA", "Provider can manage CAA records"},
			{"PTR", "Provider supports adding PTR records for reverse lookup zones"},
			{"NAPTR", "Provider can manage NAPTR records"},
			{"SRV", "Driver has explicitly implemented SRV record management"},
			{"SSHFP", "Provider can manage SSHFP records"},
			{"TLSA", "Provider can manage TLSA records"},
			{"TXTMulti", "Provider can manage TXT records with multiple strings"},
			{"R53_ALIAS", "Provider supports Route 53 limited ALIAS"},
			{"AZURE_ALIAS", "Provider supports Azure DNS limited ALIAS"},
			{"DS", "Provider supports adding DS records"},

			{"dual host", "This provider is recommended for use in 'dual hosting' scenarios. Usually this means the provider allows full control over the apex NS records"},
			{"create-domains", "This means the provider can automatically create domains that do not currently exist on your account. The 'dnscontrol create-domains' command will initialize any missing domains"},
			{"no_purge", "indicates you can use NO_PURGE macro to prevent deleting records not managed by dnscontrol. A few providers that generate the entire zone from scratch have a problem implementing this."},
			{"get-zones", "indicates the dnscontrol get-zones subcommand is implemented."},
		},
	}
	for _, p := range providerTypes {
		if p == "NONE" {
			continue
		}
		fm := FeatureMap{}
		notes := providers.Notes[p]
		if notes == nil {
			notes = providers.DocumentationNotes{}
		}
		setCap := func(name string, cap providers.Capability) {
			if notes[cap] != nil {
				fm[name] = notes[cap]
				return
			}
			fm.SetSimple(name, true, func() bool { return providers.ProviderHasCapability(p, cap) })
		}
		setDoc := func(name string, cap providers.Capability, defaultNo bool) {
			if notes[cap] != nil {
				fm[name] = notes[cap]
			} else if defaultNo {
				fm[name] = &providers.DocumentationNote{
					HasFeature: false,
				}
			}
		}
		setDoc("Official Support", providers.DocOfficiallySupported, true)
		fm.SetSimple("DNS Provider", false, func() bool { return providers.DNSProviderTypes[p] != nil })
		fm.SetSimple("Registrar", false, func() bool { return providers.RegistrarTypes[p] != nil })
		setCap("ALIAS", providers.CanUseAlias)
		setCap("AUTODNSSEC", providers.CanAutoDNSSEC)
		setCap("CAA", providers.CanUseCAA)
		setCap("NAPTR", providers.CanUseNAPTR)
		setCap("PTR", providers.CanUsePTR)
		setCap("R53_ALIAS", providers.CanUseRoute53Alias)
		setCap("AZURE_ALIAS", providers.CanUseAzureAlias)
		setCap("SRV", providers.CanUseSRV)
		setCap("SSHFP", providers.CanUseSSHFP)
		setCap("TLSA", providers.CanUseTLSA)
		setCap("TXTMulti", providers.CanUseTXTMulti)
		setCap("get-zones", providers.CanGetZones)
		setCap("DS", providers.CanUseDS)
		setDoc("dual host", providers.DocDualHost, false)
		setDoc("create-domains", providers.DocCreateDomains, true)

		// no purge is a freaky double negative
		cap := providers.CantUseNOPURGE
		if notes[cap] != nil {
			fm["no_purge"] = notes[cap]
		} else {
			fm.SetSimple("no_purge", false, func() bool { return !providers.ProviderHasCapability(p, cap) })
		}
		matrix.Providers[p] = fm
	}
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, matrix)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("docs/_includes/matrix.html", buf.Bytes(), 0644)
}

// FeatureDef describes features.
type FeatureDef struct {
	Name, Desc string
}

// FeatureMap maps provider names to compliance documentation.
type FeatureMap map[string]*providers.DocumentationNote

// SetSimple configures a provider's setting in fm.
func (fm FeatureMap) SetSimple(name string, unknownsAllowed bool, f func() bool) {
	if f() {
		fm[name] = &providers.DocumentationNote{HasFeature: true}
	} else if !unknownsAllowed {
		fm[name] = &providers.DocumentationNote{HasFeature: false}
	}
}

// FeatureMatrix describes features and which providers support it.
type FeatureMatrix struct {
	Features  []FeatureDef
	Providers map[string]FeatureMap
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"safe": func(s string) template.HTML { return template.HTML(s) },
}).Parse(`
	{% comment %}
    Matrix generated by build/generate/featureMatrix.go. DO NOT HAND EDIT!
{% endcomment %}{{$providers := .Providers}}
<table class="table-header-rotated">
<thead>
	<tr>
	<th></th>
	{{range $key,$val := $providers}}<th class="rotate"><div><span>{{$key}}</span></div></th>
	{{end -}}
	</tr>
</thead>
<tbody>
	{{range .Features}}{{$name := .Name}}<tr>
		<th class="row-header" style="text-decoration: underline;" data-toggle="tooltip" data-container="body" data-placement="top" title="{{.Desc}}">{{$name}}</th>
		{{range $pname, $features := $providers}}{{$f := index $features $name}}{{if $f -}}
		<td class="{{if $f.HasFeature}}success{{else if $f.Unimplemented}}info{{else}}danger{{end}}"
			{{- if $f.Comment}} data-toggle="tooltip" data-container="body" data-placement="top" title="{{$f.Comment}}"{{end}}>
			{{if $f.Link}}<a href="{{$f.Link}}">{{end}}<i class="fa {{if and $f.Comment (not $f.Unimplemented)}}has-tooltip {{end}}
				{{- if $f.HasFeature}}fa-check text-success{{else if $f.Unimplemented}}fa-circle-o text-info{{else}}fa-times text-danger{{end}}" aria-hidden="true"></i>{{if $f.Link}}</a>{{end}}
		</td>
		{{- else}}<td><i class="fa fa-minus dim"></i></td>{{end}}
		{{end -}}
	</tr>
	{{end -}}
</tbody>
</table>
`))
