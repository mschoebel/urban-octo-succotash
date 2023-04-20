package uos

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

// LogPanicContext logs the specified message and context at log level 'panic'.
func LogPanicContext(message string, context LogContext) {
	appendLogContext(log.Panic(), message, context)
}

// LogFatalContext logs the specified message and context at log level 'fatal'.
func LogFatalContext(message string, context LogContext) {
	appendLogContext(log.Fatal(), message, context)
}

// LogErrorContext logs the specified message and context at log level 'error'.
func LogErrorContext(message string, context LogContext) {
	appendLogContext(log.Error(), message, context)
}

// LogWarnContext logs the specified message and context at log level 'warning'.
func LogWarnContext(message string, context LogContext) {
	appendLogContext(log.Warn(), message, context)
}

// LogInfoContext logs the specified message and context at log level 'info'.
func LogInfoContext(message string, context LogContext) {
	appendLogContext(log.Info(), message, context)
}

// LogDebugContext logs the specified message and context at log level 'debug'.
func LogDebugContext(message string, context LogContext) {
	appendLogContext(log.Debug(), message, context)
}

// LogTraceContext logs the specified message and context at log level 'trace'.
func LogTraceContext(message string, context LogContext) {
	appendLogContext(log.Trace(), message, context)
}

// LogPanic logs the specified message at log level 'panic'.
func LogPanic(message string) {
	LogPanicContext(message, nil)
}

// LogFatal logs the specified message at log level 'fatal'.
func LogFatal(message string) {
	LogFatalContext(message, nil)
}

// LogError logs the specified message at log level 'error'.
func LogError(message string) {
	LogErrorContext(message, nil)
}

// LogWarn logs the specified message at log level 'warning'.
func LogWarn(message string) {
	LogWarnContext(message, nil)
}

// LogInfo logs the specified message at log level 'info'.
func LogInfo(message string) {
	LogInfoContext(message, nil)
}

// LogDebug logs the specified message at log level 'debug'.
func LogDebug(message string) {
	LogDebugContext(message, nil)
}

// LogTrace logs the specified message at log level 'trace'.
func LogTrace(message string) {
	LogTraceContext(message, nil)
}

// LogPanicError logs the specified message and error at log level 'panic'. Panics.
func LogPanicError(message string, err error) {
	LogPanicContext(message, errorLogContext(err))
	panic(message)
}

// LogFatalError logs the specified message and error at log level 'fatal'.
func LogFatalError(message string, err error) {
	LogFatalContext(message, errorLogContext(err))
}

// LogErrorObj logs the specified message and error at log level 'error'.
func LogErrorObj(message string, err error) {
	LogErrorContext(message, errorLogContext(err))
}

// LogWarn logs the specified message at and error log level 'warning'.
func LogWarnError(message string, err error) {
	LogWarnContext(message, errorLogContext(err))
}

// LogInfo logs the specified message at and error log level 'info'.
func LogInfoError(message string, err error) {
	LogInfoContext(message, errorLogContext(err))
}

// LogDebug logs the specified message and error at log level 'debug'.
func LogDebugError(message string, err error) {
	LogDebugContext(message, errorLogContext(err))
}
