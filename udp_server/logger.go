package udp_server

import (
	"fmt"
)

type Logger struct {
	level int
}

const (
	LOG_LEVEL_DEBUG   int = 0
	LOG_LEVEL_INFO    int = 1
	LOG_LEVEL_WARNING int = 2
	LOG_LEVEL_ERROR   int = 3
)

var logArray = []string{"DEBUG", "INFO", "WARN", "ERROR"}

var logger Logger // Package-level variable to hold the logger instance

func init() {
	logger = Logger{
		level: LOG_LEVEL_INFO, // Initialize the logger with the default level
	}
}

func (log *Logger) setLogLevel(level int) {
	log.level = level
}

func (log *Logger) log(level int, msg string, args ...any) {
	if level >= log.level {
		formattedMsg := fmt.Sprintf(msg, args...)
		fmt.Printf("[%s] %s\n", logArray[level], formattedMsg)
	}
}

func (log *Logger) warn(msg string, args ...any) {
	log.log(LOG_LEVEL_WARNING, msg, args...)
}

func (log *Logger) info(msg string, args ...any) {
	log.log(LOG_LEVEL_INFO, msg, args...)
}

func (log *Logger) debug(msg string, args ...any) {
	log.log(LOG_LEVEL_INFO, msg, args...)
}
