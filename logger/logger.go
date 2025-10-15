package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

/**
logger has levels: Panic, Fatal, Error, Warn, Info, and Debug
Panic will log a message and then panic.  This is not to be used in production.
Fatal will log a message and Goexit(), but osExit() if the caller is context.Background()
Fatal should also be used in rare cases.
Error will log a message an return, but it is expected that the transaction will fail.
Warn, Info, and Debug log messages and continue.
*/

type LogLevel int

const (
	LOGWARN LogLevel = iota + 3
	LOGINFO
	LOGDEBUG
)

var logLevel LogLevel = LOGDEBUG
var dumpSkipLines = 3 // number of dump lines to skip

var panicLog *log.Logger
var fatalLog *log.Logger
var errorLog *log.Logger
var warnLog *log.Logger
var infoLog *log.Logger
var debugLog *log.Logger

func init() {
	setFile(os.Stderr)
	logLevel := os.Getenv("FCBH_DATASET_LOG_LEVEL")
	if logLevel != `` {
		SetLevel(logLevel)
	} else {
		SetLevel(`INFO`)
	}
	logFile := os.Getenv("FCBH_DATASET_LOG_FILE")
	if logFile != `` {
		SetOutput(logFile)
	} else {
		SetOutput(`stderr`)
	}
	_ = os.Setenv("TZ", "America/New_York")
	//_ = os.Setenv("TZ", "America/Denver")
}

// SetOutput accepts: stdout, stderr, or a filePath
func SetOutput(filePath string) {
	if filePath == "stdout" {
		setFile(os.Stdout)
	} else if filePath == "stderr" {
		setFile(os.Stderr)
	} else {
		file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Warn(context.TODO(), 0, "log.SetOutput failed", err)
		} else {
			setFile(file)
		}
	}
}

func setFile(file *os.File) {
	panicLog = log.New(file, "PANIC ", log.Ldate|log.Ltime)
	fatalLog = log.New(file, "FATAL ", log.Ldate|log.Ltime)
	errorLog = log.New(file, "ERROR ", log.Ldate|log.Ltime)
	warnLog = log.New(file, "WARN ", log.Ldate|log.Ltime)
	infoLog = log.New(file, "INFO ", log.Ldate|log.Ltime)
	debugLog = log.New(file, "DEBUG ", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// SetLevel set an error reporting level. The logger will process messages of this level and higher.
func SetLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		logLevel = LOGDEBUG
	case "info":
		logLevel = LOGINFO
	case "warn":
		logLevel = LOGWARN
	}
}

func SetDumpSkipLines(lines int) {
	dumpSkipLines = lines
}

// Panic will log the message with Println and then call panic()
func Panic(ctx context.Context, param ...any) {
	panicLog.Panicln(param, requestInfo(ctx))
}

// Fatal will log the message with Println and call os.Exit is in background
func Fatal(ctx context.Context, param ...any) {
	fatalLog.Println(param, requestInfo(ctx))
	fatalLog.Println(dumpLines())
	runType := ctx.Value(`runType`)
	if runType != `cli` {
		runtime.Goexit()
	} else {
		os.Exit(1)
	}
}

func Error(ctx context.Context, http int, err error, param ...any) *Status {
	status, ok := err.(*Status)
	if ok {
		return status
	} else {
		return errorImpl(ctx, http, err.Error(), param...)
	}
}

func ErrorNoErr(ctx context.Context, http int, param ...any) *Status {
	return errorImpl(ctx, http, ``, param...)
}

func ExecError(ctx context.Context, http int, stderr string, param ...any) *Status {
	stderr = strings.TrimSpace(stderr)
	if strings.HasPrefix(stderr, `{"status"`) {
		var s Status
		jsonErr := json.Unmarshal([]byte(stderr), &s)
		if jsonErr != nil {
			Warn(ctx, jsonErr.Error())
			Warn(ctx, param)
			return nil
		}
		s.Status = http
		s.Request = requestInfo(ctx)
		errorLog.Printf("%+v", s)
		return &s
	} else {
		param = append([]any{stderr}, param...)
		Warn(ctx, param...)
		return nil
	}
}

func errorImpl(ctx context.Context, http int, err string, param ...any) *Status {
	var result []byte
	for _, p := range param {
		result = fmt.Append(result, p)
		result = fmt.Append(result, ` `)
	}
	var e Status
	e.Status = http
	e.Err = err
	e.Message = strings.TrimSpace(string(result))
	e.Trace = dumpLines()
	e.Request = requestInfo(ctx)
	errorLog.Printf("%+v", e)
	return &e
}

// Warn will log the message with Println and then continue
func Warn(ctx context.Context, param ...any) {
	if logLevel >= LOGWARN {
		warnLog.Println(param...)
		//warnLog.Println(dumpLines())
	}
}

// Info will log the message with Println and then continue
func Info(ctx context.Context, param ...any) {
	if logLevel >= LOGINFO {
		infoLog.Println(param...)
	}
}

// Debug will log the message with Println and then continue
func Debug(ctx context.Context, param ...any) {
	if logLevel >= LOGDEBUG {
		// For info on each, see: https://golang.org/pkg/runtime/#MemStats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		var msg []string
		msg = append(msg, fmt.Sprintf("Alloc = %v mb,", bToMb(m.Alloc)))
		msg = append(msg, fmt.Sprintf("Malloca = %v mb,", bToMb(m.Mallocs)))
		msg = append(msg, fmt.Sprintf("Frees = %v mb,", bToMb(m.Frees)))
		msg = append(msg, fmt.Sprintf("NumGC = %v", m.NumGC))
		param = append(param, strings.Join(msg, " "))
		debugLog.Println(param)
	}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func requestInfo(ctx context.Context) string {
	if ctx != nil {
		request := ctx.Value("request")
		if request != nil {
			return request.(string)
		}
	}
	return ""
}

func dumpLines() string {
	var results []string
	var file string
	var line int
	var ok = true
	for i := 2; ok; i++ {
		_, file, line, ok = runtime.Caller(i)
		pos := strings.Index(file, `fcbh-dataset-io`)
		if pos >= 0 {
			results = append(results, file[pos:]+":"+strconv.Itoa(line))
		}
	}
	return strings.Join(results, "\n")
}
