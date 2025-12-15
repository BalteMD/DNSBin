package core

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog/log"
)

var Transport http.Transport

// the xml for the autodiscover service
const autodiscoverXML = `<?xml version="1.0" encoding="utf-8"?><Autodiscover xmlns="http://schemas.microsoft.com/exchange/autodiscover/outlook/requestschema/2006">
<Request><EMailAddress>{{.Email}}</EMailAddress>
<AcceptableResponseSchema>http://schemas.microsoft.com/exchange/autodiscover/outlook/responseschema/2006a</AcceptableResponseSchema>
</Request></Autodiscover>`

func parseTemplate(tmpl string, email string) (string, error) {
	t := template.Must(template.New("tmpl").Parse(tmpl))
	var buff bytes.Buffer
	data := struct{ Email string }{Email: email}
	err := t.Execute(&buff, data)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

// InsecureRedirectsO365 allows forwarding the Authorization header even when we shouldn't
type InsecureRedirectsO365 struct {
	Transport http.RoundTripper
	User      string
	Pass      string
	Insecure  bool
	Email     string
	UserAgent string
}

// RoundTrip custom redirector that allows us to forward the auth header, even when the domain changes.
// This is needed as some Office 365 domains will redirect from autodiscover.domain.com to autodiscover.outlook.com
// and Go does not forward Sensitive headers such as Authorization (https://golang.org/src/net/http/client.go#41)
func (l InsecureRedirectsO365) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t := l.Transport

	if t == nil {
		t = &Transport
	}
	resp, err = t.RoundTrip(req)
	if err != nil {
		return
	}

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect:
		location := resp.Header.Get("Location")
		log.Info().Str("request_url", req.URL.String()).Str("redirect_location", location).Msg("Request redirected")

		URL, err := url.Parse(location)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse redirect URL: %w", err)
		}

		r, err := parseTemplate(autodiscoverXML, l.Email)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse template: %w", err)
		}

		newReq, err := http.NewRequest("POST", URL.String(), strings.NewReader(r))
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create new request: %w", err)
		}

		newReq.Header.Add("Content-Type", "text/xml")
		newReq.Header.Add("User-Agent", l.UserAgent)
		newReq.Header.Add("X-MapiHttpCapability", "1") //we want MAPI info
		newReq.Header.Add("X-AnchorMailbox", l.User)   //we want MAPI info
		newReq.SetBasicAuth(l.User, l.Pass)

		client := http.Client{Transport: t, Timeout: 10 * time.Second}
		return client.Do(newReq)

	}

	return resp, nil
}
