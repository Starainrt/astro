package basic

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestPlanetOccultationSaturnFootprintsContainCenterLine(t *testing.T) {
	config, ok := planetOccultationConfigFor(OccultationSaturn)
	if !ok {
		t.Fatal("Saturn occultation config is unavailable")
	}
	seedTT := occultationTimeToTT(time.Date(2025, time.February, 1, 4, 0, 48, 0, time.UTC))
	for _, contact := range []struct {
		name    string
		frameAt occultationPathFrameFunc
	}{
		{name: "partial", frameAt: func(tt float64) (occultationPathFrame, bool) {
			return planetOccultationPathFrameAt(tt, config)
		}},
		{name: "total", frameAt: func(tt float64) (occultationPathFrame, bool) {
			return planetOccultationTotalPathFrameAt(tt, config)
		}},
	} {
		startTT, endTT, found := occultationPathWindowForFrame(
			seedTT, seedTT-occultationPathSearchSpanDays, seedTT+occultationPathSearchSpanDays,
			contact.frameAt, true,
		)
		if !found {
			t.Fatalf("%s center interval is unavailable", contact.name)
		}
		checked := 0
		for tt := startTT + 2.0/1440.0; tt < endTT-2.0/1440.0; tt += 5.0 / 1440.0 {
			frame, frameOK := contact.frameAt(tt)
			center, _, centerOK := occultationEarthLineIntersection(frame.moon, frame.axis)
			footprint, footprintOK := planetOccultationFootprintAt(tt, contact.frameAt, time.UTC)
			if !frameOK || !centerOK || !footprintOK {
				t.Fatalf("%s center or footprint is unavailable at TT %.9f", contact.name, tt)
			}
			longitude, latitude := occultationPathGeodetic(tt, center)
			if !planetOccultationFootprintContains(footprint, longitude, latitude) {
				t.Fatalf("%s footprint does not contain center %.6f, %.6f at TT %.9f",
					contact.name, longitude, latitude, tt)
			}
			checked++
		}
		if checked < 10 {
			t.Fatalf("%s checked only %d center samples", contact.name, checked)
		}
	}
}

func TestPlanetOccultationPathRejectsExcessiveAggregateSampling(t *testing.T) {
	start := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	_, err := FindPlanetOccultationPaths(
		start, start.Add(24*time.Hour), OccultationSaturn,
		OccultationPathOptions{Step: time.Second},
	)
	if !errors.Is(err, ErrOccultationPathSamplingLimit) {
		t.Fatalf("FindPlanetOccultationPaths() error = %v, want ErrOccultationPathSamplingLimit", err)
	}
}

func TestPlanetOccultationFootprintsHaveIndependentSampleBudget(t *testing.T) {
	times := occultationPathSampleTimesWithLimit(
		0, 1, 0.500001, 1.0/86400.0, planetOccultationFootprintMaxSamples,
	)
	if len(times) > planetOccultationFootprintMaxSamples {
		t.Fatalf("footprint sample count = %d, maximum %d", len(times), planetOccultationFootprintMaxSamples)
	}
	foundGreatest := false
	for _, sample := range times {
		if sample == 0.500001 {
			foundGreatest = true
			break
		}
	}
	if !foundGreatest {
		t.Fatal("bounded footprint samples omitted greatest")
	}
}

func planetOccultationFootprintContains(
	footprint PlanetOccultationFootprint,
	longitude, latitude float64,
) bool {
	for _, polygon := range footprint.Polygons {
		inside := false
		for current, previous := 0, len(polygon)-1; current < len(polygon); previous, current = current, current+1 {
			currentX := math.Remainder(polygon[current].Longitude-longitude, 360)
			previousX := math.Remainder(polygon[previous].Longitude-longitude, 360)
			currentY := polygon[current].Latitude
			previousY := polygon[previous].Latitude
			if math.Abs(currentX-previousX) > 180 {
				if currentX < previousX {
					currentX += 360
				} else {
					previousX += 360
				}
			}
			if (currentY > latitude) == (previousY > latitude) {
				continue
			}
			intersectionX := previousX + (latitude-previousY)*(currentX-previousX)/(currentY-previousY)
			if intersectionX > 0 {
				inside = !inside
			}
		}
		if inside {
			return true
		}
	}
	return false
}
