package logging

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"time"
)

type Record struct {
	Time       time.Time
	Level      logLevel
	Message    string
	LoggerName string
}

type Emitter interface {
	Emit(string, *Record)
}

type Logger struct {
	Name     string
	Handlers map[string]Emitter
}

func NewLogger() *Logger {
	return &Logger{Handlers: make(map[string]Emitter)}
}

var DefaultLogger = NewLogger()

func (l *Logger) AddHandler(name string, h Emitter) {
	oldHandler, ok := l.Handlers[name]
	if ok {
		closer, ok := oldHandler.(io.Closer)
		if ok {
			_ = closer.Close()
		}
	}
	l.Handlers[name] = h
}

func (l *Logger) Log(level logLevel, format string, values ...interface{}) {
	rd := &Record{
		Time:       time.Now(),
		Level:      level,
		Message:    fmt.Sprintf(format, values...),
		LoggerName: l.Name,
	}
	for name, h := range l.Handlers {
		h.Emit(name, rd)
	}
}

func (l *Logger) Debug(format string, values ...interface{}) {
	l.Log(DEBUG, format, values...)
}

func (l *Logger) Info(format string, values ...interface{}) {
	l.Log(INFO, format, values...)
}

func (l *Logger) Warning(format string, values ...interface{}) {
	l.Log(WARNING, format, values...)
}

func (l *Logger) Error(format string, values ...interface{}) {
	l.Log(ERROR, format, values...)
}

func (l *Logger) ResetLogLevel(level string) {
	for _, e := range l.Handlers {
		if h, ok := e.(*Handler); ok {
			h.SetLevelString(level)
		}
	}
}

//打印日志用，根据回退堆栈层级获取文件名和行号信息
//参数：需要回退的堆栈层数
func GetLogBtInfo(level int) string {
	return ""
	if level < 0 { //参数错误
		return ""
	}
	format := ""
	level += 1 //函数自身占一层
	_, file, line, ok := runtime.Caller(level)
	if ok == true {
		file = filepath.Base(file)
		prefix := fmt.Sprintf("[%s:%d] ", file, line)
		format = prefix + format
	}
	return format
}

func AddHandler(name string, h Emitter) {
	DefaultLogger.AddHandler(name, h)
}

func Log(level logLevel, format string, values ...interface{}) {
	DefaultLogger.Log(level, format, values...)
}

func Warning(format string, values ...interface{}) {
	DefaultLogger.Log(WARNING, format, values...)
}

func ResetLogLevel(level string) {
	DefaultLogger.ResetLogLevel(level)
}

//带当前堆栈信息的日志接口;如果堆栈信息比较复杂，就用GetLogBtInfo+原始的日志接口
func Debug(format string, values ...interface{}) {
	format = GetLogBtInfo(1) + format //回退一层到原始栈
	DefaultLogger.Log(DEBUG, format, values...)
}

func Info(format string, values ...interface{}) {
	format = GetLogBtInfo(1) + format //回退一层到原始栈
	DefaultLogger.Log(INFO, format, values...)
}

func Error(format string, values ...interface{}) {
	format = GetLogBtInfo(1) + format //回退一层到原始栈
	DefaultLogger.Log(ERROR, format, values...)
}
