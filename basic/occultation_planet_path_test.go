package basic

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestPlanetOccultationFiniteDiskExpandsOuterAndContractsTotalPath(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2024, time.August, 21, 2, 41, 36, 0, time.UTC))
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
	}
	_, _, finiteWidth, finiteOK := occultationPathLimitsAndWidthForFrame(tt, frameAt)
	if !finiteOK {
		t.Fatal("finite-disk path limits are unavailable")
	}
	pointFrameAt := func(tt float64) (occultationPathFrame, bool) {
		frame, valid := planetOccultationPathFrameAt(tt, config)
		frame.targetRadius = 0
		return frame, valid
	}
	_, _, pointWidth, pointOK := occultationPathLimitsAndWidthForFrame(tt, pointFrameAt)
	if !pointOK {
		t.Fatal("point-source comparison limits are unavailable")
	}
	if finiteWidth <= pointWidth {
		t.Fatalf("finite-disk outer width = %.6f km, want greater than point-source width %.6f km", finiteWidth, pointWidth)
	}
	if finiteWidth-pointWidth < 1 {
		t.Fatalf("finite-disk expansion = %.6f km, want a measurable planetary-radius contribution", finiteWidth-pointWidth)
	}
	innerFrameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationTotalPathFrameAt(tt, config)
	}
	_, _, totalWidth, totalOK := occultationPathLimitsAndWidthForFrame(tt, innerFrameAt)
	if !totalOK {
		t.Fatal("finite-disk total-occultation limits are unavailable")
	}
	if totalWidth >= pointWidth {
		t.Fatalf("finite-disk total width = %.6f km, want less than point-source width %.6f km", totalWidth, pointWidth)
	}
	if pointWidth-totalWidth < 1 {
		t.Fatalf("finite-disk contraction = %.6f km, want a measurable planetary-radius contribution", pointWidth-totalWidth)
	}
}

func TestPlanetOccultationConesUseTwoSphereCommonTangents(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	outer, ok := planetOccultationPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn outer-contact cone is unavailable")
	}
	inner, ok := planetOccultationTotalPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn inner-contact cone is unavailable")
	}

	planetRA, planetDec := config.apparentRaDecN(tt, -1)
	planetDistance := config.earthDistanceN(tt, -1) * occultationPathAstronomicalUnitKM
	target := occultationPathRaDecVector(planetRA, planetDec, planetDistance)
	moonToTargetDistance := occultationPathNorm(occultationPathSub(target, outer.moon))
	moonRadiusKM := occultationPathNorm(outer.moon) * math.Sin(outer.moonRadius)
	wantOuter := math.Asin((moonRadiusKM + config.equatorialRadiusKM) / moonToTargetDistance)
	wantInner := math.Asin((moonRadiusKM - config.equatorialRadiusKM) / moonToTargetDistance)
	if difference := math.Abs(outer.targetRadius - wantOuter); difference > 1e-15 {
		t.Fatalf("outer-contact cone angle = %.15g rad, want %.15g (difference %.3g)", outer.targetRadius, wantOuter, difference)
	}
	if difference := math.Abs(inner.targetRadius - wantInner); difference > 1e-15 {
		t.Fatalf("inner-contact cone angle = %.15g rad, want %.15g (difference %.3g)", inner.targetRadius, wantInner, difference)
	}
	for _, contact := range []struct {
		name  string
		frame occultationPathFrame
	}{
		{name: "outer", frame: outer},
		{name: "inner", frame: inner},
	} {
		origin, direction, rayOK := occultationPathBoundaryRay(contact.frame, 0.73)
		if !rayOK {
			t.Fatalf("%s-contact boundary ray is unavailable", contact.name)
		}
		moonNormal := occultationPathSub(origin, contact.frame.moon)
		if difference := math.Abs(occultationPathNorm(moonNormal) - moonRadiusKM); difference > 1e-6 {
			t.Fatalf("%s-contact lunar tangency radius differs by %.9f km", contact.name, difference)
		}
		if residual := math.Abs(occultationPathDot(moonNormal, direction)); residual > 1e-6 {
			t.Fatalf("%s-contact ray/lunar-radius dot product = %.9f km", contact.name, residual)
		}
		targetParameter := occultationPathDot(occultationPathSub(target, origin), direction)
		targetTangent := occultationPathAdd(origin, occultationPathScale(direction, targetParameter))
		targetNormal := occultationPathSub(targetTangent, target)
		if difference := math.Abs(occultationPathNorm(targetNormal) - config.equatorialRadiusKM); difference > 1e-5 {
			t.Fatalf("%s-contact planetary tangency radius differs by %.9f km", contact.name, difference)
		}
		if residual := math.Abs(occultationPathDot(targetNormal, direction)); residual > 1e-5 {
			t.Fatalf("%s-contact ray/planet-radius dot product = %.9f km", contact.name, residual)
		}
	}
}

