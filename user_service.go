package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
)

// HeaderProvider provides values for headers
type HeaderProvider interface {
	Get(key string) (val string) // Get a header value
}

// RoleLookup finds all roles a given user has
type RoleLookup interface {
	Lookup(u *User) ([]Role, error)
}

// UserService provides the identity and information associated with a User by inspecting
// Http headers
type UserService struct {
	UserBase      string      // BaseURI for user IDs, e.g. http://archive.local/fcrepo/rest/users/
	JsonldContext string      // JSON-LD context URI for User resources
	HeaderDefs    ShibHeaders // Header definitions
	Roles         RoleLookup  // Role lookup service
}

func (u UserService) FromHeaders(headers HeaderProvider) (*User, error) {
	eppn := headers.Get(oneOf(u.HeaderDefs.Eppn, DefaultShibHeaders.Eppn))

	if !strings.Contains(eppn, "@") {
		log.Printf("Bad Eppn (from %s).  Headers are: \n%+v", u.HeaderDefs.Eppn, headers)
		return nil, ErrorBadInput(fmt.Sprintf("Eppn is expected to be user@domain, instead got '%s'", eppn))
	}

	user := &User{
		ID:          u.UserBase + eppn,
		Type:        "User",
		Context:     u.JsonldContext,
		Displayname: headers.Get(oneOf(u.HeaderDefs.Displayname, DefaultShibHeaders.Displayname)),
		Firstname:   headers.Get(oneOf(u.HeaderDefs.GivenName, DefaultShibHeaders.GivenName)),
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
		headerName := locator
		locatorName := locator

		if strings.Contains(locator, ":") {
			parts := strings.Split(locator, ":")
			headerName = parts[0]
			locatorName = parts[1]
		}

		val := headers.Get(headerName)
		if val != "" {
			locatorIds = append(locatorIds, domain+":"+locatorName+":"+val)
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
