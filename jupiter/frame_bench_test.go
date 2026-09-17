package jupiter

import (
	"testing"
	"time"
)

func BenchmarkSatellites(b *testing.B) {
	date := time.Date(2026, 4, 2, 18, 0, 0, 0, time.UTC)
	var sink GalileanSatellitesInfo
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = Satellites(date)
	}
	_ = sink
}

func BenchmarkSatellitePhenomena(b *testing.B) {
	date := time.Date(2026, 4, 2, 18, 0, 0, 0, time.UTC)
	var sink GalileanPhenomenaInfo
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = SatellitePhenomena(date)
	}
	_ = sink
}

func BenchmarkGalileanPhenomenonEventLast(b *testing.B) {
	date := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	var sink GalileanPhenomenonEvent
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = LastGalileanPhenomenonEvent(date, GalileanSatelliteIo, GalileanPhenomenonTransit)
	}
	_ = sink
}

func BenchmarkGalileanPhenomenonEventNext(b *testing.B) {
	date := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	var sink GalileanPhenomenonEvent
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = NextGalileanPhenomenonEvent(date, GalileanSatelliteIo, GalileanPhenomenonTransit)
	}
	_ = sink
}

func BenchmarkGalileanPhenomenonEventClosest(b *testing.B) {
	date := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	var sink GalileanPhenomenonEvent
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = ClosestGalileanPhenomenonEvent(date, GalileanSatelliteIo, GalileanPhenomenonTransit)
	}
	_ = sink
}

func BenchmarkGalileanPhenomenonContactEventClosest(b *testing.B) {
	date := time.Date(2026, 4, 2, 20, 0, 0, 0, time.UTC)
	var sink GalileanPhenomenonContactEvent
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = ClosestGalileanPhenomenonContactEvent(date, GalileanSatelliteIo, GalileanPhenomenonTransit)
	}
	_ = sink
}
