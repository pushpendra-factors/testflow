package utils

/* 
Util for Log operations with Log Rotation
*/

import(
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"fmt"
	"os"
	"sync"
)

var fLogger *log.Logger
var uLogger *log.Logger
type LogWriter struct{}

var l sync.Mutex

func (f LogWriter) RegisterOutputFile(FileName string){
	fLogger = RegisterFile(FileName)
}

func (f LogWriter) RegisterUserDataFile(FileName string){
	uLogger = RegisterFile(FileName)
}

func RegisterFile(FileName string)*log.Logger{
	File, err := os.OpenFile(FileName,  os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        fmt.Printf("error opening file: %v", err)
        os.Exit(1)
    }

	_logger := log.New(File, "", log.LstdFlags)
    _logger.SetOutput(&lumberjack.Logger{
		Filename:   FileName,
		MaxSize:    1, // megabytes
		MaxAge:     10, // days
		Compress:   true, // disabled by default
    })
	// SET MaxBackups: 2 if required
	return _logger
}

func (f LogWriter) WriteOutput(data string){
	Write(data, fLogger)
}

func (f LogWriter) WriteUserData(data string){
	Write(data, uLogger)
}

func Write(data string, _logger *log.Logger){
	m.Lock()
		_logger.Printf("|%s\n",data)
	m.Unlock()
}