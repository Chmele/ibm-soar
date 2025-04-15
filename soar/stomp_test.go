package soar

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"
)

type FakeConn struct {
	ReadData  []byte
	WriteData bytes.Buffer
	closed    bool
}

func (f *FakeConn) Read(b []byte) (int, error) {
	if len(f.ReadData) == 0 {
		return 0, io.EOF
	}
	n := copy(b, f.ReadData)
	f.ReadData = f.ReadData[n:]
	return n, nil
}

func (f *FakeConn) Write(b []byte) (int, error) {
	return f.WriteData.Write(b)
}

func (f *FakeConn) Close() error {
	f.closed = true
	return nil
}

type dummyAddr struct {
	addr string
}

func (d *dummyAddr) Network() string { return "tcp" }
func (d *dummyAddr) String() string  { return d.addr }

func (f *FakeConn) LocalAddr() net.Addr  { return &dummyAddr{"local"} }
func (f *FakeConn) RemoteAddr() net.Addr { return &dummyAddr{"remote"} }

func (f *FakeConn) SetDeadline(t time.Time) error      { return nil }
func (f *FakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (f *FakeConn) SetWriteDeadline(t time.Time) error { return nil }

func (l *StompListener) ConnectMock(prewritten []byte) (*FakeConn, error) {
	fakeConn := &FakeConn{
		ReadData: prewritten,
	}

	err := l.ConnectSTOMP(fakeConn)
	return fakeConn, err
}

func TestStompConnection(t *testing.T) {
	client := &HTTPClient{
		Hostname:  "test-host",
		KeyId:     "test-id",
		KeySecret: "test-secret",
		Org:       &Org{ID: 123},
	}

	mockResponse := []byte("CONNECTED\nversion:1.2\n\n\x00")

	listener := &StompListener{
		HTTPClient:         client,
		StompPort:          "65001",
		Ctx:                context.Background(),
		Done:               make(chan struct{}),
		Insecure:           true,
		MessageDestination: "unit-test",
		Logger: &StompLogger{
			slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})),
		},
	}

	fakeConn, err := listener.ConnectMock(mockResponse)
	if err != nil {
		t.Fatalf("Failed to create mock connection: %v", err)
	}

	actual, err := io.ReadAll(&fakeConn.WriteData)
	if err != nil {
		t.Fatalf("Failed reading client sent payload: %v", err)
	}
	expected := "CONNECT\nhost:test-host\nheart-beat:0,0\nlogin:test-id\npasscode:test-secret\naccept-version:1.2"
	expectedBytes := append([]byte(expected), []byte{10, 10, 0}...)
	if len(expectedBytes) != len(actual) {
		t.Fatalf("Length not matching: %v\nACTUAL: %v", expectedBytes, actual)
	}
	for i := range actual {
		if actual[i] != expectedBytes[i] {
			t.Fatalf("Payload byte mismatch, index %d: %v\nACTUAL: %v", i, expectedBytes, actual)
		}
	}
}
