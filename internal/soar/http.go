package soar

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

type HTTPClient struct {
	Client    http.Client
	Session   *SessionResponseJson
	Org       *Org
	KeyId     string
	KeySecret string
	Hostname  string
	Ctx       context.Context
}

func NewHTTPClient(ctx context.Context, hostname, keyId, keySecret string, insecure bool) (*HTTPClient, error) {
	ret := &HTTPClient{
		KeyId:     keyId,
		KeySecret: keySecret,
		Hostname:  hostname,
		Ctx:       ctx,
		Client: http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
			},
		},
	}
	session, err := ret.GetOrg()
	if err != nil {
		return nil, err
	}
	ret.Session = session
	ret.Org = &session.Orgs[0]
	return ret, nil
}

func (s *HTTPClient) Request(method, url string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://%s/rest/", s.Hostname)+url, data)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(s.Ctx)
	auth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", s.KeyId, s.KeySecret))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	return s.Client.Do(req)
}

func (s *HTTPClient) GetOrg() (session *SessionResponseJson, err error) {
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
	return body, nil
}

func (s *HTTPClient) OrgRequest(method, url string, data io.Reader) (*http.Response, error) {
	return s.Request(method, fmt.Sprintf("orgs/%d/%s", s.Org.ID, url), data)
}

func (s *HTTPClient) GetMessageDestinationAvailable(name string) (bool, error) {
	resp, err := s.OrgRequest("GET", "message_destinations/"+name, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var md MessageDestination
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		return false, err
	}
	return slices.Contains(md.APIKeys, s.Session.APIKeyHandle), nil
}

func (s *HTTPClient) GetInboundDestinationAvailable(name string) (bool, error) {
	resp, err := s.OrgRequest("GET", "inbound_destinations/"+name, nil)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	var md InboundDestination
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		return false, err
	}
	return slices.Contains(md.ReadPrincipals, s.Session.APIKeyHandle) && slices.Contains(md.WritePrincipals, s.Session.APIKeyHandle), nil
}
