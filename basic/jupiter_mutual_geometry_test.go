package basic

import (
	"math"
	"sort"
	"testing"
)

const (
	mutualIndependentFieldTolerance      = 1e-7
	mutualIndependentDistanceToleranceKM = 0.01
)

type mutualEclipseIndependentGeometry struct {
	valid                                                                   bool
	active, passive                                                         int
	separationKM, axialKM, distanceKM, umbraKM, penumbraKM, passiveRadiusKM float64
}

func mutualEclipseIndependentAt(jde float64, first, second int) mutualEclipseIndependentGeometry {
	jupiter := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(jde), orbitJ2000Obliquity)
	positions := [2]Vector3{}
	for index, satellite := range []int{first, second} {
		state := jupiterGalileanSatelliteCorrectedStateAtET(jde-jupiterGalileanReferenceJD, satellite-1, jupiterGalileanUseSolarLongPeriod(jde))
		positions[index] = Vector3{jupiter[0] + state.X, jupiter[1] + state.Y, jupiter[2] + state.Z}
	}
	if vectorDot(positions[0], positions[0]) > vectorDot(positions[1], positions[1]) {
		positions[0], positions[1], first, second = positions[1], positions[0], second, first
	}
	distanceAU := math.Sqrt(vectorDot(positions[0], positions[0]))
	delta := Vector3{positions[1][0] - positions[0][0], positions[1][1] - positions[0][1], positions[1][2] - positions[0][2]}
	axialAU := vectorDot(delta, positions[0]) / distanceAU
	if axialAU <= 0 {
		return mutualEclipseIndependentGeometry{}
	}
	separationKM := math.Sqrt(math.Max(0, vectorDot(delta, delta)-axialAU*axialAU)) * astronomicalUnitKM
	distanceKM, axialKM := distanceAU*astronomicalUnitKM, axialAU*astronomicalUnitKM
	activeRadiusKM := jupiterGalileanSatelliteRadiusJupiterRadii(first) * jupiterGalileanEquatorialRadiusKM
	passiveRadiusKM := jupiterGalileanSatelliteRadiusJupiterRadii(second) * jupiterGalileanEquatorialRadiusKM
	sunRadiusKM := solarRadiusAU * astronomicalUnitKM
	return mutualEclipseIndependentGeometry{
		valid: true, active: first, passive: second, separationKM: separationKM, axialKM: axialKM, distanceKM: distanceKM,
		umbraKM:         activeRadiusKM - axialKM*(sunRadiusKM-activeRadiusKM)/distanceKM,
		penumbraKM:      activeRadiusKM + axialKM*(sunRadiusKM+activeRadiusKM)/distanceKM,
		passiveRadiusKM: passiveRadiusKM,
	}
}

func mutualIndependentOverlap(separation, firstRadius, secondRadius float64) float64 {
	if firstRadius <= 0 || secondRadius <= 0 || separation >= firstRadius+secondRadius {
		return 0
	}
	if separation <= math.Abs(firstRadius-secondRadius) {
		radius := math.Min(firstRadius, secondRadius)
		return math.Pi * radius * radius
	}
	firstAngleRad := math.Acos(math.Max(-1, math.Min(1, (separation*separation+firstRadius*firstRadius-secondRadius*secondRadius)/(2*separation*firstRadius))))
	secondAngleRad := math.Acos(math.Max(-1, math.Min(1, (separation*separation+secondRadius*secondRadius-firstRadius*firstRadius)/(2*separation*secondRadius))))
	product := (-separation + firstRadius + secondRadius) * (separation + firstRadius - secondRadius) * (separation - firstRadius + secondRadius) * (separation + firstRadius + secondRadius)
	return firstRadius*firstRadius*firstAngleRad + secondRadius*secondRadius*secondAngleRad - math.Sqrt(math.Max(0, product))/2
}

