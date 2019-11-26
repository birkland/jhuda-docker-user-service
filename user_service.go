package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// HeaderProvider provides values for headers
type HeaderProvider interface {
	Get(key string) (val string) // Get a header value
}

type RoleLookup interface {
	Lookup(u *User) ([]Role, error)
}

// UserService provides the identity and information associated with a User by inspecting
// Http headers
type UserService struct {
	UserBase   string
	HeaderDefs ShibHeaders
	Roles      RoleLookup
}

func (u UserService) FromHeaders(headers HeaderProvider) (*User, error) {
	eppn := headers.Get(oneOf(u.HeaderDefs.Eppn, DefaultShibHeaders.Eppn))

	if !strings.Contains(eppn, "@") {
		return nil, ErrorBadInput(fmt.Sprintf("Eppn is expected to be user@domain, instead got '%s'", eppn))
	}

	user := &User{
		ID:          u.UserBase + eppn,
		Displayname: headers.Get(oneOf(u.HeaderDefs.Displayname, DefaultShibHeaders.Displayname)),
		Firstname:   headers.Get(oneOf(u.HeaderDefs.GivenName, DefaultShibHeaders.GivenName)),
		Middlename:  headers.Get(oneOf(u.HeaderDefs.MiddleName, DefaultShibHeaders.MiddleName)),
		Lastname:    headers.Get(oneOf(u.HeaderDefs.LastName, DefaultShibHeaders.LastName)),
		Email:       headers.Get(oneOf(u.HeaderDefs.Email, DefaultShibHeaders.Email)),
		Locatorids:  u.locatorIds(u.HeaderDefs.LocatorIDs, headers),
	}

	return u.addRoles(user)
}

func (u UserService) locatorIds(locators []string, headers HeaderProvider) []string {

	// If locator headers slice is nil (undefined), then use the defaults.
	// Note: this differs from an explicitly allocated empty slice, which is used
	// to signal intent for zero locators
	if locators == nil {
		locators = DefaultShibHeaders.LocatorIDs
	}

	var locatorIds []string

	domain := strings.Split(headers.Get(oneOf(u.HeaderDefs.Eppn, DefaultShibHeaders.Eppn)), "@")[1]

	for _, locator := range locators {
		val := headers.Get(locator)
		if val != "" {
			locatorIds = append(locatorIds, domain+":"+locator+":"+val)
		}
	}

	return locatorIds
}

func (u UserService) addRoles(user *User) (*User, error) {

	if u.Roles == nil {
		return user, nil
	}

	uniqueRoles := map[string]bool{}

	for _, role := range user.Roles {
		uniqueRoles[role] = true
	}

	roles, err := u.Roles.Lookup(user)
	if err != nil {
		return nil, errors.Errorf("Error determining roles for %s", user.ID)
	}

	for _, r := range roles {
		role := r.Simple()
		if !uniqueRoles[role] {
			uniqueRoles[role] = true
			user.Roles = append(user.Roles, role)
		}
	}

	return user, nil
}

func oneOf(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}

	return val
}
