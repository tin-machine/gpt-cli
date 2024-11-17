package main

import (
	"log"
)

// Logger はロギングのためのインターフェースです
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// ConsoleLogger はコンソールに出力するロガーです
type ConsoleLogger struct {
	debug bool
}

func NewConsoleLogger(debug bool) *ConsoleLogger {
	return &ConsoleLogger{debug: debug}
}

func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	log.Printf("INFO: "+msg, args...)
}

func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	log.Printf("ERROR: "+msg, args...)
}

func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if l.debug {
		if l.debug {
			log.Printf("DEBUG: "+msg, args...) // デバッグモードが有効な場合は情報を出力
		} else {
			// デバッグモードが無効な場合は何も出力しない
			return
		}
	}
}
