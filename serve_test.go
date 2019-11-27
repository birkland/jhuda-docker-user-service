package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-test/deep"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestDefaultServe(t *testing.T) {

	port := strconv.Itoa(randomPort(t))
	go run([]string{os.Args[0], "serve", "-port", port})

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%s/whoami", port), nil)
	req.Header.Add(DefaultShibHeaders.Eppn, "foo@example.org")
	req.Header.Add(DefaultShibHeaders.Email, "me@example.org")

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

	expected := User{
		ID:         "foo@example.org",
		Email:      "me@example.org",
		Locatorids: []string{"example.org:Eppn:foo@example.org"},
	}

	diffs := deep.Equal(user, expected)
	if len(diffs) > 0 {
		t.Fatalf("returned user does not match expected:\n%s", strings.Join(diffs, "\n"))
	}
}

func attempt(t *testing.T, req *http.Request) *http.Response {
	var err error
	var resp *http.Response
	client := &http.Client{}
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err == nil {
			return resp
		}
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
