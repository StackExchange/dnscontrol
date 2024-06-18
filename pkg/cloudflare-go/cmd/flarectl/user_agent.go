package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

func formatUserAgentRule(rule cloudflare.UserAgentRule) []string {
	return []string{
		rule.ID,
		rule.Description,
		rule.Mode,
		rule.Configuration.Value,
		strconv.FormatBool(rule.Paused),
	}
}

func userAgentCreate(c *cli.Context) error {
	if err := checkFlags(c, "zone", "mode", "value"); err != nil {
		fmt.Println(err)
		return err
	}

	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Println(err)
		return err
	}

	userAgentRule := cloudflare.UserAgentRule{
		Description: c.String("description"),
		Mode:        c.String("mode"),
		Paused:      c.Bool("paused"),
		Configuration: cloudflare.UserAgentRuleConfig{
			Target: "ua",
			Value:  c.String("value"),
		},
	}

	resp, err := api.CreateUserAgentRule(context.Background(), zoneID, userAgentRule)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating User-Agent block rule: ", err)
		return err
	}

	output := [][]string{
		formatUserAgentRule(resp.Result),
	}

	writeTable(c, output, "ID", "Description", "Mode", "Value", "Paused")

	return nil
}

func userAgentUpdate(c *cli.Context) error {
	if err := checkFlags(c, "zone", "id", "mode", "value"); err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	userAgentRule := cloudflare.UserAgentRule{
		Description: c.String("description"),
		Mode:        c.String("mode"),
		Paused:      c.Bool("paused"),
		Configuration: cloudflare.UserAgentRuleConfig{
			Target: "ua",
			Value:  c.String("value"),
		},
	}

	resp, err := api.UpdateUserAgentRule(context.Background(), zoneID, c.String("id"), userAgentRule)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error updating User-Agent block rule: ", err)
		return err
	}

	output := [][]string{
		formatUserAgentRule(resp.Result),
	}

	writeTable(c, output, "ID", "Description", "Mode", "Value", "Paused")

	return nil
}

func userAgentDelete(c *cli.Context) error {
	if err := checkFlags(c, "zone", "id"); err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	resp, err := api.DeleteUserAgentRule(context.Background(), zoneID, c.String("id"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error deleting User-Agent block rule: ", err)
		return err
	}

	output := [][]string{
		formatUserAgentRule(resp.Result),
	}

	writeTable(c, output, "ID", "Description", "Mode", "Value", "Paused")

	return nil
}

func userAgentList(c *cli.Context) error {
	if err := checkFlags(c, "zone", "page"); err != nil {
		return err
	}

	zoneID, err := api.ZoneIDByName(c.String("zone"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}

	resp, err := api.ListUserAgentRules(context.Background(), zoneID, c.Int("page"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listing User-Agent block rules: ", err)
		return err
	}

	output := make([][]string, 0, len(resp.Result))
	for _, rule := range resp.Result {
		output = append(output, formatUserAgentRule(rule))
	}

	writeTable(c, output, "ID", "Description", "Mode", "Value", "Paused")

	return nil
}
