package eclipse

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/moon/limb"
)

type lunarLimbCase struct {
	name     string
	date     time.Time
	lon, lat float64
	total    bool
}

var lunarLimbCases = []lunarLimbCase{
	{"2024-04-08 Dallas total", time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC), -96.8, 32.8, true},
	{"2024-04-08 Torreon total", time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC), -103.43, 25.54, true},
	{"2023-10-14 Albuquerque annular", time.Date(2023, 10, 14, 16, 0, 0, 0, time.UTC), -106.65, 35.08, false},
}

func lunarLimbContacts(info LocalSolarEclipseInfo) [4]time.Time {
	return [4]time.Time{info.PartialStart, info.CentralStart, info.CentralEnd, info.PartialEnd}
}

// 月缘只应让接触时刻移动几秒：既要确实移动，又不能超过 15 s；食型与默认结果都不变。
func TestLocalSolarEclipseLunarLimbShiftsContacts(t *testing.T) {
	limbed := NewWithInstance(nil).WithLunarLimb(limb.Default())
	for _, c := range lunarLimbCases {
		circular, ok := defaultModule.LocalSolarEclipseOnDate(c.date, c.lon, c.lat, 0)
		if !ok || !circular.HasCentral || circular.HasTotal != c.total {
			t.Fatalf("%s: circular type %v", c.name, circular.Type)
		}
		corrected, ok := limbed.LocalSolarEclipseOnDate(c.date, c.lon, c.lat, 0)
		if !ok || corrected.Type != circular.Type {
			t.Fatalf("%s: limb changed type %v -> %v", c.name, circular.Type, corrected.Type)
		}
		moved := false
		a, b := lunarLimbContacts(circular), lunarLimbContacts(corrected)
		for _, point := range corrected.ContactPoints {
			want := b[map[string]int{"C1": 0, "C2": 1, "C3": 2, "C4": 3}[point.Label]]
			if gap := point.Time.Sub(want).Seconds(); math.Abs(gap) > 0.05 {
				t.Errorf("%s %s contact point off phase time by %.3f s", c.name, point.Label, gap)
			}
		}
		if len(corrected.ContactPoints) != 4 {
			t.Errorf("%s: %d contact points", c.name, len(corrected.ContactPoints))
		}
		for index := range a {
			shift := b[index].Sub(a[index]).Seconds()
			t.Logf("%s contact %d shift %+.3f s", c.name, index, shift)
			if math.Abs(shift) > 15 {
				t.Errorf("%s contact %d moved %.3f s", c.name, index, shift)
			}
			if math.Abs(shift) > 0.01 {
				moved = true
			}
		}
		if !moved {
			t.Errorf("%s: lunar limb did not move any contact", c.name)
		}
		t.Logf("%s central duration circular %.3f s, limb %.3f s", c.name,
			circular.CentralEnd.Sub(circular.CentralStart).Seconds(),
			corrected.CentralEnd.Sub(corrected.CentralStart).Seconds())

		// 选项里的月缘与模块设置等价；模块设置不影响原模块。
		viaOptions, _ := defaultModule.LocalSolarEclipseOnDateWithOptions(c.date, c.lon, c.lat, 0,
			SolarEclipseOptions{LunarLimb: limb.Default()})
		if lunarLimbContacts(viaOptions) != b {
			t.Errorf("%s: options limb differs from module limb", c.name)
		}
		again, _ := defaultModule.LocalSolarEclipseOnDate(c.date, c.lon, c.lat, 0)
		if lunarLimbContacts(again) != a {
			t.Errorf("%s: WithLunarLimb changed the original module", c.name)
		}
	}
}

