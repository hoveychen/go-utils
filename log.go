package goutils

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
)

var (
	debug = flag.Bool("debug", false, "True to turn into debug mode.")

	infoLog, debugLog, errLog, fatalLog *log.Logger
)

func init() {
	infoLog = log.New(os.Stdout, "[INFO]", log.LstdFlags)
	debugLog = log.New(os.Stdout, "[DEBUG]", log.LstdFlags|log.Lshortfile)
	errLog = log.New(os.Stderr, "[ERROR]", log.LstdFlags|log.Llongfile)
	fatalLog = log.New(os.Stderr, "[FATAL]", log.LstdFlags|log.Llongfile)
}

// Check provide a quick way to check unexpected errors that should never happen.
// It's basically an assertion that once err != nil, fatal panic is thrown.
func Check(err error) {
	if err != nil {
		LogError(err)
		panic(err)
	}
}

// Check provide a quick way to check unexpected errors that should never happen.
// It's almost the same as Check(), except only in debug mode will throw panic.
func DCheck(err error) {
	if err != nil {
		LogError(err)
		if *debug {
			panic(err)
		}
	}
}

// LogError prints error to error output with [ERROR] prefix.
func LogError(v ...interface{}) {
	errLog.Output(2, fmt.Sprintln(v...))
}

// LogInfo prints info to standard output with [INFO] prefix.
func LogInfo(v ...interface{}) {
	infoLog.Output(2, fmt.Sprintln(v...))
}

// LogDebug prints info to standard output with [DEBUG] prefix in debug mode.
func LogDebug(v ...interface{}) {
	if *debug {
		debugLog.Output(2, fmt.Sprintln(v...))
	}
}

// LogFatal prints error to error output with [FATAL] prefix, and terminate the
// application.
func LogFatal(v ...interface{}) {
	fatalLog.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// PrintJson outputs any varible in Json format to console. Useful for debuging.
func PrintJson(v interface{}) {
	fmt.Println(Jsonify(v))
}

// Jsonify provides shortcut to return an json format string of any varible.
func Jsonify(v interface{}) string {
	d, err := json.MarshalIndent(v, "", "  ")
	DCheck(err)
	return string(d)
}

// GetFuncName provides shortcut to print the name of any function.
func GetFuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// NewError returns an error composed like fmt.Sprintf().
func NewError(v ...interface{}) error {
	return errors.New(fmt.Sprintln(v...))
}
