package logger

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/rs/zerolog"
)

var (
	log  *zerolog.Logger
	once sync.Once
)

// init 初始化日志实例
func init() {
	once.Do(func() {
		// 创建一个 ConsoleWriter
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		// 创建一个新的日志实例，使用 ConsoleWriter
		logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

		log = &logger
	})
}

// 打印信息日志，带有自定义格式
func baseMsg(msg string, e *zerolog.Event, v ...interface{}) {
	// 获取调用者信息
	_, file, line, ok := runtime.Caller(2) // 获取调用 Info 的文件和行号
	if !ok {
		file = "unknown"
		line = 0
	}
	// 格式化日志消息
	formattedMsg := fmt.Sprintf(msg, v...)
	// 打印自定义格式的日志，包含调用者信息
	e.Msgf("%s:%d] %s", file, line, formattedMsg)
}

func Info(msg string, v ...interface{}) {
	baseMsg(msg, log.Info(), v...)
}

func Warning(msg string, v ...interface{}) {
	baseMsg(msg, log.Warn(), v...)
}

func Error(msg string, v ...interface{}) {
	baseMsg(msg, log.Error(), v...)
}

func Debug(msg string, v ...interface{}) {
	baseMsg(msg, log.Debug(), v)
}

func Fatal(msg string, v ...interface{}) {
	baseMsg(msg, log.Fatal(), v)
}
