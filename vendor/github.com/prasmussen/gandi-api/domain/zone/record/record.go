package record

import (
	"github.com/prasmussen/gandi-api/client"
)

type Record struct {
	*client.Client
}

func New(c *client.Client) *Record {
	return &Record{c}
}

// Count number of records for a given zone/version
func (self *Record) Count(zoneId, version int64) (int64, error) {
	var result int64
	params := []interface{}{self.Key, zoneId, version}
	if err := self.Call("domain.zone.record.count", params, &result); err != nil {
		return -1, err
	}
	return result, nil
}

// List records of a version of a DNS zone
func (self *Record) List(zoneId, version int64) ([]*RecordInfo, error) {
	opts := &struct {
		Page int `xmlrpc:"page"`
	}{0}
	const perPage = 100
	params := []interface{}{self.Key, zoneId, version, opts}
	records := make([]*RecordInfo, 0)
	for {
		var res []interface{}
		if err := self.Call("domain.zone.record.list", params, &res); err != nil {
			return nil, err
		}
		for _, r := range res {
			record := ToRecordInfo(r.(map[string]interface{}))
			records = append(records, record)
		}
		if len(res) < perPage {
			break
		}
		opts.Page++
	}
	return records, nil
}

// Add a new record to zone
func (self *Record) Add(args RecordAdd) (*RecordInfo, error) {
	var res map[string]interface{}
	createArgs := map[string]interface{}{
		"name":  args.Name,
		"type":  args.Type,
		"value": args.Value,
		"ttl":   args.Ttl,
	}

	params := []interface{}{self.Key, args.Zone, args.Version, createArgs}
	if err := self.Call("domain.zone.record.add", params, &res); err != nil {
		return nil, err
	}
	return ToRecordInfo(res), nil
}

// Remove a record from a zone/version
func (self *Record) Delete(zoneId, version, recordId int64) (bool, error) {
	var res int64
	deleteArgs := map[string]interface{}{"id": recordId}
	params := []interface{}{self.Key, zoneId, version, deleteArgs}
	if err := self.Call("domain.zone.record.delete", params, &res); err != nil {
		return false, err
	}
	return (res == 1), nil
}

// Update a record from zone/version
func (self *Record) Update(args RecordUpdate) ([]*RecordInfo, error) {
	var res []interface{}
	updateArgs := map[string]interface{}{
		"name":  args.Name,
		"type":  args.Type,
		"value": args.Value,
		"ttl":   args.Ttl,
	}
	updateOpts := map[string]string{
		"id": args.Id,
	}

	params := []interface{}{self.Key, args.Zone, args.Version, updateOpts, updateArgs}
	if err := self.Call("domain.zone.record.update", params, &res); err != nil {
		return nil, err
	}

	records := make([]*RecordInfo, 0)
	for _, r := range res {
		record := ToRecordInfo(r.(map[string]interface{}))
		records = append(records, record)
	}
	return records, nil
}

// SetRecords replaces the entire zone with new records.
func (self *Record) SetRecords(zone_id, version_id int64, args []RecordSet) ([]*RecordInfo, error) {
	var res []interface{}

	params := []interface{}{self.Key, zone_id, version_id, args}
	if err := self.Call("domain.zone.record.set", params, &res); err != nil {
		return nil, err
	}

	records := make([]*RecordInfo, 0)
	for _, r := range res {
		record := ToRecordInfo(r.(map[string]interface{}))
		records = append(records, record)
	}
	return records, nil
}

//// Set the current zone of a domain
//func (self *Record) Set(domainName string, id int64) (*domain.DomainInfo, error) {
//    var res map[string]interface{}
//    params := []interface{}{self.Key, domainName, id}
//    if err := self.zone.set", params, &res); err != nil {
//        return nil, err
//    }
//    return domain.ToDomainInfo(res), nil
//}
