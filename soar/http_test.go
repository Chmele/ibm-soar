package soar

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"github.com/chmele/ibm-soar/soar/structures"
)

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func TestNewHTTPClient(t *testing.T) {
	mockSession := &structures.SessionResponseJson{
		APIKeyHandle: 444,
		Orgs:         []structures.Org{{ID: 1234}},
	}
	mockBody, _ := json.Marshal(mockSession)

	client := &HTTPClient{
		KeyId:     "id",
		KeySecret: "secret",
		Hostname:  "test.local",
		Client: http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) *http.Response {
					if !strings.Contains(req.URL.Path, "session") {
						t.Fatalf("unexpected path: %s", req.URL.Path)
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader(mockBody)),
						Header:     make(http.Header),
					}
				},
			},
		},
		Ctx: context.Background(),
	}

	session, err := client.GetOrg()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if session.APIKeyHandle != 444 {
		t.Errorf("expected APIKeyHandle to be 'test-key', got: %d", session.APIKeyHandle)
	}
}

func TestGetMessageDestinationAvailable(t *testing.T) {
	md := structures.MessageDestination{
		APIKeys: []int{42},
	}
	body, _ := json.Marshal(md)

	client := &HTTPClient{
		Session:   &structures.SessionResponseJson{APIKeyHandle: 42},
		Org:       &structures.Org{ID: 1234},
		KeyId:     "id",
		KeySecret: "secret",
		Hostname:  "test.local",
		Client: http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) *http.Response {
					if !strings.Contains(req.URL.Path, "message_destinations/test") {
						t.Fatalf("unexpected path: %s", req.URL.Path)
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(bytes.NewReader(body)),
						Header:     make(http.Header),
					}
				},
			},
		},
		Ctx: context.Background(),
	}

	ok, err := client.GetMessageDestinationAvailable("test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected message destination to be available")
	}
}

func TestOrgRequest(t *testing.T) {
	client := &HTTPClient{
		Org: &structures.Org{ID: 99},
		Client: http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) *http.Response {
					expected := "orgs/99/test"
					if !strings.Contains(req.URL.String(), expected) {
						t.Errorf("expected path to contain %s, got %s", expected, req.URL.String())
					}
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(strings.NewReader("ok")),
						Header:     make(http.Header),
					}
				},
			},
		},
		Ctx: context.Background(),
	}
	resp, err := client.OrgRequest("GET", "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body 'ok', got '%s'", body)
	}
}

func TestGetInboundDestinationAvailable(t *testing.T) {
	tests := []struct {
		name             string
		read             []int
		write            []int
		expected         bool
		expectedHttpCode int
	}{
		{"available", []int{42}, []int{42}, true, 200},
		{"missing read", []int{}, []int{42}, false, 200},
		{"missing write", []int{42}, []int{}, false, 200},
		{"unauthorized", nil, nil, false, 401},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &HTTPClient{
				Session: &structures.SessionResponseJson{APIKeyHandle: 42},
				Org:     &structures.Org{ID: 1},
				Client: http.Client{
					Transport: &mockRoundTripper{
						roundTripFunc: func(req *http.Request) *http.Response {
							if tt.expectedHttpCode != 200 {
								return &http.Response{
									StatusCode: tt.expectedHttpCode,
									Body:       io.NopCloser(strings.NewReader("unauthorized")),
								}
							}
							resp := structures.InboundDestination{
								ReadPrincipals:  tt.read,
								WritePrincipals: tt.write,
							}
							b, _ := json.Marshal(resp)
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(bytes.NewReader(b)),
								Header:     make(http.Header),
							}
						},
					},
				},
				Ctx: context.Background(),
			}
			ok, err := client.GetInboundDestinationAvailable("inbound")
			if tt.expectedHttpCode != 200 && err == nil {
				t.Errorf("expected error for HTTP %d", tt.expectedHttpCode)
			}
			if ok != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, ok)
			}
		})
	}
}

func TestGetMessageDestinationAvailable_Error(t *testing.T) {
	client := &HTTPClient{
		Session: &structures.SessionResponseJson{APIKeyHandle: 444},
		Org:     &structures.Org{ID: 42},
		Client: http.Client{
			Transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader("not found")),
					}
				},
			},
		},
		Ctx: context.Background(),
	}

	ok, err := client.GetMessageDestinationAvailable("missing")
	if err == nil {
		t.Error("expected error from GetMessageDestinationAvailable, got nil")
	}
	if ok {
		t.Error("expected false availability for missing destination")
	}
}