func TestJupiterGalileanMutualEclipseConsistencyIndependentFields(t *testing.T) {
	pairs := [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}}
	expected := 214
	if testing.Short() {
		pairs = pairs[1:2]
		expected = 43
	}
	maximumCoverage, maximumImpact, maximumGap, maximumMinimum := 0.0, 0.0, 0.0, 0.0
	count := 0
	for _, pair := range pairs {
		for _, event := range collectMutualEclipsesBySearch(JDCalc(2026, 1, 1), JDCalc(2028, 1, 1), pair[0], pair[1]) {
			greatest := jupiterGalileanEclipseEmission(event.Greatest, event.Passive, nil)
			geometry := mutualEclipseIndependentAt(greatest, pair[0], pair[1])
			if !geometry.valid || geometry.active != event.Active || geometry.passive != event.Passive {
				t.Fatalf("independent identity pair=%v event=%+v", pair, event)
			}
			coverage := mutualIndependentOverlap(geometry.separationKM, geometry.umbraKM, geometry.passiveRadiusKM) / (math.Pi * geometry.passiveRadiusKM * geometry.passiveRadiusKM)
			impact := geometry.separationKM / (geometry.penumbraKM + geometry.passiveRadiusKM)
			annular := geometry.umbraKM > 0 && geometry.separationKM+geometry.umbraKM <= geometry.passiveRadiusKM
			maximumCoverage = math.Max(maximumCoverage, math.Abs(coverage-event.Coverage))
			maximumImpact = math.Max(maximumImpact, math.Abs(impact-event.ImpactParameter))
			if !isFinite(coverage) || !isFinite(impact) || math.Abs(coverage-event.Coverage) > mutualIndependentFieldTolerance || math.Abs(impact-event.ImpactParameter) > mutualIndependentFieldTolerance || annular != event.Annular {
				t.Fatalf("independent fields pair=%v greatest=%.12f coverage=%.12g/%.12g b=%.12g/%.12g annular=%v/%v", pair, event.Greatest, coverage, event.Coverage, impact, event.ImpactParameter, annular, event.Annular)
			}
			start := jupiterGalileanEclipseEmission(event.Start, event.Passive, nil)
			end := jupiterGalileanEclipseEmission(event.End, event.Passive, nil)
			for _, contact := range []float64{start, end} {
				boundary := mutualEclipseIndependentAt(contact, pair[0], pair[1])
				gap := math.Abs(boundary.separationKM - boundary.penumbraKM - boundary.passiveRadiusKM)
				maximumGap = math.Max(maximumGap, gap)
				if !boundary.valid || gap > mutualIndependentDistanceToleranceKM {
					t.Fatalf("independent contact pair=%v gap=%.12f km", pair, gap)
				}
			}
			for sample := 0; ; sample++ {
				candidate := math.Min(end, start+float64(sample)*10/86400)
				other := mutualEclipseIndependentAt(candidate, pair[0], pair[1])
				difference := geometry.separationKM - other.separationKM
				maximumMinimum = math.Max(maximumMinimum, difference)
				if !other.valid || other.active != geometry.active || difference > mutualIndependentDistanceToleranceKM {
					t.Fatalf("independent global minimum pair=%v sample=%.12f improvement=%.12f km", pair, candidate, difference)
				}
				if candidate >= end {
					break
				}
			}
			state := JupiterGalileanSatellitePhenomenaWithInstance(event.Greatest, nil)
			if !mutualPhenomenaExactlyEqual(state, event.GreatestState) {
				t.Fatalf("state at own Greatest differs pair=%v", pair)
			}
			count++
		}
	}
	if count != expected {
		t.Fatalf("independent fields events=%d want=%d", count, expected)
	}
	t.Logf("events=%d max coverage=%.12g impact=%.12g annular mismatches=0 contact gap=%.12g km sampled minimum improvement=%.12g km; states at own Greatest exact", count, maximumCoverage, maximumImpact, maximumGap, maximumMinimum)
}

