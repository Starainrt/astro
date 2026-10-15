package basic

import (
	"fmt"
	"math"
	"testing"

	"github.com/starainrt/astro/internal/geodata"
)

func TestSolarEclipsePathTopologyAcrossSarosAnchors(t *testing.T) {
	if testing.Short() {
		t.Skip("slow regression")
	}
	for _, year := range []int{1526, 1600, 1700, 1800, 1900, 2000, 2100, 2200, 2300, 2400, 2526} {
		events := solarEclipseScanEvents(JDCalc(year, 1, 1), JDCalc(year+1, 1, 1))
		if len(events) == 0 {
			t.Fatalf("%d has no solar eclipse candidate", year)
		}
		eventJDE := events[0]
		name := JD2Date(eventJDE).Format("2006-01-02")
		t.Run(name, func(t *testing.T) {
			assertSolarEclipsePathTopology(t, eventJDE, 60.0/1440.0)
		})
	}
}

func assertSolarEclipsePathTopology(t *testing.T, eventJDE, stepDays float64) {
	t.Helper()
	issues := solarEclipsePathTopologyIssues(eventJDE, stepDays)
	if len(issues) > 0 {
		t.Fatalf("path topology issues: %s", issues[0])
	}
}

func solarEclipsePathTopologyIssues(eventJDE, stepDays float64) []string {
	result := SolarEclipsePartialFootprints(eventJDE, SolarEclipsePartialFootprintOptions{
		StepDays: stepDays, BoundaryPoints: 24, DisableRiseSetCurves: true,
	})
	issues := make([]string, 0)
	if !result.Eclipse.HasPartial {
		return []string{"partial eclipse result is missing"}
	}
	contacts := []SolarEclipsePathPoint{result.P1, result.P2, result.P3, result.P4}
	previous := 0.0
	for index, contact := range contacts {
		if contact.JDE == 0 {
			continue
		}
		if previous != 0 && contact.JDE <= previous {
			issues = append(issues, fmt.Sprintf("penumbral contacts not increasing at %d", index))
		}
		previous = contact.JDE
	}
	centralContacts := []SolarEclipsePathPoint{result.U1, result.U2, result.U3, result.U4}
	previous = 0
	for index, contact := range centralContacts {
		if contact.JDE == 0 {
			continue
		}
		if previous != 0 && contact.JDE <= previous {
			issues = append(issues, fmt.Sprintf("central contacts not increasing at %d", index))
		}
		previous = contact.JDE
	}
	for footprintIndex, footprint := range append(append([]SolarEclipsePartialFootprint(nil), result.Footprints...), result.CentralShadowFootprints...) {
		totalBoundaryPoints := 0
		for _, boundary := range footprint.Boundaries {
			totalBoundaryPoints += len(boundary)
		}
		for boundaryIndex, boundary := range footprint.Boundaries {
			if len(boundary) < 2 && (footprint.Closed || totalBoundaryPoints != 1) {
				issues = append(issues, fmt.Sprintf("footprint %d boundary %d is too short", footprintIndex, boundaryIndex))
				continue
			}
			for pointIndex, point := range boundary {
				if !finite(point.JDE) || !finite(point.Longitude) || !finite(point.Latitude) ||
					point.Longitude < -180 || point.Longitude > 180 || point.Latitude < -90 || point.Latitude > 90 {
					issues = append(issues, fmt.Sprintf("footprint %d boundary %d point %d is invalid", footprintIndex, boundaryIndex, pointIndex))
				}
				if pointIndex > 0 && solarEclipsePathDistanceKM(boundary[pointIndex-1], point) > 12000 {
					issues = append(issues, fmt.Sprintf("footprint %d boundary %d has a discontinuity", footprintIndex, boundaryIndex))
				}
			}
		}
		if footprint.Closed {
			if len(footprint.Boundaries) == 0 {
				issues = append(issues, fmt.Sprintf("closed footprint %d has no boundary", footprintIndex))
			}
		}
	}
	if result.Eclipse.HasCentral {
		// Limit derivation needs the normal one-minute tangent samples. A
		// coarse diagnostic footprint step can leave only two center points and
		// is intentionally not reused for the paired cross-sections.
		centralStepDays := solarEclipsePathDefaultStepDays
		path := SolarEclipseCentralPath(eventJDE, SolarEclipsePathOptions{StepDays: centralStepDays})
		if len(path.CenterLine) < 2 || len(path.NorthernLimit) < 2 || len(path.NorthernLimit) != len(path.SouthernLimit) {
			issues = append(issues, "central path or paired limits are incomplete")
		}
		for index := 1; index < len(path.CenterLine); index++ {
			if path.CenterLine[index].JDE <= path.CenterLine[index-1].JDE {
				issues = append(issues, "center-line times are not increasing")
				break
			}
		}
		for index := 1; index < len(path.NorthernLimit); index++ {
			if path.NorthernLimit[index].JDE <= path.NorthernLimit[index-1].JDE ||
				path.SouthernLimit[index].JDE <= path.SouthernLimit[index-1].JDE {
				issues = append(issues, "central-limit times are not increasing")
				break
			}
			insideTwoLimitPath := result.Eclipse.Centrality == SolarEclipseCentralTwoLimits &&
				index > 1 && index < len(path.NorthernLimit)-1
			northGap := solarEclipsePathDistanceKM(path.NorthernLimit[index-1], path.NorthernLimit[index]) > 500 &&
				!(math.Abs(path.NorthernLimit[index-1].Latitude) > 80 && math.Abs(path.NorthernLimit[index].Latitude) > 80)
			southGap := solarEclipsePathDistanceKM(path.SouthernLimit[index-1], path.SouthernLimit[index]) > 500 &&
				!(math.Abs(path.SouthernLimit[index-1].Latitude) > 80 && math.Abs(path.SouthernLimit[index].Latitude) > 80)
			if insideTwoLimitPath && (northGap || southGap) {
				issues = append(issues, "central-limit branch has a gap above 500 km")
				break
			}
		}
		if len(path.CenterLine) >= 2 {
			first := path.CenterLine[0]
			last := path.CenterLine[len(path.CenterLine)-1]
			if math.Abs(first.SunAltitude) > 0.02 || math.Abs(last.SunAltitude) > 0.02 {
				issues = append(issues, fmt.Sprintf(
					"center-line limits are off the horizon: first=%.6f last=%.6f",
					first.SunAltitude, last.SunAltitude,
				))
			}
			if !finite(first.WidthKM) || !finite(last.WidthKM) ||
				first.WidthKM < 0 || last.WidthKM < 0 || first.WidthKM > 5000 || last.WidthKM > 5000 {
				issues = append(issues, fmt.Sprintf(
					"center-line limit widths are invalid: first=%.3f last=%.3f",
					first.WidthKM, last.WidthKM,
				))
			}
		}
	} else if result.Eclipse.Type != SolarEclipsePartial && result.Eclipse.Centrality == SolarEclipseNonCentral {
		if len(result.CentralBandSegments) == 0 {
			if err := auditSolarEclipseOpenBandSweep(result.CentralBandFootprints); err != nil {
				issues = append(issues, "non-central eclipse has no usable central band: "+err.Error())
			}
		}
		for index, segment := range result.CentralBandSegments {
			if len(segment) < 4 || solarEclipsePathDistanceKM(segment[0], segment[len(segment)-1]) > 0.1 {
				issues = append(issues, fmt.Sprintf("non-central band segment %d is not closed", index))
			}
		}
		if len(result.CentralBandSegments) > 0 {
			if err := auditSolarEclipseBandContainsFootprints(result.CentralBandSegments, result.CentralBandFootprints); err != nil {
				issues = append(issues, "non-central band excludes an instantaneous central-shadow boundary: "+err.Error())
			}
		}
	}
	return issues
}

