package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
	"github.com/urfave/cli/v2"
)

func formatAccessRule(rule cloudflare.AccessRule) []string {
	return []string{
		rule.ID,
		rule.Configuration.Value,
		rule.Scope.Type,
		rule.Mode,
		rule.Notes,
	}
}

func firewallAccessRules(c *cli.Context) error {
	accountID, zoneID, err := getScope(c)
	if err != nil {
		return err
	}

	// Create an empty access rule for searching for rules
	rule := cloudflare.AccessRule{
		Configuration: getConfiguration(c),
	}
	if c.String("scope-type") != "" {
		rule.Scope.Type = c.String("scope-type")
	}
	if c.String("notes") != "" {
		rule.Notes = c.String("notes")
	}
	if c.String("mode") != "" {
		rule.Mode = c.String("mode")
	}

	var response *cloudflare.AccessRuleListResponse
	switch {
	case accountID != "":
		response, err = api.ListAccountAccessRules(context.Background(), accountID, rule, 1)
	case zoneID != "":
		response, err = api.ListZoneAccessRules(context.Background(), zoneID, rule, 1)
	default:
		response, err = api.ListUserAccessRules(context.Background(), rule, 1)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	totalPages := response.ResultInfo.TotalPages
	rules := make([]cloudflare.AccessRule, 0, response.ResultInfo.Total)
	rules = append(rules, response.Result...)
	if totalPages > 1 {
		for page := 2; page <= totalPages; page++ {
			switch {
			case accountID != "":
				response, err = api.ListAccountAccessRules(context.Background(), accountID, rule, page)
			case zoneID != "":
				response, err = api.ListZoneAccessRules(context.Background(), zoneID, rule, page)
			default:
				response, err = api.ListUserAccessRules(context.Background(), rule, page)
			}
			if err != nil {
				fmt.Println(err)
				return err
			}
			rules = append(rules, response.Result...)
		}
	}

	output := make([][]string, 0, len(rules))
	for _, rule := range rules {
		output = append(output, formatAccessRule(rule))
	}
	writeTable(c, output, "ID", "Value", "Scope", "Mode", "Notes")

	return nil
}

func firewallAccessRuleCreate(c *cli.Context) error {
	if err := checkFlags(c, "mode", "value"); err != nil {
		fmt.Println(err)
		return err
	}
	accountID, zoneID, err := getScope(c)
	if err != nil {
		return err
	}
	configuration := getConfiguration(c)
	mode := c.String("mode")
	notes := c.String("notes")

	rule := cloudflare.AccessRule{
		Configuration: configuration,
		Mode:          mode,
		Notes:         notes,
	}

	var (
		rules []cloudflare.AccessRule
	)

	switch {
	case accountID != "":
		resp, err := api.CreateAccountAccessRule(context.Background(), accountID, rule)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating account access rule: ", err)
			return err
		}
		rules = append(rules, resp.Result)
	case zoneID != "":
		resp, err := api.CreateZoneAccessRule(context.Background(), zoneID, rule)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating zone access rule: ", err)
			return err
		}
		rules = append(rules, resp.Result)
	default:
		resp, err := api.CreateUserAccessRule(context.Background(), rule)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error creating user access rule: ", err)
			return err
		}
		rules = append(rules, resp.Result)
	}

	output := make([][]string, 0, len(rules))
	for _, rule := range rules {
		output = append(output, formatAccessRule(rule))
	}
	writeTable(c, output, "ID", "Value", "Scope", "Mode", "Notes")

	return nil
}

func firewallAccessRuleUpdate(c *cli.Context) error {
	if err := checkFlags(c, "id"); err != nil {
		fmt.Println(err)
		return err
	}
	id := c.String("id")
	accountID, zoneID, err := getScope(c)
	if err != nil {
		return err
	}
	mode := c.String("mode")
	notes := c.String("notes")

	rule := cloudflare.AccessRule{
		Mode:  mode,
		Notes: notes,
	}

	var (
		rules       []cloudflare.AccessRule
		errUpdating = "error updating firewall access rule"
	)
	switch {
	case accountID != "":
		resp, err := api.UpdateAccountAccessRule(context.Background(), accountID, id, rule)
		if err != nil {
			return fmt.Errorf(errUpdating+": %w", err)
		}
		rules = append(rules, resp.Result)
	case zoneID != "":
		resp, err := api.UpdateZoneAccessRule(context.Background(), zoneID, id, rule)
		if err != nil {
			return fmt.Errorf(errUpdating+": %w", err)
		}
		rules = append(rules, resp.Result)
	default:
		resp, err := api.UpdateUserAccessRule(context.Background(), id, rule)
		if err != nil {
			return fmt.Errorf(errUpdating+": %w", err)
		}
		rules = append(rules, resp.Result)
	}

	output := make([][]string, 0, len(rules))
	for _, rule := range rules {
		output = append(output, formatAccessRule(rule))
	}
	writeTable(c, output, "ID", "Value", "Scope", "Mode", "Notes")

	return nil
}

