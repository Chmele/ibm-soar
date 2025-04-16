package soar

import (
	"fmt"
	"log/slog"
)

type StompLogger struct {
	Logger *slog.Logger
}

func (l *StompLogger) Debugf(format string, v ...any) {
	l.Logger.Debug(fmt.Sprintf(format, v...))
}
func (l *StompLogger) Infof(format string, v ...any) {
	l.Logger.Info(fmt.Sprintf(format, v...))
}
func (l *StompLogger) Warningf(format string, v ...any) {
	l.Logger.Warn(fmt.Sprintf(format, v...))
}
func (l *StompLogger) Errorf(format string, v ...any) {
	l.Logger.Error(fmt.Sprintf(format, v...))
}
func (l *StompLogger) Debug(msg string)   { l.Logger.Debug(msg) }
func (l *StompLogger) Info(msg string)    { l.Logger.Info(msg) }
func (l *StompLogger) Warning(msg string) { l.Logger.Warn(msg) }
func (l *StompLogger) Error(msg string)   { l.Logger.Error(msg) }