func TestJupiterGalileanMutualEclipseConsistencyIndependentScan(t *testing.T) {
	if testing.Short() {
		t.Skip("two-year six-pair independent gap scan")
	}
	const coarseStepDays = 10.0 / 1440
	const fineStepDays = 30.0 / 86400
	start, end := JDCalc(2026, 1, 1), JDCalc(2028, 1, 1)
	totalIntervals, totalEvents, totalSearchOnly := 0, 0, 0
	for _, pair := range [][2]int{{1, 2}, {1, 3}, {1, 4}, {2, 3}, {2, 4}, {3, 4}} {
		firstSpeed := 2 * math.Pi * jupiterGalileanMutualSemimajorAxis(pair[0]) / jupiterGalileanOrbitPeriodDays(pair[0])
		secondSpeed := 2 * math.Pi * jupiterGalileanMutualSemimajorAxis(pair[1]) / jupiterGalileanOrbitPeriodDays(pair[1])
		thresholdKM := 3 * (firstSpeed + secondSpeed) * jupiterGalileanEquatorialRadiusKM * coarseStepDays * 1.5
		gap := func(jde float64) float64 {
			geometry := mutualEclipseIndependentAt(jde, pair[0], pair[1])
			if !geometry.valid {
				return math.Inf(1)
			}
			return geometry.separationKM - geometry.penumbraKM - geometry.passiveRadiusKM
		}
		var intervals [][2]float64
		inside, entry := false, 0.0
		for block := 0; ; block++ {
			left := start + float64(block)*coarseStepDays
			if left >= end {
				break
			}
			coarseGap := gap(left)
			if !inside && isFinite(coarseGap) && coarseGap > thresholdKM {
				continue
			}
			for sample := 0; sample < 20; sample++ {
				candidate := left + float64(sample)*fineStepDays
				if candidate >= end {
					break
				}
				value := gap(candidate)
				if value < 0 && !inside {
					inside, entry = true, candidate
				}
				if value >= 0 && inside {
					intervals = append(intervals, [2]float64{entry, candidate})
					inside = false
				}
			}
		}
		if inside {
			intervals = append(intervals, [2]float64{entry, end})
		}
		events := collectMutualEclipsesBySearch(start, end, pair[0], pair[1])
		matched := make([]bool, len(events))
		maximumStartSeconds := 0.0
		for _, interval := range intervals {
			found := -1
			for index, event := range events {
				emission := jupiterGalileanEclipseEmission(event.Start, event.Passive, nil)
				difference := math.Abs(emission - interval[0])
				if !matched[index] && difference <= 1.5*fineStepDays {
					found = index
					maximumStartSeconds = math.Max(maximumStartSeconds, difference*86400)
					break
				}
			}
			if found < 0 {
				t.Fatalf("independent interval missing from search pair=%v interval=%v", pair, interval)
			}
			matched[found] = true
		}
		searchOnly := 0
		for index, event := range events {
			if matched[index] {
				continue
			}
			startGeometry := jupiterGalileanEclipseEmission(event.Start, event.Passive, nil)
			endGeometry := jupiterGalileanEclipseEmission(event.End, event.Passive, nil)
			if endGeometry-startGeometry >= 2*fineStepDays {
				t.Fatalf("search-only event exceeds two fine steps pair=%v event=%+v", pair, event)
			}
			searchOnly++
			t.Logf("search-only pair=%v active=%d passive=%d greatest=%.12f duration=%.9f s", pair, event.Active, event.Passive, event.Greatest, (endGeometry-startGeometry)*86400)
		}
		totalIntervals += len(intervals)
		totalEvents += len(events)
		totalSearchOnly += searchOnly
		t.Logf("pair=%v independent intervals=%d search events=%d search-only=%d maximum Start difference=%.9f s", pair, len(intervals), len(events), searchOnly, maximumStartSeconds)
	}
	if totalEvents != 214 {
		t.Fatalf("independent scan search count=%d want=214", totalEvents)
	}
	t.Logf("independent intervals=%d search events=%d search-only=%d", totalIntervals, totalEvents, totalSearchOnly)
}
func TestJupiterGalileanMutualEclipseSyntheticFrame(t *testing.T) {
	for _, jde := range []float64{JDCalc(2026, 1, 1), JDCalc(2027, 7, 1)} {
		frame := newJupiterGalileanEclipseFrame(jde, nil)
		wantHelio := rotateEclipticToEquatorial(jupiterHeliocentricVectorJ2000(jde), orbitJ2000Obliquity)
		poleRA, poleDec, _ := jupiterPoleRotation(jde)
		if frame.jupiterHelio != wantHelio || frame.pole != raDecToVector(poleRA, poleDec) {
			t.Fatalf("geometric frame at %.9f differs from direct evaluation", jde)
		}
		for satellite := 1; satellite <= 4; satellite++ {
			state := jupiterGalileanSatelliteCorrectedStateAtET(jde-jupiterGalileanReferenceJD, satellite-1, jupiterGalileanUseSolarLongPeriod(jde))
			want := Vector3{wantHelio[0] + state.X, wantHelio[1] + state.Y, wantHelio[2] + state.Z}
			if frame.satellite(jde, satellite) != want {
				t.Fatalf("satellite %d geometric frame mismatch", satellite)
			}
		}
	}
}

