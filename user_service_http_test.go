package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestMethodNotAllowed(t *testing.T) {

	for _, method := range []string{http.MethodPost, http.MethodDelete, http.MethodPut} {
		resp := httptest.NewRecorder()
		httpUserService(nil).ServeHTTP(resp, httptest.NewRequest(method, "/whoami", nil))

		if resp.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method should not be allowed: %s", method)
		}
	}
}

type FakeUserProvider func() (*User, error)

func (f FakeUserProvider) FromHeaders(headers HeaderProvider) (*User, error) {
	return f()
}

func TestBadRequest(t *testing.T) {
	cases := map[string]struct {
		err          error
		expectedCode int
	}{
		"bad requst": {
			err:          ErrorBadInput("Nooo"),
			expectedCode: http.StatusBadRequest,
		},
		"internal error": {
			err:          errors.New("Boooo"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			resp := httptest.NewRecorder()

			httpUserService(FakeUserProvider(func() (*User, error) {
				return nil, tc.err
			})).ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/whoami", nil))

			if resp.Code != tc.expectedCode {
				t.Fatalf("Got code %d, but expected %d", resp.Code, tc.expectedCode)
			}
		})
	}
}

func TestResponse(t *testing.T) {
	user := &User{
		ID:    "foo:/bar",
		Email: "foo@example.org",
		Roles: []string{"butcher", "baker"},
	}

	resp := httptest.NewRecorder()
	httpUserService(FakeUserProvider(func() (*User, error) {
		return user, nil
	})).ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/whoami", nil))

	var returnedUser User
	err := json.Unmarshal(resp.Body.Bytes(), &returnedUser)
	if err != nil {
		t.Fatalf("Encountered error reading response: %v", err)
	}

	diffs := deep.Equal(user, &returnedUser)
	if len(diffs) > 0 {
		t.Fatalf("Got different response than expected:\n%s", strings.Join(diffs, "\n"))
	}

	if !strings.Contains(resp.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("Bad content type: %s", resp.Header().Get("Content-Type"))
	}
}

type CannotWrite struct {
	code int
}

func (c *CannotWrite) Header() http.Header {
	return http.Header(map[string][]string{})
}

func (c *CannotWrite) Write([]byte) (int, error) {
	return 0, errors.New("Couldn't write")
}

func (c *CannotWrite) WriteHeader(statusCode int) {
	c.code = statusCode
}

func TestSerializationError(t *testing.T) {
	resp := &CannotWrite{}
	httpUserService(FakeUserProvider(func() (*User, error) {
		return &User{}, nil
	})).ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/whoami", nil))

	if resp.code != http.StatusInternalServerError {
		t.Fatalf("Got wrong response code: %d, expected: %d", resp.code, http.StatusInternalServerError)
	}
}
