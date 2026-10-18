package astro_test

import (
	"runtime"
	"strings"
)

// Go 1.16 的旧逃逸分析/386 浮点代码生成与新版本不同，只放行已知的 1 次分配与 1 ULP 差异。
func go116() bool {
	return strings.HasPrefix(runtime.Version(), "go1.16")
}

func go116386() bool {
	return runtime.GOARCH == "386" && go116()
}