func TestJupiterGalileanMutualEclipseSyntheticCone(t *testing.T) {
	for _, distance := range []float64{0, 1e6} {
		umbra, penumbra := jupiterGalileanEclipseConeRadii(distance, 7e8, 1829.4)
		wantUmbra, wantPenumbra := 1829.4, 1829.4
		if distance > 0 {
			wantUmbra, wantPenumbra = 838.1562857142859, 2825.8705714285716
		}
		if math.Abs(umbra-wantUmbra) > 1e-9 || math.Abs(penumbra-wantPenumbra) > 1e-9 {
			t.Fatalf("cone at %.1f km: %.12f/%.12f want %.12f/%.12f", distance, umbra, penumbra, wantUmbra, wantPenumbra)
		}
	}
	for _, test := range []struct {
		name                        string
		umbra, separation, coverage float64
		annular                     bool
	}{
		{name: "positive total", umbra: 3, separation: 0, coverage: 1},
		{name: "positive annular", umbra: 1, separation: 0, coverage: 0.25, annular: true},
		{name: "positive partial", umbra: 1, separation: 2, coverage: 0.111652479681166},
		{name: "zero", umbra: 0, separation: 0},
		{name: "negative", umbra: -1, separation: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			coverage, annular := jupiterGalileanEclipseDiskMetrics(test.separation, test.umbra, 2)
			if math.Abs(coverage-test.coverage) > 1e-12 || annular != test.annular {
				t.Fatalf("metrics %.16g/%v want %.16g/%v", coverage, annular, test.coverage, test.annular)
			}
		})
	}
	if _, ok := jupiterGalileanEclipseProjection(Vector3{5, 0, 0}, Vector3{4.99, 0, 0}, 1, 2); ok {
		t.Fatal("a satellite behind the passive satellite casts a forward shadow")
	}
}

func TestJupiterGalileanMutualEclipseConsistencyGeometryTime(t *testing.T) {
	const geometryJD = 2461171.6
	for passive := 1; passive <= 4; passive++ {
		reception := jupiterGalileanEclipseReception(geometryJD, passive, nil)
		emission := jupiterGalileanEclipseEmission(reception, passive, nil)
		if reception <= geometryJD || math.Abs(emission-geometryJD)*86400 > 0.001 {
			t.Fatalf("satellite %d light-time round trip %.9fs", passive, (emission-geometryJD)*86400)
		}
	}
	geometry, ok := jupiterGalileanMutualEclipseGeometryAt(geometryJD, 1, 2, nil)
	reversed, reversedOK := jupiterGalileanMutualEclipseGeometryAt(geometryJD, 2, 1, nil)
	if !ok || !reversedOK || geometry != reversed || geometry.PenumbraKM < geometry.UmbraKM {
		t.Fatalf("invalid pair geometry: %+v / %+v", geometry, reversed)
	}
}

func TestJupiterGalileanMutualDiskMetricsSynthetic(t *testing.T) {
	tests := []struct {
		name          string
		separation    float64
		activeRadius  float64
		passiveRadius float64
		impact        float64
		coverage      float64
		annular       bool
	}{
		{
			name:          "separated",
			separation:    3,
			activeRadius:  1,
			passiveRadius: 1,
			impact:        1.5,
			coverage:      0,
		},
		{
			name:          "external tangent",
			separation:    2,
			activeRadius:  1,
			passiveRadius: 1,
			impact:        1,
			coverage:      0,
		},
		{
			name:          "intersecting",
			separation:    1,
			activeRadius:  1,
			passiveRadius: 1,
			impact:        0.5,
			coverage:      0.3910022189557706,
		},
		{
			name:          "contained",
			separation:    0.5,
			activeRadius:  2,
			passiveRadius: 1,
			impact:        1.0 / 6.0,
			coverage:      1,
		},
		{
			name:          "coincident",
			separation:    0,
			activeRadius:  1,
			passiveRadius: 1,
			impact:        0,
			coverage:      1,
			annular:       true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			impact, coverage, annular := jupiterGalileanDiskMetrics(test.separation, test.activeRadius, test.passiveRadius)
			if impact != test.impact {
				t.Fatalf("impact = %.16g, want %.16g", impact, test.impact)
			}
			if math.Abs(coverage-test.coverage) > 1e-15 {
				t.Fatalf("coverage = %.16g, want %.16g", coverage, test.coverage)
			}
			if annular != test.annular {
				t.Fatalf("annular = %v, want %v", annular, test.annular)
			}
		})
	}
}

