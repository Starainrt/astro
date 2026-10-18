package basic

import (
	"runtime"
	"strings"
)

func go116386() bool {
	return runtime.GOARCH == "386" && strings.HasPrefix(runtime.Version(), "go1.16")
}
