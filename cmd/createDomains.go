package cmd

import "github.com/urfave/cli"

var createDomainsCommand = &cli.Command{
	Name:  "create-domains",
	Usage: "ensures that all domains in your configuration are present in all providers.",
	Action: func(ctx *cli.Context) error {
		return exit(CreateDomains(globalCreateDomainsArgs))
	},
	Category: catMain,
	Flags:    globalPreviewArgs.flags(),
}

type CreateDomainsArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
}

func (args *CreateDomainsArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

var globalCreateDomainsArgs CreateDomainsArgs

func CreateDomains(args CreateDomainsArgs) error {

	return nil
}
