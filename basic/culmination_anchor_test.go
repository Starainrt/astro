package basic

import (
	"math"
	"testing"
	"time"
)

// 本文件核实太阳中天（floor 正午锚点）与月亮中天（floor+0.5 午夜锚点）两套锚点：
// 调用方按各自约定补偿后，两者都必须落在同一个本地民用日，并接近当日本地日的真实中天。
//
// 框架约定（由各自调用方固定下来）：
//   - SunHeight/HMoonHeight 的 jd 是本地民用时框架，本地民用日从 Date2JD(本地 0 时) 起算；
//   - basic.CulminationTime 的返回值是本地民用时框架，sun/sun.go 减 tz/24 得到 UT；
//   - basic.MoonCulminationTime 的返回值就是本地民用时框架，moon/moon.go 按 byZone=true 使用。

type culminationAnchorSite struct {
	name string
	lon  float64
	lat  float64
	tz   float64
}

var culminationAnchorSites = []culminationAnchorSite{
	{"beijing", 116.4074, 39.9042, 8},
	{"utc", 0, 0, 0},
	{"newyork", -74, 40.7, -5},
	{"adelaide", 138.6, -34.9, 9.5},
	{"chatham", -176.5, -43.9, 12.75},
	{"kiritimati", -157.4, 1.9, 14},
}

func culminationAnchorLocalMidnight(site culminationAnchorSite, timestamp time.Time) (time.Time, float64) {
	location := time.FixedZone("anchor", int(site.tz*3600))
	local := time.Date(timestamp.Year(), timestamp.Month(), timestamp.Day(), 0, 0, 0, 0, location)
	return local, Date2JD(local)
}

func TestSunCulminationAnchorKeepsLocalDay(t *testing.T) {
	timestamp := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, site := range culminationAnchorSites {
		local, midnightJD := culminationAnchorLocalMidnight(site, timestamp)
		location := local.Location()
		// sun/sun.go 的补偿：本地 0 时 JD 再加半天，使 floor 落在同一本地日；随后减 tz/24 得 UT。
		got := JD2DateByZone(CulminationTime(midnightJD+0.5, site.lon, site.tz)-site.tz/24, location, false)
		truthJD, truthAltitude := 0.0, -999.0
		for minute := 0; minute <= 24*60; minute++ {
			jd := midnightJD + float64(minute)/1440.0
			if altitude := SunHeight(jd, site.lon, site.lat, site.tz); altitude > truthAltitude {
				truthAltitude, truthJD = altitude, jd
			}
		}
		truth := JD2DateByZone(truthJD-site.tz/24, location, false)
		if got.Day() != local.Day() || got.Month() != local.Month() {
			t.Fatalf("%s: sun culmination = %s, want local day %s", site.name, got.Format("2006-01-02 15:04"), local.Format("2006-01-02"))
		}
		if deviation := math.Abs(got.Sub(truth).Minutes()); deviation > 1.5 {
			t.Fatalf("%s: sun culmination deviates %.2f min from the daily altitude maximum", site.name, deviation)
		}
	}
}

func TestMoonCulminationAnchorKeepsLocalDay(t *testing.T) {
	timestamp := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, site := range culminationAnchorSites {
		local, midnightJD := culminationAnchorLocalMidnight(site, timestamp)
		location := local.Location()
		got := JD2DateByZone(MoonCulminationTime(midnightJD, site.lon, site.lat, site.tz), location, true)
		truthJD, truthAltitude := 0.0, -999.0
		for minute := 0; minute <= 24*60; minute++ {
			jd := midnightJD + float64(minute)/1440.0
			if altitude := HMoonHeight(jd, site.lon, site.lat, site.tz); altitude > truthAltitude {
				truthAltitude, truthJD = altitude, jd
			}
		}
		truth := JD2DateByZone(truthJD, location, true)
		if got.Day() != local.Day() || got.Month() != local.Month() {
			t.Fatalf("%s: moon culmination = %s, want local day %s", site.name, got.Format("2006-01-02 15:04"), local.Format("2006-01-02"))
		}
		if deviation := math.Abs(got.Sub(truth).Minutes()); deviation > 6.0 {
			t.Fatalf("%s: moon culmination deviates %.2f min from the daily altitude maximum", site.name, deviation)
		}
	}
}

func TestSunCulminationAnchorDiffersFromMidnightAnchor(t *testing.T) {
	site := culminationAnchorSites[0]
	_, midnightJD := culminationAnchorLocalMidnight(site, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	// 不补半天时 floor 落在相邻的 UT 日：这正是调用方必须补偿 +0.5 的原因。
	uncompensated := CulminationTime(midnightJD, site.lon, site.tz) - site.tz/24
	compensated := CulminationTime(midnightJD+0.5, site.lon, site.tz) - site.tz/24
	if math.Abs(uncompensated-compensated) < 0.4 {
		t.Fatalf("expected the uncompensated anchor to land on an adjacent day: %.6f vs %.6f", uncompensated, compensated)
	}
}
