package opensrs

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Header struct {
	XMLName xml.Name `xml:"header"`
	Version string   `xml:"version"`
}

type Item struct {
	XMLName xml.Name `xml:"item"`
	Key     string   `xml:"key,attr"`
	DtArray *DtArray `xml:"dt_array,omitempty"`
	DtAssoc *DtAssoc `xml:"dt_assoc,omitempty"`
	Value   string   `xml:",chardata"`
}

func (i *Item) decode() interface{} {
	if i.DtAssoc != nil {
		return i.DtAssoc.decode()
	}
	if i.DtArray != nil {
		return i.DtArray.decode()
	}
	return i.Value
}

type DtArray struct {
	XMLName  xml.Name `xml:"dt_array"`
	ItemList []Item   `xml:"item,omitempty"`
}

func (d *DtArray) decode() []interface{} {
	m := make([]interface{}, 0)
	for _, element := range d.ItemList {
		m = append(m, element.decode())
	}
	return m
}

type DtAssoc struct {
	XMLName  xml.Name `xml:"dt_assoc"`
	ItemList []Item   `xml:"item,omitempty"`
}

func (d *DtAssoc) decode() Map {
	m := make(Map)
	for _, element := range d.ItemList {
		m[element.Key] = element.decode()
	}
	return m
}

type DataBlock struct {
	XMLName xml.Name `xml:"data_block"`
	DtAssoc *DtAssoc `xml:"dt_assoc,omitempty"`
	//DtArray DtArray `xml:"dt_array,omitempty"`
}

type Map map[string]interface{}

func (d *DataBlock) decode() Map {
	m := make(Map)
	if d.DtAssoc != nil {
		return d.DtAssoc.decode()
	}
	return m
}

func encodeItem(key string, value reflect.Value) Item {
	item := Item{}
	item.Key = key
	v := internalEncode(value)
	s, ok := v.(string)
	if ok {
		item.Value = s
	}
	dtass, ok := v.(DtAssoc)
	if ok {
		item.DtAssoc = &dtass
	}
	dtarr, ok := v.(DtArray)
	if ok {
		item.DtArray = &dtarr
	}
	return item
}

func internalEncode(v reflect.Value) (p interface{}) {
	t := v.Type()
	switch t.Kind() {
	case reflect.Interface:
		return internalEncode(v.Elem())
	case reflect.String:
		return v.Interface().(string)
	case reflect.Struct:
		dt := DtAssoc{}

		for i := 0; i < t.NumField(); i++ {
			key := strings.ToLower(t.Field(i).Name)
			value := v.Field(i)
			item := encodeItem(key, value)
			dt.ItemList = append(dt.ItemList, item)
		}
		return dt
	case reflect.Map: // DtAssoc
		dt := DtAssoc{}

		for _, k := range v.MapKeys() {
			v := v.MapIndex(k)
			key := k.String()
			item := encodeItem(key, v)
			dt.ItemList = append(dt.ItemList, item)
		}
		return dt
	case reflect.Slice: //DtArray
		dt := DtArray{}
		for i := 0; i < v.Len(); i++ {
			key := strconv.Itoa(i)
			value := v.Index(i)
			item := encodeItem(key, value)
			dt.ItemList = append(dt.ItemList, item)
		}
		return dt
	default:
		log.Println("FAIL, unknown type", t.Kind())
	}
	return nil
}

type Body struct {
	XMLName   xml.Name  `xml:"body"`
	DataBlock DataBlock `xml:"data_block"`
}

type OPSEnvelope struct {
	XMLName xml.Name `xml:"OPS_envelope"`
	Header  Header   `xml:"header"`
	Body    Body     `xml:"body"`
}

func FromXml(b []byte, v interface{}) error {
	var q OPSEnvelope
	err := xml.Unmarshal(b, &q)
	if err != nil {
		return err
	}
	m := q.Body.DataBlock.decode()
	jsonString, _ := json.Marshal(m)

	return json.Unmarshal(jsonString, &v)

}

func ToXml(v interface{}) (b []byte, err error) {
	jsonString, _ := json.Marshal(v)
	var m interface{}
	json.Unmarshal(jsonString, &m)

	q := OPSEnvelope{Header: Header{Version: "0.9"}, Body: Body{}}
	dtass, ok := internalEncode(reflect.ValueOf(m)).(DtAssoc)
	if ok {
		q.Body.DataBlock.DtAssoc = &dtass
	} else {
		return nil, errors.New("Encoding failed")
	}
	return xml.MarshalIndent(q, "", " ")
}
