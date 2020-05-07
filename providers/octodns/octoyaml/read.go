package octoyaml

/*
This module handles reading OctoDNS yaml files.  Sadly the YAML files
are so entirely flexible that parsing them is a nighmare.  We UnMarshalYAML
them into a slice of interfaces mapped to interfaces, then use reflection
to walk the tree, interpreting what we find along the way.  As we collect
data we output models.RecordConfig objects.
*/

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"

	yaml "gopkg.in/yaml.v2"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// ReadYaml parses a yaml input and returns a list of RecordConfigs
func ReadYaml(r io.Reader, origin string) (models.Records, error) {
	results := models.Records{}

	// Slurp the YAML into a string.
	ydata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("can not read yaml filehandle: %w", err)
	}

	// Unmarshal the mystery data into a structure we can relect into.
	var mysterydata map[string]interface{}
	err = yaml.Unmarshal(ydata, &mysterydata)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml: %w", err)
	}
	//fmt.Printf("ReadYaml: mysterydata == %v\n", mysterydata)

	// Traverse every key/value pair.
	for k, v := range mysterydata { // Each label
		// k, v: k is the label, v is everything we know about the label.
		// In other code, k1, v2 refers to one level deeper, k3, k3 refers to
		// one more level deeper, and so on.
		//fmt.Printf("ReadYaml: NEXT KEY\n")
		//fmt.Printf("ReadYaml:  KEY=%s v.(type)=%s\n", k, reflect.TypeOf(v).String())
		switch v.(type) {
		case map[interface{}]interface{}:
			// The value is itself a map. This means we have a label with
			// with one or more records, each of them are all the same rtype.
			// parseLeaf will handle both of these forms:
			// For example, this:
			// 'www':
			//    type: A
			//    values:
			//      - 1.2.3.4
			//      - 1.2.3.5
			// or
			// 'www':
			//    type: CNAME
			//    value: foo.example.com.
			results, err = parseLeaf(results, k, v, origin)
			if err != nil {
				return results, fmt.Errorf("leaf (%v) error: %w", v, err)
			}
		case []interface{}:
			// The value is a list. This means we have a label with
			// multiple records, each of them may be different rtypes.
			// We need to call parseLeaf() once for each rtype.
			// For example, this:
			// 'www':
			// - type: A
			//   values:
			// 	- 1.2.3.4
			// 	- 1.2.3.5
			// - type: MX
			//   values:
			// 	- priority: 10
			// 	  value: mx1.example.com.
			// 	- priority: 10
			// 	  value: mx2.example.com.
			for i, v3 := range v.([]interface{}) { // All the label's list
				_ = i
				//fmt.Printf("ReadYaml:   list key=%s i=%d v3.(type)=%s\n", k, i, typeof(v3))
				switch v3.(type) {
				case map[interface{}]interface{}:
					//fmt.Printf("ReadYaml:   v3=%v\n", v3)
					results, err = parseLeaf(results, k, v3, origin)
					if err != nil {
						return results, fmt.Errorf("leaf v3=%v: %w", v3, err)
					}
				default:
					return nil, fmt.Errorf("unknown type in list3: k=%s v.(type)=%T v=%v", k, v, v)
				}
			}

		default:
			return nil, fmt.Errorf("unknown type in list1: k=%s v.(type)=%T v=%v", k, v, v)
		}
	}

	sortRecs(results, origin)
	//fmt.Printf("ReadYaml: RESULTS=%v\n", results)
	return results, nil
}