func firewallAccessRuleCreateOrUpdate(c *cli.Context) error {
	if err := checkFlags(c, "mode", "value"); err != nil {
		fmt.Println(err)
		return err
	}
	accountID, zoneID, err := getScope(c)
	if err != nil {
		return err
	}
	configuration := getConfiguration(c)
	mode := c.String("mode")
	notes := c.String("notes")

	// Look for an existing record
	rule := cloudflare.AccessRule{
		Configuration: configuration,
	}
	var response *cloudflare.AccessRuleListResponse
	switch {
	case accountID != "":
		response, err = api.ListAccountAccessRules(context.Background(), accountID, rule, 1)
	case zoneID != "":
		response, err = api.ListZoneAccessRules(context.Background(), zoneID, rule, 1)
	default:
		response, err = api.ListUserAccessRules(context.Background(), rule, 1)
	}
	if err != nil {
		fmt.Println("Error creating or updating firewall access rule:", err)
		return err
	}

	rule.Mode = mode
	rule.Notes = notes
	if len(response.Result) > 0 {
		for _, r := range response.Result {
			if mode == "" {
				rule.Mode = r.Mode
			}
			if notes == "" {
				rule.Notes = r.Notes
			}
			switch {
			case accountID != "":
				_, err = api.UpdateAccountAccessRule(context.Background(), accountID, r.ID, rule)
			case zoneID != "":
				_, err = api.UpdateZoneAccessRule(context.Background(), zoneID, r.ID, rule)
			default:
				_, err = api.UpdateUserAccessRule(context.Background(), r.ID, rule)
			}
			if err != nil {
				fmt.Println("Error updating firewall access rule:", err)
			}
		}
	} else {
		switch {
		case accountID != "":
			_, err = api.CreateAccountAccessRule(context.Background(), accountID, rule)
		case zoneID != "":
			_, err = api.CreateZoneAccessRule(context.Background(), zoneID, rule)
		default:
			_, err = api.CreateUserAccessRule(context.Background(), rule)
		}
		if err != nil {
			fmt.Println("Error creating firewall access rule:", err)
		}
	}

	return nil
}

func firewallAccessRuleDelete(c *cli.Context) error {
	if err := checkFlags(c, "id"); err != nil {
		fmt.Println(err)
		return err
	}
	ruleID := c.String("id")

	accountID, zoneID, err := getScope(c)
	if err != nil {
		return err
	}

	var (
		rules       []cloudflare.AccessRule
		errDeleting = "error deleting firewall access rule"
	)
	switch {
	case accountID != "":
		resp, err := api.DeleteAccountAccessRule(context.Background(), accountID, ruleID)
		if err != nil {
			return fmt.Errorf(errDeleting+": %w", err)
		}
		rules = append(rules, resp.Result)
	case zoneID != "":
		resp, err := api.DeleteZoneAccessRule(context.Background(), zoneID, ruleID)
		if err != nil {
			return fmt.Errorf(errDeleting+": %w", err)
		}
		rules = append(rules, resp.Result)
	default:
		resp, err := api.DeleteUserAccessRule(context.Background(), ruleID)
		if err != nil {
			return fmt.Errorf(errDeleting+": %w", err)
		}
		rules = append(rules, resp.Result)
	}
	if err != nil {
		fmt.Println("Error deleting firewall access rule:", err)
	}

	output := make([][]string, 0, len(rules))
	for _, rule := range rules {
		output = append(output, formatAccessRule(rule))
	}
	writeTable(c, output, "ID", "Value", "Scope", "Mode", "Notes")

	return nil
}

func getScope(c *cli.Context) (string, string, error) {
	var account, accountID string
	if c.String("account") != "" {
		account = c.String("account")
		params := cloudflare.AccountsListParams{}
		accounts, _, err := api.Accounts(context.Background(), params)
		if err != nil {
			fmt.Println(err)
			return "", "", err
		}
		for _, acc := range accounts {
			if acc.Name == account {
				accountID = acc.ID
				break
			}
		}
		if accountID == "" {
			err := errors.New("account could not be found")
			fmt.Println(err)
			return "", "", err
		}
	}

	var zone, zoneID string
	if c.String("zone") != "" {
		zone = c.String("zone")
		id, err := api.ZoneIDByName(zone)
		if err != nil {
			fmt.Println(err)
			return "", "", err
		}
		zoneID = id
	}

	if zoneID != "" && accountID != "" {
		err := errors.New("Cannot specify both --zone and --account")
		fmt.Println(err)
		return "", "", err
	}

	return accountID, zoneID, nil
}

func getConfiguration(c *cli.Context) cloudflare.AccessRuleConfiguration {
	configuration := cloudflare.AccessRuleConfiguration{}
	if c.String("value") != "" {
		ip := net.ParseIP(c.String("value"))
		_, cidr, cidrErr := net.ParseCIDR(c.String("value"))
		_, asnErr := strconv.ParseInt(c.String("value"), 10, 32)
		if ip != nil {
			configuration.Target = "ip"
			configuration.Value = ip.String()
		} else if cidrErr == nil {
			cidr.IP = cidr.IP.Mask(cidr.Mask)
			configuration.Target = "ip_range"
			configuration.Value = cidr.String()
		} else if asnErr == nil {
			configuration.Target = "asn"
			configuration.Value = c.String("value")
		} else {
			configuration.Target = "country"
			configuration.Value = c.String("value")
		}
	}
	return configuration
}
