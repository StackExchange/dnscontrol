package contact

import "github.com/prasmussen/gandi-api/client"

type Contact struct {
	*client.Client
}

func New(c *client.Client) *Contact {
	return &Contact{c}
}

// Get contact financial balance
func (self *Contact) Balance() (*BalanceInformation, error) {
	var res map[string]interface{}
	params := []interface{}{self.Key}
	if err := self.Call("contact.balance", params, &res); err != nil {
		return nil, err
	}
	return toBalanceInformation(res), nil
}

// Get contact information
func (self *Contact) Info(handle string) (*ContactInformation, error) {
	var res map[string]interface{}

	var params []interface{}
	if handle == "" {
		params = []interface{}{self.Key}
	} else {
		params = []interface{}{self.Key, handle}
	}
	if err := self.Call("contact.info", params, &res); err != nil {
		return nil, err
	}
	return toContactInformation(res), nil
}

// Create a contact
func (self *Contact) Create(opts ContactCreate) (*ContactInformation, error) {
	var res map[string]interface{}
	createArgs := map[string]interface{}{
		"given":      opts.Firstname,
		"family":     opts.Lastname,
		"email":      opts.Email,
		"password":   opts.Password,
		"streetaddr": opts.Address,
		"zip":        opts.Zipcode,
		"city":       opts.City,
		"country":    opts.Country,
		"phone":      opts.Phone,
		"type":       opts.ContactType(),
	}

	params := []interface{}{self.Key, createArgs}
	if err := self.Call("contact.create", params, &res); err != nil {
		return nil, err
	}
	return toContactInformation(res), nil
}

// Delete a contact
func (self *Contact) Delete(handle string) (bool, error) {
	var res bool

	var params []interface{}
	if handle == "" {
		params = []interface{}{self.Key}
	} else {
		params = []interface{}{self.Key, handle}
	}
	if err := self.Call("contact.delete", params, &res); err != nil {
		return false, err
	}
	return res, nil
}