func TestPlanetOccultationInnerConeUsesSignedTargetRadius(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	tt := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	frame, ok := planetOccultationTotalPathFrameAt(tt, config)
	if !ok {
		t.Fatal("Saturn inner-contact cone is unavailable")
	}

	for index := 0; index < occultationPathBoundaryScanPoints; index++ {
		theta := 2 * math.Pi * float64(index) / float64(occultationPathBoundaryScanPoints)
		want, _, wantOK := occultationPathBoundaryVector(frame, theta)
		if !wantOK {
			continue
		}
		discriminant, _, scale, lineOK := occultationPathBoundaryLine(frame, theta)
		if !lineOK || discriminant < 0 {
			continue
		}
		got, _, gotOK := occultationPathBoundaryIntersection(frame, theta, 1e-12*math.Max(scale, 1))
		if !gotOK {
			t.Fatalf("signed inner-cone intersection is unavailable at theta %.9f", theta)
		}
		if difference := occultationPathNorm(occultationPathSub(got, want)); difference > 1e-6 {
			t.Fatalf("inner-cone intersection differs by %.6f km at theta %.9f", difference, theta)
		}
		return
	}
	t.Fatal("no comparable Saturn inner-cone boundary point found")
}

func TestOccultationPathBoundaryTangentFindsBetweenSamples(t *testing.T) {
	const boundaryRadiusKM = 1737.4
	theta := math.Pi / float64(occultationPathBoundaryScanPoints)
	offset := occultationPathEarthEquatorialRadiusKM + boundaryRadiusKM - 0.01
	moon := occultationPathVector{
		x: 384000,
		y: -offset * math.Cos(theta),
		z: -offset * math.Sin(theta),
	}
	frame := occultationPathFrame{
		moon:       moon,
		axis:       occultationPathVector{x: -1},
		first:      occultationPathVector{y: 1},
		second:     occultationPathVector{z: 1},
		moonRadius: math.Asin(boundaryRadiusKM / occultationPathNorm(moon)),
	}
	for _, sampledTheta := range []float64{0, 2 * math.Pi / float64(occultationPathBoundaryScanPoints)} {
		if _, _, ok := occultationPathBoundaryVector(frame, sampledTheta); ok {
			t.Fatalf("fixture is not narrower than the old sample spacing at theta %.9f", sampledTheta)
		}
	}
	point, tangentTheta, ok := occultationPathBoundaryTangent(frame)
	if !ok {
		t.Fatal("continuous boundary tangency was not found between scan points")
	}
	if math.Abs(tangentTheta-theta) > 5e-5 {
		t.Fatalf("tangent theta = %.9f, want %.9f", tangentTheta, theta)
	}
	polarRatioSquared := occultationPathEarthPolarRatio * occultationPathEarthPolarRatio
	ellipsoidResidual := point.x*point.x + point.y*point.y + point.z*point.z/polarRatioSquared -
		occultationPathEarthEquatorialRadiusKM*occultationPathEarthEquatorialRadiusKM
	if math.Abs(ellipsoidResidual) > 1e-3 {
		t.Fatalf("tangent point ellipsoid residual = %.9f", ellipsoidResidual)
	}
	frameAt := func(float64) (occultationPathFrame, bool) { return frame, true }
	if _, _, centerOK := occultationEarthLineIntersection(frame.moon, frame.axis); centerOK {
		t.Fatal("synthetic center line unexpectedly intersects Earth")
	}
	north, south, width, limitsOK := occultationPathLimitsAndWidthForFrame(2451545, frameAt)
	if !limitsOK {
		t.Fatal("boundary-only event did not produce path limits")
	}
	if separation := occultationPathNorm(occultationPathSub(north, south)); separation <= 1e-6 {
		t.Fatalf("boundary-only path limits collapsed to one point: separation=%.12f km", separation)
	}
	if width <= 0 {
		t.Fatalf("boundary-only path width = %.12f km, want positive", width)
	}
	greatest, greatestOK := occultationPathBoundaryPointForFrame(2451545, frameAt, time.UTC)
	if !greatestOK {
		t.Fatal("boundary-only event did not produce a greatest surface point")
	}
	if greatest.WidthKM <= 0 {
		t.Fatalf("boundary-only greatest width = %.12f km, want positive", greatest.WidthKM)
	}
}

func TestPlanetOccultationSaturnLimitsRemainContinuous(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	paths, err := FindPlanetOccultationPaths(
		time.Date(2025, time.February, 1, 0, 0, 0, 0, location),
		time.Date(2025, time.February, 2, 0, 0, 0, 0, location),
		OccultationSaturn,
		OccultationPathOptions{Step: 2 * time.Minute},
	)
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths() error = %v", err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() returned %d paths, want 1", len(paths))
	}

	for _, limit := range []struct {
		name   string
		points []OccultationPathPoint
	}{
		{name: "outer northern", points: paths[0].NorthernLimit},
		{name: "outer southern", points: paths[0].SouthernLimit},
		{name: "total northern", points: paths[0].NorthernTotalLimit},
		{name: "total southern", points: paths[0].SouthernTotalLimit},
	} {
		for index := 1; index < len(limit.points); index++ {
			distance := occultationPathDistanceKM(limit.points[index-1], limit.points[index])
			if distance > 1000 {
				t.Fatalf("%s limit jumps %.1f km between %v and %v", limit.name, distance,
					limit.points[index-1].Time, limit.points[index].Time)
			}
		}
	}
}

