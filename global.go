package slf4go

import (
	"sync"

	"github.com/libs4go/scf4go"
)

var once sync.Once
var factory *loggerFactory

func getLoggerFactor() *loggerFactory {
	once.Do(func() {
		factory = newLoggerFactory()
	})

	return factory
}

// RegisterBackend .
func RegisterBackend(name string, backend Backend) {
	getLoggerFactor().registerBackend(name, backend)
}

// RegisterFilter create filter chain with call order
func RegisterFilter(filter Filter) {
	getLoggerFactor().registerFilter(filter)
}

// Sync sync flush all logger event
func Sync() {
	getLoggerFactor().sync()
}

// Get create or get logger with name
func Get(name string) Logger {
	return getLoggerFactor().createLogger(name)
}

// Config config loggers with scf4go
func Config(config scf4go.Config) error {
	return getLoggerFactor().setConfig(config)
}
