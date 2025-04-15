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
	net.Conn // Optional: embed to satisfy interface

	ReadData  []byte
	WriteData bytes.Buffer

	WriteNotify chan struct{} // Signal channel
	closed      bool
}

func (f *FakeConn) Read(p []byte) (n int, err error) {
	if len(f.ReadData) == 0 {
		return 0, io.EOF
	}
	n = copy(p, f.ReadData)
	f.ReadData = f.ReadData[n:]
	return n, nil
}

func (f *FakeConn) Write(p []byte) (n int, err error) {
	n, err = f.WriteData.Write(p)
	if f.WriteNotify != nil {
		// non-blocking notify
		select {
		case f.WriteNotify <- struct{}{}:
		default:
		}
	}
	return n, err
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
	writeNotify := make(chan struct{}, 1)
	fakeConn := &FakeConn{
		ReadData:    prewritten,
		WriteNotify: writeNotify,
	}

	err := l.ConnectSTOMP(fakeConn)
	return fakeConn, err
}

func CompareBytes(t *testing.T, expected, actual []byte) {
	if len(expected) != len(actual) {
		t.Fatalf("Length not matching: %v\nACTUAL: %v", expected, actual)
	}
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Payload byte mismatch, index %d: %v\nACTUAL: %v", i, expected, actual)
		}
	}

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
	expected := []byte("CONNECT\nhost:test-host\nheart-beat:0,0\nlogin:test-id\npasscode:test-secret\naccept-version:1.2\n\n\x00")
	if len(expected) != len(actual) {
		t.Fatalf("Length not matching: %v\nACTUAL: %v", expected, actual)
	}
	for i := range actual {
		if actual[i] != expected[i] {
			t.Fatalf("Payload byte mismatch, index %d: %v\nACTUAL: %v", i, expected, actual)
		}
	}
}

func TestStompSubscription(t *testing.T) {
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
	if err := listener.Subscribe(); err != nil {
		t.Fatalf("Failed to subscibe %v", err)
	}
	var actual []byte
	select {
	case <-fakeConn.WriteNotify:
		actual, _ = io.ReadAll(&fakeConn.WriteData)

		// Proceed
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for CONNECT frame to be written")
	}
	select {
	case <-fakeConn.WriteNotify:
		// Proceed
	case <-time.After(1 * time.Second):
		t.Fatal("Timed out waiting for SUBSCRIBE frame to be written")
	}
	actual, err = io.ReadAll(&fakeConn.WriteData)
	expected := []byte("SUBSCRIBE\ndestination:actions.123.unit-test\nack:auto\nactivemq.prefetchSize:50\nid:actions.123.unit-test\n\n\x00")
	CompareBytes(t, expected, actual)
}
