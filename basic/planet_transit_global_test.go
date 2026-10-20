package basic

import (
	"math"
	"testing"
	"time"

	"github.com/starainrt/astro/internal/ephem"
)

func TestPlanetTransitGlobalGrazes(tester *testing.T) {
	for _, model := range []ephem.Model{ephem.Semianalytic, ephem.Legacy} {
		inst := ephem.NewModelInstance(model)
		for _, year := range []int{-540, 2854} {
			transit := NextVenusTransitWithInstance(JDCalc(year, 1, 1), inst)
			if !transit.Valid || JD2DateByZone(transit.Greatest, time.UTC, false).Year() != year {
				tester.Fatalf("model=%d year=%d: missed global event: %+v", model, year, transit)
			}
			if transit.HasExternal || transit.HasInternal || transit.ExternalIngress != 0 || transit.ExternalEgress != 0 {
				tester.Fatalf("global graze fabricated geocentric contacts: %+v", transit)
			}
			if math.Abs(transit.Gamma) <= transit.PenumbralRadius || math.Abs(transit.Gamma) >= transit.PenumbralRadius+1.01 {
				tester.Fatalf("unexpected shadow geometry: %+v", transit)
			}
			observer := Observer{Longitude: -135, Latitude: -60}
			local := PlanetTransitLocalWithInstance(transit, observer, inst)
			if !local.Valid || !local.Visible || local.HasInternal || !local.ExternalIngress.Valid || !local.ExternalEgress.Valid {
				tester.Fatalf("model=%d year=%d: missing visible local graze: %+v", model, year, local)
			}
			if year == 2854 {
				if other := PlanetTransitLocalWithInstance(transit, Observer{Latitude: 45}, inst); other.Valid {
					tester.Fatal("northern site must miss the 2854 event")
				}
			}
		}
	}
}

func TestPlanetTransitSeriesNumbers(tester *testing.T) {
	for _, sample := range []struct{ planet, year, number, period int }{
		{2, 1631, 6, 243}, {2, 1639, 4, 243}, {2, 2004, 3, 243}, {2, 2012, 5, 243}, {2, 2611, 4, 243}, {2, 2854, 4, 243},
		{1, 1957, 9, 46}, {1, 2003, 9, 46}, {1, 2006, 8, 46}, {1, 2016, 7, 46}, {1, 2019, 6, 46}, {1, 2174, 11, 46},
	} {
		query := JDCalc(sample.year, 1, 1)
		result := NextVenusTransit(query)
		if sample.planet == 1 {
			result = NextMercuryTransit(query)
		}
		if result.Series.Number != sample.number || result.Series.PeriodYears != sample.period {
			tester.Fatalf("planet=%d year=%d: series=%+v, want %d/%d", sample.planet, sample.year, result.Series, sample.number, sample.period)
		}
	}
}

func TestPlanetTransitSeriesUnassigned(tester *testing.T) {
	for _, sample := range []struct {
		cfg  planetTransitConfig
		year int
	}{
		{mercuryTransitConfig(), 1500}, {mercuryTransitConfig(), 2400},
		{venusTransitConfig(), -2100}, {venusTransitConfig(), 4100},
	} {
		series := planetTransitSeriesAt(JDCalc(sample.year, 1, 1), sample.cfg)
		if series.Number != 0 || series.PeriodYears == 0 {
			tester.Fatalf("unassigned series: %+v", series)
		}
	}
}

func TestPlanetTransitNearInternalEdge(tester *testing.T) {
	for _, year := range []int{1631, 2611} {
		transit := NextVenusTransit(JDCalc(year, 1, 1))
		if !transit.Valid || !transit.HasExternal || !transit.HasInternal {
			tester.Fatalf("year=%d: missing geocentric contacts", year)
		}
		full, partial := false, false
		for latitude := -75; latitude <= 75; latitude += 30 {
			for longitude := -180; longitude < 180; longitude += 45 {
				local := PlanetTransitLocal(transit, Observer{Longitude: float64(longitude), Latitude: float64(latitude)})
				if !local.Valid || !local.Visible {
					continue
				}
				if local.HasInternal {
					full = true
				} else {
					partial = true
				}
			}
		}
		if !full || !partial {
			tester.Fatalf("year=%d: full=%v partial=%v", year, full, partial)
		}
	}
}

