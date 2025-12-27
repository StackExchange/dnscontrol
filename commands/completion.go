package commands

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"strings"
	"text/template"
	"unicode/utf8"

	// "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v3"
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
		ArgsUsage:   fmt.Sprintf("[ %s ]", strings.Join(supportedShells, " | ")),
		Description: fmt.Sprintf("Generate shell completion script for [ %s ]", strings.Join(supportedShells, " | ")),
		// BashComplete: func(ctx *cli.Context) {  // BashComplete renamed to ShellComplete in v3
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			for _, shell := range supportedShells {
				if strings.HasPrefix(shell, cmd.Args().First()) {
					if _, err := cmd.Root().Writer.Write([]byte(shell + "\n")); err != nil {
						panic(err)
					}
				}
			}
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var inputShell string
			if inputShell = cmd.Args().First(); inputShell == "" {
				if inputShell = os.Getenv("SHELL"); inputShell == "" {
					return cli.Exit(errors.New("shell not specified"), 1)
				}
			}
			shellName := path.Base(inputShell) // necessary if using $SHELL, noop otherwise

			template := templates[shellName]
			if template == nil {
				return cli.Exit(fmt.Errorf("unknown shell: %s", inputShell), 1)
			}

			err = template.Execute(cmd.Root().Writer, struct {
				App *cli.Command
			}{cmd.Root()})
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

func dnscontrolPrintCommandSuggestions(commands []*cli.Command, writer io.Writer) {
	for _, command := range commands {
		if command.Hidden {
			continue
		}
		if strings.HasSuffix(os.Getenv("SHELL"), "zsh") {
			for _, name := range command.Names() {
				_, _ = fmt.Fprintf(writer, "%s:%s\n", name, command.Usage)
			}
		} else {
			for _, name := range command.Names() {
				_, _ = fmt.Fprintf(writer, "%s\n", name)
			}
		}
	}
}

func dnscontrolCliArgContains(flagName string) bool {
	for name := range strings.SplitSeq(flagName, ",") {
		name = strings.TrimSpace(name)
		count := min(utf8.RuneCountInString(name), 2)
		flag := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
		if slices.Contains(os.Args, flag) {
			return true
		}
	}
	return false
}

func dnscontrolPrintFlagSuggestions(lastArg string, flags []cli.Flag, writer io.Writer) {
	cur := strings.TrimPrefix(lastArg, "-")
	cur = strings.TrimPrefix(cur, "-")
	for _, flag := range flags {
		if bflag, ok := flag.(*cli.BoolFlag); ok && bflag.Hidden {
			continue
		}
		for _, name := range flag.Names() {
			name = strings.TrimSpace(name)
			// this will get total count utf8 letters in flag name
			count := min(utf8.RuneCountInString(name),
				// reuse this count to generate single - or -- in flag completion
				2)
			// if flag name has more than one utf8 letter and last argument in cli has -- prefix then
			// skip flag completion for short flags example -v or -x
			if strings.HasPrefix(lastArg, "--") && count == 1 {
				continue
			}
			// match if last argument matches this flag and it is not repeated
			if strings.HasPrefix(name, cur) && cur != name && !dnscontrolCliArgContains(name) {
				flagCompletion := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
				_, _ = fmt.Fprintln(writer, flagCompletion)
			}
		}
	}
}

func islastFlagComplete(lastArg string, flags []cli.Flag) bool {
	cur := strings.TrimPrefix(lastArg, "-")
	cur = strings.TrimPrefix(cur, "-")
	for _, flag := range flags {
		for _, name := range flag.Names() {
			name = strings.TrimSpace(name)
			if strings.HasPrefix(name, cur) && cur != name && !dnscontrolCliArgContains(name) {
				return false
			}
		}
	}
	return true
}