// 2025-09-07 月全食紫金山天文台发布时刻（北京时间列换算 UT，0.1 分钟精度），只用于观察月缘修正的方向与量级。
func TestLunarEclipseLunarLimbAgainstPMO(t *testing.T) {
	pmo := []struct {
		label string
		ut    time.Time
	}{
		{"P1", time.Date(2025, 9, 7, 15, 26, 54, 0, time.UTC)},
		{"U1", time.Date(2025, 9, 7, 16, 26, 48, 0, time.UTC)},
		{"U2", time.Date(2025, 9, 7, 17, 30, 24, 0, time.UTC)},
		{"U3", time.Date(2025, 9, 7, 18, 53, 12, 0, time.UTC)},
		{"U4", time.Date(2025, 9, 7, 19, 56, 54, 0, time.UTC)},
		{"P4", time.Date(2025, 9, 7, 20, 56, 36, 0, time.UTC)},
	}
	date := time.Date(2025, 9, 7, 18, 0, 0, 0, time.UTC)
	plain := NewWithInstance(nil)
	limbed := plain.WithLunarLimb(limb.Default())
	for _, chauvenet := range []bool{false, true} {
		var circular, corrected LunarEclipseInfo
		if chauvenet {
			circular, corrected = plain.ClosestLunarEclipseChauvenet(date), limbed.ClosestLunarEclipseChauvenet(date)
		} else {
			circular, corrected = plain.ClosestLunarEclipseDanjon(date), limbed.ClosestLunarEclipseDanjon(date)
		}
		if !corrected.Maximum.Equal(circular.Maximum) || corrected.Type != circular.Type {
			t.Fatalf("chauvenet=%v: maximum or type changed", chauvenet)
		}
		c := []time.Time{circular.PenumbralStart, circular.PartialStart, circular.TotalStart, circular.TotalEnd, circular.PartialEnd, circular.PenumbralEnd}
		l := []time.Time{corrected.PenumbralStart, corrected.PartialStart, corrected.TotalStart, corrected.TotalEnd, corrected.PartialEnd, corrected.PenumbralEnd}
		var circErr, limbErr float64
		for i, ref := range pmo {
			ce := LunarEclipseInfoInUT1(circular)
			le := LunarEclipseInfoInUT1(corrected)
			cu := []time.Time{ce.PenumbralStart, ce.PartialStart, ce.TotalStart, ce.TotalEnd, ce.PartialEnd, ce.PenumbralEnd}[i]
			lu := []time.Time{le.PenumbralStart, le.PartialStart, le.TotalStart, le.TotalEnd, le.PartialEnd, le.PenumbralEnd}[i]
			shift := l[i].Sub(c[i]).Seconds()
			de, dl := cu.Sub(ref.ut).Seconds(), lu.Sub(ref.ut).Seconds()
			circErr += math.Abs(de)
			limbErr += math.Abs(dl)
			t.Logf("chauvenet=%v %s shift %+6.2f s  PMO err circular %+6.1f s limb %+6.1f s", chauvenet, ref.label, shift, de, dl)
			if math.Abs(shift) > 20 {
				t.Errorf("chauvenet=%v %s: limb shift %.2f s too large", chauvenet, ref.label, shift)
			}
		}
		t.Logf("chauvenet=%v PMO MAE circular %.2f s, limb %.2f s", chauvenet, circErr/6, limbErr/6)
		// 紫金山按 Chauvenet 口径发布，只有该口径可比精度；Danjon 与之差约 90 s 属口径差。
		if chauvenet && (limbErr > circErr || limbErr/6 > 2) {
			t.Errorf("PMO MAE: limb %.2f s, circular %.2f s; want the limb no worse and within 2 s", limbErr/6, circErr/6)
		}
	}
}

// 2024-04-08 全球日食紫金山天文台发布 P1/P4（UT，整秒），只用于观察月缘修正的方向与量级。
func TestGlobalSolarEclipseLunarLimbAgainstPMO(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	plain := NewWithInstance(nil)
	limbed := plain.WithLunarLimb(limb.Default())
	circular, corrected := plain.ClosestSolarEclipse(date), limbed.ClosestSolarEclipse(date)
	if !corrected.GreatestEclipse.Equal(circular.GreatestEclipse) || corrected.Type != circular.Type {
		t.Fatal("greatest or type changed")
	}
	cu, lu := SolarEclipseInfoInUT1(circular), SolarEclipseInfoInUT1(corrected)
	for _, c := range []struct {
		label    string
		ref      time.Time
		circ, lb time.Time
	}{
		{"P1", time.Date(2024, 4, 8, 15, 42, 13, 0, time.UTC), cu.PartialBeginOnEarth, lu.PartialBeginOnEarth},
		{"P4", time.Date(2024, 4, 8, 20, 52, 21, 0, time.UTC), cu.PartialEndOnEarth, lu.PartialEndOnEarth},
	} {
		shift := c.lb.Sub(c.circ).Seconds()
		t.Logf("%s shift %+6.2f s  PMO err circular %+6.2f s limb %+6.2f s", c.label, shift,
			c.circ.Sub(c.ref).Seconds(), c.lb.Sub(c.ref).Seconds())
		if math.Abs(shift) > 20 {
			t.Errorf("%s: limb shift %.2f s too large", c.label, shift)
		}
		// 发布值取整秒，±0.5 s 的舍入淹没圆月面与月缘之差，只约束月缘结果贴近发布值。
		if d := c.lb.Sub(c.ref).Seconds(); math.Abs(d) > 1.5 {
			t.Errorf("%s: limb result %.2f s from PMO", c.label, d)
		}
	}
	t.Logf("central duration circular %.2f s limb %.2f s", circular.CentralDuration.Seconds(), corrected.CentralDuration.Seconds())
}

