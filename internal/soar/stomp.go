package soar

import (
	"bytes"
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
}

func NewStompListener(h *HTTPClient, stompPort string, messageDestination string) *StompListener {
	return &StompListener{
		HTTPClient:         h,
		StompPort:          stompPort,
		MessageDestination: messageDestination,
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
	)
	if err != nil {
		return err
	}

	sub, err := l.subscribe(stompConn)
	if err != nil {
		return err
	}
	go l.STOMPLoop(sub, stompConn)
	return nil
}

func (l *StompListener) STOMPLoop(sub *stomp.Subscription, stompConn *stomp.Conn) {
	var err error
	defer stompConn.Disconnect()
	for msg := range sub.C {
		if msg.Err != nil {
			log.Printf("Error receiving message: %v", msg.Err)
			continue
		}
		go l.processFunctionMessage(msg)
		go func() {
			err = l.respondToFunctionMessage(msg, stompConn, FuncResponse{
				MessageType: 0,
				Message:     "Starting App Function ㊡",
				Complete:    false,
			})
			if err != nil {
				log.Printf("Error sending message %v", err)
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
			}
		}()
	}
}

func (l *StompListener) subscribe(conn *stomp.Conn) (*stomp.Subscription, error) {
	return conn.Subscribe(
		fmt.Sprintf("actions.%d.%s", l.HTTPClient.Org.ID, l.MessageDestination),
		stomp.AckAuto,
		stomp.SubscribeOpt.Header("activemq.prefetchSize", "50"),
		stomp.SubscribeOpt.Id("repl"),
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
	return conn.Send("acks.201.test", "application/json", body, stomp.SendOpt.Header("correlation-id", correlationID))
}
