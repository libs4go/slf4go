package slf4go

import "github.com/libs4go/scf4go"

type nullBackend struct {
}

func (null *nullBackend) Send(*EventEntry) {

}

func (null *nullBackend) Sync() {

}

func (null *nullBackend) Config(config scf4go.Config) error {
	return nil
}
