package basic

import "testing"

func TestSolarEclipse43290612PolarAnnularCentralBandCloses(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		JDCalc(4329, 6, 12),
		SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440.0, BoundaryPoints: 96},
	)
	if result.Eclipse.Type != SolarEclipseAnnular || result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
		t.Fatalf("type=%s centrality=%s, want two-limit annular eclipse", result.Eclipse.Type, result.Eclipse.Centrality)
	}
	if len(result.CentralBandHorizonClosures) != 2 {
		t.Fatalf("horizon closures=%d, want two", len(result.CentralBandHorizonClosures))
	}
	if len(result.CentralBandSegments) != 1 {
		t.Fatalf("central-band segments=%d, want one closed envelope", len(result.CentralBandSegments))
	}
	ring := result.CentralBandSegments[0]
	if len(ring) < 4 || solarEclipsePathDistanceKM(ring[0], ring[len(ring)-1]) > 0.01 {
		t.Fatal("central-band envelope is not closed")
	}
}
