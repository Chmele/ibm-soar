package soar

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type SOARClient struct {
	Client    http.Client
	OrgId     string
	BaseUrl   string
	KeyId     string
	KeySecret string
}

func NewSOARClient(hostname, keyId, keySecret, orgId string) *SOARClient {
	return &SOARClient{
		BaseUrl:   fmt.Sprintf("https://%s/rest/", hostname),
		KeyId:     keyId,
		KeySecret: keySecret,
		OrgId:     orgId,
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (s *SOARClient) Request(method, url string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, s.BaseUrl+url, data)
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", s.KeyId, s.KeySecret))
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	return s.Client.Do(req)
}

func (s *SOARClient) Session() error {
	resp, err := s.Request("GET", "session", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("%v\n\n", resp)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("%s", body)
	return nil
}
