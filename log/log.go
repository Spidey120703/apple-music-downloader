package log

import (
	"log"
	"os"
)

var Info *log.Logger
var Warn *log.Logger
var Error *log.Logger

func init() {
	Info = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	Warn = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	Error = log.New(os.Stdout, "[ERROR] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
}
