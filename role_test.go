package main_test

import (
	"testing"

	jhuda "github.com/jhu-sheridan-libraries/jhuda-docker-user-service"
)

func TestRole(t *testing.T) {
	role := jhuda.Role{
		Base: "info:test/",
		Name: "foo",
	}

	if role.URL() != "info:test/foo" {
		t.Fatalf("Bad role URI: %s", role.URL())
	}

	if role.Simple() != "foo" {
		t.Fatalf("Bad simple role name: %s", role.Simple())
	}
}
