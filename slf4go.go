package slf4go

import (
	"encoding/json"
	"fmt"

	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/libs4go/errors"
	"github.com/libs4go/scf4go"
)

const errVendor = "slf4go"

// Errors
var (
	ErrArgs  = errors.New("args number errors", errors.WithVendor(errVendor), errors.WithCode(-1))
	ErrLevel = errors.New("invalid error level", errors.WithVendor(errVendor), errors.WithCode(-2))
)

// Logger .
type Logger interface {
	Name() string
	T(message string, args ...interface{})
	D(message string, args ...interface{})
	I(message string, args ...interface{})
	W(message string, args ...interface{})
	E(message string, args ...interface{})
}

// Backend .
type Backend interface {
	Config(config scf4go.Config) error
	Send(entry *EventEntry)
	Sync()
}

// Filter .
type Filter interface {
	Name() string
	Config(config scf4go.Config)
	MakeChain(backend Backend) Backend
}

// Level logger level
type Level int

func (l Level) String() string {
	switch l {
	case TRACE:
		return "trace"
	case DEBUG:
		return "debug"
	case INFO:
		return "info"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	}

	panic(errors.Wrap(ErrLevel, "unknown level %d", l))
}

// MarshalJSON .
func (l Level) MarshalJSON() ([]byte, error) {
	switch l {
	case TRACE:
		return json.Marshal("trace")
	case DEBUG:
		return json.Marshal("debug")
	case INFO:
		return json.Marshal("info")
	case WARN:
		return json.Marshal("warn")
	case ERROR:
		return json.Marshal("error")
	}

	return nil, errors.Wrap(ErrLevel, "unknown level %d", l)
}

// UnmarshalJSON .
func (l *Level) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	case "trace":
		*l = TRACE
		return nil
	case "debug":
		*l = DEBUG
		return nil
	case "info":
		*l = INFO
		return nil
	case "warn":
		*l = WARN
		return nil
	case "error":
		*l = ERROR
		return nil
	}

	return errors.Wrap(ErrLevel, "unknown level %s", s)
}

// logger levels .
const (
	TRACE = Level(iota)
	DEBUG
	INFO
	WARN
	ERROR
)

// EventEntry .
type EventEntry struct {
	Timestamp time.Time              `json:"@t"`
	Level     Level                  `json:"@l"`
	Message   string                 `json:"@m"`
	Attrs     map[string]interface{} `json:"@a"`
	Source    string                 `json:"@s"`
	File      string                 `json:"@f"`
	Line      int                    `json:"@line"`
	Function  string                 `json:"@func"`
}

type loggerFactory struct {
	sync.RWMutex
	filter        []Filter
	backend       map[string]Backend
	configs       map[string]*loggerConfig
	loggers       map[string]*loggerFacade
	defaultConfig *loggerConfig
}

type loggerConfig struct {
	Backend string `json:"backend"`
	Level   Level  `json:"level"`
}

func newLoggerFactory() *loggerFactory {
	factory := &loggerFactory{
		backend: make(map[string]Backend),
		configs: make(map[string]*loggerConfig),
		loggers: make(map[string]*loggerFacade),
	}

	factory.backend["null"] = &nullBackend{}

	factory.defaultConfig = &loggerConfig{
		Level:   DEBUG,
		Backend: "null",
	}

	return factory
}

func (factory *loggerFactory) registerBackend(name string, backend Backend) {
	factory.Lock()
	defer factory.Unlock()

	for _, filter := range factory.filter {
		backend = filter.MakeChain(backend)
	}

	factory.backend[name] = backend
}

func (factory *loggerFactory) registerFilter(filter Filter) {
	factory.Lock()
	defer factory.Unlock()

	factory.filter = append(factory.filter, filter)

	cached := make(map[string]Backend)

	for name, backend := range factory.backend {
		println(fmt.Sprintf("filter %s backend %s", filter.Name(), name))
		cached[name] = filter.MakeChain(backend)
	}

	factory.backend = cached
}

func (factory *loggerFactory) configLogger(logger string, backend string, level Level) {
	factory.Lock()
	defer factory.Unlock()

	factory.configs[logger] = &loggerConfig{
		Backend: backend,
		Level:   level,
	}
}

func (factory *loggerFactory) config(backend string, level Level) {
	factory.Lock()
	defer factory.Unlock()

	factory.defaultConfig = &loggerConfig{
		Backend: backend,
		Level:   level,
	}
}