func auditSolarEclipseBandContainsFootprints(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
) error {
	if solarEclipseNonCentralBandContainsFootprints(segments, footprints) {
		return nil
	}
	miss := solarEclipseBandFootprintMissDistanceKM(segments, footprints)
	return fmt.Errorf("maximum boundary miss %.6f km exceeds %.3f km tolerance",
		miss, solarEclipseNonCentralBandContainmentToleranceKM)
}

func solarEclipseBandFootprintMissDistanceKM(
	segments [][]SolarEclipsePathPoint,
	footprints []SolarEclipsePartialFootprint,
) float64 {
	polygons := make([][]geodata.GeoPoint, 0, len(segments))
	for _, segment := range segments {
		polygon := make([]geodata.GeoPoint, len(segment))
		for index, point := range segment {
			polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		polygons = append(polygons, polygon)
	}
	paths := make([][]geodata.GeoPoint, 0, len(footprints))
	for _, footprint := range footprints {
		for _, boundary := range footprint.Boundaries {
			path := make([]geodata.GeoPoint, len(boundary))
			for index, point := range boundary {
				path[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			paths = append(paths, path)
		}
	}
	return geodata.SphericalPolygonsPathMissDistanceKM(polygons, paths, false)
}

func auditSolarEclipseOpenBandSweep(footprints []SolarEclipsePartialFootprint) error {
	samples := make([]geodata.OpenBoundarySweepSample, 0, len(footprints))
	for _, footprint := range footprints {
		boundaries := make([][]geodata.GeoPoint, 0, len(footprint.Boundaries))
		for _, boundary := range footprint.Boundaries {
			points := make([]geodata.GeoPoint, len(boundary))
			for index, point := range boundary {
				if !finite(point.Longitude) || !finite(point.Latitude) {
					return fmt.Errorf("footprint contains a non-finite point")
				}
				points[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			boundaries = append(boundaries, points)
		}
		samples = append(samples, geodata.OpenBoundarySweepSample{
			Boundaries: boundaries,
			Closed:     footprint.Closed,
		})
	}
	polygons, err := geodata.OpenBoundarySweep(samples)
	if err != nil {
		return err
	}
	if len(polygons) == 0 {
		return fmt.Errorf("open footprint sweep contains no polygon")
	}
	for polygonIndex, polygon := range polygons {
		if len(polygon) < 3 {
			return fmt.Errorf("open footprint sweep polygon %d has only %d points", polygonIndex, len(polygon))
		}
		for pointIndex, point := range polygon {
			if !finite(point.Longitude) || !finite(point.Latitude) ||
				point.Longitude < -180 || point.Longitude > 180 || point.Latitude < -90 || point.Latitude > 90 {
				return fmt.Errorf("open footprint sweep polygon %d point %d is invalid", polygonIndex, pointIndex)
			}
		}
	}
	return nil
}

func solarEclipseScanEvents(startJDE, endJDE float64) []float64 {
	seed := CalcMoonSHByJDE(startJDE, 0)
	if seed < startJDE {
		seed = CalcMoonSHByJDE(seed+25, 0)
	}
	var events []float64
	for seed < endJDE {
		result := SolarEclipse(seed)
		if result.HasPartial {
			events = append(events, result.GreatestEclipse)
		}
		next := CalcMoonSHByJDE(seed+25, 0)
		if !finite(next) || next <= seed+20 {
			break
		}
		seed = next
	}
	return events
}
