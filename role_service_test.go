package main_test

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
	jhuda "github.com/jhu-sheridan-libraries/jhuda-user-service"
)

func TestDefaultRoles(t *testing.T) {
	svc := jhuda.RoleService{
		RoleBase:     "info:test/",
		DefaultRoles: []string{"foo", "bar"},
	}

	roles, _ := svc.Lookup(nil)
	expected := []jhuda.Role{
		{
			Base: "info:test/",
			Name: "foo",
		}, {
			Base: "info:test/",
			Name: "bar",
		},
	}

	diffs := deep.Equal(expected, roles)

	if len(diffs) > 0 {
		t.Fatalf("Did not get expected roles:\n %s", strings.Join(diffs, "\n"))
	}
}
