package basic

import "testing"

// TestMoonStateMatchesHMoonHeight 固定 MoonState 与 HMoonHeight 逐位同口径。
func TestMoonStateMatchesHMoonHeight(t *testing.T) {
	for _, jd := range []float64{2462502.5, 2416745.5, 2469807.75, 2378496.25} {
		state := MoonStateAt(jd)
		for lon := -180.0; lon < 180; lon += 23.5 {
			for lat := -89.5; lat <= 89.5; lat += 7.5 {
				if got, want := state.HMoonHeight(lon, lat), HMoonHeight(jd, lon, lat, 0); got != want {
					t.Fatalf("jd=%v lon=%v lat=%v got %v want %v", jd, lon, lat, got, want)
				}
			}
		}
	}
}
