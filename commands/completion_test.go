package commands

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"

	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

type shellTestDataItem struct {
	shellName                string
	shellPath                string
	completionScriptTemplate *template.Template
}

// setupTestShellCompletionCommand resets the buffers used to capture output and errors from the app.
func setupTestShellCompletionCommand(app *cli.App) func(t *testing.T) {
	return func(t *testing.T) {
		app.Writer.(*bytes.Buffer).Reset()
		cli.ErrWriter.(*bytes.Buffer).Reset()
	}
}

func TestShellCompletionCommand(t *testing.T) {
	app := cli.NewApp()
	app.Name = "testing"

	var appWriterBuffer bytes.Buffer
	app.Writer = &appWriterBuffer // capture output from app

	var appErrWriterBuffer bytes.Buffer
	cli.ErrWriter = &appErrWriterBuffer // capture errors from app (apparently, HandleExitCoder doesn't use app.ErrWriter!?)

	cli.OsExiter = func(int) {} // disable os.Exit call

	app.Commands = []*cli.Command{
		shellCompletionCommand(),
	}

	shellsAndCompletionScripts, err := testHelperGetShellsAndCompletionScripts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shellsAndCompletionScripts) == 0 {
		t.Fatal("no shells found")
	}

	invalidShellTestDataItem := shellTestDataItem{
		shellName: "invalid",
		shellPath: "/bin/invalid",
	}
	for _, tt := range shellsAndCompletionScripts {
		if tt.shellName == invalidShellTestDataItem.shellName {
			t.Fatalf("invalidShellTestDataItem.shellName (%s) is actually a valid shell name", invalidShellTestDataItem.shellName)
		}
	}

	// Test shell argument
	t.Run("shellArg", func(t *testing.T) {
		for _, tt := range shellsAndCompletionScripts {
			t.Run(tt.shellName, func(t *testing.T) {
				tearDownTest := setupTestShellCompletionCommand(app)
				defer tearDownTest(t)

				err := app.Run([]string{app.Name, "shell-completion", tt.shellName})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				got := appWriterBuffer.String()
				want, err := testHelperRenderTemplateFromApp(app, tt.completionScriptTemplate)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}

				stderr := appErrWriterBuffer.String()
				if stderr != "" {
					t.Errorf("want no stderr, got %q", stderr)
				}
			})
		}

		t.Run(invalidShellTestDataItem.shellName, func(t *testing.T) {
			tearDownTest := setupTestShellCompletionCommand(app)
			defer tearDownTest(t)

			err := app.Run([]string{app.Name, "shell-completion", "invalid"})

			if err == nil {
				t.Fatal("expected error, but didn't get one")
			}

			want := fmt.Sprintf("unknown shell: %s", invalidShellTestDataItem.shellName)
			got := strings.TrimSpace(appErrWriterBuffer.String())
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

			stdout := appWriterBuffer.String()
			if stdout != "" {
				t.Errorf("want no stdout, got %q", stdout)
			}
		})
	})

	// Test $SHELL envar
	t.Run("$SHELL", func(t *testing.T) {
		for _, tt := range shellsAndCompletionScripts {
			t.Run(tt.shellName, func(t *testing.T) {
				tearDownTest := setupTestShellCompletionCommand(app)
				defer tearDownTest(t)

				t.Setenv("SHELL", tt.shellPath)

				err := app.Run([]string{app.Name, "shell-completion"})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				got := appWriterBuffer.String()
				want, err := testHelperRenderTemplateFromApp(app, tt.completionScriptTemplate)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}

				stderr := appErrWriterBuffer.String()
				if stderr != "" {
					t.Errorf("want no stderr, got %q", stderr)
				}
			})
		}

		t.Run(invalidShellTestDataItem.shellName, func(t *testing.T) {
			tearDownTest := setupTestShellCompletionCommand(app)
			defer tearDownTest(t)

			t.Setenv("SHELL", invalidShellTestDataItem.shellPath)

			err := app.Run([]string{app.Name, "shell-completion"})
			if err == nil {
				t.Fatal("expected error, but didn't get one")
			}

			want := fmt.Sprintf("unknown shell: %s", invalidShellTestDataItem.shellPath)
			got := strings.TrimSpace(appErrWriterBuffer.String())
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

			stdout := appWriterBuffer.String()
			if stdout != "" {
				t.Errorf("want no stdout, got %q", stdout)
			}
		})
	})

	// Test shell argument completion (meta)
	t.Run("shell-name-completion", func(t *testing.T) {
		type testCase struct {
			shellArg string
			expected []string
		}
		testCases := []testCase{
			{shellArg: ""}, // empty 'shell' argument, returns all known shells (expected is filled later)
			{shellArg: "invalid", expected: []string{""}}, // invalid shell, returns none
		}

		for _, tt := range shellsAndCompletionScripts {
			testCases[0].expected = append(testCases[0].expected, tt.shellName)
			for i := range tt.shellName {
				testCases = append(testCases, testCase{
					shellArg: tt.shellName[:i+1],
					expected: []string{tt.shellName},
				})
			}
		}

		for _, tC := range testCases {
			t.Run(tC.shellArg, func(t *testing.T) {
				tearDownTest := setupTestShellCompletionCommand(app)
				defer tearDownTest(t)
				app.EnableBashCompletion = true
				defer func() {
					app.EnableBashCompletion = false
				}()

				err := app.Run([]string{app.Name, "shell-completion", tC.shellArg, "--generate-bash-completion"})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				for _, line := range strings.Split(strings.TrimSpace(appWriterBuffer.String()), "\n") {
					if !slices.Contains(tC.expected, line) {
						t.Errorf("%q found, but not expected", line)
					}
				}
			})
		}
	})
}

// testHelperGetShellsAndCompletionScripts collects all supported shells and their completion scripts and returns them
// as a slice of shellTestDataItem.
// The completion scripts are sourced with getCompletionSupportedShells
func testHelperGetShellsAndCompletionScripts() ([]shellTestDataItem, error) {
	shells, templates, err := getCompletionSupportedShells()
	if err != nil {
		return nil, err
	}

	var shellsAndValues []shellTestDataItem
	for shellName, t := range templates {
		if !slices.Contains(shells, shellName) {
			return nil, fmt.Errorf(
				`"%s" is not present in slice of shells from getCompletionSupportedShells`, shellName)
		}
		shellsAndValues = append(
			shellsAndValues,
			shellTestDataItem{
				shellName:                shellName,
				shellPath:                fmt.Sprintf("/bin/%s", shellName),
				completionScriptTemplate: t,
			},
		)
	}
	return shellsAndValues, nil
}

// testHelperRenderTemplateFromApp renders a given template with a given app.
// This is used to test the output of the CLI command against a 'known good' value.
func testHelperRenderTemplateFromApp(app *cli.App, scriptTemplate *template.Template) (string, error) {
	var scriptBytes bytes.Buffer
	err := scriptTemplate.Execute(&scriptBytes, struct {
		App *cli.App
	}{app})

	return scriptBytes.String(), err
}
