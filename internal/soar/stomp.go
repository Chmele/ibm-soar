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

type StompListener struct {
	HTTPClient         *HTTPClient
	StompPort          string
	MessageDestination string
	Ctx                context.Context
	Done               chan struct{}
	Insecure           bool
	Subscription       *stomp.Subscription
	Conn               *stomp.Conn
}

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

func (l *StompListener) Listen(f ...FunctionCallHandler) error {
	netConn, err := l.ConnectTLS()
	if err != nil {
		return err
	}

	if err := l.ConnectSTOMP(netConn); err != nil {
		return err
	}
	slog.Info("Connected to STOMP")

	if err := l.Subscribe(); err != nil {
		return err
	}
	slog.Info("Subscribed to queue",
		slog.String("message_destination", l.MessageDestination))
	go func() {
		defer close(l.Done)
		l.STOMPLoop(f...)
	}()
	return nil
}

func (l *StompListener) ConnectTLS() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return tls.DialWithDialer(
		dialer,
		"tcp",
		fmt.Sprintf("%s:%s", l.HTTPClient.Hostname, l.StompPort),
		&tls.Config{InsecureSkipVerify: l.Insecure},
	)

}

func (l *StompListener) ConnectSTOMP(connection net.Conn) error {
	conn, err := stomp.Connect(connection,
		stomp.ConnOpt.Login(l.HTTPClient.KeyId, l.HTTPClient.KeySecret),
		stomp.ConnOpt.AcceptVersion(stomp.V12),
		stomp.ConnOpt.Host(l.HTTPClient.Hostname),
		stomp.ConnOpt.HeartBeat(0, 0),
	)
	if err != nil {
		return err
	}
	l.Conn = conn
	return nil
}

func (l *StompListener) STOMPLoop(f ...FunctionCallHandler) error {
	defer func() {
		slog.Info("STOMP Disconnecting")
		l.Conn.Disconnect()
	}()
	defer func() {
		slog.Info("STOMP Unsubscribing")
		l.Subscription.Unsubscribe()
	}()
	errCh := make(chan error)
	for {
		select {
		case <-l.Ctx.Done():
			slog.Info("STOMP is shutting down")
			return nil
		case msg, ok := <-l.Subscription.C:
			if !ok {
				return errors.New("Attempted to read closed STOMP channel")
			}
			go func() {
				errCh <- l.HandleFunc(f...)(msg)
			}()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

type ProcessFunc func(*stomp.Message) error
type FunctionCallHandler func(*FunctionCall) (*FuncResponse, error)

func (l *StompListener) Subscribe() error {
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

func (l *StompListener) HandleFunc(functions ...FunctionCallHandler) ProcessFunc {
	return func(msg *stomp.Message) error {
		fc, err := ParseFunctionMessage(msg.Body)
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
			if err := l.SendFunctionResponse(msg, body); err != nil {
				return err
			}
		}
		return nil
	}
}

func (l *StompListener) SendFunctionResponse(msg *stomp.Message, body []byte) error {
	correlationID := msg.Header.Get("correlation-id")
	return l.Conn.Send(
		fmt.Sprintf("acks.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination),
		"application/json",
		body,
		stomp.SendOpt.Header("correlation-id", correlationID),
	)
}

func ParseFunctionMessage(b []byte) (*FunctionCall, error) {
	call := new(FunctionCall)
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(call); err != nil {
		return nil, err
	}
	return call, nil
}
