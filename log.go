package main

import (
	"log"
	"os"
)

var Info *log.Logger
var Warn *log.Logger

func init() {
	Info = log.New(os.Stdout, "[INFO] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
	Warn = log.New(os.Stdout, "[WARN] ", log.LstdFlags|log.Lshortfile|log.Lmicroseconds)
}
