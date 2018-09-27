package easynet2

import (
	"io"
	"log"
)

//NetLog omit
var NetLog *EasyLogger = newEasyLogger()

type nullWriter struct {
}

func (thls *nullWriter) Write(p []byte) (n int, err error) {
	if p != nil {
		n = len(p)
	}
	return
}

//EasyLogger omit
type EasyLogger struct {
	nullWriter
	DEBUG *log.Logger
	INFO  *log.Logger
	WARN  *log.Logger
	ERROR *log.Logger
}

func newEasyLogger() *EasyLogger {
	curData := new(EasyLogger)
	curData.DEBUG = log.New(&curData.nullWriter, "[DEBUG]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	curData.INFO = log.New(&curData.nullWriter, "[INFO]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	curData.WARN = log.New(&curData.nullWriter, "[WARN]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	curData.ERROR = log.New(&curData.nullWriter, "[ERROR]", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	return curData
}

//SetOutput omit
func (thls *EasyLogger) SetOutput(w io.Writer) {
	thls.DEBUG.SetOutput(w)
	thls.INFO.SetOutput(w)
	thls.WARN.SetOutput(w)
	thls.ERROR.SetOutput(w)
}
