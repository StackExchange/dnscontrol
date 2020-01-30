package commands

import (
	"fmt"

	"github.com/urfave/cli"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args GetZoneArgs
	return &cli.Command{
		Name:  "get-zone",
		Usage: "downloads a zone from a provider",
		Action: func(ctx *cli.Context) error {
			return exit(GetZone(args))
		},
		Flags: args.flags(),
	}
}())

// GetZoneArgs args required for the create-domain subcommand.
type GetZoneArgs struct {
	foo int
}

func (args *GetZoneArgs) flags() []cli.Flag {
	flags := args.GetZoneArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

// GetZone contains all data/flags needed to run create-domains, independently of CLI.
func GetZone(args GetZoneArgs) error {
	fmt.Println("GET ZONE!")
	return nil
}
