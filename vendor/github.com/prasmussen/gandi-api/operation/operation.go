package operation

import "github.com/prasmussen/gandi-api/client"

type Operation struct {
	*client.Client
}

func New(c *client.Client) *Operation {
	return &Operation{c}
}

// Count operations created by this contact
func (self *Operation) Count() (int64, error) {
	var result int64
	// params := Params{Params: []interface{}{self.Key}}
	params := []interface{}{self.Key}
	if err := self.Call("operation.count", params, &result); err != nil {
		return -1, err
	}
	return result, nil
}

// Get operation information
func (self *Operation) Info(id int64) (*OperationInfo, error) {
	var res map[string]interface{}
	// params := Params{Params: []interface{}{self.Key, id}}
	params := []interface{}{self.Key, id}
	if err := self.Call("operation.info", params, &res); err != nil {
		return nil, err
	}
	return ToOperationInfo(res), nil
}

// Cancel an operation
func (self *Operation) Cancel(id int64) (bool, error) {
	var res bool
	// params := Params{Params: []interface{}{self.Key, id}}
	params := []interface{}{self.Key, id}
	if err := self.Call("operation.cancel", params, &res); err != nil {
		return false, err
	}
	return res, nil
}

// List operations created by this contact
func (self *Operation) List() ([]*OperationInfo, error) {
	var res []interface{}
	// params := Params{Params: []interface{}{self.Key}}
	params := []interface{}{self.Key}
	if err := self.Call("operation.list", params, &res); err != nil {
		return nil, err
	}

	operations := make([]*OperationInfo, len(res), len(res))
	for i, r := range res {
		operations[i] = ToOperationInfo(r.(map[string]interface{}))
	}
	return operations, nil
}
