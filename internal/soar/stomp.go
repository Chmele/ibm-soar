package soar

import (
	"bytes"
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
}

func NewStompListener(ctx context.Context, h *HTTPClient, stompPort string, messageDestination string) *StompListener {
	return &StompListener{
		HTTPClient:         h,
		StompPort:          stompPort,
		MessageDestination: messageDestination,
		Ctx:                ctx,
		Done:               make(chan struct{}),
	}
}

func (l *StompListener) Listen() error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	netConn, err := tls.DialWithDialer(
		dialer,
		"tcp",
		fmt.Sprintf("%s:%s", l.HTTPClient.Hostname, l.StompPort),
		&tls.Config{InsecureSkipVerify: true},
	)
	if err != nil {
		return err
	}

	stompConn, err := stomp.Connect(netConn,
		stomp.ConnOpt.Login(l.HTTPClient.KeyId, l.HTTPClient.KeySecret),
		stomp.ConnOpt.AcceptVersion(stomp.V12),
		stomp.ConnOpt.Host(l.HTTPClient.Hostname),
		stomp.ConnOpt.HeartBeat(0, 0),
	)
	if err != nil {
		return err
	}

	log.Println("󱘖 Connected to STOMP")
	sub, err := l.subscribe(stompConn)
	if err != nil {
		return err
	}
	log.Println("󱚣 Subscribed to queue", l.MessageDestination)
	go func() {
		defer close(l.Done)
		l.STOMPLoop(sub, stompConn)
	}()
	return nil
}

func (l *StompListener) STOMPLoop(sub *stomp.Subscription, stompConn *stomp.Conn) error {
	defer func() {
		log.Println("STOMP Disconnecting")
		stompConn.Disconnect()
	}()
	defer func() {
		log.Println("STOMP Unsubscribing")
		sub.Unsubscribe()
	}()
	errCh := make(chan error)
	for {
		select {
		case <-l.Ctx.Done():
			log.Println("STOMP is shutting down")
			return nil
		case msg, ok := <-sub.C:
			if !ok {
				log.Println("STOMP channel closed")
				return nil
			}
			go func() {
				errCh <- l.CompleteFunc(msg, stompConn)
			}()
		case err := <- errCh:
			if err != nil {
				return err
			}
		}
	}
}

type ProcessFunc func(*stomp.Message, *stomp.Conn)

func (l *StompListener) CompleteFunc(msg *stomp.Message, stompConn *stomp.Conn) error {
	if msg.Err != nil {
		log.Printf("Received error message: %v", msg.Err)
		return msg.Err
	}
	l.processFunctionMessage(msg)
	err := l.respondToFunctionMessage(msg, stompConn, FuncResponse{
		MessageType: 0,
		Message:     "Starting App Function ㊡",
		Complete:    false,
	})
	if err != nil {
		log.Printf("Error sending message %v", err)
		return err
	}
	err = l.respondToFunctionMessage(msg, stompConn, FuncResponse{
		MessageType: 2,
		Message:     "App function completed",
		Complete:    true,
		Results: &Results{
			Version: 2.0,
			Success: true,
			Reason:  nil,
			Content: "Plain text content",
			Raw:     nil,
		},
	})
	if err != nil {
		log.Printf("Error sending message %v", err)
		return err
	}
	return nil
}

func (l *StompListener) subscribe(conn *stomp.Conn) (*stomp.Subscription, error) {
	Id := fmt.Sprintf("actions.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination)
	return conn.Subscribe(
		Id,
		stomp.AckClientIndividual,
		stomp.SubscribeOpt.Header("activemq.prefetchSize", "50"),
		stomp.SubscribeOpt.Id(Id),
	)
}

func (l *StompListener) processFunctionMessage(msg *stomp.Message) error {
	fc := new(FunctionCall)
	if err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(fc); err != nil {
		return err
	}
	log.Printf("󰊕 Got function call STOMP message: %s", fc.Inputs)
	return nil
}

func (l *StompListener) respondToFunctionMessage(msg *stomp.Message, conn *stomp.Conn, fr FuncResponse) error {
	correlationID := msg.Header.Get("correlation-id")
	body, err := json.Marshal(fr)
	if err != nil {
		return err
	}
	return conn.Send(
		fmt.Sprintf("acks.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination),
		"application/json",
		body,
		stomp.SendOpt.Header("correlation-id", correlationID),
	)
}
