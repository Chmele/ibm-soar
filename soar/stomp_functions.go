package soar

import (
	"fmt"
	"log/slog"
)

// A FunctionCallHandler that is logging the fact of a message recieving
func LoggingResponse(c *FunctionCall) (*FuncResponse, error) {
	slog.Info("Recieved function call", slog.String("function_name", c.Function.Name))
	return nil, nil
}

// A FunctionCallHandler updating the status of a playbook instance currently running a function that a function run is started
func StartedResponse(*FunctionCall) (*FuncResponse, error) {
	return &FuncResponse{
		MessageType: 0,
		Message:     "Starting App Function",
		Complete:    false,
	}, nil
}

// A FunctionCallHandler updating the status of a playbook instance currently running a function with result response
func CompletedResponse(c *FunctionCall) (*FuncResponse, error) {
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

// Not implemented
type FunctionLookup struct {
	mapping map[string]FunctionCallHandler
}

func (l *FunctionLookup) Register(name string, handler FunctionCallHandler) {
	l.mapping[name] = handler
}

func(l *FunctionLookup) Handler(c *FunctionCall) (*FuncResponse, error) {
	f, ok := l.mapping[c.Function.Name]
	if !ok {
		return nil, fmt.Errorf("Got a call with unregistered function name: %s", c.Function.Name)
	}
	return f(c)
}
