package uos

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type internalLogger struct{}

// Log provides the application logger
var Log *internalLogger

func setupLogging() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if config.Logging.UseConsole {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	switch config.Logging.Level {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	Log = &internalLogger{}
}

// LogContext specifies key-value context information for a log entry.
type LogContext map[string]interface{}

func errorLogContext(err error) LogContext {
	return LogContext{"error": err}
}

func appendLogContext(l *zerolog.Event, message string, context LogContext) {
	event := l

	for k, v := range context {
		switch vType := v.(type) {
		case error:
			event = event.AnErr(k, vType)
		case bool:
			event = event.Bool(k, vType)
		case time.Duration:
			event = event.Dur(k, vType)
		case float32:
			event = event.Float32(k, vType)
		case float64:
			event = event.Float64(k, vType)
		case net.IP:
			event = event.IPAddr(k, vType)
		case net.IPNet:
			event = event.IPPrefix(k, vType)
		case int:
			event = event.Int(k, vType)
		case int16:
			event = event.Int16(k, vType)
		case int32:
			event = event.Int32(k, vType)
		case int64:
			event = event.Int64(k, vType)
		case int8:
			event = event.Int8(k, vType)
		case net.HardwareAddr:
			event = event.MACAddr(k, vType)
		case string:
			event = event.Str(k, vType)
		case time.Time:
			event = event.Time(k, vType)
		case uint:
			event = event.Uint(k, vType)
		case uint16:
			event = event.Uint16(k, vType)
		case uint32:
			event = event.Uint32(k, vType)
		case uint64:
			event = event.Uint64(k, vType)
		case uint8:
			event = event.Uint8(k, vType)
		default:
			// convert to string
			event = event.Str(k, fmt.Sprintf("%v", v))
		}
	}

	l.Msg(message)
}

// PanicContext logs the specified message and context at log level 'panic'.
func (internalLogger) PanicContext(message string, context LogContext) {
	Metrics.CounterInc(mLogMessage)
	Metrics.CounterInc(mLogMessagePanic)
	appendLogContext(log.Panic(), message, context)
}

// FatalContext logs the specified message and context at log level 'fatal'.
func (internalLogger) FatalContext(message string, context LogContext) {
	Metrics.CounterInc(mLogMessage)
	Metrics.CounterInc(mLogMessagePanic)
	appendLogContext(log.Fatal(), message, context)
}

// ErrorContext logs the specified message and context at log level 'error'.
func (internalLogger) ErrorContext(message string, context LogContext) {
	Metrics.CounterInc(mLogMessage)
	Metrics.CounterInc(mLogMessageError)
	appendLogContext(log.Error(), message, context)
}

// WarnContext logs the specified message and context at log level 'warning'.
func (internalLogger) WarnContext(message string, context LogContext) {
	Metrics.CounterInc(mLogMessage)
	Metrics.CounterInc(mLogMessageWarning)
	appendLogContext(log.Warn(), message, context)
}

// InfoContext logs the specified message and context at log level 'info'.
func (internalLogger) InfoContext(message string, context LogContext) {
	Metrics.CounterInc(mLogMessage)
	appendLogContext(log.Info(), message, context)
}

// DebugContext logs the specified message and context at log level 'debug'.
func (internalLogger) DebugContext(message string, context LogContext) {
	appendLogContext(log.Debug(), message, context)
}

// TraceContext logs the specified message and context at log level 'trace'.
func (internalLogger) TraceContext(message string, context LogContext) {
	appendLogContext(log.Trace(), message, context)
}

// Panic logs the specified message at log level 'panic'.
func (l *internalLogger) Panic(message string) {
	l.PanicContext(message, nil)
}

// Fatal logs the specified message at log level 'fatal'.
func (l *internalLogger) Fatal(message string) {
	l.FatalContext(message, nil)
}

// Error logs the specified message at log level 'error'.
func (l *internalLogger) Error(message string) {
	l.ErrorContext(message, nil)
}

// Warn logs the specified message at log level 'warning'.
func (l *internalLogger) Warn(message string) {
	l.WarnContext(message, nil)
}

// Info logs the specified message at log level 'info'.
func (l *internalLogger) Info(message string) {
	l.InfoContext(message, nil)
}

// Debug logs the specified message at log level 'debug'.
func (l *internalLogger) Debug(message string) {
	l.DebugContext(message, nil)
}

// Trace logs the specified message at log level 'trace'.
func (l *internalLogger) Trace(message string) {
	l.TraceContext(message, nil)
}

// PanicError logs the specified message and error at log level 'panic'. Panics.
func (l *internalLogger) PanicError(message string, err error) {
	l.PanicContext(message, errorLogContext(err))
	panic(message)
}

// FatalError logs the specified message and error at log level 'fatal'.
func (l *internalLogger) FatalError(message string, err error) {
	l.FatalContext(message, errorLogContext(err))
}

// ErrorObj logs the specified message and error at log level 'error'.
func (l *internalLogger) ErrorObj(message string, err error) {
	l.ErrorContext(message, errorLogContext(err))
}

// Warn logs the specified message at and error log level 'warning'.
func (l *internalLogger) WarnError(message string, err error) {
	l.WarnContext(message, errorLogContext(err))
}

// Info logs the specified message at and error log level 'info'.
func (l *internalLogger) InfoError(message string, err error) {
	l.InfoContext(message, errorLogContext(err))
}

// Debug logs the specified message and error at log level 'debug'.
func (l *internalLogger) DebugError(message string, err error) {
	l.DebugContext(message, errorLogContext(err))
}
