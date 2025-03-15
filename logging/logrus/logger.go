package logrus

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	dateTimeLayout = "2006-01-02 15.04.05"
)

type Hook struct {
	sync.Mutex
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	hook.Lock()
	defer hook.Unlock()

	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		_, err = w.Write([]byte(line))
	}
	if err != nil {
		return err
	}
	return nil
}
func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry
var once sync.Once

type Logger struct {
	*logrus.Entry
}

func GetLogger() (*Logger, error) {
	var er error
	once.Do(func() {
		l := logrus.New()
		l.SetReportCaller(true)
		l.Formatter = &logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s:%d", filename, f.Line), (f.Function + "()")
			},
			DisableColors: false,
			FullTimestamp: true,
		}

		err := os.MkdirAll("logs", 0644)
		if err != nil || os.IsExist(err) {
			panic("can't create logs directory")
		}
		t := time.Now().UTC().Format(dateTimeLayout)
		logFile, err := os.OpenFile(fmt.Sprintf("logs/%s.log", t), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if er = err; er != nil {
			return
		}
		l.SetOutput(io.Discard)

		l.AddHook(&Hook{
			Writer:    []io.Writer{logFile, os.Stdout},
			LogLevels: logrus.AllLevels,
		})

		l.SetLevel(logrus.TraceLevel)

		e = logrus.NewEntry(l)
	})
	if er != nil {
		return nil, er
	}
	return &Logger{e}, nil
}
