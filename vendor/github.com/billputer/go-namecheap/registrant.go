package namecheap

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// Registrant is a struct that contains all the data necesary to register a domain.
// That is to say, every field in this struct is REQUIRED by the namecheap api to
// crate a new domain.
// In order for `addValues` method to work, all fields must remain strings.
type Registrant struct {
	RegistrantFirstName, RegistrantLastName,
	RegistrantAddress1, RegistrantAddress2, RegistrantCity,
	RegistrantStateProvince, RegistrantPostalCode, RegistrantCountry,
	RegistrantPhone, RegistrantEmailAddress,

	TechFirstName, TechLastName,
	TechAddress1, TechAddress2,
	TechCity, TechStateProvince, TechPostalCode, TechCountry,
	TechPhone, TechEmailAddress,

	AdminFirstName, AdminLastName,
	AdminAddress1, AdminAddress2,
	AdminCity, AdminStateProvince, AdminPostalCode, AdminCountry,
	AdminPhone, AdminEmailAddress,

	AuxBillingFirstName, AuxBillingLastName,
	AuxBillingAddress1, AuxBillingAddress2,
	AuxBillingCity, AuxBillingStateProvince, AuxBillingPostalCode, AuxBillingCountry,
	AuxBillingPhone, AuxBillingEmailAddress string
}

// newRegistrant return a new registrant where all the required fields are the same.
// Feel free to change them as needed
func newRegistrant(
	firstName, lastName,
	addr1, addr2,
	city, state, postalCode, country,
	phone, email string,
) *Registrant {
	return &Registrant{
		RegistrantFirstName:     firstName,
		RegistrantLastName:      lastName,
		RegistrantAddress1:      addr1,
		RegistrantAddress2:      addr2,
		RegistrantCity:          city,
		RegistrantStateProvince: state,
		RegistrantPostalCode:    postalCode,
		RegistrantCountry:       country,
		RegistrantPhone:         phone,
		RegistrantEmailAddress:  email,
		TechFirstName:           firstName,
		TechLastName:            lastName,
		TechAddress1:            addr1,
		TechAddress2:            addr2,
		TechCity:                city,
		TechStateProvince:       state,
		TechPostalCode:          postalCode,
		TechCountry:             country,
		TechPhone:               phone,
		TechEmailAddress:        email,
		AdminFirstName:          firstName,
		AdminLastName:           lastName,
		AdminAddress1:           addr1,
		AdminAddress2:           addr2,
		AdminCity:               city,
		AdminStateProvince:      state,
		AdminPostalCode:         postalCode,
		AdminCountry:            country,
		AdminPhone:              phone,
		AdminEmailAddress:       email,
		AuxBillingFirstName:     firstName,
		AuxBillingLastName:      lastName,
		AuxBillingAddress1:      addr1,
		AuxBillingAddress2:      addr2,
		AuxBillingCity:          city,
		AuxBillingStateProvince: state,
		AuxBillingPostalCode:    postalCode,
		AuxBillingCountry:       country,
		AuxBillingPhone:         phone,
		AuxBillingEmailAddress:  email,
	}
}

// addValues adds the fields of this struct to the passed in url.Values.
// It is important that all the fields of Registrant remain string type.
func (reg *Registrant) addValues(u url.Values) error {
	if u == nil {
		return errors.New("nil value passed as url.Values")
	}

	val := reflect.ValueOf(*reg)
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldName := t.Field(i).Name
		field := val.Field(i).String()
		if ty := val.Field(i).Kind(); ty != reflect.String {
			return fmt.Errorf(
				"Registrant cannot have types that aren't string; %s is type %s",
				fieldName, ty,
			)
		}
		if field == "" {
			if strings.Contains(fieldName, "ddress2") {
				continue
			}

			return fmt.Errorf("Field %s cannot be empty", fieldName)
		}

		u.Set(fieldName, fmt.Sprintf("%v", field))
	}

	return nil
}
