package basic

import (
	"math"
	"testing"
	"time"
)

func TestMoonHorizonMatchesTopocentricAltitude(t *testing.T) {
	for _, jd := range []float64{JDCalc(2026, 3, 3), JDCalc(2025, 9, 7), JDCalc(2024, 12, 15)} {
		points := MoonHorizon(jd, 360)
		if len(points) != 360 {
			t.Fatalf("horizon points=%d", len(points))
		}
		for _, point := range points {
			if altitude := HMoonHeight(jd, point[0], point[1], 0); math.Abs(altitude) > 1e-9 {
				t.Fatalf("JD=%v point=%v altitude=%g, want zero", jd, point, altitude)
			}
		}
	}
	if MoonHorizon(math.NaN(), 360) != nil {
		t.Fatal("invalid JD accepted")
	}
}

// TestMoonHorizonUsesUTCInput 固定地平圈的时标口径：入参是 UTC 儒略日，站心恒星时按 UTC→UT1 换算。
// 取 DUT1 明显的两个时刻，确认圈上点只在 UTC 口径下高度角为零，且两种口径在本地可区分。
func TestMoonHorizonUsesUTCInput(t *testing.T) {
	for _, at := range []time.Time{
		time.Date(1980, 3, 15, 18, 0, 0, 0, time.UTC),
		time.Date(2035, 3, 15, 18, 0, 0, 0, time.UTC),
	} {
		jdUTC := Date2JD(at)
		dut1Seconds := (UTC2UT1(jdUTC) - jdUTC) * 86400
		if math.Abs(dut1Seconds) < 0.3 {
			t.Fatalf("%s DUT1=%.3f s 太小，区分不出两种口径", at.Format("2006-01-02"), dut1Seconds)
		}
		jdAsUT1 := jdUTC + dut1Seconds/86400
		distinguishable := false
		for _, point := range MoonHorizon(jdUTC, 360) {
			if altitude := HMoonHeight(jdUTC, point[0], point[1], 0); math.Abs(altitude) > 1e-9 {
				t.Fatalf("%s UTC 口径下圈上点高度角=%g，应为零", at.Format("2006-01-02"), altitude)
			}
			if altitude := HMoonHeight(jdAsUT1, point[0], point[1], 0); math.Abs(altitude) > 1e-4 {
				distinguishable = true
			}
		}
		if !distinguishable {
			t.Fatalf("%s 两种口径不可区分，用例失去意义", at.Format("2006-01-02"))
		}
		if MoonHorizon(jdAsUT1, 360)[0] == MoonHorizon(jdUTC, 360)[0] {
			t.Fatalf("%s UTC 与 UT1 两种读法给出了同一条圈", at.Format("2006-01-02"))
		}
	}
}

// TestMoonStateHorizonMatchesPackageHorizon 固定 MoonState 派生量与包级函数逐位一致，以及采样数与非法入参的兜底。
func TestMoonStateHorizonMatchesPackageHorizon(t *testing.T) {
	jd := JDCalc(2026, 3, 3)
	state := MoonStateAt(jd)
	for _, samples := range []int{0, 1, 12, 360, 5000} {
		got := state.MoonHorizon(samples)
		want := MoonHorizon(jd, samples)
		if len(got) != len(want) {
			t.Fatalf("samples=%d 点数 %d，包级 %d", samples, len(got), len(want))
		}
		for index := range got {
			if got[index] != want[index] {
				t.Fatalf("samples=%d 第 %d 点 %v，包级 %v", samples, index, got[index], want[index])
			}
		}
	}
	for _, invalid := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if points := MoonStateAt(invalid).MoonHorizon(360); points != nil {
			t.Fatalf("非有限 UTC 儒略日 %v 仍给出地平圈", invalid)
		}
		if points := MoonHorizon(invalid, 360); points != nil {
			t.Fatalf("非有限 UTC 儒略日 %v 仍给出地平圈", invalid)
		}
	}
	if altitude := MoonStateAt(math.NaN()).HMoonHeight(0, 0); !math.IsNaN(altitude) {
		t.Fatalf("非有限状态的高度角=%v，期望 NaN", altitude)
	}
}

// TestHMoonHeightUsesLocalCivilFrame 固定 HMoonHeight 的时标框架：jd 是当地民用时（墙上时刻），
// tz 是时区偏移小时数。同一物理时刻写成「当地民用时 + tz」与「UTC 数值 + tz=0」必须一致；
// 把 UTC 数值再配非零 tz 会多减一次时区，必须能区分。
func TestHMoonHeightUsesLocalCivilFrame(t *testing.T) {
	utc := time.Date(2029, 1, 1, 16, 0, 0, 0, time.UTC)
	zone := time.FixedZone("CST", 8*3600)
	utcJD := Date2JD(utc)
	localJD := Date2JD(utc.In(zone))
	// JD 约 2.5e6 天，双精度下 8 小时之差本身带 ~1e-8 h 的表示误差。
	if offset := (localJD - utcJD) * 24; math.Abs(offset-8) > 1e-6 {
		t.Fatalf("当地民用时框架与 UTC 相差 %.9f h，期望 8 h", offset)
	}
	for _, site := range []struct{ lon, lat float64 }{{108.729, -59.937}, {0, 0}, {-70, 45}} {
		local := HMoonHeight(localJD, site.lon, site.lat, 8)
		reference := HMoonHeight(utcJD, site.lon, site.lat, 0)
		if math.Abs(local-reference) > 1e-5 {
			t.Fatalf("(%.3f,%.3f) 当地民用时 %.9f，UTC 数值配 tz=0 %.9f", site.lon, site.lat, local, reference)
		}
	}
	state := MoonStateAt(utcJD)
	for _, site := range []struct{ lon, lat float64 }{{108.729, -59.937}, {0, 0}, {-70, 45}} {
		if state.HMoonHeight(site.lon, site.lat) != HMoonHeight(utcJD, site.lon, site.lat, 0) {
			t.Fatalf("(%.3f,%.3f) MoonState 与 HMoonHeight 不一致", site.lon, site.lat)
		}
	}
	if mixed := HMoonHeight(utcJD, 0, 0, 8) - HMoonHeight(utcJD, 0, 0, 0); math.Abs(mixed) < 0.1 {
		t.Fatalf("UTC 数值配非零 tz 的差异只有 %.6f°，用例失去区分度", mixed)
	}
}
