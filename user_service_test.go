package main_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/go-test/deep"
	jhuda "github.com/jhu-sheridan-libraries/jhuda-docker-user-service"
)

func TestLocatorIDs(t *testing.T) {

	userBase := "http://example.org/fcrepo/rest/users/"

	cases := []struct {
		name       string
		headerDefs jhuda.ShibHeaders
		headers    map[string][]string
		expected   *jhuda.User
	}{{
		name: "custom headers",
		headerDefs: jhuda.ShibHeaders{
			Eppn:        "Custom-Eppn",
			Email:       "Custom-Email",
			GivenName:   "Custom-Given-Name",
			LastName:    "Custom-Last-Name",
			Displayname: "Custom-Display-Name",
			LocatorIDs: []string{
				"Foo",
				"Bar",
			},
		},
		headers: map[string][]string{
			"Custom-Eppn":         {"foo@example.org"},
			"Custom-Email":        {"me@example.org"},
			"Custom-Given-Name":   {"Bo"},
			"Custom-Last-Name":    {"Vine"},
			"Custom-Display-Name": {"MOOO"},
			"Foo":                 {"foo"},
			"Bar":                 {"bar"},
		},
		expected: &jhuda.User{
			ID:          "http://example.org/fcrepo/rest/users/foo@example.org",
			Type:        "User",
			Firstname:   "Bo",
			Lastname:    "Vine",
			Displayname: "MOOO",
			Email:       "me@example.org",
			Locatorids:  []string{"example.org:Foo:foo", "example.org:Bar:bar"},
		},
	}, {
		name: "default headers",
		headers: map[string][]string{
			"Eppn":           {"foo@example.org"},
			"Mail":           {"me@example.org"},
			"Givenname":      {"Bo"},
			"Sn":             {"Vine"},
			"Displayname":    {"MOOO"},
			"Employeenumber": {"foo"},
			"Unique-Id":      {"bar"},
		},
		expected: &jhuda.User{
			ID:          "http://example.org/fcrepo/rest/users/foo@example.org",
			Type:        "User",
			Firstname:   "Bo",
			Lastname:    "Vine",
			Displayname: "MOOO",
			Email:       "me@example.org",
			Locatorids: []string{
				"example.org:Employeenumber:foo",
				"example.org:unique-id:bar",
				"example.org:Eppn:foo@example.org",
			},
		},
	}, {
		name: "sparse",
		headerDefs: jhuda.ShibHeaders{
			LocatorIDs: []string{}, // Explicitly empty.  We do not want any locators
		},
		headers: map[string][]string{
			"Eppn":           {"foo@example.org"},
			"Mail":           {"me@example.org"},
			"Displayname":    {"MOOO"},
			"Employeenumber": {"foo"},
			"Unique-Id":      {"bar"},
		},
		expected: &jhuda.User{
			ID:          "http://example.org/fcrepo/rest/users/foo@example.org",
			Type:        "User",
			Displayname: "MOOO",
			Email:       "me@example.org",
		},
	}}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			user, err := jhuda.UserService{
				HeaderDefs: c.headerDefs,
				UserBase:   userBase,
			}.FromHeaders(http.Header(c.headers))

			if err != nil {
				t.Fatalf("Got unexpected error: %v", err)
			}

			diffs := deep.Equal(user, c.expected)

			if len(diffs) > 0 {
				t.Fatalf("Did not get back expected user!\n%s", strings.Join(diffs, "\n"))
			}
		})
	}
}

func TestBadEppn(t *testing.T) {
	cases := map[string]map[string][]string{
		"Malformed eppn": {
			"Eppn": {"FooBar"},
		},
		"No eppn": {
			"Foo": {"Bar"},
		},
	}

	for name, headers := range cases {
		headers := http.Header(headers)
		t.Run(name, func(t *testing.T) {
			_, err := jhuda.UserService{
				HeaderDefs: jhuda.DefaultShibHeaders,
			}.FromHeaders(headers)

			if err == nil {
				t.Fatalf("Expected error!")
			}
		})
	}
}

type FakeRoleLookup struct {
	roles []jhuda.Role
	err   error
}

func (l FakeRoleLookup) Lookup(u *jhuda.User) ([]jhuda.Role, error) {
	if l.err != nil {
		return nil, l.err
	}

	return l.roles, nil
}

func TestRoles(t *testing.T) {
	cases := map[string]struct {
		roles       []jhuda.Role
		expected    []string
		expectedErr error
	}{
		"basic": {
			roles: []jhuda.Role{{
				Name: "foo",
			}, {
				Name: "bar",
			}},
			expected: []string{"foo", "bar"},
		},
		"non redundant": {
			roles: []jhuda.Role{{
				Name: "foo",
			}, {
				Name: "bar",
			}, {
				Name: "bar",
			}},
			expected: []string{"foo", "bar"},
		},
		"error": {
			roles:       []jhuda.Role{{}},
			expectedErr: errors.New("An error"),
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			user, err := jhuda.UserService{
				Roles: FakeRoleLookup{
					roles: tc.roles,
					err:   tc.expectedErr,
				},
			}.FromHeaders(http.Header(map[string][]string{
				"Eppn": {"foo@example.org"},
			}))

			if err != nil && tc.expectedErr == nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				return
			}

			diffs := deep.Equal(user.Roles, tc.expected)
			if len(diffs) > 0 {
				t.Fatalf("Roles different than expected!\nExpected:\n%s\n\nGot:\n%s",
					strings.Join(tc.expected, "\n"),
					strings.Join(user.Roles, "\n"))
			}
		})
	}
}

func TestContext(t *testing.T) {
	cases := map[string]struct {
		context  string
		expected *jhuda.User
	}{
		"no context": {
			expected: &jhuda.User{
				ID:         "foo@example.org",
				Type:       "User",
				Locatorids: []string{"example.org:Eppn:foo@example.org"},
			},
		},
		"defined context": {
			context: "http://example.org/context/",
			expected: &jhuda.User{
				ID:         "foo@example.org",
				Type:       "User",
				Context:    "http://example.org/context/",
				Locatorids: []string{"example.org:Eppn:foo@example.org"},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			user, err := jhuda.UserService{
				JsonldContext: tc.context,
			}.FromHeaders(http.Header(map[string][]string{
				"Eppn": {"foo@example.org"},
			}))

			if err != nil {
				t.Fatalf("Got an error: %v", err)
			}

			diffs := deep.Equal(tc.expected, user)
			if len(diffs) > 0 {
				t.Fatalf("Got User that differs from expected:\n%s", strings.Join(diffs, "\n"))
			}
		})
	}
}