func TestPlanetTransitInputBounds(tester *testing.T) {
	for _, query := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64} {
		if result := NextVenusTransit(query); result.Valid {
			tester.Fatal("non-finite query produced a transit")
		}
	}
}

func TestPlanetTransitResultRebuild(tester *testing.T) {
	greatest := time.Date(2854, 12, 14, 13, 0, 0, 0, time.UTC)
	if PlanetTransitResultFromTimes(2, time.Time{}, time.Time{}, greatest, time.Time{}, time.Time{}).Valid {
		tester.Fatal("contact rebuild accepted missing C1/C4")
	}
	for _, window := range [][2]time.Time{
		{{}, greatest.Add(time.Hour)}, {greatest, greatest.Add(time.Hour)}, {greatest.Add(-time.Hour), greatest},
	} {
		if PlanetTransitResultFromWindow(2, window[0], greatest, window[1]).Valid {
			tester.Fatalf("window rebuild accepted %v", window)
		}
	}
	transit := NextVenusTransit(JDCalc(2854, 1, 1))
	rebuilt := PlanetTransitResultFromWindow(2, JD2DateByZone(transit.WindowStart, time.UTC, false),
		JD2DateByZone(transit.Greatest, time.UTC, false), JD2DateByZone(transit.WindowEnd, time.UTC, false))
	observer := Observer{Longitude: -135, Latitude: -60}
	want, got := PlanetTransitLocal(transit, observer), PlanetTransitLocal(rebuilt, observer)
	if !got.Valid || math.Abs(got.ExternalIngress.JD-want.ExternalIngress.JD)*86400 > 0.01 || math.Abs(got.ExternalEgress.JD-want.ExternalEgress.JD)*86400 > 0.01 {
		tester.Fatalf("window rebuild changed local contacts: got %+v want %+v", got, want)
	}
}

// boundedCoverageEngine 把内置星历限制在一个时间窗内，窗外的请求按越界报错。
type boundedCoverageEngine struct {
	ephem.Engine
	start, end float64
}

func (bounded boundedCoverageEngine) within(jd float64) bool {
	return jd >= bounded.start && jd <= bounded.end
}

func (bounded boundedCoverageEngine) State(target, center ephem.Body, jdTDB float64) (pos, vel [3]float64, err error) {
	if !bounded.within(jdTDB) {
		return [3]float64{}, [3]float64{}, ephem.ErrOutOfRange
	}
	return bounded.Engine.State(target, center, jdTDB)
}

func (bounded boundedCoverageEngine) Apparent(target ephem.Body, jdTDB float64) (pos [3]float64, lightTime, distance float64, err error) {
	if !bounded.within(jdTDB) {
		return [3]float64{}, 0, 0, ephem.ErrOutOfRange
	}
	return bounded.Engine.(ephem.ApparentEngine).Apparent(target, jdTDB)
}

func (bounded boundedCoverageEngine) Coverage(ephem.Body) (start, end float64, ok bool) {
	return bounded.start, bounded.end, true
}

// 有限覆盖星历：搜索起点前的候选落在覆盖之外时，不得因此丢掉覆盖内的凌日。
func TestPlanetTransitBoundedEphemerisCoverage(tester *testing.T) {
	base := ephem.NewModelInstance(ephem.Semianalytic).Engine()
	start := JDCalc(2019, 1, 1)
	bounded := ephem.NewInstance(boundedCoverageEngine{Engine: base, start: start, end: JDCalc(2021, 1, 1)})
	query := start + 60
	transit := NextMercuryTransitWithInstance(query, bounded)
	if !transit.Valid {
		tester.Fatal("bounded ephemeris lost the 2019-11-11 transit inside its coverage")
	}
	if year := JD2DateByZone(transit.Greatest, time.UTC, false).Year(); year != 2019 {
		tester.Fatalf("bounded ephemeris returned year %d, want 2019", year)
	}
	if transit.HasExternal != true || transit.HasInternal != true {
		tester.Fatalf("bounded ephemeris returned an incomplete transit: %+v", transit)
	}
}
