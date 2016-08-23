package web

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/StackExchange/dnscontrol/js"
	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/normalize"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
)

var (
	jsFilename, credsFilename string
)

var apiMux = http.NewServeMux()

func Serve(jsFile, creds string, devMode bool) {
	if devMode {
		runWebpack()
	}
	jsFilename, credsFilename = jsFile, creds
	http.HandleFunc("/helpers.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		helpers := js.GetHelpers(devMode)
		w.Write([]byte(helpers))
	})

	http.Handle("/bundle.js", gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(FSMustByte(devMode, "/bundle.js"))
	})))
	http.Handle("/api/", apiMux)
	api("/api/save", save, "POST")
	api("/api/preview", preview, "POST")
	api("/api/run", runCorrection, "POST")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := func() error {
			script, err := ioutil.ReadFile(jsFile)
			if err != nil {
				return err
			}
			b64 := base64.StdEncoding.EncodeToString(script)
			buf := &bytes.Buffer{}
			err = index.Execute(buf, b64)
			if err != nil {
				return err
			}
			w.Write(buf.Bytes())
			return nil
		}()
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func save(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	err = ioutil.WriteFile(jsFilename, body, os.FileMode(0660))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type correction struct {
	*models.Correction
	ID      string
	created time.Time
}

var cache = map[string]*correction{}

func newCorrection(c *models.Correction) *correction {
	corr := &correction{
		Correction: c,
		created:    time.Now(),
		ID:         randomString(),
	}
	cache[corr.ID] = corr
	return corr
}

func randomString() string {
	buf := make([]byte, 6)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func runCorrection(r *http.Request) (interface{}, error) {
	id := r.FormValue("id")
	c, ok := cache[id]
	if !ok {
		return nil, fmt.Errorf("Cannot find correction %s.", id)
	}
	return nil, c.Correction.F()
}

func preview(r *http.Request) (interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	data := &struct {
		Config *models.DNSConfig
		Query  struct {
			Domain    string
			Registrar bool
			Dsps      []bool
		}
	}{}
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}

	//limit to only the domain we care about
	data.Config.Domains = []*models.DomainConfig{data.Config.FindDomain(data.Query.Domain)}

	errs := normalize.NormalizeAndValidateConfig(data.Config)
	if len(errs) > 0 {
		buf := &bytes.Buffer{}
		for _, err := range errs {
			fmt.Fprintln(buf, err)
		}
		return nil, fmt.Errorf(buf.String())
	}

	configs, err := config.LoadProviderConfigs(credsFilename)
	if err != nil {
		return nil, err
	}

	registrars, err := providers.CreateRegistrars(data.Config, configs)
	if err != nil {
		return nil, err
	}

	dsps, err := providers.CreateDsps(data.Config, configs)
	if err != nil {
		return nil, err
	}

	domain := data.Config.FindDomain(data.Query.Domain)
	if domain == nil {
		return nil, fmt.Errorf("Didn't find domain %s in config.", data.Query.Domain)
	}

	corrections := map[string][]*correction{}
	for i, dspName := range domain.Dsps {
		readOnly := false
		if i >= len(data.Query.Dsps) || !data.Query.Dsps[i] {
			if len(domain.Nameservers) == 0 {
				readOnly = true
			} else {
				continue
			}
		}
		dsp, ok := dsps[dspName]
		if !ok {
			return nil, fmt.Errorf("DSP %s not found", dspName)
		}
		dom, err := domain.Copy()
		if err != nil {
			return nil, err
		}
		cs := []*correction{}
		domCorrections, err := dsp.GetDomainCorrections(dom)
		if err != nil {
			return nil, err
		}
		if len(domain.Nameservers) == 0 && len(dom.Nameservers) > 0 {
			domain.Nameservers = dom.Nameservers
		}
		for _, c := range domCorrections {
			cs = append(cs, newCorrection(c))
		}
		if !readOnly {
			corrections["DSP: "+dspName] = cs
		}
	}

	if data.Query.Registrar {
		reg, ok := registrars[domain.Registrar]
		if !ok {
			return nil, fmt.Errorf("Registrar %s not found", domain.Registrar)
		}
		dom, err := domain.Copy()
		if err != nil {
			return nil, err
		}
		cs := []*correction{}
		regCorrections, err := reg.GetRegistrarCorrections(dom)
		if err != nil {
			return nil, err
		}
		for _, c := range regCorrections {
			cs = append(cs, newCorrection(c))
		}
		corrections["Registrar: "+domain.Registrar] = cs
	}
	return corrections, nil
}

var index = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>DNSControl IDE</title>
    <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
	</head>
<body>
	<app></app>
    <script>var initialScript = decodeURIComponent(escape(window.atob("{{.}}")))</script>
	<script src="./helpers.js" type="text/javascript"></script>
	<script src="./bundle.js" type="text/javascript"></script>
</body>
</html>
`))

type apiHandler func(r *http.Request) (interface{}, error)

func api(route string, f apiHandler, methods ...string) {
	apiMux.Handle(route, gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(methods) > 0 {
			ok := false
			for _, meth := range methods {
				if meth == r.Method {
					ok = true
					break
				}
			}
			if !ok {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		}
		ret, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if ret == nil {
			return
		}
		m := json.NewEncoder(w)
		if err = m.Encode(ret); err != nil {
			http.Error(w, err.Error(), 500)
		}

	})))
}

func runWebpack() {
	cmd := exec.Command("npm", "run", "watch")
	cmd.Dir = "web"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	go func() { log.Fatal(cmd.Run()) }()
}
