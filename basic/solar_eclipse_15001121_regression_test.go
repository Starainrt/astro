package basic

import "testing"

func TestSolarEclipse15001121ShallowTotalBandCloses(t *testing.T) {
	result := SolarEclipsePartialFootprints(
		JDCalc(1500, 11, 21),
		SolarEclipsePartialFootprintOptions{StepDays: 2.0 / 1440.0, BoundaryPoints: 96},
	)
	if result.Eclipse.Type != SolarEclipseTotal || result.Eclipse.Centrality != SolarEclipseCentralTwoLimits {
		t.Fatalf("type=%s centrality=%s, want two-limit total eclipse", result.Eclipse.Type, result.Eclipse.Centrality)
	}
	if len(result.CentralBandSegments) != 1 {
		t.Logf("footprints=%d closures=%d", len(result.CentralBandFootprints), len(result.CentralBandHorizonClosures))
		t.Fatalf("central-band segments=%d, want one closed envelope", len(result.CentralBandSegments))
	}
	if len(result.CentralBandHorizonClosures) != 2 {
		t.Fatalf("horizon closures=%d, want two", len(result.CentralBandHorizonClosures))
	}
}
