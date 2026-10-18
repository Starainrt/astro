package geojson_test

import (
	"runtime"
	"strings"
	"testing"
)

func go116386() bool {
	return runtime.GOARCH == "386" && strings.HasPrefix(runtime.Version(), "go1.16")
}

func skipHeavyOnGo116386(t *testing.T) {
	t.Helper()
	if go116386() {
		t.Skip("Go 1.16/386 skips the heavy integration regressions; Go 1.17+ covers them")
	}
}
