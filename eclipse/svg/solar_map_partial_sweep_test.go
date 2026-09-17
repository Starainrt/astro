package svg

import (
	"testing"
	"time"
)

// 偏食足迹扫描的密步长请求被夹到两分钟：并集需要一整条自洽的扫描序列，
// 更密的扫描成倍放大成本却不改产物，事后抽稀又会破坏并集的覆盖面判定。
func TestSolarEclipseMapPartialStepUsesDefaultSweep(t *testing.T) {
	date := time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		requested time.Duration
		want      time.Duration
	}{
		{requested: 0, want: 2 * time.Minute},
		{requested: -time.Minute, want: 2 * time.Minute},
		{requested: time.Second, want: 2 * time.Minute},
		{requested: 10 * time.Second, want: 2 * time.Minute},
		{requested: time.Minute, want: 2 * time.Minute},
		{requested: 119 * time.Second, want: 2 * time.Minute},
		{requested: 2 * time.Minute, want: 2 * time.Minute},
		{requested: 3 * time.Minute, want: 3 * time.Minute},
		{requested: 10 * time.Minute, want: 10 * time.Minute},
	} {
		got := normalizeSolarEclipseMapSVGOptions(date, SolarEclipseMapSVGOptions{
			PartialStep: test.requested,
		}).PartialStep
		if got != test.want {
			t.Fatalf("PartialStep %s normalized to %s, want %s", test.requested, got, test.want)
		}
	}

	// 一秒请求与默认两分钟请求必须逐字节相同，代价也回到默认量级。
	fast, fastOK := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{
		Width: 900, Height: 620, Location: time.UTC, PartialStep: time.Second,
	})
	defaulted, defaultOK := SolarEclipseMapSVG(date, SolarEclipseMapSVGOptions{
		Width: 900, Height: 620, Location: time.UTC, PartialStep: 2 * time.Minute,
	})
	if !fastOK || !defaultOK {
		t.Fatal("missing 2024-04-08 map")
	}
	if fast != defaulted {
		t.Fatalf("a one-second sweep request changed the rendered document (%d vs %d bytes)", len(fast), len(defaulted))
	}
}
