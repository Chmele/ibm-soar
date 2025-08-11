package soar

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/go-stomp/stomp/v3"
)

// Structure representing a single STOMP connection to a single SOAR message destination
type StompListener struct {
	HTTPClient         *HTTPClient
	StompPort          string
	MessageDestination string
	Ctx                context.Context
	Done               chan struct{}
	Insecure           bool
	Subscription       *stomp.Subscription
	Conn               *stomp.Conn
	Logger             *slog.Logger
}

// Creates a stomp listener with connectivity and access check
func NewStompListener(h *HTTPClient, opts ...StompOption) (*StompListener, error) {
	ret := &StompListener{
		HTTPClient: h,
		StompPort:  "65001",
		Ctx:        context.Background(),
		Done:       make(chan struct{}),
		Insecure:   false,
	}
	for _, opt := range opts {
		err := opt(ret)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// Main entry point for stomp listening
func (l *StompListener) Listen(f ...FunctionCallHandler) error {
	netConn, err := l.connectTLS()
	if err != nil {
		return err
	}

	if err := l.connectSTOMP(netConn); err != nil {
		return err
	}
	l.Logger.Info("Connected to STOMP")

	if err := l.subscribe(); err != nil {
		return err
	}
	l.Logger.Info("Subscribed to queue",
		slog.String("message_destination", l.MessageDestination))
	go func() {
		defer close(l.Done)
		l.stompLoop(f...)
	}()
	return nil
}

// Fancy stuff for supporting insecure connections
func (l *StompListener) connectTLS() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return tls.DialWithDialer(
		dialer,
		"tcp",
		fmt.Sprintf("%s:%s", l.HTTPClient.Hostname, l.StompPort),
		&tls.Config{InsecureSkipVerify: l.Insecure},
	)

}

// Use stomp library in this func to set up Conn field
func (l *StompListener) connectSTOMP(connection net.Conn) error {
	conn, err := stomp.Connect(connection,
		stomp.ConnOpt.Login(l.HTTPClient.KeyId, l.HTTPClient.KeySecret),
		stomp.ConnOpt.AcceptVersion(stomp.V12),
		stomp.ConnOpt.Host(l.HTTPClient.Hostname),
		stomp.ConnOpt.HeartBeat(0, 0),
		stomp.ConnOpt.Logger(&StompLogger{l.Logger}),
	)
	if err != nil {
		return err
	}
	l.Conn = conn
	return nil
}

// Endless listening loop for message channel
func (l *StompListener) stompLoop(f ...FunctionCallHandler) error {
	defer func() {
		l.Logger.Info("STOMP Disconnecting")
		l.Conn.Disconnect()
	}()
	defer func() {
		l.Logger.Info("STOMP Unsubscribing")
		l.Subscription.Unsubscribe()
	}()
	errCh := make(chan error)
	for {
		select {
		case <-l.Ctx.Done():
			l.Logger.Info("STOMP is shutting down")
			return nil
		case msg, ok := <-l.Subscription.C:
			if !ok {
				return errors.New("Attempted to read closed STOMP channel")
			}
			go func() {
				errCh <- l.handleFunc(f...)(msg)
			}()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

// The function that is either writes the response to a message or returns an error
type ProcessFunc func(*stomp.Message) error

// The function used as a part of a response to message (updates the processing status, logs the message, etc.)
type FunctionCallHandler func(*FunctionCall) (*FuncResponse, error)

// Fancy hardcoded internal queue names here and "just working" constants
func (l *StompListener) subscribe() error {
	Id := fmt.Sprintf("actions.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination)
	sub, err := l.Conn.Subscribe(
		Id,
		stomp.AckAuto,
		stomp.SubscribeOpt.Header("activemq.prefetchSize", "50"),
		stomp.SubscribeOpt.Id(Id),
	)
	if err != nil {
		return err
	}
	l.Subscription = sub
	return nil
}

// Calling handlers one-by-one, responding with updated run statuses and result as handlers suggest (JSON)
func (l *StompListener) handleFunc(functions ...FunctionCallHandler) ProcessFunc {
	return func(msg *stomp.Message) error {
		fc, err := parseFunctionMessage(msg.Body)
		if err != nil {
			return err
		}
		for _, f := range functions {
			fr, err := f(fc)
			// no-op handler
			if fr == nil {
				continue
			}
			if err != nil {
				return err
			}
			body, err := json.Marshal(fr)
			if err != nil {
				return err
			}
			if err := l.sendFunctionResponse(msg, body); err != nil {
				return err
			}
		}
		return nil
	}
}

// Responds, specifing the message it responds to with the bytes provided
func (l *StompListener) sendFunctionResponse(msg *stomp.Message, body []byte) error {
	correlationID := msg.Header.Get("correlation-id")
	return l.Conn.Send(
		fmt.Sprintf("acks.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination),
		"application/json",
		body,
		stomp.SendOpt.Header("correlation-id", correlationID),
	)
}

// Decodes received function call STOMP message
func parseFunctionMessage(b []byte) (*FunctionCall, error) {
	call := new(FunctionCall)
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(call); err != nil {
		return nil, err
	}
	return call, nil
}
