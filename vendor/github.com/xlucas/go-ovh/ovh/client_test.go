package ovh_test

import (
	"testing"

	"github.com/xlucas/go-ovh/ovh"
)

func TestPollTimeshiftCaOvhCom(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_CA_OVHCOM, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftCaKimsufi(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_CA_KIMSUFI, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftCaRunAbove(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_CA_RUNABOVE, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftCaSoYouStart(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_CA_SOYOUSTART, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftEuOvhCom(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_EU_OVHCOM, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftEuKimsufi(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_EU_KIMSUFI, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}

func TestPollTimeshiftEuSoYouStart(t *testing.T) {
	c := ovh.NewClient(ovh.ENDPOINT_EU_SOYOUSTART, "", "", "", false)
	if err := c.PollTimeshift(); err != nil {
		t.Fatal(err)
	}
}
