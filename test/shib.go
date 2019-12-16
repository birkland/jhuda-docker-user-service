package test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type Requester interface {
	Do(req *http.Request) (*http.Response, error)
}

type ShibClient struct {
	IdpBaseURI string
	Username   string
	Password   string
	Client     Requester
	cookies    []*http.Cookie
}

type formData struct {
	action      string
	inputFields map[string]string
}

func (c *ShibClient) Do(req *http.Request) (*http.Response, error) {

	for _, c := range c.cookies {
		req.AddCookie(c)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return resp, err
	}

	if c.needsLogin(resp) {
		return c.loginAndDo(req, resp)
	}

	return resp, err
}

func (c *ShibClient) needsLogin(resp *http.Response) bool {
	return len(c.IdpBaseURI) > 1 &&
		strings.HasPrefix(resp.Request.URL.String(), c.IdpBaseURI) &&
		strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html")
}

func (c *ShibClient) loginAndDo(req *http.Request, resp *http.Response) (*http.Response, error) {

	form, err := parseForm(resp)
	if err != nil {
		return nil, errors.Wrapf(err, "could not find idp login URL")
	}

	resp, err = c.submitCreds(form)
	if err != nil {
		return nil, errors.Wrapf(err, "could not submit shib credentials")
	}

	err = c.completeSAML(resp)
	if err != nil {
		return nil, errors.Wrapf(err, "could not finish SAML handshake")
	}

	for _, c := range c.cookies {
		req.AddCookie(c)
	}

	return c.Client.Do(req)
}

func (c *ShibClient) submitCreds(form formData) (*http.Response, error) {

	body := url.Values{}
	body.Add("j_username", c.Username)
	body.Add("j_password", c.Password)
	body.Add("_eventId_proceed", "")

	submit, _ := http.NewRequest(http.MethodPost, form.action, strings.NewReader(body.Encode()))
	submit.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return c.Client.Do(submit)
}

func (c *ShibClient) extractCookies(resp *http.Response) error {
	c.cookies = resp.Cookies()

	if len(c.cookies) == 0 {
		return errors.New("did not find any auth cookies")
	}

	return nil
}

func (c *ShibClient) completeSAML(resp *http.Response) error {
	if err := c.extractCookies(resp); err != nil {
		return errors.Wrapf(err, "response should have had a cookie")
	}

	form, err := parseForm(resp)
	if err != nil {
		return errors.Wrapf(err, "could not parse form from shib response")
	}

	body := url.Values{}
	for k, v := range form.inputFields {
		body.Add(k, v)
	}

	submit, _ := http.NewRequest(http.MethodPost, form.action, strings.NewReader(body.Encode()))
	submit.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range c.cookies {
		submit.AddCookie(c)
	}

	if httpc, ok := c.Client.(*http.Client); ok {
		prev := httpc.CheckRedirect
		defer func() {
			httpc.CheckRedirect = prev
		}()

		httpc.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	resp, err = c.Client.Do(submit)
	if err != nil {
		return errors.Wrapf(err, "could not perform final redirect request")
	}
	drain(resp)

	return c.extractCookies(resp)
}

func drain(resp *http.Response) {
	defer resp.Body.Close()
	_, _ = io.Copy(ioutil.Discard, resp.Body)
}

func parseForm(resp *http.Response) (formData, error) {
	defer resp.Body.Close()

	form := formData{
		inputFields: make(map[string]string),
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return form, errors.Wrapf(err, "could not parse shib form")
	}

	doc.Find("form").Each(func(i int, f *goquery.Selection) {
		raw, _ := f.Attr("action")
		action, _ := url.Parse(raw)
		form.action = resp.Request.URL.ResolveReference(action).String()

		f.Find("input").Each(func(i int, input *goquery.Selection) {
			name, _ := input.Attr("name")
			value, _ := input.Attr("value")

			form.inputFields[name] = value
		})

		f.Find("button").Each(func(i int, input *goquery.Selection) {
			name, _ := input.Attr("name")
			form.inputFields[name] = ""
		})
	})

	if !strings.HasPrefix(form.action, "http") {
		return form, errors.Errorf("malformed or missing form action uri: '%s'", form.action)
	}

	return form, nil
}
