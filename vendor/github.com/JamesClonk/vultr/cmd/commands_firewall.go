package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/jawher/mow.cli"
)

func firewallGroupCreate(cmd *cli.Cmd) {
	cmd.Spec = "[DESCRIPTION]"

	desc := cmd.StringArg("DESCRIPTION", "", "Optional description for the new group")

	cmd.Action = func() {
		id, err := GetClient().CreateFirewallGroup(*desc)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Firewall group created\n\n")
		lengths := []int{10, 64}
		tabsPrint(columns{"GROUP_ID", "DESCRIPTION"}, lengths)
		tabsPrint(columns{id, *desc}, lengths)
		tabsFlush()
	}
}

func firewallGroupDelete(cmd *cli.Cmd) {
	cmd.Spec = "GROUP_ID"

	gid := cmd.StringArg("GROUP_ID", "", "Firewall group ID")

	cmd.Action = func() {
		if err := GetClient().DeleteFirewallGroup(*gid); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Firewall group %s deleted\n", *gid)
	}
}

func firewallGroupSetDescription(cmd *cli.Cmd) {
	cmd.Spec = "GROUP_ID DESCRIPTION"

	gid := cmd.StringArg("GROUP_ID", "", "Firewall group ID")
	desc := cmd.StringArg("DESCRIPTION", "", "New description for the firewall group")

	cmd.Action = func() {
		if err := GetClient().SetFirewallGroupDescription(*gid, *desc); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Set description for firewall group %s: %s\n", *gid, *desc)
	}
}

func firewallGroupList(cmd *cli.Cmd) {
	cmd.Action = func() {
		groups, err := GetClient().GetFirewallGroups()
		if err != nil {
			log.Fatal(err)
		}

		if len(groups) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{10, 64, 12, 16}
		tabsPrint(columns{"GROUP_ID", "DESCRIPTION", "RULE_COUNT", "INSTANCE_COUNT"}, lengths)
		for _, g := range groups {
			tabsPrint(columns{
				g.ID,
				g.Description,
				g.RuleCount,
				g.InstanceCount,
			}, lengths)
		}
		tabsFlush()
	}
}

func firewallRuleCreate(cmd *cli.Cmd) {
	cmd.Spec = "-g -n ((--tcp --port) | (--udp --port) | --icmp | --gre)"
	gid := cmd.StringOpt("g group-id", "", "Firewall group ID (see <firewall group list>)")
	cidr := cmd.StringOpt("n network", "0.0.0.0/0", "IPv4/IPv6 network in CIDR notation")
	tcp := cmd.BoolOpt("tcp", false, "TCP protocol")
	udp := cmd.BoolOpt("udp", false, "UDP protocol")
	icmp := cmd.BoolOpt("icmp", false, "ICMP protocol")
	gre := cmd.BoolOpt("gre", false, "GRE protocol")
	port := cmd.StringOpt("port", "", "Port number or port range (TCP/UDP only)")

	cmd.Action = func() {
		var protocol string
		switch {
		case *tcp:
			protocol = "tcp"
		case *udp:
			protocol = "udp"
		case *icmp:
			protocol = "icmp"
		case *gre:
			protocol = "gre"
		}

		_, network, err := net.ParseCIDR(*cidr)
		if err != nil {
			log.Fatalf("Invalid network CIDR: %s", *cidr)
		}

		ruleNum, err := GetClient().CreateFirewallRule(*gid, protocol, *port, network)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Firewall rule created\n\n")
		lengths := []int{10, 10, 10, 12, 20}
		tabsPrint(columns{"GROUP_ID", "RULE_NUM", "PROTOCOL", "PORT", "NETWORK"}, lengths)
		tabsPrint(columns{*gid, ruleNum, protocol, *port, network}, lengths)
		tabsFlush()
	}
}

func firewallRuleDelete(cmd *cli.Cmd) {
	cmd.Spec = "GROUP_ID RULE_NUM"

	gid := cmd.StringArg("GROUP_ID", "", "Firewall group ID")
	rule := cmd.IntArg("RULE_NUM", 0, "Firewall rule number")

	cmd.Action = func() {
		if err := GetClient().DeleteFirewallRule(*rule, *gid); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Firewall rule %d in group %s deleted\n", *rule, *gid)
	}
}

func firewallRuleList(cmd *cli.Cmd) {
	cmd.Spec = "GROUP_ID"

	gid := cmd.StringArg("GROUP_ID", "", "Firewall group ID (see <firewall group list>)")

	cmd.Action = func() {
		rules, err := GetClient().GetFirewallRules(*gid)
		if err != nil {
			log.Fatal(err)
		}

		if len(rules) == 0 {
			fmt.Println()
			return
		}

		lengths := []int{10, 10, 8, 12, 20}
		tabsPrint(columns{"RULE_NUM", "ACTION", "PROTOCOL", "PORT", "NETWORK"}, lengths)
		for _, r := range rules {
			tabsPrint(columns{
				r.RuleNumber,
				r.Action,
				r.Protocol,
				r.Port,
				r.Network.String(),
			}, lengths)
		}
		tabsFlush()
	}
}