func TestJupiterGalileanMutualDiskMetricsUnequalPartialOverlap(t *testing.T) {
	tests := []struct {
		name          string
		separation    float64
		activeRadius  float64
		passiveRadius float64
		coverage      float64
	}{
		{name: "larger foreground disk", separation: 2, activeRadius: 2, passiveRadius: 1, coverage: 0.4466099187246639},
		{name: "smaller foreground disk", separation: 2, activeRadius: 1, passiveRadius: 2, coverage: 0.111652479681166},
		{name: "Ganymede occulting Europa", separation: 1154.1996709941025, activeRadius: 2631.2, passiveRadius: 1562.6, coverage: 0.988377027615301},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, coverage, _ := jupiterGalileanDiskMetrics(test.separation, test.activeRadius, test.passiveRadius)
			if math.Abs(coverage-test.coverage) > 1e-12 {
				t.Fatalf("coverage = %.16g, want %.16g", coverage, test.coverage)
			}

			area := jupiterGalileanDiskOverlapArea(test.separation, test.activeRadius, test.passiveRadius)
			reversedArea := jupiterGalileanDiskOverlapArea(test.separation, test.passiveRadius, test.activeRadius)
			areaTolerance := math.Max(1, area) * 1e-12
			if math.Abs(area-reversedArea) > areaTolerance {
				t.Fatalf("overlap area is not symmetric: %.16g vs %.16g", area, reversedArea)
			}
			if area < 0 || area > math.Pi*math.Pow(math.Min(test.activeRadius, test.passiveRadius), 2)+areaTolerance {
				t.Fatalf("overlap area is outside physical bounds: %.16g", area)
			}
		})
	}
}

func TestJupiterGalileanMutualGeometryConsistencyWithSatelliteObservations(t *testing.T) {
	tests := []struct {
		name   string
		jde    float64
		first  int
		second int
	}{
		{name: "j2000 io europa", jde: 2451545.0, first: 1, second: 2},
		{name: "modern europa callisto", jde: 2459001.5, first: 2, second: 4},
		{name: "future ganymede callisto", jde: 2461000.5, first: 3, second: 4},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			geometry, ok := jupiterGalileanMutualGeometryAt(test.jde, test.first, test.second)
			if !ok {
				t.Fatal("mutual geometry is invalid")
			}
			observations := JupiterGalileanSatelliteObservations(test.jde)
			active := test.first
			passive := test.second
			if observations[active-1].OffsetZJupiterRadii > observations[passive-1].OffsetZJupiterRadii {
				active, passive = passive, active
			}
			if geometry.Active != active || geometry.Passive != passive {
				t.Fatalf("active/passive = %d/%d, want %d/%d", geometry.Active, geometry.Passive, active, passive)
			}
			activeObservation := observations[active-1]
			passiveObservation := observations[passive-1]
			if geometry.ActiveX != activeObservation.OffsetXJupiterRadii || geometry.ActiveY != activeObservation.OffsetYJupiterRadii || geometry.ActiveZ != activeObservation.OffsetZJupiterRadii {
				t.Fatalf("active projection = (%.16g, %.16g, %.16g), want (%.16g, %.16g, %.16g)", geometry.ActiveX, geometry.ActiveY, geometry.ActiveZ, activeObservation.OffsetXJupiterRadii, activeObservation.OffsetYJupiterRadii, activeObservation.OffsetZJupiterRadii)
			}
			if geometry.PassiveX != passiveObservation.OffsetXJupiterRadii || geometry.PassiveY != passiveObservation.OffsetYJupiterRadii || geometry.PassiveZ != passiveObservation.OffsetZJupiterRadii {
				t.Fatalf("passive projection = (%.16g, %.16g, %.16g), want (%.16g, %.16g, %.16g)", geometry.PassiveX, geometry.PassiveY, geometry.PassiveZ, passiveObservation.OffsetXJupiterRadii, passiveObservation.OffsetYJupiterRadii, passiveObservation.OffsetZJupiterRadii)
			}
			reversed, reversedOK := jupiterGalileanMutualGeometryAt(test.jde, test.second, test.first)
			if !reversedOK || reversed != geometry {
				t.Fatalf("reversed pair geometry differs: %#v vs %#v", reversed, geometry)
			}
		})
	}
}

