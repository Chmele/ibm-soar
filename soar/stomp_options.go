package soar

import (
	"context"
	"fmt"
)


type StompOption func(*StompListener) error
type StompOpts struct {}
var Stomp = StompOpts{}
func (StompOpts) MessageDestination(md string) func(*StompListener) error {
	return func(l *StompListener) error {
		l.MessageDestination = md
		return nil
	}
}
func (StompOpts) Port(port int) func(*StompListener) error {
	return func(l *StompListener) error {
		l.StompPort = fmt.Sprint(port)
		return nil
	}
}
func (StompOpts) Context(ctx context.Context) func(*StompListener) error {
	return func(l *StompListener) error {
		l.Ctx = ctx
		return nil
	}
}
func (StompOpts) Insecure(insecure bool) func(*StompListener) error {
	return func(l *StompListener) error {
		l.Insecure = insecure
		return nil
	}
}