func parseLeaf(results models.Records, k string, v interface{}, origin string) (models.Records, error) {
	var rType, rTarget string
	var rTTL uint32
	rTargets := []string{}
	var someresults models.Records
	for k2, v2 := range v.(map[interface{}]interface{}) { // All  the label's items
		// fmt.Printf("ReadYaml: ifs tk2=%s tv2=%s len(rTargets)=%d\n", typeof(k2), typeof(v2), len(rTargets))
		if typeof(k2) == "string" && (typeof(v2) == "string" || typeof(v2) == "int") {
			// The 2nd level key is a string, and the 2nd level value is a string or int.
			// Here are 3 examples:
			// type: CNAME
			// value: foo.example.com.
			// ttl: 3
			//fmt.Printf("parseLeaf:   k2=%s v2=%v\n", k2, v2)
			switch k2.(string) {
			case "type":
				rType = v2.(string)
			case "ttl":
				var err error
				rTTL, err = decodeTTL(v2)
				if err != nil {
					return nil, fmt.Errorf("parseLeaf: can not parse ttl (%v)", v2)
				}
			case "value":
				rTarget = v2.(string)
			case "values":
				switch v2.(type) {
				case string:
					rTarget = v2.(string)
				default:
					return nil, fmt.Errorf("parseLeaf: unknown type in values: rtpe=%s k=%s k2=%s v2.(type)=%T v2=%v", rType, k, k2, v2, v2)
				}
			default:
				panic("Should not happen")
			}
		} else if typeof(k2) == "string" && typeof(v2) == "[]interface {}" {
			// The 2nd level key is a string, and the 2nd level value is a list.
			someresults = nil
			for _, v3 := range v2.([]interface{}) {
				switch v3.(type) {
				case string:
					// Example:
					// values:
					//   - 1.2.3.1
					//   - 1.2.3.2
					//   - 1.2.3.3
					// We collect all the values for later, when we'll need to generate
					// one RecordConfig for each value.
					//fmt.Printf("parseLeaf: s-append %s\n", v3.(string))
					rTargets = append(rTargets, v3.(string))
				case map[interface{}]interface{}:
					// Example:
					// values:
					// - priority: 10
					//   value: mx1.example.com.
					// - priority: 10
					//   value: mx2.example.com.
					// We collect the individual values. When we are done with this level,
					// we should have enough to generate a single RecordConfig.
					newRc := newRecordConfig(k, rType, "", rTTL, origin)
					for k4, v4 := range v3.(map[interface{}]interface{}) {
						//fmt.Printf("parseLeaf: k4=%s v4=%s\n", k4, v4)
						switch k4.(string) {
						case "priority": // MX,SRV
							priority := uint16(v4.(int))
							newRc.MxPreference = priority
							newRc.SrvPriority = priority
							// Assign it to both places. We'll zap the wrong one later.
						case "weight": // SRV
							newRc.SrvWeight = uint16(v4.(int))
						case "port": // SRV
							newRc.SrvPort = uint16(v4.(int))
						case "value": // MX
							newRc.SetTarget(v4.(string))
						}
					}
					//fmt.Printf("parseLeaf: append %v\n", newRc)
					someresults = append(someresults, newRc)
				default:
					return nil, fmt.Errorf("parseLeaf: unknown type in map: rtype=%s k=%s v3.(type)=%T v3=%v", rType, k, v3, v3)
				}
			}
		} else {
			return nil, fmt.Errorf("parseLeaf: unknown type in level 2: k=%s k2=%s v.2(type)=%T v2=%v", k, k2, v2, v2)
		}
	}
	// fmt.Printf("parseLeaf: Target=(%v)\n", rTarget)
	// fmt.Printf("parseLeaf: len(rTargets)=%d\n", len(rTargets))
	// fmt.Printf("parseLeaf: len(someresults)=%d\n", len(someresults))

	// We've now looped through everything about one label. Make the RecordConfig(s).

	if len(someresults) > 0 {
		// We have many results. Generate a RecordConfig for each one.
		for _, r := range someresults {
			r.Type = rType
			r.TTL = rTTL
			results = append(results, r)
			// Earlier we didn't know what the priority was for. Now that  we know the rType,
			// we zap the wrong one.
			switch r.Type {
			case "MX":
				r.SrvPriority = 0
			case "SRV":
				r.MxPreference = 0
			default:
				panic("ugh")
			}
		}
	} else if rTarget != "" && len(rTargets) == 0 {
		// The file used "value".  Generate a single RecordConfig
		//fmt.Printf("parseLeaf: 1-newRecordConfig(%v, %v, %v, %v, %v)\n", k, rType, rTarget, rTTL, origin)
		results = append(results, newRecordConfig(k, rType, rTarget, rTTL, origin))
	} else {
		// The file used "values" so now we must generate a RecordConfig for each value.
		for _, target := range rTargets {
			//fmt.Printf("parseLeaf: 3-newRecordConfig(%v, %v, %v, %v, %v)\n", k, rType, target, rTTL, origin)
			results = append(results, newRecordConfig(k, rType, target, rTTL, origin))
		}
	}
	return results, nil
}

// newRecordConfig is a RecordConfig factory.
func newRecordConfig(rname, rtype, target string, ttl uint32, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type: rtype,
		TTL:  ttl,
	}
	rc.SetLabel(rname, origin)
	switch rtype {
	case "TXT":
		rc.SetTargetTXT(target)
	default:
		rc.SetTarget(target)
	}
	return rc
}

// typeof returns a string that indicates v's type:
func typeof(v interface{}) string {
	// Cite: https://stackoverflow.com/a/20170555/71978
	return reflect.TypeOf(v).String()
}

// decodeTTL decodes an interface into a TTL value.
// This is useful when you don't know if a TTL arrived as a string or int.
func decodeTTL(ttl interface{}) (uint32, error) {
	switch ttl.(type) {
	case uint32:
		return ttl.(uint32), nil
	case string:
		s := ttl.(string)
		t, err := strconv.ParseUint(s, 10, 32)
		return uint32(t), fmt.Errorf("decodeTTL failed to parse (%s): %w", s, err)
	case int:
		i := ttl.(int)
		if i < 0 {
			return 0, fmt.Errorf("ttl cannot be negative (%d)", i)
		}
		return uint32(i), nil
	}
	panic("I don't know what type this TTL is")
}
