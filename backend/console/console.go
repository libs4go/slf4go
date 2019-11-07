package console

import (
	"strings"
	"time"

	"github.com/libs4go/scf4go"
	"github.com/libs4go/slf4go"
)

type formatter struct {
	Timestamp string `json:"timestamp"`
	Output    string `json:"output"`
}

type consoleImpl struct {
	formatter *formatter
}

func (console *consoleImpl) Send(entry *slf4go.EventEntry) {
	message := strings.ReplaceAll(console.formatter.Output, "@t", entry.Timestamp.Format(console.formatter.Timestamp))

	message = strings.ReplaceAll(message, "@l", entry.Level.String())

	message = strings.ReplaceAll(message, "@m", entry.Message)

	switch entry.Level {
	case slf4go.TRACE:
		tracep(message, "\n")
	case slf4go.DEBUG:
		debugp(message, "\n")
	case slf4go.INFO:
		infop(message, "\n")
	case slf4go.WARN:
		warnp(message, "\n")
	case slf4go.ERROR:
		errorp(message, "\n")
	}
}

func (console *consoleImpl) Sync() {

}

func (console *consoleImpl) Config(config scf4go.Config) error {
	var formatter *formatter
	err := config.Get("formatter").Scan(&formatter)

	if formatter == nil || err != nil {
		return nil
	}

	console.formatter = formatter

	return nil
}

func init() {
	slf4go.RegisterBackend("console", &consoleImpl{
		formatter: &formatter{
			Timestamp: time.RFC3339,
			Output:    "@t |@l| @m",
		},
	})
}
