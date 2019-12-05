package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
)

func TestServe(t *testing.T) {

	cases := map[string]struct {
		args     []string
		headers  map[string]string
		expected User
	}{
		"defaults": {
			headers: map[string]string{
				DefaultShibHeaders.Eppn:        "foo@example.org",
				DefaultShibHeaders.Email:       "me@example.org",
				DefaultShibHeaders.Displayname: "Moo",
				DefaultShibHeaders.GivenName:   "Bos",
				DefaultShibHeaders.LastName:    "Taurus",
			},
			expected: User{
				ID:          "foo@example.org",
				Type:        "User",
				Email:       "me@example.org",
				Displayname: "Moo",
				Firstname:   "Bos",
				Lastname:    "Taurus",
				Locatorids:  []string{"example.org:Eppn:foo@example.org"},
			},
		},
		"custom headers": {
			args: []string{
				"-eppnHeader", "Test-Eppn",
				"-displayNameHeader", "Test-Displayname",
				"-emailHeader", "Test-Email",
				"-givenNameHeader", "Test-Givenname",
				"-lastNameHeader", "Test-Lastname",
				"-locatorHeaders", "Foo,Bar"},
			headers: map[string]string{
				"Test-Eppn":        "foo@example.org",
				"Test-Email":       "me@example.org",
				"Test-Displayname": "MOO",
				"Test-Givenname":   "Bos",
				"Test-Lastname":    "Taurus",
				"Foo":              "foo",
				"Bar":              "bar",
			},
			expected: User{
				ID:          "foo@example.org",
				Type:        "User",
				Email:       "me@example.org",
				Displayname: "MOO",
				Firstname:   "Bos",
				Lastname:    "Taurus",
				Locatorids:  []string{"example.org:Foo:foo", "example.org:Bar:bar"},
			},
		},
		"user baseurl": {
			args: []string{"-userBaseUrl", "http://example.org/fcrepo/rest/"},
			headers: map[string]string{
				DefaultShibHeaders.Eppn: "foo@example.org",
			},
			expected: User{
				ID:         "http://example.org/fcrepo/rest/foo@example.org",
				Type:       "User",
				Locatorids: []string{"example.org:Eppn:foo@example.org"},
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			port := strconv.Itoa(randomPort(t))
			args := append([]string{os.Args[0], "serve", "-port", port}, tc.args...)

			go run(args)

			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%s/whoami", port), nil)
			for k, v := range tc.headers {
				req.Header.Add(k, v)
			}

			resp := attempt(t, req)
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				t.Fatalf("Did not get a 200 response, got %d", resp.StatusCode)
			}

			var user User
			err := json.NewDecoder(resp.Body).Decode(&user)
			if err != nil {
				t.Fatalf("Bad JSON User response: %s", err)
			}

			diffs := deep.Equal(user, tc.expected)
			if len(diffs) > 0 {
				t.Fatalf("returned user does not match expected:\n%s", strings.Join(diffs, "\n"))
			}

			// Basically, send our server a ^C and let it stop itself gracefully
			proc, _ := os.FindProcess(os.Getpid())
			_ = proc.Signal(os.Interrupt)
		})
	}
}

func attempt(t *testing.T, req *http.Request) *http.Response {
	var err error
	var resp *http.Response
	client := &http.Client{}
	for i := 0; i < 100; i++ {
		resp, err = client.Do(req)
		if err == nil {
			return resp
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Connect to user service failed: %s", err)
	return nil
}

func randomPort(t *testing.T) int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("could not resolve tcp: %v", err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("Could not resolve port:%v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
