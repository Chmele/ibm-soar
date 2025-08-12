package main

import (
	"context"
	"flag"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/chmele/ibm-soar/soar"
	"github.com/chmele/ibm-soar/soar/structures"
)

type HTTPInputs struct {
	Method string `json:"http_method"`
	Url    string `json:"http_url"`
}


func HTTPRequest(fc *structures.FunctionCall) (*structures.FuncResponse, error) {
	inputs, err := soar.LoadInputs[HTTPInputs](fc)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(inputs.Method, inputs.Url, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return soar.SuccessResponse(string(body)), nil
}

func main() {
	tokenId := flag.String("tokenId", "", "Token ID")
	tokenSecret := flag.String("tokenSecret", "", "Token Secret")
	ipAddr := flag.String("ip", "", "Server IP address")
	insecure := flag.Bool("insecure", true, "Use insecure connection")
	destination := flag.String("destination", "http", "STOMP message destination")

	flag.Parse()

	if *tokenId == "" || *tokenSecret == "" || *ipAddr == "" {
		log.Fatalf("All flags -tokenId, -tokenSecret and -ip are required %s, %s, %s", *tokenId, *tokenSecret, *
ipAddr)
	}

	ctx := context.Background()
	client, err := soar.NewHTTPClient(ctx, *ipAddr, *tokenId, *tokenSecret, *insecure)
	if err != nil {
		log.Fatalf("Error while checking connectivity: %v", err)
	}

	stompListener, err := soar.NewStompListener(client,
		soar.Stomp.Context(ctx),
		soar.Stomp.Insecure(*insecure),
		soar.Stomp.Logger(slog.Default()),
		soar.Stomp.MessageDestination(*destination),
	)
	if err != nil {
		log.Fatalf("Error creating STOMP listener: %v", err)
	}

	err = stompListener.Listen(soar.LoggingResponse, HTTPRequest)
	if err != nil {
		log.Fatalf("Error while STOMPing: %v", err)
	}
	<-stompListener.Done
}
