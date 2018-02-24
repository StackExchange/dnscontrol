package contact

import (
    "time"
)

type PrepaidInformation struct {
    Id int64
    Amount string
    Currency string
    DateCreated time.Time
    DateUpdated time.Time
}

type BalanceInformation struct {
    AnnualBalance string
    Grid string
    OutstandingAmount float64
    Prepaid *PrepaidInformation
}

type ContactInformation struct {
    Firstname string
    Lastname string
    Email string
    Address string
    Zipcode string
    City string
    Country string
    Phone string
    ContactType int64
    Handle string
}

func (self ContactInformation) ContactTypeString() string {
    switch self.ContactType {
        case 0:
            return "Person"
        case 1:
            return "Company"
        case 2:
            return "Association"
        case 3:
            return "Public Body"
        case 4:
            return "Reseller"
    }
    return ""
}

type ContactCreate struct {
    Firstname string `goptions:"--firstname, obligatory, description='First name'"`
    Lastname string `goptions:"--lastname, obligatory, description='Last name'"`
    Email string `goptions:"--email, obligatory, description='Email address'"`
    Password string `goptions:"--password, obligatory, description='Password'"`
    Address string `goptions:"--address, obligatory, description='Street address'"`
    Zipcode string `goptions:"--zipcode, obligatory, description='Zip code'"`
    City string `goptions:"--city, obligatory, description='City'"`
    Country string `goptions:"--country, obligatory, description='Country'"`
    Phone string `goptions:"--phone, obligatory, description='Phone number'"`

    // Contact types
    IsPerson bool `goptions:"--person, obligatory, mutexgroup='type', description='Contact type person'"`
    IsCompany bool `goptions:"--company, obligatory, mutexgroup='type', description='Contact type company'"`
    IsAssociation bool `goptions:"--association, obligatory, mutexgroup='type', description='Contact type association'"`
    IsPublicBody bool `goptions:"--publicbody, obligatory, mutexgroup='type', description='Contact type public body'"`
    IsReseller bool `goptions:"--reseller, obligatory, mutexgroup='type', description='Contact type reseller'"`
}

func (self ContactCreate) ContactType() int {
    if self.IsPerson { return 0 }
    if self.IsCompany { return 1 }
    if self.IsAssociation { return 2 }
    if self.IsPublicBody { return 3 }
    if self.IsReseller { return 4 }

    // Default to person
    return 0
}