func TestJupiterGalileanMutualSeasonGateSynthetic(t *testing.T) {
	tests := []struct {
		name string
		jde  float64
		want bool
	}{
		{name: "2015 season", jde: JDCalc(2015, 1, 1.5), want: true},
		{name: "2021 season", jde: JDCalc(2021, 1, 1.5), want: true},
		{name: "2026 season", jde: JDCalc(2026, 1, 1.5), want: true},
		{name: "2018 outside season", jde: JDCalc(2018, 1, 1.5), want: false},
		{name: "2024 outside season", jde: JDCalc(2024, 1, 1.5), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := jupiterGalileanMutualSeasonAllows(test.jde, 1, 2); got != test.want {
				t.Fatalf("season gate = %v, want %v", got, test.want)
			}
		})
	}
}

func TestJupiterGalileanMutualSeasonMeasurementsSynthetic(t *testing.T) {
	const step = 0.1
	start := JDCalc(2000, 1, 1.5)
	end := JDCalc(2040, 1, 1.5)

	semiMajorAxes := [4][]float64{}
	inclinations := [4]float64{}
	for jd := start; jd <= end+step/2; jd += step {
		poleRA, poleDec, _ := jupiterPoleRotation(jd)
		pole := raDecToVector(poleRA, poleDec)
		states := JupiterGalileanSatelliteStates(jd)
		for satellite, state := range states {
			rho := math.Sqrt(state.X*state.X + state.Y*state.Y + state.Z*state.Z)
			if rho == 0 {
				t.Fatalf("satellite %d has zero radius at JDE %.1f", satellite+1, jd)
			}
			semiMajorAxes[satellite] = append(semiMajorAxes[satellite], rho/(jupiterGalileanEquatorialRadiusKM/astronomicalUnitKM))
			sine := math.Abs((state.X*pole[0] + state.Y*pole[1] + state.Z*pole[2]) / rho)
			if sine > 1 {
				sine = 1
			}
			inclination := math.Asin(sine) * deg
			if inclination > inclinations[satellite] {
				inclinations[satellite] = inclination
			}
		}
	}

	var medians [4]float64
	for satellite := range semiMajorAxes {
		sort.Float64s(semiMajorAxes[satellite])
		medians[satellite] = semiMajorAxes[satellite][len(semiMajorAxes[satellite])/2]
	}
	t.Logf("satellite a medians (Jupiter radii): Io=%.12f Europa=%.12f Ganymede=%.12f Callisto=%.12f", medians[0], medians[1], medians[2], medians[3])
	t.Logf("satellite maximum inclinations (deg): Io=%.12f Europa=%.12f Ganymede=%.12f Callisto=%.12f", inclinations[0], inclinations[1], inclinations[2], inclinations[3])
	for first := 0; first < 4; first++ {
		for second := first + 1; second < 4; second++ {
			sine := 1.2 * (jupiterGalileanSatelliteRadiusJupiterRadii(first+1) + jupiterGalileanSatelliteRadiusJupiterRadii(second+1) + medians[first]*math.Sin(inclinations[first]*rad) + medians[second]*math.Sin(inclinations[second]*rad)) / math.Abs(medians[second]-medians[first])
			if sine < 1 {
				sine = math.Asin(sine) * deg
			} else {
				sine = 90
			}
			t.Logf("DEmax %d-%d = %.12f deg", first+1, second+1, sine)
		}
	}
}

func TestJupiterGalileanMutualSeasonGateConsistency(t *testing.T) {
	maxDifference := 0.0
	for jd := JDCalc(2000, 1, 1.5); jd <= JDCalc(2040, 1, 1.5); jd++ {
		iauDeclination := jupiterGalileanMutualEarthDeclination(jd)
		_, meeusDeclination := JupiterDSDE(jd)
		difference := math.Abs(iauDeclination - meeusDeclination)
		if difference > maxDifference {
			maxDifference = difference
		}
	}
	t.Logf("IAU-pole DE vs JupiterDSDE maximum difference: %.12f deg", maxDifference)
}
