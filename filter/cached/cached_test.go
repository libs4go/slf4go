package cached

import (
	"testing"

	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec"
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"

	_ "github.com/libs4go/slf4go/backend/console"
)

func init() {
	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./slf4go.yaml")))

	if err != nil {
		panic(err)
	}

	err = slf4go.Config(config)

	if err != nil {
		panic(err)
	}
}

func TestLog(t *testing.T) {

	slf4go.Get("test").D("test a")

	slf4go.Sync()
}

func BenchmarkWrite(t *testing.B) {

	t.StopTimer()
	logger := slf4go.Get("test")
	t.StartTimer()

	for i := 0; i < t.N; i++ {
		logger.D("test")
	}

}
