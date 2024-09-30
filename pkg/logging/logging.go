package logging

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type GeneralFormatter struct{}

func (f *GeneralFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &strings.Builder{}

	if entry.Buffer != nil {
		b.Write(entry.Buffer.Bytes())
	}

	now := time.Now().Format("15:04:05")
	level := strings.ToUpper(entry.Level.String())
	message := entry.Message

	_, file, line, ok := runtime.Caller(6)
	if !ok {
		file = "unknown_file"
		line = 0
	} else {
		file = file[strings.LastIndex(file, "/")+1:]
	}

	// 构建格式化的日志输出
	fmt.Fprintf(b, "[%s %s] [%s:%d]: %s\n", now, level, file, line, message)

	return []byte(b.String()), nil
}

var logger = logrus.New()

func init() {
	logger.SetFormatter(&GeneralFormatter{})
}

func GetLogger() *logrus.Logger {
	return logger
}
