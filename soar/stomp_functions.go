package soar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/chmele/ibm-soar/soar/structures"
)

// A FunctionCallHandler that is logging the fact of a message recieving
func LoggingResponse(c *structures.FunctionCall) (*structures.FuncResponse, error) {
	slog.Info("Recieved function call", slog.String("function_name", c.Function.Name))
	return nil, nil
}

// A FunctionCallHandler updating the status of a playbook instance currently running a function that a function run is started
func StartedResponse(*structures.FunctionCall) (*structures.FuncResponse, error) {
	return &structures.FuncResponse{
		MessageType: 0,
		Message:     "Starting App Function",
		Complete:    false,
	}, nil
}

// A FunctionCallHandler updating the status of a playbook instance currently running a function with result response
func CompletedResponse(c *structures.FunctionCall) (*structures.FuncResponse, error) {
	return &structures.FuncResponse{
		MessageType: 2,
		Message:     "App function completed",
		Complete:    true,
		Results: &structures.Results{
			Version: 2.0,
			Success: true,
			Content: "Plain text content",
		},
	}, nil
}

func ErrorResponse(c *structures.FunctionCall, err error) *structures.FuncResponse {
	return &structures.FuncResponse{
		MessageType: 3,
		Message:     fmt.Sprintf("Error occured: %v", err),
		Complete:    true,
		Results: &structures.Results{
			Version: 2.0,
			Success: false,
			Content: nil,
		},
	}
}

func SuccessResponse(text string) *structures.FuncResponse {
	return &structures.FuncResponse{
		MessageType: 0,
		Message:     "App function completed",
		Complete:    true,
		Results: &structures.Results{
			Version: 2.0,
			Success: true,
			Content: text,
		},
	}
}

func LoadInputs[Input any](fc *structures.FunctionCall) (*Input, error) {
	var inputMap map[string]any
	if m, ok := fc.Inputs.(map[string]any); ok {
		inputMap = m
	} else {
		return nil, fmt.Errorf("Function input is not a map of strings")
	}
	data, err := json.Marshal(inputMap)
	if err != nil {
		return nil, err
	}
	var ret Input
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

// Not implemented
type FunctionLookup struct {
	mapping map[string]FunctionCallHandler
}

func (l *FunctionLookup) Register(name string, handler FunctionCallHandler) {
	l.mapping[name] = handler
}

func (l *FunctionLookup) Handler(c *structures.FunctionCall) (*structures.FuncResponse, error) {
	f, ok := l.mapping[c.Function.Name]
	if !ok {
		return nil, fmt.Errorf("Got a call with unregistered function name: %s", c.Function.Name)
	}
	return f(c)
}
