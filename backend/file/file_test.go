package file

import (
	"testing"

	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./slf4go.yaml")))
	require.NoError(t, err)

	err = slf4go.Config(config)

	require.NoError(t, err)

	slf4go.Get("test").D("test a {@test}", 1)
}
