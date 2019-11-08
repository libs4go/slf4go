package slf4go

import (
	"runtime"
	"strings"
)

func getCallFrame() runtime.Frame {
	pcs := make([]uintptr, 2)

	count := runtime.Callers(4, pcs)

	frames := runtime.CallersFrames(pcs[:count])

	frame, _ := frames.Next()

	if index := strings.Index(frame.File, "src"); index != -1 {
		// trim GOPATH or GOROOT prifix
		frame.File = string(frame.File[index+4:])
	}

	return frame
}
