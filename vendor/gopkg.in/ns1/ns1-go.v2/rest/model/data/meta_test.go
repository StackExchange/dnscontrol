package data

import "testing"

func TestMeta_StringMap(t *testing.T) {
	meta := &Meta{}
	meta.Up = true
	meta.Latitude = 0.50
	meta.Note = "hello!"
	meta.Longitude = FeedPtr{FeedID: "12345678"}
	meta.Georegion = []interface{}{"US-EAST"}
	meta.Priority = 10
	meta.Weight = 10.0
	meta.Country = []string{"US", "UK"}
	m := meta.StringMap()

	if m["up"].(string) != "1" {
		t.Fatal("up should be 1")
	}

	if m["latitude"].(string) != "0.5" {
		t.Fatal("latitude should be '0.5'")
	}

	if m["georegion"].(string) != "US-EAST" {
		t.Fatal("georegion should be 'US-EAST")
	}

	if m["note"].(string) != "hello!" {
		t.Fatal("note should be 'hello!'")
	}

	if m["priority"].(string) != "10" {
		t.Fatal("priority should be 10")
	}

	if m["weight"].(string) != "10" {
		t.Fatal("weight should be 10")
	}

	if m["country"].(string) != "US,UK" {
		t.Fatal("country should be 'US,UK'")
	}

	expected := `{"feed":"12345678"}`
	if m["longitude"].(string) != expected {
		t.Fatal("longitude should be", expected, "was", m["longitude"].(string))
	}

	meta.Up = false
	m = meta.StringMap()

	if m["up"].(string) != "0" {
		t.Fatal("up should be 0")
	}

	meta.Up = struct{}{}
	defer func(t *testing.T) {
		if r := recover(); r == nil {
			t.Fatal("meta should have panicked but did not")
		}
	}(t)
	meta.StringMap()

}

func TestParseType(t *testing.T) {
	fs := "3.14"

	v := ParseType(fs)

	if _, ok := v.(float64); !ok {
		t.Fatal("value should be float64, was", v)
	}

	cs := "hello,goodbye"

	v = ParseType(cs)

	if _, ok := v.([]string); !ok {
		t.Fatal("value should be []string, was", v)
	}

	is := "42"

	v = ParseType(is)

	if _, ok := v.(int); !ok {
		t.Fatal("value should be int, was", v)
	}

	s := "string value"

	v = ParseType(s)
	if _, ok := v.(string); !ok {
		t.Fatal("value should be string, was", v)
	}
}

func TestMetaFromMap(t *testing.T) {
	m := make(map[string]interface{})

	m["latitude"] = "0.50"
	m["up"] = "1"
	m["connections"] = "5"
	m["longitude"] = `{"feed":"12345678"}`
	meta := MetaFromMap(m)

	if !meta.Up.(bool) {
		t.Fatal("meta.Up should be true")
	}

	if meta.Latitude.(float64) != 0.5 {
		t.Fatal("meta.Latitude should equal 0.5, was", meta.Latitude)
	}

	if meta.Connections.(int) != 5 {
		t.Fatal("meta.Connections should equal 5, was", meta.Connections)
	}

	if meta.Longitude.(FeedPtr).FeedID != "12345678" {
		t.Fatal("meta.Longitude should be a feed ptr with id 12345678, was", meta.Longitude)
	}

	m["up"] = "0"
	meta = MetaFromMap(m)

	if meta.Up.(bool) {
		t.Fatal("meta.Up should be false")
	}
}

func TestGeokeyString(t *testing.T) {
	expected := "AFRICA,ASIAPAC,EUROPE,SOUTH-AMERICA,US-CENTRAL,US-EAST,US-WEST"
	got := geoKeyString()
	if expected != got {
		t.Fatalf("expected '%s', got '%s'", expected, got)
	}
}

func TestMeta_Validate(t *testing.T) {
	m := &Meta{}
	m.Up = true
	m.Georegion = "US-EAST"
	m.Connections = 5
	m.Longitude = 0.80
	m.Latitude = 0.80
	m.Country = "US"
	m.USState = "MA"
	m.CAProvince = "ON"
	m.Note = "Testing out this cool meta validation"
	m.LoadAvg = 3.14
	m.Weight = 40.0
	m.Requests = 10
	m.IPPrefixes = "10.0.0.1/24"
	m.Priority = 1
	errs := m.Validate()
	if len(errs) > 0 {
		t.Fatal("there should be 0 errors, but there were", len(errs), ":", errs)
	}

	m.IPPrefixes = []interface{}{"10.0.0.1/24", "10.0.0.2/24"}
	errs = m.Validate()
	if len(errs) > 0 {
		t.Fatal("there should be 0 errors, but there were", len(errs), ":", errs)
	}

	m.Georegion = "fantasy region"
	m.Up = "bad value"
	m.Connections = -5
	m.Longitude = 10000.0
	m.Latitude = -10000.0
	m.Country = "fantasy land"
	m.USState = "fantasy state"
	m.CAProvince = "quebec"
	m.Note = string(make([]rune, 257))
	m.LoadAvg = -3.14
	m.Weight = -40.0
	m.Requests = -1
	m.IPPrefixes = "1234567"
	m.Priority = -1
	errs = m.Validate()
	if len(errs) != 14 {
		t.Fatal("expected 14 errors, but there were", len(errs), ":", errs)
	}

	m = &Meta{}
	m.Georegion = []interface{}{"US-EAST", "fantasy land"}
	m.Country = []interface{}{"US", "CANADA"}
	m.Up = struct{}{}
	errs = m.Validate()
	if len(errs) != 3 {
		t.Fatal("expected 3 errors, but there were", len(errs), ":", errs)
	}
}