func TestRefinedPlanetOccultationCenterLineRespectsWidthTolerance(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	paths, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: 5 * time.Minute, TargetSpacingKM: 50},
	)
	if err != nil || len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths() paths=%d err=%v, want one", len(paths), err)
	}
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	frameAt := func(tt float64) (occultationPathFrame, bool) {
		return planetOccultationPathFrameAt(tt, config)
	}
	for index, point := range paths[0].CenterLine {
		exact, pointOK := occultationPathCenterPointForFrame(centerTimeTT(point.Time), frameAt, time.UTC)
		if !pointOK {
			t.Fatalf("exact center point %d is unavailable", index)
		}
		if difference := math.Abs(point.WidthKM - exact.WidthKM); difference > occultationPathWidthToleranceKM {
			t.Fatalf("center point %d width differs from exact value by %.9f km: got %.9f want %.9f",
				index, difference, point.WidthKM, exact.WidthKM)
		}
	}
}

func TestPlanetOccultationSaturnLimitsDoNotDependOnStep(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*3600)
	start := time.Date(2024, time.August, 21, 0, 0, 0, 0, location)
	end := time.Date(2024, time.August, 22, 0, 0, 0, 0, location)
	fine := findSinglePlanetOccultationPath(t, start, end, 30*time.Second)
	coarse := findSinglePlanetOccultationPath(t, start, end, 2*time.Minute)

	for _, limits := range []struct {
		name         string
		fine, coarse []OccultationPathPoint
	}{
		{name: "outer northern", fine: fine.NorthernLimit, coarse: coarse.NorthernLimit},
		{name: "outer southern", fine: fine.SouthernLimit, coarse: coarse.SouthernLimit},
		{name: "total northern", fine: fine.NorthernTotalLimit, coarse: coarse.NorthernTotalLimit},
		{name: "total southern", fine: fine.SouthernTotalLimit, coarse: coarse.SouthernTotalLimit},
	} {
		assertOccultationPathCommonSamplesEqual(t, limits.name, limits.fine, limits.coarse)
		for index := 1; index+1 < len(limits.coarse); index++ {
			paired := coarse.SouthernLimit
			if strings.HasPrefix(limits.name, "total") {
				paired = coarse.SouthernTotalLimit
			}
			if strings.HasSuffix(limits.name, "southern") {
				continue
			}
			if distance := occultationPathDistanceKM(limits.coarse[index], paired[index]); distance < 0.001 {
				t.Fatalf("%s and southern limit collapse at %v", limits.name, limits.coarse[index].Time)
			}
		}
	}
}

func findSinglePlanetOccultationPath(t *testing.T, start, end time.Time, step time.Duration) PlanetOccultationPath {
	t.Helper()
	paths, err := FindPlanetOccultationPaths(start, end, OccultationSaturn, OccultationPathOptions{Step: step})
	if err != nil {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) error = %v", step, err)
	}
	if len(paths) != 1 {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) returned %d paths, want 1", step, len(paths))
	}
	if !paths[0].HasTotalBand {
		t.Fatalf("FindPlanetOccultationPaths(step=%v) has no total band", step)
	}
	return paths[0]
}

func assertOccultationPathCommonSamplesEqual(t *testing.T, name string, fine, coarse []OccultationPathPoint) {
	t.Helper()
	matched := 0
	fineIndex := 0
	for _, coarsePoint := range coarse[1 : len(coarse)-1] {
		for fineIndex+1 < len(fine) && fine[fineIndex].Time.Before(coarsePoint.Time.Add(-20*time.Millisecond)) {
			fineIndex++
		}
		nearest := -1
		nearestDelta := math.Inf(1)
		for candidateIndex := fineIndex - 2; candidateIndex <= fineIndex+2; candidateIndex++ {
			if candidateIndex < 0 || candidateIndex >= len(fine) {
				continue
			}
			delta := math.Abs(fine[candidateIndex].Time.Sub(coarsePoint.Time).Seconds())
			if delta < nearestDelta {
				nearest = candidateIndex
				nearestDelta = delta
			}
		}
		if nearest < 0 || nearestDelta > 0.00001 {
			continue
		}
		matched++
		if distance := occultationPathDistanceKM(fine[nearest], coarsePoint); distance > 5 {
			t.Fatalf("%s differs by %.1f km at common time %v (sample delta %.6f s)",
				name, distance, coarsePoint.Time, nearestDelta)
		}
	}
	if matched < 10 {
		t.Fatalf("%s compared only %d common samples, want at least 10", name, matched)
	}
}
