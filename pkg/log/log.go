package log

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

const (
	HttpXRequestId = "X-Request-Id"
	CtxRequestId   = "requestId"
)

func InitLog(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Errorf("failed to parse log level: %v, err: %v", logLevel, err)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		DisableColors:   true,
		DisableQuote:    true,
		CallerPrettyfier: func(frame *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", path.Base(frame.File), frame.Line)
		},
	})
}

func GetLogger(c context.Context) *logrus.Entry {
	v := c.Value(CtxRequestId)
	if v != nil {
		return logrus.WithFields(logrus.Fields{
			CtxRequestId: v,
		})
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func NewLogger() *logrus.Entry {
	return logrus.NewEntry(logrus.StandardLogger())
}
