package octoyaml

import (
	"io"
	"io/ioutil"
	"math"
	"reflect"
	"strconv"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// ReadYaml parses a yaml input and returns a list of RecordConfigs
func ReadYaml(r io.Reader, origin string) (models.Records, error) {
	results := models.Records{}

	ydata, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrapf(err, "can not read yaml filehandle")
	}

	var mysterydata map[string]interface{}
	err = yaml.Unmarshal(ydata, &mysterydata)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal yaml")
	}
	//fmt.Printf("ReadYaml: mysterydata == %v\n", mysterydata)

	for k, v := range mysterydata { // Each label
		//fmt.Printf("ReadYaml: NEXT KEY\n")
		//fmt.Printf("ReadYaml:  KEY=%s v.(type)=%s\n", k, reflect.TypeOf(v).String())
		switch v.(type) {
		case map[interface{}]interface{}:
			results, err = parseLeaf(results, k, v, origin)
			if err != nil {
				return results, errors.Wrapf(err, "leaf (%v) error", v)
			}
		case []interface{}:
			for i, v3 := range v.([]interface{}) { // All the label's list
				_ = i
				//fmt.Printf("ReadYaml:   list key=%s i=%d v3.(type)=%s\n", k, i, typeof(v3))
				switch v3.(type) {
				case map[interface{}]interface{}:
					//fmt.Printf("ReadYaml:   v3=%v\n", v3)
					results, err = parseLeaf(results, k, v3, origin)
					if err != nil {
						return results, errors.Wrapf(err, "leaf v3=%v", v3)
					}
				default:
					return nil, errors.Errorf("unknown type in list3: k=%s v.(type)=%T v=%v", k, v, v)
				}
			}

		default:
			return nil, errors.Errorf("unknown type in list1: k=%s v.(type)=%T v=%v", k, v, v)
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
		//fmt.Printf("ReadYaml: ifs tk2=%s tv2=%s len(rTargets)=%d\n", typeof(k2), typeof(v2), len(rTargets))
		if typeof(k2) == "string" && (typeof(v2) == "string" || typeof(v2) == "int") {
			//fmt.Printf("parseLeaf:   k2=%s v2=%v\n", k2, v2)
			switch k2.(string) {
			case "type":
				rType = v2.(string)
			case "ttl":
				var err error
				rTTL, err = decodeTTL(v2)
				if err != nil {
					return nil, errors.Errorf("parseLeaf: can not parse ttl (%v)", v2)
				}
			case "value":
				rTarget = v2.(string)
			case "values":
				switch v2.(type) {
				case string:
					rTarget = v2.(string)
				default:
					return nil, errors.Errorf("parseLeaf: unknown type in values: rtpe=%s k=%s k2=%s v2.(type)=%T v2=%v", rType, k, k2, v2, v2)
				}
			}
		} else if typeof(k2) == "string" && typeof(v2) == "[]interface {}" {
			someresults = nil
			for _, v3 := range v2.([]interface{}) {
				switch v3.(type) {
				case string:
					rTargets = append(rTargets, v3.(string))
				case map[interface{}]interface{}:
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
							newRc.Target = v4.(string)
						}
					}
					someresults = append(someresults, newRc)
				default:
					return nil, errors.Errorf("parseLeaf: unknown type in map: rtype=%s k=%s v3.(type)=%T v3=%v", rType, k, v3, v3)
				}
			}
		} else {
			return nil, errors.Errorf("parseLeaf: unknown type in level 2: k=%s k2=%s v.2(type)=%T v2=%v", k, k2, v2, v2)
		}
	}
	// We've now looped through everything about one label. Make the RecordConfig(s).
	//fmt.Printf("parseLeaf: len(rTargets)=%d\n", len(rTargets))
	if len(someresults) > 0 {
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
		results = append(results, newRecordConfig(k, rType, rTarget, rTTL, origin))
	} else if len(rTargets) == 1 {
		results = append(results, newRecordConfig(k, rType, rTargets[0], rTTL, origin))
	} else if len(rTargets) > 1 {
		for _, target := range rTargets {
			results = append(results, newRecordConfig(k, rType, target, rTTL, origin))
		}
	}
	return results, nil
}

func newRecordConfig(rname, rtype, target string, ttl uint32, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:   rtype,
		Target: target,
		TTL:    ttl,
	}
	rc.SetLabel(rname, origin)
	return rc
}

// typeof returns a string that indicates v's type:
func typeof(v interface{}) string {
	// Cite: https://stackoverflow.com/a/20170555/71978
	return reflect.TypeOf(v).String()
}

// decodeTTL decodes an interface into a TTL value.
func decodeTTL(ttl interface{}) (uint32, error) {
	switch ttl.(type) {
	case uint32:
		return ttl.(uint32), nil
	case string:
		s := ttl.(string)
		t, err := strconv.ParseUint(s, 10, 32)
		return uint32(t), errors.Wrapf(err, "decodeTTL failed to parse (%s)", s)
	case int:
		i := ttl.(int)
		if i < 0 || i > math.MaxUint32 {
			return 0, errors.Errorf("ttl won't fit in 32-bits (%d)", i)
		}
		return uint32(i), nil
	}
	panic("I don't know what type this TTL is")
}
