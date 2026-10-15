package ephemtest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// Values 按模型取期望值：旧模型用内联常量，新模型用 testdata/semianalytic/<name>.json。
// 新模型且 Update() 时记录实际值，测试结束写回文件 / per-model expected values for inline-constant tests.
type Values struct {
	tb     testing.TB
	path   string
	vals   map[string]interface{}
	semi   bool
	update bool
	dirty  bool
}

// NewValues 打开名为 name 的期望值表 / opens the expected-value table called name.
func NewValues(tb testing.TB, name string) *Values {
	tb.Helper()
	v := &Values{
		tb:     tb,
		path:   filepath.Join("testdata", "semianalytic", name+".json"),
		vals:   map[string]interface{}{},
		semi:   ephem.Active() == ephem.Semianalytic,
		update: Update(),
	}
	if !v.semi {
		return v
	}
	if data, err := os.ReadFile(v.path); err == nil {
		if err := json.Unmarshal(data, &v.vals); err != nil {
			tb.Fatal(err)
		}
	} else if !v.update {
		tb.Fatalf("%v (regenerate with ASTRO_UPDATE_GOLDEN=1 go test)", err)
	}
	if v.update {
		tb.Cleanup(v.flush)
	}
	return v
}

// Want 返回 key 的期望值；legacy 为 0 表示该项不检查，原样返回 0 / expected value for key.
func (v *Values) Want(key string, legacy, got float64) float64 {
	v.tb.Helper()
	if !v.semi || legacy == 0 {
		return legacy
	}
	if v.update {
		v.vals[key], v.dirty = got, true
		return got
	}
	want, ok := v.vals[key].(float64)
	if !ok {
		v.tb.Fatalf("%s: no numeric expected value for %q", v.path, key)
	}
	return want
}

// WantString 同 Want，用于文本期望值；legacy 为空表示不检查 / string variant of Want.
func (v *Values) WantString(key, legacy, got string) string {
	v.tb.Helper()
	if !v.semi || legacy == "" {
		return legacy
	}
	if v.update {
		v.vals[key], v.dirty = got, true
		return got
	}
	want, ok := v.vals[key].(string)
	if !ok {
		v.tb.Fatalf("%s: no string expected value for %q", v.path, key)
	}
	return want
}

func (v *Values) flush() {
	if !v.dirty {
		return
	}
	data, err := json.MarshalIndent(v.vals, "", "  ")
	if err != nil {
		v.tb.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(v.path), 0o755); err != nil {
		v.tb.Fatal(err)
	}
	if err := os.WriteFile(v.path, append(data, '\n'), 0o644); err != nil {
		v.tb.Fatal(err)
	}
	v.tb.Logf("rewrote %s: %d values", v.path, len(v.vals))
}
