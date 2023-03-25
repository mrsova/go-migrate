package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"runtime"
	"strconv"
)

var Logger log.Logger

func New() {
	Logger = log.Logger{}

	Logger.SetReportCaller(true)
	Logger.SetFormatter(&log.JSONFormatter{
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			fileName := path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
			return "", fileName
		},
	})
	Logger.SetOutput(os.Stdout)
	Logger.SetLevel(log.InfoLevel)
}
