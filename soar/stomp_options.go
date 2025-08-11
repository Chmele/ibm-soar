package soar

import (
	"context"
	"fmt"
	"log/slog"
)

type StompOption func(*StompListener) error
type StompOpts struct{}

var Stomp = StompOpts{}

// Message destination api name
func (StompOpts) MessageDestination(md string) func(*StompListener) error {
	return func(l *StompListener) error {
		l.MessageDestination = md
		return nil
	}
}

// Port to be used in a STOMP connection, 65001 by default
func (StompOpts) Port(port int) func(*StompListener) error {
	return func(l *StompListener) error {
		l.StompPort = fmt.Sprint(port)
		return nil
	}
}

// Context to be used in a listening
func (StompOpts) Context(ctx context.Context) func(*StompListener) error {
	return func(l *StompListener) error {
		l.Ctx = ctx
		return nil
	}
}

// Whether to trust the self-signed certificate or not
func (StompOpts) Insecure(insecure bool) func(*StompListener) error {
	return func(l *StompListener) error {
		l.Insecure = insecure
		return nil
	}
}

// Structured logger to be used for runtime logs
func (StompOpts) Logger(logger *slog.Logger) func(*StompListener) error {
	return func(l *StompListener) error {
		l.Logger = logger
		return nil
	}
}
