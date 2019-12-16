package test_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/jhu-sheridan-libraries/jhuda-docker-user-service/test"
)

func TestFoo(t *testing.T) {
	client := test.ShibClient{
		IdpBaseURI: "https://archive.local/idp/",
		Username:   "depositor",
		Password:   "moo",
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}

	request, _ := http.NewRequest(http.MethodGet, "https://archive.local/whoami", nil)

	resp, err := client.Do(request)
	if err != nil {
		t.Fatalf("Request errored: %v", err)
	}

	// do stuff with resp
	t.Fatalf(resp.Header.Get("Content-Type"))
}
