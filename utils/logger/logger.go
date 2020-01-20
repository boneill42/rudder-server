package logger

import (
	"bytes"
	"github.com/rudderlabs/rudder-server/config"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

/*
Using levels(like Debug, Info etc.) in logging is a way to categorize logs based on their importance.
The idea is to have the option of running the application in different logging levels based on
how verbose we want the logging to be.
For example, using Debug level of logging, logs everything and it might slow the application, so we run application
in DEBUG level for local development or when we want to look through the entire flow of events in detail.
We use 4 logging levels here Debug, Info, Error and Fatal.
*/

const (
	levelDebug = iota + 1 // Most verbose logging level
	levelInfo             // Logs about state of the application
	levelError            // Logs about errors which dont immediately halt the application
	levelFatal            // Logs which crashes the application
)

var levelMap = map[string]int{
	"DEBUG": levelDebug,
	"INFO":  levelInfo,
	"ERROR": levelError,
	"FATAL": levelFatal,
}

var (
	enableConsole       bool
	enableFile          bool
	consoleJsonFormat   bool
	fileJsonFormat      bool
	level               int
	enableTimestamp     bool
	enableFileNameInLog bool
	enableStackTrace    bool
	logFileLocation     string
	logFileSize         int
)

var Log *zap.SugaredLogger

func loadConfig() {
	level = levelMap[config.GetEnv("LOG_LEVEL", "INFO")]
	enableConsole = config.GetBool("Zapper.enableConsole", true)
	enableFile = config.GetBool("Zapper.enableFile", false)
	consoleJsonFormat = config.GetBool("Zapper.consoleJsonFormat", true)
	fileJsonFormat = config.GetBool("Zapper.fileJsonFormat", false)
	logFileLocation = config.GetString("Zapper.logFileLocation", "/tmp/rudder_log.txt")
	logFileSize = config.GetInt("Zapper.logFileSize", 100)
	enableTimestamp = config.GetBool("Zapper.enableTimestamp", false)
	enableFileNameInLog = config.GetBool("Zapper.enableFileNameInLog", false)
	enableStackTrace = config.GetBool("Zapper.enableStackTrace", false)
}

var options []zap.Option

// Setup sets up the logger initially
func Setup() {
	loadConfig()
	Log = configureLogger()
}

func IsDebugLevel() bool {
	return levelDebug >= level
}

// Debug level logging.
// Most verbose logging level.
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Info level logging.
// Use this to log the state of the application. Dont use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Error level logging.
// Use this to log errors which dont immediately halt the application.
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Fatal level logging.
// Use this to log errors which crash the application.
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

// Debugf does debug level logging similar to fmt.Printf.
// Most verbose logging level
func Debugf(format string, args ...interface{}) {
	Log.Debug(args...)
}

// Infof does info level logging similar to fmt.Printf.
// Use this to log the state of the application. Dont use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Errorf does error level logging similar to fmt.Printf.
// Use this to log errors which dont immediately halt the application.
func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

// Fatalf does fatal level logging similar to fmt.Printf.
// Use this to log errors which crash the application.
func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

// LogRequest reads and logs the request body and resets the body to original state.
func LogRequest(req *http.Request) {
	if levelDebug >= level {
		defer req.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(req.Body)
		bodyString := string(bodyBytes)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		//print raw request body for debugging purposes
		Log.Debug("Request Body: ", bodyString)
	}
}
