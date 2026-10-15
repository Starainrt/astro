// 事件回归基线按星历模型分文件：旧模型用 testdata/<name>.json，新模型用 testdata/semianalytic/<name>.json。
// ASTRO_UPDATE_GOLDEN=1 且当前为新模型时，以旧基线的输入时刻重算并重写新模型基线；旧模型基线从不改写。
package basic

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
	"github.com/starainrt/astro/internal/ephemtest"
)

type eventBaselineSample struct {
	InputUTC string            `json:"input_utc"`
	TTJDBits uint64            `json:"tt_jd_bits"`
	Events   map[string]uint64 `json:"events"`
}

type eventBaselineFile struct {
	Samples []eventBaselineSample `json:"samples"`
}

type namedEventFunc struct {
	name string
	fn   func(float64) float64
}

func readEventBaseline(t *testing.T, path string) eventBaselineFile {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var b eventBaselineFile
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}
	if len(b.Samples) == 0 {
		t.Fatalf("empty event baseline %s", path)
	}
	return b
}

// eventBaselineData 当前模型的基线原始字节；需要时先重写新模型基线。
func eventBaselineData(t *testing.T, legacyPath string, cases []namedEventFunc) []byte {
	t.Helper()
	path := ephemtest.Golden(legacyPath)
	if ephemtest.Update() && ephem.Active() == ephem.Semianalytic {
		src := readEventBaseline(t, legacyPath)
		out := eventBaselineFile{Samples: make([]eventBaselineSample, len(src.Samples))}
		for i, s := range src.Samples {
			jd := math.Float64frombits(s.TTJDBits)
			ev := make(map[string]uint64, len(cases))
			for _, c := range cases {
				ev[c.name] = math.Float64bits(c.fn(jd))
			}
			out.Samples[i] = eventBaselineSample{InputUTC: s.InputUTC, TTJDBits: s.TTJDBits, Events: ev}
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		names := make([]string, 0, len(cases))
		for _, c := range cases {
			names = append(names, c.name)
		}
		sort.Strings(names)
		t.Logf("rewrote %s: %d samples × %d events", path, len(out.Samples), len(names))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (regenerate with ASTRO_UPDATE_GOLDEN=1 go test)", err)
	}
	return data
}
