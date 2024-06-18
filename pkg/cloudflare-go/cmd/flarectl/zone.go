package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

func zoneCerts(*cli.Context) error {
	return nil
}

func zoneKeyless(*cli.Context) error {
	return nil
}

func zoneRailgun(*cli.Context) error {
	return nil
}

func zoneCreate(c *cli.Context) error {
	if err := checkFlags(c, "zone"); err != nil {
		return err
	}
	zone := c.String("zone")
	jumpstart := c.Bool("jumpstart")
	accountID := c.String("account-id")
	zoneType := c.String("type")
	var account cloudflare.Account
	if accountID != "" {
		account.ID = accountID
	}

	if zoneType != "partial" {
		zoneType = "full"
	}

	_, err := api.CreateZone(context.Background(), zone, jumpstart, account, zoneType)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		return err
	}

	return nil
}

func zoneCheck(c *cli.Context) error {
	if err := checkFlags(c, "zone"); err != nil {
		return err
	}
	zone := c.String("zone")

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		fmt.Println(err)
		return err
	}

	res, err := api.ZoneActivationCheck(context.Background(), zoneID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("%s\n", res.Messages[0].Message)

	return nil
}

func zoneList(c *cli.Context) error {
	zones, err := api.ListZones(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}
	output := make([][]string, 0, len(zones))
	for _, z := range zones {
		output = append(output, []string{
			z.ID,
			z.Name,
			z.Plan.Name,
			z.Status,
		})
	}
	writeTable(c, output, "ID", "Name", "Plan", "Status")

	return nil
}

func zoneDelete(c *cli.Context) error {
	if err := checkFlags(c, "zone"); err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	_, err = api.DeleteZone(context.Background(), zoneID)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		return err
	}

	return nil
}

func zoneCreateLockdown(c *cli.Context) error {
	if err := checkFlags(c, "zone", "urls", "targets", "values"); err != nil {
		return err
	}
	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Println(err)
		return err
	}
	targets := c.StringSlice("targets")
	values := c.StringSlice("values")
	if len(targets) != len(values) {
		cli.ShowCommandHelp(c, "targets and values does not match") //nolint
		return nil
	}
	var zonelockdownconfigs = []cloudflare.ZoneLockdownConfig{}
	for index := 0; index < len(targets); index++ {
		zonelockdownconfigs = append(zonelockdownconfigs, cloudflare.ZoneLockdownConfig{
			Target: c.StringSlice("targets")[index],
			Value:  c.StringSlice("values")[index],
		})
	}
	params := cloudflare.ZoneLockdownCreateParams{
		Description:    c.String("description"),
		URLs:           c.StringSlice("urls"),
		Configurations: zonelockdownconfigs,
	}

	resp, err := api.CreateZoneLockdown(context.Background(), cloudflare.ZoneIdentifier(zoneID), params)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating ZONE lock down: ", err)
		return err
	}

	output := make([][]string, 0, 1)

	format := []string{
		resp.ID,
	}

	output = append(output, format)

	writeTable(c, output, "ID")

	return nil
}

func zoneInfo(c *cli.Context) error {
	var zone string
	if c.NArg() > 0 {
		zone = c.Args().First()
	} else if c.String("zone") != "" {
		zone = c.String("zone")
	} else {
		cli.ShowSubcommandHelp(c) //nolint
		return nil
	}
	zones, err := api.ListZones(context.Background(), zone)
	if err != nil {
		fmt.Println(err)
		return err
	}
	output := make([][]string, 0, len(zones))
	for _, z := range zones {
		var nameservers []string
		if len(z.VanityNS) > 0 {
			nameservers = z.VanityNS
		} else {
			nameservers = z.NameServers
		}
		output = append(output, []string{
			z.ID,
			z.Name,
			z.Plan.Name,
			z.Status,
			strings.Join(nameservers, ", "),
			fmt.Sprintf("%t", z.Paused),
			z.Type,
		})
	}
	writeTable(c, output, "ID", "Zone", "Plan", "Status", "Name Servers", "Paused", "Type")

	return nil
}

func zonePlan(*cli.Context) error {
	return nil
}

