package svg

import (
	"html"
	"math"
	"regexp"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
	eclipsecore "github.com/starainrt/astro/eclipse"
)

// 详细版"食甚时的月亮（地心坐标）"块必须用食甚时刻的力学时求值，与 NASA 月食图同一时刻；
// 输入日期的时刻（当天 0 时或 12 时）会让赤经赤纬整体偏掉。
func TestLunarEclipseDetailedGeocentricBlockUsesGreatestEclipse(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	cases := []struct {
		name string
		date time.Time
		ra   string
		dec  string
	}{
		// 对应 NASA 图上的 Moon at Greatest Eclipse：2026-03-03 为 10h56m15.0s / +06°24'05.2"，
		// 2025-03-14 为 11h38m23.0s / +02°40'54.6"，与本库的差在 0.3 角秒内。
		{"2026-03-03", time.Date(2026, 3, 3, 0, 0, 0, 0, cst), "10h56m15.1s", "+06°24&#39;05.2&#34;"},
		{"2025-03-14", time.Date(2025, 3, 14, 0, 0, 0, 0, cst), "11h38m23.0s", "+02°40&#39;54.3&#34;"},
	}
	for _, want := range cases {
		doc, ok := LunarEclipseDetailedSVG(want.date, LunarEclipseDetailedSVGOptions{
			Language: "zh", Location: cst})
		if !ok {
			t.Fatalf("%s: missing chart", want.name)
		}
		ra, dec := lunarGeocentricMoonBlock(t, doc)
		if ra != want.ra || dec != want.dec {
			t.Errorf("%s: 食甚坐标 RA=%s Dec=%s, want RA=%s Dec=%s", want.name, ra, dec, want.ra, want.dec)
		}
		info, ok := eclipsecore.LunarEclipseOnDate(want.date)
		if !ok {
			t.Fatalf("%s: missing eclipse", want.name)
		}
		maximumRA, maximumDec := basic.HMoonTrueRaDec(timeToTTJDE(info.Maximum))
		if got, wantValue := ra, formatSolarEclipseRA(maximumRA); got != wantValue {
			t.Errorf("%s: RA=%s 不是食甚时刻的 %s", want.name, got, wantValue)
		}
		if got, wantValue := dec, html.EscapeString(formatSolarEclipseDec(maximumDec)); got != wantValue {
			t.Errorf("%s: Dec=%s 不是食甚时刻的 %s", want.name, got, wantValue)
		}
		inputRA, inputDec := basic.HMoonTrueRaDec(timeToTTJDE(want.date))
		if math.Abs(maximumRA-inputRA) < 1 {
			t.Fatalf("%s: 该用例区分不出输入日期与食甚时刻", want.name)
		}
		if dec == html.EscapeString(formatSolarEclipseDec(inputDec)) {
			t.Errorf("%s: Dec 仍等于输入日期时刻的值 %s", want.name, formatSolarEclipseDec(inputDec))
		}
	}
}

func lunarGeocentricMoonBlock(t *testing.T, doc string) (string, string) {
	t.Helper()
	block := regexp.MustCompile(`食甚时的月亮（地心坐标）(.*?)</g>`).FindStringSubmatch(doc)
	if block == nil {
		t.Fatal("geocentric Moon block not found")
	}
	value := func(label string) string {
		pattern := regexp.MustCompile(regexp.QuoteMeta(label) + `</text><text[^>]*text-anchor="end">([^<]*)</text>`)
		match := pattern.FindStringSubmatch(block[1])
		if match == nil {
			t.Fatalf("row %s not found in the geocentric Moon block", label)
		}
		return match[1]
	}
	return value("赤经 R.A."), value("赤纬 Dec.")
}
