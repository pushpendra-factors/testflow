package logger

/*
Log registration for debug and error logs
*/

import(
	"os"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
    "log"
    "data_simulator/constants"
)

var Debug *log.Logger
var Error *log.Logger

func RegisterLogFiles(filePattern string, jobname string){
    debugLogFile := fmt.Sprintf("%s/%s_%s_%s.log",
        constants.LOCALOUTPUTFOLDER, jobname, "debugLogs",filePattern)
    errorLogFile := fmt.Sprintf("%s/%s_%s_%s.log",
        constants.LOCALOUTPUTFOLDER, jobname, "errorLogs",filePattern)
    fmt.Printf("Check For all debug logs in: %v\n", debugLogFile)
	debuglog, err := os.OpenFile(debugLogFile,  os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Printf("error opening file: %v\n", err)
        os.Exit(1)
    }
	Debug = log.New(debuglog, "", log.LstdFlags)
    Debug.SetOutput(&lumberjack.Logger{
		Filename:   debugLogFile,
		MaxSize:    1, // megabytes
		MaxAge:     10, // days
		Compress:   true, // disabled by default
    })
    
    fmt.Printf("Check For all error logs in: %v\n", errorLogFile)
	errorlog, err := os.OpenFile(errorLogFile,  os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Printf("error opening file: %v", err)
        os.Exit(1)
    }

    Error = log.New(errorlog, "", log.LstdFlags)
    Error.SetOutput(&lumberjack.Logger{
		Filename:   errorLogFile,
		MaxSize:    1, // megabytes
		MaxAge:     10, // days
		Compress:   true, // disabled by default
    })   
}