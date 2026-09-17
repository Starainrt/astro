package svgmap

import (
	"strings"
	"testing"
)

func BenchmarkWriteLandOrthographic(b *testing.B) {
	frame := Frame{
		X: 57, Y: 184, Width: 360, Height: 360,
		Projection:      ProjectionOrthographic,
		CenterLongitude: 104.145, CenterLatitude: -1.5,
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		var builder strings.Builder
		frame.WriteLand(&builder, "clip")
	}
}

func BenchmarkWriteLandCenteredEquirectangular(b *testing.B) {
	frame := Frame{
		X: 57, Y: 184, Width: 360, Height: 180,
		Projection: ProjectionEquirectangular, CenterLongitude: -104.145,
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		var builder strings.Builder
		frame.WriteLand(&builder, "clip")
	}
}
