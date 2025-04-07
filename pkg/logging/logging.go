package logging

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"club.asynclab/asrp/pkg/config"
	"github.com/sirupsen/logrus"
)

type GeneralFormatter struct {
	IsVerbose bool
}

func (f *GeneralFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &strings.Builder{}
	if entry.Buffer != nil {
		b.Write(entry.Buffer.Bytes())
	}

	now := time.Now().Format("15:04:05")
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message

	showPath := ""
	_, file, line, ok := runtime.Caller(6)
	if !ok {
		showPath = "unknown_file"
		line = 0
	} else {
		showPath = file
	}

	if f.IsVerbose {
		fmt.Fprintf(b, "[%s %5s] [%s:%d]: %s\n", now, level, showPath, line, message)
	} else {
		fmt.Fprintf(b, "[%s %5s]: %s\n", now, level, message)
	}
	return []byte(b.String()), nil
}

var logger = logrus.New()

func Init(isVerbose bool) {
	logger.SetFormatter(&GeneralFormatter{IsVerbose: isVerbose})
	if config.IsVerbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
}

func GetLogger() *logrus.Logger {
	return logger
}

func init() {
	Init(false)
}
