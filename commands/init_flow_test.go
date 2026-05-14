package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stubAsker drives the init flow from a pre recorded script for
// deterministic tests.
type stubAsker struct {
	t          *testing.T
	selects    []string
	inputs     []string
	secrets    []string
	multilines []string
	confirm    []bool
}

func (stub *stubAsker) Select(_, _ string, options []string, _ string) (string, error) {
	if len(stub.selects) == 0 {
		stub.t.Fatalf("unexpected Select (options %v)", options)
	}
	value := stub.selects[0]
	stub.selects = stub.selects[1:]
	return value, nil
}

func (stub *stubAsker) Input(_, _, defaultValue string) (string, error) {
	if len(stub.inputs) == 0 {
		return defaultValue, nil
	}
	value := stub.inputs[0]
	stub.inputs = stub.inputs[1:]
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func (stub *stubAsker) Secret(_, _ string) (string, error) {
	if len(stub.secrets) == 0 {
		stub.t.Fatalf("unexpected Secret call")
	}
	value := stub.secrets[0]
	stub.secrets = stub.secrets[1:]
	return value, nil
}

func (stub *stubAsker) Multiline(_, _ string) (string, error) {
	if len(stub.multilines) == 0 {
		stub.t.Fatalf("unexpected Multiline call")
	}
	value := stub.multilines[0]
	stub.multilines = stub.multilines[1:]
	return value, nil
}

func (stub *stubAsker) Confirm(_ string, _ bool) (bool, error) {
	if len(stub.confirm) == 0 {
		stub.t.Fatalf("unexpected Confirm call")
	}
	value := stub.confirm[0]
	stub.confirm = stub.confirm[1:]
	return value, nil
}

// TestRunInit_NoneBindFlow walks the full init flow using only the built
// in providers (NONE registrar + BIND DNS). It asserts the generated
// creds.json and dnsconfig.js parse cleanly.
func TestRunInit_NoneBindFlow(t *testing.T) {
	dir := t.TempDir()

	stub := &stubAsker{
		t: t,
		selects: []string{
			"BIND", // DNS provider (asked first)
			"NONE", // registrar
		},
		inputs: []string{
			"",            // BIND: directory (accept default)
			"",            // BIND: filenameformat (accept default)
			"",            // BIND: entry name (accept default "bind_primary")
			"",            // NONE: entry name (accept default "none_primary")
			"example.com", // first domain
		},
		confirm: []bool{
			false, // "Add another domain?"
			true,  // Write these files?
			false, // Compare domains with zones at provider?
			false, // Run preview now?
		},
	}

	// Replace the subprocess seam so the test does not actually exec the
	// dnscontrol binary for `preview` or `get-zones`.
	origRun := runSubcommand
	runSubcommand = func(*exec.Cmd) error { return nil }
	t.Cleanup(func() { runSubcommand = origRun })

	args := InitArgs{
		CredsFile:  filepath.Join(dir, "creds.json"),
		ConfigFile: filepath.Join(dir, "dnsconfig.js"),
	}
	if err := runInit(args, stub); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	credsBytes, err := os.ReadFile(args.CredsFile)
	if err != nil {
		t.Fatalf("read creds: %v", err)
	}
	if !strings.Contains(string(credsBytes), `"bind_primary"`) {
		t.Errorf("creds.json missing bind_primary entry: %s", credsBytes)
	}
	if !strings.Contains(string(credsBytes), `"TYPE": "BIND"`) {
		t.Errorf("creds.json missing BIND TYPE: %s", credsBytes)
	}

	configBytes, err := os.ReadFile(args.ConfigFile)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(configBytes), `NewDnsProvider("bind_primary")`) {
		t.Errorf("config missing bind_primary provider: %s", configBytes)
	}
	if !strings.Contains(string(configBytes), `D("example.com"`) {
		t.Errorf("config missing example.com domain: %s", configBytes)
	}
}
