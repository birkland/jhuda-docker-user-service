package main

type ShibHeaders struct {
	Displayname string
	Email       string
	Eppn        string
	GivenName   string
	LastName    string
	LocatorIDs  []string
}

var DefaultShibHeaders = ShibHeaders{
	Displayname: "Displayname",
	Email:       "Mail",
	Eppn:        "Eppn",
	GivenName:   "Givenname",
	LastName:    "Sn",
	LocatorIDs:  []string{"Employeenumber", "unique-id", "Eppn"},
}