func zoneSettings(*cli.Context) error {
	return nil
}

func zoneCachePurge(c *cli.Context) error {
	if err := checkFlags(c, "zone"); err != nil {
		cli.ShowSubcommandHelp(c) //nolint
		return err
	}

	zoneName := c.String("zone")
	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	var resp cloudflare.PurgeCacheResponse

	// Purge everything
	if c.Bool("everything") {
		resp, err = api.PurgeEverything(context.Background(), zoneID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error purging all from zone %q: %s\n", zoneName, err)
			return err
		}
	} else {
		var (
			files    = c.StringSlice("files")
			tags     = c.StringSlice("tags")
			hosts    = c.StringSlice("hosts")
			prefixes = c.StringSlice("prefixes")
		)

		if len(files) == 0 && len(tags) == 0 && len(hosts) == 0 && len(prefixes) == 0 {
			fmt.Fprintln(os.Stderr, "You must provide at least one of the --files, --tags, --prefixes or --hosts flags")
			return nil
		}

		// Purge selectively
		purgeReq := cloudflare.PurgeCacheRequest{
			Files:    c.StringSlice("files"),
			Tags:     c.StringSlice("tags"),
			Hosts:    c.StringSlice("hosts"),
			Prefixes: c.StringSlice("prefixes"),
		}

		resp, err = api.PurgeCache(context.Background(), zoneID, purgeReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error purging the cache from zone %q: %s\n", zoneName, err)
			return err
		}
	}

	output := make([][]string, 0, 1)
	output = append(output, formatCacheResponse(resp))

	writeTable(c, output, "ID")

	return nil
}

func zoneRecords(c *cli.Context) error {
	var zone string
	if c.NArg() > 0 {
		zone = c.Args().First()
	} else if c.String("zone") != "" {
		zone = c.String("zone")
	} else {
		cli.ShowSubcommandHelp(c) //nolint
		return nil
	}

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Create an empty record for searching for records
	rr := cloudflare.ListDNSRecordsParams{}
	var records []cloudflare.DNSRecord
	if c.String("id") != "" {
		rec, err := api.GetDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), c.String("id"))
		if err != nil {
			fmt.Println(err)
			return err
		}
		records = append(records, rec)
	} else {
		if c.String("type") != "" {
			rr.Type = c.String("type")
		}
		if c.String("name") != "" {
			rr.Name = c.String("name")
		}
		if c.String("content") != "" {
			rr.Content = c.String("content")
		}
		var err error
		records, _, err = api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), rr)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	output := make([][]string, 0, len(records))
	for _, r := range records {
		switch r.Type {
		case "MX":
			r.Content = fmt.Sprintf("%d %s", *r.Priority, r.Content)
		case "SRV":
			dp := r.Data.(map[string]interface{})
			r.Content = fmt.Sprintf("%.f %s", dp["priority"], r.Content)
			// Cloudflare's API, annoyingly, automatically prepends the weight
			// and port into content, separated by tabs.
			// XXX: File this as a bug. LOC doesn't do this.
			r.Content = strings.Replace(r.Content, "\t", " ", -1)
		}
		output = append(output, []string{
			r.ID,
			r.Type,
			r.Name,
			r.Content,
			strconv.FormatBool(*r.Proxied),
			fmt.Sprintf("%d", r.TTL),
		})
	}
	writeTable(c, output, "ID", "Type", "Name", "Content", "Proxied", "TTL")

	return nil
}

func formatCacheResponse(resp cloudflare.PurgeCacheResponse) []string {
	return []string{
		resp.Result.ID,
	}
}

func zoneExport(c *cli.Context) error {
	var zone string
	if c.NArg() > 0 {
		zone = c.Args().First()
	} else if c.String("zone") != "" {
		zone = c.String("zone")
	} else {
		cli.ShowSubcommandHelp(c) //nolint
		return nil
	}

	zoneID, err := api.ZoneIDByName(zone)
	if err != nil {
		fmt.Println(err)
		return err
	}

	res, err := api.ZoneExport(context.Background(), zoneID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Print(res)

	return nil
}
