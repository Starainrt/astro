package timenote

import (
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro"
)

// 时标声明里同一个 UTC 偏移只写一次：时区名本身是偏移标签（FixedZone("UTC+08:00") 之类）时不再追加偏移。
func TestScaleWritesTheOffsetOnce(t *testing.T) {
	instant := time.Date(2025, time.June, 5, 20, 2, 6, 0, time.UTC)
	for _, test := range []struct {
		name     string
		location *time.Location
		language string
		want     string
	}{
		{name: "偏移型时区名", location: time.FixedZone("UTC+08:00", 8*3600), want: "图中时刻为 UTC（显示时区 UTC+08:00）"},
		{name: "偏移型时区名英文", location: time.FixedZone("UTC+08:00", 8*3600), language: "en", want: "All times are UTC (shown in UTC+08:00)"},
		{name: "整点时区名", location: time.FixedZone("UTC+8", 8*3600), want: "图中时刻为 UTC（显示时区 UTC+8）"},
		{name: "半小时偏移名", location: time.FixedZone("+05:30", 5*3600+1800), want: "图中时刻为 UTC（显示时区 +05:30）"},
		{name: "地名缩写中文", location: time.FixedZone("CST", 8*3600), want: "图中时刻为 UTC（显示时区 CST，UTC+08:00）"},
		{name: "地名缩写英文", location: time.FixedZone("EST", -5*3600), language: "en", want: "All times are UTC (shown in EST, UTC-05:00)"},
		{name: "偏移与名字不符", location: time.FixedZone("UTC+3", 8*3600), want: "图中时刻为 UTC（显示时区 UTC+3，UTC+08:00）"},
		{name: "UTC", location: time.UTC, want: "图中时刻为 UTC"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := Scale(astro.TimeScaleUTC, instant, test.location, test.language)
			if got != test.want {
				t.Fatalf("Scale() = %q, want %q", got, test.want)
			}
			if strings.Count(got, "UTC+08:00") > 1 || strings.Count(got, "UTC+8") > 1 {
				t.Fatalf("Scale() = %q, want the offset written once", got)
			}
		})
	}
}

func TestNamedZoneOffset(t *testing.T) {
	for _, test := range []struct {
		name   string
		value  string
		offset int
		ok     bool
	}{
		{name: "UTC 前缀", value: "UTC+08:00", offset: 8 * 3600, ok: true},
		{name: "GMT 前缀", value: "GMT-5", offset: -5 * 3600, ok: true},
		{name: "无前缀", value: "+0800", offset: 8 * 3600, ok: true},
		{name: "带秒", value: "UTC+05:30:30", offset: 5*3600 + 30*60 + 30, ok: true},
		{name: "地名", value: "CST", ok: false},
		{name: "UTC 本身", value: "UTC", ok: false},
		{name: "小数偏移", value: "UTC+8.5", ok: false},
		{name: "分钟越界", value: "UTC+0870", ok: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			offset, ok := namedZoneOffset(test.value)
			if ok != test.ok || offset != test.offset {
				t.Fatalf("namedZoneOffset(%q) = (%d, %v), want (%d, %v)", test.value, offset, ok, test.offset, test.ok)
			}
		})
	}
}
