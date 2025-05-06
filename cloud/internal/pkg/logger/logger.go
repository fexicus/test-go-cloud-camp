package logger

import (
	"context"
	"fmt"
	"time"
)

type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelCheck Level = "CHECK"
	LevelFatal Level = "FATAL"
)

type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Check(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Stop()
}

type ThreadLogger struct {
	out    chan string
	stopCh chan struct{}
}

func NewThreadLogger(ctx context.Context) *ThreadLogger {
	l := &ThreadLogger{
		out:    make(chan string, 100),
		stopCh: make(chan struct{}),
	}
	go func() {
		for {
			select {
			case msg := <-l.out:
				fmt.Println(msg)
			case <-l.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return l
}

func (l *ThreadLogger) logWithLevel(level Level, format string, a ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s: %s", level, timestamp, fmt.Sprintf(format, a...))
	select {
	case l.out <- line:
	default:
		// если канал полон, отпечатаем напрямую
		fmt.Println(line)
	}
}

func (l *ThreadLogger) Info(format string, a ...interface{}) { l.logWithLevel(LevelInfo, format, a...) }
func (l *ThreadLogger) Warn(format string, a ...interface{}) { l.logWithLevel(LevelWarn, format, a...) }
func (l *ThreadLogger) Error(format string, a ...interface{}) {
	l.logWithLevel(LevelError, format, a...)
}
func (l *ThreadLogger) Check(format string, a ...interface{}) {
	l.logWithLevel(LevelCheck, format, a...)
}
func (l *ThreadLogger) Fatal(format string, a ...interface{}) {
	l.logWithLevel(LevelFatal, format, a...)
	l.Stop()
}

func (l *ThreadLogger) Stop() {
	close(l.stopCh)
}