func (factory *loggerFactory) sync() {
	factory.Lock()
	defer factory.Unlock()

	for _, backend := range factory.backend {
		backend.Sync()
	}
}

func (factory *loggerFactory) setConfig(config scf4go.Config) error {
	factory.Lock()
	defer factory.Unlock()

	for name, backend := range factory.backend {
		backend.Config(config.SubConfig("backend", name))
	}

	for _, filter := range factory.filter {
		filter.Config(config.SubConfig("filter", filter.Name()))
	}

	var defaultConfig loggerConfig

	if err := config.Get("default").Scan(&defaultConfig); err != nil {
		return errors.Wrap(err, "parse default logger error")
	}

	factory.defaultConfig = &defaultConfig

	var configs map[string]*loggerConfig

	if err := config.Get("logger").Scan(&configs); err != nil {
		return errors.Wrap(err, "parse logger error")
	}

	factory.configs = configs

	return nil
}

func (factory *loggerFactory) createLogger(name string) Logger {

	factory.Lock()
	defer factory.Unlock()

	logger, ok := factory.loggers[name]

	if ok {
		return logger
	}

	logger = newLoggerFacade(name, factory)
	factory.loggers[name] = logger

	return logger
}

func (factory *loggerFactory) getBackend(name string) (Backend, Level) {
	factory.RLock()
	defer factory.RUnlock()

	config, ok := factory.configs[name]

	if !ok {
		config = factory.defaultConfig
	}

	backend, ok := factory.backend[config.Backend]

	if !ok {
		println(fmt.Sprintf("logger '%s' backend '%s' not found", name, config.Backend))
	}

	return backend, config.Level
}

type loggerFacade struct {
	factory *loggerFactory
	name    string
}

func newLoggerFacade(name string, factory *loggerFactory) *loggerFacade {
	return &loggerFacade{
		factory: factory,
		name:    name,
	}
}

func (facade *loggerFacade) Name() string {
	return facade.name
}

func (facade *loggerFacade) process(wl Level) (Backend, bool) {

	backend, level := facade.factory.getBackend(facade.name)

	if backend == nil {
		return nil, false
	}

	if level > wl {

		return nil, false
	}

	return backend, true
}

var messageRegx = regexp.MustCompile(`{@[a-zA-Z0-9]*}`)

func (facade *loggerFacade) createEventEntry(message string, level Level, args ...interface{}) *EventEntry {

	placeholders := messageRegx.FindAllString(message, -1)

	if len(placeholders) != len(args) {
		panic(errors.Wrap(ErrArgs, "expect args(%d) got(%d)", len(placeholders), len(args)))
	}

	attrs := make(map[string]interface{})

	for i, placeholder := range placeholders {

		data, err := json.Marshal(args[i])

		if err != nil {
			panic(errors.Wrap(err, "marshal arg %d error", i))
		}

		message = strings.Replace(message, placeholder, string(data), -1)

		placeholder = strings.TrimSuffix(strings.TrimPrefix(placeholder, "{"), "}")

		attrs[placeholder] = args[i]
	}

	callframe := getCallFrame()

	entry := &EventEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Attrs:     attrs,
		Source:    facade.name,
		File:      callframe.File,
		Line:      callframe.Line,
		Function:  callframe.Function,
	}

	return entry
}

func (facade *loggerFacade) T(message string, args ...interface{}) {
	if backend, ok := facade.process(TRACE); ok {
		backend.Send(facade.createEventEntry(message, TRACE, args...))
	}
}

func (facade *loggerFacade) D(message string, args ...interface{}) {
	if backend, ok := facade.process(DEBUG); ok {
		backend.Send(facade.createEventEntry(message, DEBUG, args...))
	}
}

func (facade *loggerFacade) I(message string, args ...interface{}) {
	if backend, ok := facade.process(INFO); ok {
		backend.Send(facade.createEventEntry(message, INFO, args...))
	}
}

func (facade *loggerFacade) W(message string, args ...interface{}) {
	if backend, ok := facade.process(WARN); ok {
		backend.Send(facade.createEventEntry(message, WARN, args...))
	}
}

func (facade *loggerFacade) E(message string, args ...interface{}) {
	if backend, ok := facade.process(ERROR); ok {
		backend.Send(facade.createEventEntry(message, ERROR, args...))
	}
}
