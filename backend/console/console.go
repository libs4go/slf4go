package console

import (
	"fmt"
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

	message := console.formatter.Output

	message = strings.ReplaceAll(message, "@line", fmt.Sprintf("%d", entry.Line))
	message = strings.ReplaceAll(message, "@func", entry.Function)
	message = strings.ReplaceAll(message, "@t", entry.Timestamp.Format(console.formatter.Timestamp))
	message = strings.ReplaceAll(message, "@l", entry.Level.String())
	message = strings.ReplaceAll(message, "@m", entry.Message)
	message = strings.ReplaceAll(message, "@s", entry.Source)
	// message = strings.ReplaceAll(message, "@f", "..."+entry.File[len(entry.File)/2:])

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

var defaultOutput = "@t |@s| |@l| @m \n from: @func:@line"

func (console *consoleImpl) Config(config scf4go.Config) error {

	console.formatter.Timestamp = config.Get("formatter", "timestamp").String(time.RFC3339)
	console.formatter.Output = config.Get("formatter", "output").String(defaultOutput)

	return nil
}

func init() {
	slf4go.RegisterBackend("console", &consoleImpl{
		formatter: &formatter{
			Timestamp: time.RFC3339,
			Output:    defaultOutput,
		},
	})
}
