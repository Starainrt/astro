package orbit

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func TestParallacticAngleMatchesHourAngleForm(t *testing.T) {
	elements := sampleObservationElements()
	date := time.Date(2025, 11, 21, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))

	_, dec, _ := basic.OrbitHourAngleWithTopocentric(
		basic.Date2JD(date),
		shanghaiLon, shanghaiLat, observationTimezone(date), shanghaiHeightMeters,
		toBasicElements(elements),
	)
	got := ParallacticAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters)
	want := basic.ParallacticAngleByHourAngle(
		HourAngle(date, elements, shanghaiLon, shanghaiLat, shanghaiHeightMeters),
		dec,
		shanghaiLat,
	)
	if got != want {
		t.Fatalf("parallactic angle mismatch: got %.17g want %.17g", got, want)
	}
}

// 旧口径把时角与赤纬放在相差 1 ULP 儒略日的两条路径上求值；差分对照用它作参考实现。
func duplicatedSolveParallacticAngle(date time.Time, elements Elements, observerLon, observerLat, observerHeight float64) float64 {
	position := ApparentTopocentricEquatorial(date, elements, observerLon, observerLat, observerHeight)
	return basic.ParallacticAngleByHourAngle(
		HourAngle(date, elements, observerLon, observerLat, observerHeight),
		position.Dec,
		observerLat,
	)
}

func parallacticJDPathsAgree(date time.Time) bool {
	utcPath := basic.UTC2TT(basic.Date2JD(date.UTC()))
	hourAnglePath := basic.UTC2TT(basic.Date2JD(date) - observationTimezone(date)/24.0)
	return utcPath == hourAnglePath
}

type parallacticCase struct {
	label    string
	date     time.Time
	elements Elements
	lon, lat float64
}

func parallacticElementSets() []struct {
	name     string
	elements Elements
} {
	return []struct {
		name     string
		elements Elements
	}{
		{"mainbelt", sampleObservationElements()},
		{"neo", Elements{EpochJD: 2461000.5, A: 1.2, E: 0.35, I: 8, Omega: 200.1, W: 120.5, M0: 10}},
		{"perihelion", Elements{EpochJD: 2461000.5, Q: 0.6, E: 0.7, I: 15, Omega: 120, W: 33, TpJD: 2461005.5}},
		{"highinclination", Elements{EpochJD: 2461000.5, A: 3.2, E: 0.3, I: 160, Omega: 20, W: 260, M0: 100}},
	}
}

func parallacticFixedCases() []parallacticCase {
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("CST", 8*3600),
		time.FixedZone("EST", -5*3600),
		time.FixedZone("NPT", 5*3600+45*60),
		time.FixedZone("LMT", -7*3600-52*60-58),
	}
	sites := []struct {
		name     string
		lon, lat float64
	}{
		{"shanghai", 121.4737, 31.2304},
		{"sydney", 151.2093, -33.8688},
		{"north-pole", 0, 89.9999},
		{"south-pole", 0, -89.9999},
		{"dateline-west", -179.99, 12},
		{"dateline-east", 179.99, -12},
		{"reykjavik", -21.8174, 64.1265},
		{"equator", 0, 0},
	}
	years := []int{-500, 1000, 1582, 1900, 2025, 2100, 3000, 4000}
	hours := []int{0, 3, 7, 12, 17, 20, 23}
	cases := make([]parallacticCase, 0, len(parallacticElementSets())*len(sites)*len(years)*len(zones)*len(hours))
	for _, set := range parallacticElementSets() {
		for _, site := range sites {
			for _, year := range years {
				for _, zone := range zones {
					for _, hour := range hours {
						cases = append(cases, parallacticCase{
							label:    fmt.Sprintf("%s/%s/%d/%s/%d", set.name, site.name, year, zone, hour),
							date:     time.Date(year, 5, 17, hour, 43, 21, 123456789, zone),
							elements: set.elements,
							lon:      site.lon,
							lat:      site.lat,
						})
					}
				}
			}
		}
	}
	return cases
}

func parallacticRandomCases(count int) []parallacticCase {
	rng := rand.New(rand.NewSource(20260915))
	sets := parallacticElementSets()
	cases := make([]parallacticCase, 0, count)
	for i := 0; i < count; i++ {
		set := sets[rng.Intn(len(sets))]
		zone := time.FixedZone("random", (rng.Intn(97)-48)*1800)
		date := time.Date(
			rng.Intn(8000)-2000,
			time.Month(1+rng.Intn(12)),
			1+rng.Intn(28),
			rng.Intn(24), rng.Intn(60), rng.Intn(60), rng.Intn(1000000000),
			zone,
		)
		cases = append(cases, parallacticCase{
			label:    fmt.Sprintf("random/%d/%s", i, set.name),
			date:     date,
			elements: set.elements,
			lon:      rng.Float64()*360 - 180,
			lat:      rng.Float64()*179.8 - 89.9,
		})
	}
	return cases
}

// 固定民用时标换算政策，避免比较样本随默认值变化。
func TestParallacticAngleDifferentialAgainstDuplicatedSolve(t *testing.T) {
	previous := basic.GetTimeScaleFuturePolicy()
	basic.SetTimeScaleFuturePolicy(basic.TimeScaleLeapSecond)
	t.Cleanup(func() { basic.SetTimeScaleFuturePolicy(previous) })
	const (
		maxAbsoluteTolerance = 3e-9
		maxRelativeTolerance = 5e-11
	)
	cases := append(parallacticFixedCases(), parallacticRandomCases(240)...)
	bitwise, drifted := 0, 0
	maxAbsolute, maxRelative := 0.0, 0.0
	maxAbsoluteCase, maxRelativeCase := "", ""
	for _, tc := range cases {
		got := ParallacticAngle(tc.date, tc.elements, tc.lon, tc.lat, shanghaiHeightMeters)
		want := duplicatedSolveParallacticAngle(tc.date, tc.elements, tc.lon, tc.lat, shanghaiHeightMeters)
		if got == want {
			bitwise++
			continue
		}
		if parallacticJDPathsAgree(tc.date) {
			t.Fatalf("%s: 两条儒略日路径逐位相等，输出必须逐位相同，got %.17g want %.17g", tc.label, got, want)
		}
		drifted++
		diff := math.Abs(got - want)
		if diff > maxAbsolute {
			maxAbsolute, maxAbsoluteCase = diff, tc.label
		}
		if math.Abs(want) > 1 {
			if relative := diff / math.Abs(want); relative > maxRelative {
				maxRelative, maxRelativeCase = relative, tc.label
			}
		}
	}
	t.Logf("cases=%d bitwise=%d (%.1f%%) drifted=%d maxAbs=%.4g deg (%s) maxRel=%.4g (%s)",
		len(cases), bitwise, 100*float64(bitwise)/float64(len(cases)), drifted,
		maxAbsolute, maxAbsoluteCase, maxRelative, maxRelativeCase)
	if maxAbsolute > maxAbsoluteTolerance {
		t.Fatalf("max absolute drift %.4g deg exceeds %.4g (%s)", maxAbsolute, maxAbsoluteTolerance, maxAbsoluteCase)
	}
	if maxRelative > maxRelativeTolerance {
		t.Fatalf("max relative drift %.4g exceeds %.4g (%s)", maxRelative, maxRelativeTolerance, maxRelativeCase)
	}
}
