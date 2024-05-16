// copied and modified from https://github.com/samber/slog-formatter/blob/4ce6c72294ee982ed459d49a2faac326a257181e/formatter_error.go

package logger

import (
	"log/slog"
	"reflect"
	"regexp"
	"runtime"
)

// ErrorFormatter transforms a go error into a readable error.
//
// Example:
//
//	err := reader.Close()
//	err = fmt.Errorf("could not close reader: %v", err)
//	logger.With("error", reader.Close()).Log("error")
//
// passed to ErrorFormatter("error"), will be transformed into:
//
//	"error": {
//	  "message": "could not close reader: file already closed",
//	  "type": "*io.ErrClosedPipe"
//	}
func ErrorFormatter(fieldName string) Formatter {
	return FormatByFieldType(fieldName, func(err error) slog.Value {
		values := []slog.Attr{
			slog.String("message", err.Error()),
			slog.String("type", reflect.TypeOf(err).String()),
			slog.String("stacktrace", stacktrace()),
		}

		return slog.GroupValue(values...)
	})
}

var reStacktrace = regexp.MustCompile(`log/slog.*\n`)

func stacktrace() string {
	stackInfo := make([]byte, 1024*1024)

	if stackSize := runtime.Stack(stackInfo, false); stackSize > 0 {
		traceLines := reStacktrace.Split(string(stackInfo[:stackSize]), -1)
		if len(traceLines) > 0 {
			return traceLines[len(traceLines)-1]
		}
	}

	return ""
}
