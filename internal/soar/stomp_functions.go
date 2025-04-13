package soar

import (
	"log"
)


func LoggingResponse (c *FunctionCall) (*FuncResponse, error) {
	log.Printf("󰡱 Got Message: %s", c.Function.Name)
	return nil, nil
}

func StartedResponse (*FunctionCall) (*FuncResponse, error) {
	return &FuncResponse{
		MessageType: 0,
		Message:     "Starting App Function ㊡",
		Complete:    false,
	}, nil
}

func CompletedResponse (c *FunctionCall) (*FuncResponse, error) {
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


