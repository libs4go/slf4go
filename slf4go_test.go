package slf4go

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/stretchr/testify/require"
)

type mockBackend struct {
	events []*EventEntry
	config scf4go.Config
}

func (mock *mockBackend) Send(event *EventEntry) {
	mock.events = append(mock.events, event)
}

func (mock *mockBackend) Sync() {
}

func (mock *mockBackend) Config(config scf4go.Config) error {
	mock.config = config
	return nil
}

func TestRegex(t *testing.T) {
	regex := regexp.MustCompile(`{(@[a-zA-Z0-9]*)}`)

	fmt.Println(regex.FindAllStringSubmatch("{@test} {@test2} {@3}", -1))
}

func TestLog(t *testing.T) {
	mock := &mockBackend{}
	getLoggerFactor().config("mock", DEBUG)
	RegisterBackend("mock", mock)

	logger := Get("test")

	logger.D("test one {@one}", 1)

	require.Equal(t, 1, len(mock.events))

	getLoggerFactor().config("mock", INFO)

	logger.D("test one {@one}", 1)

	require.Equal(t, 1, len(mock.events))
	require.Equal(t, "test one 1", mock.events[0].Message)
	require.Equal(t, len(mock.events[0].Attrs), 1)

	logger.I("test {@one} {@two}", 1, 2)

	require.Equal(t, 2, len(mock.events))
	require.Equal(t, "test 1 2", mock.events[1].Message)
	require.Equal(t, len(mock.events[1].Attrs), 2)

	logger.E("error: {@error}", ErrArgs)

	require.Equal(t, 3, len(mock.events))

	buff, err := json.Marshal(mock.events[2])

	require.NoError(t, err)

	println(string(buff))
}

func TestConfig(t *testing.T) {
	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./slf4go.yaml")))
	require.NoError(t, err)
	mock := &mockBackend{}
	RegisterBackend("mock", mock)
	err = Config(config)
	require.NoError(t, err)

	require.Equal(t, getLoggerFactor().defaultConfig.Level, INFO)
	require.Equal(t, getLoggerFactor().defaultConfig.Backend, "mock")

	require.Equal(t, mock.config.Get("test").String(""), "salt")

	require.Equal(t, getLoggerFactor().configs["test"].Backend, "test")
	require.Equal(t, getLoggerFactor().configs["test"].Level, ERROR)
}

func TestFilepath(t *testing.T) {
	filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		println(path, info.Name())
		return nil
	})
}
