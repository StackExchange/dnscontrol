package commands

import (
	"embed"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"path"
	"strings"
	"text/template"
)

//go:embed completion-scripts/completion.*.gotmpl
var completionScripts embed.FS

func shellCompletionCommand() *cli.Command {
	supportedShells, templates, err := getCompletionSupportedShells()
	if err != nil {
		panic(err)
	}
	return &cli.Command{
		Name:        "shell-completion",
		Usage:       "generate shell completion scripts",
		Hidden:      true,
		ArgsUsage:   fmt.Sprintf("[ %s ]", strings.Join(supportedShells, " | ")),
		Description: fmt.Sprintf("Generate shell completion script for [ %s ]", strings.Join(supportedShells, " | ")),
		BashComplete: func(ctx *cli.Context) {
			for _, shell := range supportedShells {
				if strings.HasPrefix(shell, ctx.Args().First()) {
					ctx.App.Writer.Write([]byte(shell + "\n"))
				}
			}
		},
		Action: func(ctx *cli.Context) error {
			var inputShell string
			if inputShell = ctx.Args().First(); inputShell == "" {
				if inputShell = os.Getenv("SHELL"); inputShell == "" {
					return cli.Exit(errors.New("shell not specified"), 1)
				}
			}
			shellName := path.Base(inputShell) // necessary if using $SHELL, noop otherwise

			template := templates[shellName]
			if template == nil {
				return cli.Exit(fmt.Errorf("unknown shell: %s", inputShell), 1)
			}

			err = template.Execute(ctx.App.Writer, struct {
				App *cli.App
			}{ctx.App})
			if err != nil {
				return cli.Exit(fmt.Errorf("failed to print completion script: %w", err), 1)
			}
			return nil
		},
	}
}

var _ = cmd(catUtils, shellCompletionCommand())

// getCompletionSupportedShells returns a list of shells with available completions.
// The list is generated from the embedded completion scripts.
func getCompletionSupportedShells() (shells []string, shellCompletionScripts map[string]*template.Template, err error) {
	scripts, err := completionScripts.ReadDir("completion-scripts")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read completion scripts: %w", err)
	}

	shellCompletionScripts = make(map[string]*template.Template)

	for _, f := range scripts {
		fNameWithoutExtension := strings.TrimSuffix(f.Name(), ".gotmpl")
		shellName := strings.TrimPrefix(path.Ext(fNameWithoutExtension), ".")

		content, err := completionScripts.ReadFile(path.Join("completion-scripts", f.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read completion script %s", f.Name())
		}

		t := template.New(shellName)
		t, err = t.Parse(string(content))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse template %s", f.Name())
		}

		shells = append(shells, shellName)
		shellCompletionScripts[shellName] = t
	}
	return shells, shellCompletionScripts, nil
}
