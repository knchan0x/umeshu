package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type LogLevel int

// Log level available
const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	PanicLevel
	FatalLevel
	RecoveryLevel
	Disable
)

var (
	Debug    func(format string, v ...interface{})
	Info     func(format string, v ...interface{})
	Warning  func(format string, v ...interface{})
	Error    func(format string, v ...interface{})
	Panic    func(format string, v ...interface{}) // Panic will stop and run defer func() like recovery before exit app
	Fatal    func(format string, v ...interface{}) // Fatal will stop and exit app immedicately, no automatic recovery
	Recovery func(format string, v ...interface{}) // logging traceback for recovery
)

var (
	defaultPrefix = []string{
		"[Debug]    ",
		"[Info]     ",
		"\033[1;33m[Warning]\033[0m  ",
		"\033[1;31m[Error]\033[0m    ",
		"\033[1;31m[Panic]\033[0m    ",
		"\033[1;31m[Fatal]\033[0m    ",
		"\033[1;31m[Recovery]\033[0m ",
	}
	defaultLevel = DebugLevel
	defaultFlag  = log.LstdFlags | log.Lshortfile
)

var (
	level   LogLevel
	loggers []*log.Logger
	backups []io.Writer
	mu      sync.Mutex
)

func init() {
	loggers = make([]*log.Logger, int(Disable))
	backups = make([]io.Writer, int(Disable))

	for i := DebugLevel; i < Disable; i++ {
		loggers[i] = log.New(os.Stdout, defaultPrefix[i], defaultFlag)
		backups[i] = os.Stdout
	}

	level = defaultLevel

	Debug = loggers[DebugLevel].Printf
	Info = loggers[InfoLevel].Printf
	Warning = loggers[WarningLevel].Printf
	Error = loggers[ErrorLevel].Printf
	Panic = panicf
	Fatal = loggers[FatalLevel].Fatalf
	Recovery = loggers[RecoveryLevel].Printf
}

// Level returns current global log level.
func Level() LogLevel {
	return level
}

// SetLevel controls the global log level.
func SetLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()

	current := int(level)
	for l := range loggers {
		if l < current {
			loggers[l].SetOutput(io.Discard)
		} else {
			loggers[l].SetOutput(backups[l])
		}
	}
}

// SetWriterForAll sets the logger for all level of loggers.
func SetWriterForAll(writer io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	for l := range loggers {
		loggers[l].SetOutput(writer)
		backups[l] = writer
	}
}

// SetWriter sets the logger for specific level.
func SetWriter(writer io.Writer, level LogLevel) {
	loggers[level].SetOutput(writer)
	backups[level] = writer
}

// SetFlagsForAll sets the flag for all level of loggers.
func SetFlagsForAll(flag int) {
	mu.Lock()
	defer mu.Unlock()

	for l := range loggers {
		loggers[l].SetFlags(flag)
	}
}

// SetFlags sets the flag for all level of loggers.
func SetFlags(flag int, level LogLevel) {
	loggers[level].SetFlags(flag)
}

// SetPrefix sets the prefix for all levels.
func SetPrefixForAll(prefix string) {
	mu.Lock()
	defer mu.Unlock()

	if prefix[len(prefix)-1] != ' ' {
		prefix = prefix + " "
	}

	for l := range loggers {
		loggers[l].SetPrefix(prefix)
	}
}

// SetPrefix sets the prefix for specific level.
func SetPrefix(prefix string, level LogLevel) {
	if prefix[len(prefix)-1] != ' ' {
		prefix = prefix + " "
	}
	loggers[level].SetPrefix(prefix)
}

// SetToDefault sets all level of loggers to default.
func SetToDefault() {
	mu.Lock()
	defer mu.Unlock()

	for l := range loggers {
		loggers[l].SetOutput(os.Stdout)
		loggers[l].SetPrefix(defaultPrefix[l])
		loggers[l].SetFlags(defaultFlag)
		backups[l] = os.Stdout
	}
	level = defaultLevel
}

// panicf is the wrapper of panic().
// It will output error message before panic.
func panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	loggers[PanicLevel].Output(2, s)
	panic(s)
}
