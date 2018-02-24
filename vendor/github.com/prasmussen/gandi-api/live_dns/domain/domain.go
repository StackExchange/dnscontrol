package domain

import (
	"fmt"

	"github.com/prasmussen/gandi-api/client"
	"github.com/prasmussen/gandi-api/live_dns/record"
)

// Domain holds the domain client stucture
type Domain struct {
	*client.Client
}

// New instanciates a new Domain client
func New(c *client.Client) *Domain {
	return &Domain{c}
}

// List domains associated to the contact represented by apikey
func (d *Domain) List() (domains []*InfoBase, err error) {
	_, err = d.Get("/domains", &domains)
	return
}

// Info Gets domain information
func (d *Domain) Info(name string) (infos *Info, err error) {
	_, err = d.Get(fmt.Sprintf("/domains/%s", name), &infos)
	return
}

// Records gets a record client for the current domain
func (d *Domain) Records(name string) record.Manager {
	return record.New(d.Client, fmt.Sprintf("/domains/%s", name))
}
