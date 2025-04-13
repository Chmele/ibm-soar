package soar

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
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

func NewStompListener(ctx context.Context, h *HTTPClient, stompPort string, messageDestination string, insecure bool) *StompListener {
	return &StompListener{
		HTTPClient:         h,
		StompPort:          stompPort,
		MessageDestination: messageDestination,
		Ctx:                ctx,
		Done:               make(chan struct{}),
		Insecure:           insecure,
	}
}

func (l *StompListener) Listen(f MessageHandler) error {
	netConn, err := l.ConnectTLS()
	if err != nil {
		return err
	}
	
	if err := l.ConnectSTOMP(netConn); err != nil {
		return err
	}
	log.Println("󱘖 Connected to STOMP")

	if err := l.Subscribe(); err != nil {
		return err
	}
	log.Println("󱚣 Subscribed to queue", l.MessageDestination)
	go func() {
		defer close(l.Done)
		l.STOMPLoop(f)
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

func (l *StompListener) STOMPLoop(f MessageHandler) error {
	defer func() {
		log.Println("STOMP Disconnecting")
		l.Conn.Disconnect()
	}()
	defer func() {
		log.Println("STOMP Unsubscribing")
		l.Subscription.Unsubscribe()
	}()
	errCh := make(chan error)
	for {
		select {
		case <-l.Ctx.Done():
			log.Println("STOMP is shutting down")
			return nil
		case msg, ok := <-l.Subscription.C:
			if !ok {
				log.Println("STOMP channel closed")
				return nil
			}
			go func() {
				errCh <- l.HandleFunc(f)(msg)
			}()
		case err := <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

type MessageHandler func(*stomp.Message) (*FuncResponse, error)
type ProcessFunc func(*stomp.Message) error

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

func (l *StompListener) HandleFunc(f func(*stomp.Message) (*FuncResponse, error)) ProcessFunc {
	return func(msg *stomp.Message) error {
		fr, err := f(msg)
		if err != nil {
			return err
		}
		body, err := json.Marshal(fr)
		if err != nil {
			return err
		}
		return l.SendFunctionResponse(msg, body)
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
