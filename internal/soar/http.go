package soar

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	Client    http.Client
	Org       *Org
	KeyId     string
	KeySecret string
	Hostname string
}

func NewHTTPClient(hostname, keyId, keySecret string) (*HTTPClient, error) {
	ret := &HTTPClient{
		KeyId:     keyId,
		KeySecret: keySecret,
		Hostname: hostname,
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	org, err := ret.GetOrg()
	if err != nil {
		return nil, err
	}
	ret.Org = org
	return ret, nil
}

func (s *HTTPClient) Request(method, url string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://%s/rest/", s.Hostname)+url, data)
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", s.KeyId, s.KeySecret))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	return s.Client.Do(req)
}

func (s *HTTPClient) GetOrg() (org *Org, err error) {
	resp, err := s.Request("GET", "session", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error connecting to SOAR, status: %s", resp.Status)
	}
	body := new(SessionResponseJson)
	if err := json.NewDecoder(resp.Body).Decode(body); err != nil {
		return nil, err
	}
	if len(body.Orgs) < 1 {
		return nil, errors.New("API key is not associated with any organization")
	}
	return &body.Orgs[0], nil
}
