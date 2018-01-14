package lib

import (
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Firewall_GetGroups_Fail(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, ``)
	defer server.Close()

	_, err := client.GetFirewallGroups()
	if err == nil {
		t.Error(err)
	}
}

func Test_Firewall_GetGroups_NoGroups(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{}`)
	defer server.Close()

	list, err := client.GetFirewallGroups()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, list)
}

func Test_Firewall_GetGroups_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
    "1a":{
        "FIREWALLGROUPID":"1a","description":"test1",
        "date_created":"2017-02-14 17:48:40","date_modified": "2017-02-14 17:48:40",
        "instance_count": 1,"rule_count": 0,"max_rule_count": 50
    },
    "1b":{
        "FIREWALLGROUPID":"1b","description":"test2",
        "date_created":"2017-02-14 17:48:40","date_modified": "2017-02-14 17:48:40",
        "instance_count": 0,"rule_count": 2,"max_rule_count": 0
    }}`)
	defer server.Close()

	groups, err := client.GetFirewallGroups()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, groups) {
		assert.Equal(t, 2, len(groups))

		assert.Equal(t, groups[0].ID, "1a")
		assert.Equal(t, groups[0].Description, "test1")
		assert.Equal(t, groups[0].Created, "2017-02-14 17:48:40")
		assert.Equal(t, groups[0].Modified, "2017-02-14 17:48:40")
		assert.Equal(t, groups[0].InstanceCount, 1)
		assert.Equal(t, groups[0].RuleCount, 0)
		assert.Equal(t, groups[0].MaxRuleCount, 50)

		assert.Equal(t, groups[1].ID, "1b")
		assert.Equal(t, groups[1].Description, "test2")
		assert.Equal(t, groups[1].Created, "2017-02-14 17:48:40")
		assert.Equal(t, groups[1].Modified, "2017-02-14 17:48:40")
		assert.Equal(t, groups[1].InstanceCount, 0)
		assert.Equal(t, groups[1].RuleCount, 2)
		assert.Equal(t, groups[1].MaxRuleCount, 0)
	}
}

func Test_Firewall_CreateGroup_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"FIREWALLGROUPID":"1a"}`)
	defer server.Close()

	id, err := client.CreateFirewallGroup("")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, id, "1a")
}

func Test_Firewall_DeleteGroup_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, ``)
	defer server.Close()

	err := client.DeleteFirewallGroup("1a")
	if err != nil {
		t.Error(err)
	}
}

func Test_Firewall_SetGroupDescription_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.SetFirewallGroupDescription("123456789", "new description")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Firewall_SetGroupDescription_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.SetFirewallGroupDescription("123456789", "new description"))
}

func Test_Firewall_GetRules_Fail(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, ``)
	defer server.Close()

	_, err := client.GetFirewallRules("1a")
	if err == nil {
		t.Error(err)
	}
}

func Test_Firewall_GetRules_NoRules(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{}`)
	defer server.Close()

	list, err := client.GetFirewallRules("1a")
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, list)
}

func Test_Firewall_GetRules_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
    "1":{
        "rulenumber":1,"action":"accept","protocol": "icmp","port": "",
        "subnet": "","subnet_size": 0
    },
    "2":{
        "rulenumber":2,"action":"accept","protocol": "tcp","port": "80",
        "subnet": "10.234.22.0","subnet_size": 24
    },
    "3":{
        "rulenumber":3,"action":"accept","protocol": "tcp","port": "80",
	"subnet": "::","subnet_size": 0
    }}`)
	defer server.Close()

	rules, err := client.GetFirewallRules("1a")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, rules) {
		assert.Equal(t, 3, len(rules))

		assert.Equal(t, rules[0].RuleNumber, 1)
		assert.Equal(t, rules[0].Action, "accept")
		assert.Equal(t, rules[0].Protocol, "icmp")
		assert.Equal(t, rules[0].Port, "")
		_, netw, _ := net.ParseCIDR("0.0.0.0/0")
		assert.Equal(t, rules[0].Network, netw)

		assert.Equal(t, rules[1].RuleNumber, 2)
		assert.Equal(t, rules[1].Action, "accept")
		assert.Equal(t, rules[1].Protocol, "tcp")
		assert.Equal(t, rules[1].Port, "80")
		_, netw, _ = net.ParseCIDR("10.234.22.0/24")
		assert.Equal(t, rules[1].Network, netw)

		assert.Equal(t, rules[2].RuleNumber, 3)
		assert.Equal(t, rules[2].Action, "accept")
		assert.Equal(t, rules[2].Protocol, "tcp")
		assert.Equal(t, rules[2].Port, "80")
		_, netw, _ = net.ParseCIDR("::/0")
		assert.Equal(t, rules[2].Network, netw)
	}
}

func Test_Firewall_CreateRule_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"rulenumber": 2}`)
	defer server.Close()

	_, netw, _ := net.ParseCIDR("10.234.22.0/24")
	num, err := client.CreateFirewallRule("1a", "tcp", "80", netw)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, num, 2)
}

func Test_Firewall_DeleteRule_Ok(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, ``)
	defer server.Close()

	err := client.DeleteFirewallRule(2, "1a")
	if err != nil {
		t.Error(err)
	}
}
