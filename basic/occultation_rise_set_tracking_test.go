package basic

import (
	"testing"
	"time"
)

func TestOccultationRiseSetNewRootKeepsExistingBranch(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, offset := range []float64{0, 160, -200} {
		point := func(lon float64, at time.Time) OccultationPathPoint {
			return OccultationPathPoint{Time: at, Longitude: normalizeLongitude(lon + offset), Latitude: 58}
		}
		tracks := []*occultationRiseSetTrack{{segments: [][]OccultationPathPoint{{point(20, start)}}}}
		tracks = appendOccultationRiseSetSamples(tracks, []OccultationPathPoint{
			point(14, start.Add(time.Minute)), point(19, start.Add(time.Minute)),
		}, 1.0/1440)
		if len(tracks) != 2 || len(tracks[0].segments[0]) != 2 || len(tracks[1].segments[0]) != 1 {
			t.Fatalf("offset=%g: new root did not get its own track", offset)
		}
		if got := tracks[0].segments[0][1].Longitude; got != normalizeLongitude(19+offset) {
			t.Errorf("offset=%g: existing track jumped to longitude %g", offset, got)
		}
	}
}