// 偏食足迹里的全球结果与 P1/P4 必须与全球日食接口同源，不能只有可见域边界按月缘修正。
func TestPartialFootprintsLunarLimbMatchesGlobal(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	limbed := NewWithInstance(nil).WithLunarLimb(limb.Default())
	global, ok := limbed.SolarEclipseOnDate(date)
	if !ok || !global.LunarLimbApplied {
		t.Fatal("global limb correction not applied")
	}
	footprints, ok := limbed.SolarEclipsePartialFootprints(date, SolarEclipsePartialFootprintOptions{})
	if !ok || !footprints.PartialBoundaryLimbApplied {
		t.Fatal("partial boundary limb correction not applied")
	}
	embedded := footprints.Eclipse
	if !embedded.LunarLimbApplied ||
		!embedded.PartialBeginOnEarth.Equal(global.PartialBeginOnEarth) ||
		!embedded.PartialEndOnEarth.Equal(global.PartialEndOnEarth) ||
		embedded.CentralDuration != global.CentralDuration {
		t.Fatalf("embedded eclipse differs from global: applied=%v P1 %v/%v P4 %v/%v",
			embedded.LunarLimbApplied, embedded.PartialBeginOnEarth, global.PartialBeginOnEarth,
			embedded.PartialEndOnEarth, global.PartialEndOnEarth)
	}
	for _, c := range []struct {
		label   string
		point   SolarEclipsePathPoint
		when    time.Time
		contact SolarEclipseLimbContact
	}{
		{"P1", footprints.P1, global.PartialBeginOnEarth, global.PartialBeginContact},
		{"P4", footprints.P4, global.PartialEndOnEarth, global.PartialEndContact},
	} {
		if !c.contact.Defined {
			continue
		}
		if !c.point.Time.Equal(c.when) || c.point.Longitude != c.contact.Longitude || c.point.Latitude != c.contact.Latitude {
			t.Fatalf("%s: footprint contact %v (%.4f, %.4f) differs from global %v (%.4f, %.4f)",
				c.label, c.point.Time, c.point.Longitude, c.point.Latitude, c.when, c.contact.Longitude, c.contact.Latitude)
		}
	}
}

// 月缘只平移全食带南北界，量级应为数百米到数公里。
func TestCentralPathLunarLimbShiftsLimits(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	plain := NewWithInstance(nil)
	circular, ok := plain.SolarEclipseCentralPath(date, SolarEclipsePathOptions{})
	if !ok {
		t.Fatal("no central path")
	}
	corrected, ok := plain.WithLunarLimb(limb.Default()).SolarEclipseCentralPath(date, SolarEclipsePathOptions{})
	if !ok || len(corrected.NorthernLimit) != len(circular.NorthernLimit) || len(corrected.SouthernLimit) != len(circular.SouthernLimit) {
		t.Fatal("limit lengths changed")
	}
	for side, pair := range [][2][]SolarEclipsePathPoint{{circular.NorthernLimit, corrected.NorthernLimit}, {circular.SouthernLimit, corrected.SouthernLimit}} {
		worst, sum, moved := 0.0, 0.0, 0
		for i := range pair[0] {
			a, b := pair[0][i], pair[1][i]
			d := math.Hypot((a.Longitude-b.Longitude)*math.Cos(a.Latitude*math.Pi/180), a.Latitude-b.Latitude) * 111.2
			if d > 0 {
				moved++
			}
			sum += d
			worst = math.Max(worst, d)
		}
		t.Logf("side %d: %d/%d points moved, mean %.3f km, max %.3f km", side, moved, len(pair[0]), sum/float64(len(pair[0])), worst)
		if worst > 10 || moved == 0 {
			t.Errorf("side %d: implausible limb shift (max %.3f km, moved %d)", side, worst, moved)
		}
	}
}

// 月缘让全食带变窄、时长变化在数秒以内；带宽与足迹必须与修正后的南北界同向。
func TestCentralPathLunarLimbWidthsAndDuration(t *testing.T) {
	date := time.Date(2024, 4, 8, 18, 0, 0, 0, time.UTC)
	plain := NewWithInstance(nil)
	circular, ok := plain.SolarEclipseCentralPath(date, SolarEclipsePathOptions{})
	if !ok {
		t.Fatal("no central path")
	}
	corrected, ok := plain.WithLunarLimb(limb.Default()).SolarEclipseCentralPath(date, SolarEclipsePathOptions{})
	if !ok || len(corrected.CenterLine) != len(circular.CenterLine) {
		t.Fatal("center line changed")
	}
	narrower, total := 0, 0
	worst := 0.0
	for i := range circular.CenterLine {
		a, b := circular.CenterLine[i].WidthKM, corrected.CenterLine[i].WidthKM
		if a <= 0 || b <= 0 {
			continue
		}
		total++
		if b < a {
			narrower++
		}
		worst = math.Max(worst, math.Abs(b-a))
	}
	t.Logf("width: %d/%d samples narrower, max change %.3f km; greatest %.3f -> %.3f km",
		narrower, total, worst, circular.Greatest.WidthKM, corrected.Greatest.WidthKM)
	t.Logf("max central duration %.2f s -> %.2f s",
		circular.MaxCentralDuration.Seconds(), corrected.MaxCentralDuration.Seconds())
	if worst > 10 || total == 0 {
		t.Errorf("implausible width change %.3f km over %d samples", worst, total)
	}
	if d := math.Abs((corrected.MaxCentralDuration - circular.MaxCentralDuration).Seconds()); d > 10 {
		t.Errorf("max duration moved by %.2f s", d)
	}
	if len(corrected.CentralBandSegments) != len(circular.CentralBandSegments) {
		t.Fatal("band segment count changed")
	}
}
