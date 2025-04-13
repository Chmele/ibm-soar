package soar

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/go-stomp/stomp/v3"
)



func StartedResponse (*stomp.Message) (*FuncResponse, error) {
	return &FuncResponse{
		MessageType: 0,
		Message:     "Starting App Function ㊡",
		Complete:    false,
	}, nil
}

func CompletedResponse (m *stomp.Message) (*FuncResponse, error) {
	call, err := ParseFunctionMessage(m.Body)
	if err != nil {
		return nil, err
	}
	log.Printf("󰡱  Got Message: %s", call.Function.Name)
	return &FuncResponse{
		MessageType: 2,
		Message:     "App function completed",
		Complete:    true,
		Results: &Results{
			Version: 2.0,
			Success: true,
			Content: "Plain text content",
		},
	}, nil
}

func ParseFunctionMessage(b []byte) (*FunctionCall, error) {
	call := new(FunctionCall)
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(call); err != nil {
		return nil, err
	}
	return call, nil
}

