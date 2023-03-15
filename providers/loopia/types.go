package loopia

import (
	"encoding/xml"
	"fmt"
)

// types for XML-RPC method calls and parameters

type param interface {
	param()
}

type paramString struct {
	XMLName xml.Name `xml:"param"`
	Value   string   `xml:"value>string"`
}

func (p paramString) param() {}

type paramInt struct {
	XMLName xml.Name `xml:"param"`
	Value   uint32   `xml:"value>int"`
}

func (p paramInt) param() {}

// payload of values for a subdomain record
type paramStruct struct {
	XMLName       xml.Name       `xml:"param"`
	StructMembers []structMember `xml:"value>struct>member"`
}

func (p paramStruct) param() {}

type structMember interface {
	structMember()
}

type structMemberString struct {
	Name  string `xml:"name"`
	Value string `xml:"value>string"`
}

func (m structMemberString) structMember() {}

type structMemberInt struct {
	Name  string `xml:"name"`
	Value uint32 `xml:"value>int"`
}

func (m structMemberInt) structMember() {}

type structMemberBool struct {
	Name  string `xml:"name"`
	Value bool   `xml:"value>boolean"`
}

func (m structMemberBool) structMember() {}

type methodCall struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`
	Params     []param  `xml:"params>param"`
}

// types for XML-RPC responses

type response interface {
	faultCode() uint32
	faultString() string
}

type responseString struct {
	responseFault
	Value string `xml:"params>param>value>string"`
}

type responseFault struct {
	FaultCode   uint32 `xml:"fault>value>struct>member>value>int"`
	FaultString string `xml:"fault>value>struct>member>value>string"`
}

func (r responseFault) faultCode() uint32   { return r.FaultCode }
func (r responseFault) faultString() string { return r.FaultString }

type rpcError struct {
	faultCode   uint32
	faultString string
}

func (e rpcError) Error() string {
	return fmt.Sprintf("RPC Error: (%d) %s", e.faultCode, e.faultString)
}

type zRec struct {
	// "name"         "value"
	Type     string
	Rdata    string
	Priority uint16 // the next 4 ints are 112 bits
	TTL      uint32
	RecordID uint32
}

// zoneRecord and domainObject are synonymous (but belong to different parts of
// XML structure in req/resp). Can we unify them?
type zoneRecord struct {
	XMLName    xml.Name   `xml:"struct"`
	Properties []Property `xml:"member"`
	// Properties map[string]interface{}
}

type zoneRecordsResponse struct {
	responseFault
	XMLName     xml.Name     `xml:"methodResponse"`
	ZoneRecords []zoneRecord `xml:"params>param>value>array>data>value>struct"`
}

type Property struct {
	Key   string `xml:"name"`
	Value Value  `xml:"value"`
}

type Value struct {
	// String string `xml:",any"`
	String string `xml:"string"`
	Int    int    `xml:"int"`
	Bool   bool   `xml:"bool"`
}

func (p Property) Name() string   { return p.Key }
func (p Property) String() string { return p.Value.String }
func (p Property) Int() int       { return p.Value.Int }
func (p Property) Bool() bool     { return p.Value.Bool }

func (zr *zoneRecord) GetZR() zRec {
	record := zRec{}
	for _, prop := range zr.Properties {
		switch prop.Key {
		case "type":
			record.Type = prop.String()
		case "ttl":
			record.TTL = uint32(prop.Int())
		case "priority":
			record.Priority = uint16(prop.Int())
		case "rdata":
			record.Rdata = prop.String()
		case "record_id":
			record.RecordID = uint32(prop.Int())
		}
	}
	return record
}

func (zrec *zRec) SetZR() zoneRecord {
	//This method creates a zoneRecord to receive from responses.
	return zoneRecord{
		XMLName: xml.Name{Local: "struct"},
		Properties: []Property{
			Property{Key: "type", Value: Value{String: zrec.Type}},
			Property{Key: "ttl", Value: Value{Int: int(zrec.TTL)}},
			Property{Key: "priority", Value: Value{Int: int(zrec.Priority)}},
			Property{Key: "rdata", Value: Value{String: zrec.Rdata}},
			Property{Key: "record_id", Value: Value{Int: int(zrec.RecordID)}},
		},
	}
}

func (zrec *zRec) SetPS() paramStruct {
	//This method creates a paramStruct for sending in requests.
	return paramStruct{
		XMLName: xml.Name{Local: "struct"},
		StructMembers: []structMember{
			structMemberString{Name: "type", Value: zrec.Type},
			structMemberInt{Name: "ttl", Value: uint32(zrec.TTL)},
			structMemberInt{Name: "priority", Value: uint32(zrec.Priority)},
			structMemberString{Name: "rdata", Value: zrec.Rdata},
			structMemberInt{Name: "record_id", Value: uint32(zrec.RecordID)},
		},
	}
}

type domainObject struct {
	XMLName    xml.Name   `xml:"struct"`
	Properties []Property `xml:"member"`
}

// type LoopiaDomainObject struct {
//  "name"         "value"
// 	ReferenceNo    int32
// 	Domain         string
// 	RenewalStatus  string
// 	ExpirationDate string
// 	Paid           bool
// 	Registered     bool
// }

type domainObjectsResponse struct {
	responseFault
	XMLName xml.Name       `xml:"methodResponse"`
	Domains []domainObject `xml:"params>param>value>array>data>value>struct"`
}

type subDomainsResponse struct {
	responseFault
	XMLName xml.Name `xml:"methodResponse"`
	Params  []string `xml:"params>param>value>array>data>value>string"`
}
