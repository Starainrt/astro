package geojson_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/geojson"
)

// 本文件钉住导出路径成本：掠食带的限线派生与打包足迹都在这条路径上。
func marshalBenchmarkDate(b *testing.B, text string) time.Time {
	b.Helper()
	if strings.HasPrefix(text, "-") {
		parts := strings.Split(strings.TrimPrefix(text, "-"), "-")
		if len(parts) != 3 {
			b.Fatalf("invalid astronomical date %q", text)
		}
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			b.Fatal(err)
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			b.Fatal(err)
		}
		day, err := strconv.Atoi(parts[2])
		if err != nil {
			b.Fatal(err)
		}
		return time.Date(-year, time.Month(month), day, 12, 0, 0, 0, time.UTC)
	}
	date, err := time.Parse("2006-01-02", text)
	if err != nil {
		b.Fatal(err)
	}
	return date
}

func benchmarkSolarEclipseMarshal(b *testing.B, text string) {
	date := marshalBenchmarkDate(b, text)
	info, ok := eclipse.SolarEclipseOnDate(date)
	if !ok || !info.HasCentral {
		b.Fatalf("expected a central solar eclipse on %s", text)
	}
	partialOptions, pathOptions := grazingBandOptions(info, "overview")
	partial, ok := eclipse.SolarEclipsePartialFootprints(date, partialOptions)
	if !ok {
		b.Fatal("missing partial footprints")
	}
	central, ok := eclipse.SolarEclipseCentralPath(date, pathOptions)
	if !ok {
		b.Fatal("missing central path")
	}
	data, err := geojson.MarshalSolarEclipse(partial, &central)
	if err != nil || len(data) == 0 {
		b.Fatalf("MarshalSolarEclipse() bytes=%d err=%v", len(data), err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		value, err := geojson.MarshalSolarEclipse(partial, &central)
		if err != nil || len(value) == 0 {
			b.Fatalf("MarshalSolarEclipse() bytes=%d err=%v", len(value), err)
		}
	}
}

func BenchmarkSolarEclipseMarshalGrazing(b *testing.B) {
	for _, text := range []string{"2003-05-31", "1874-10-10", "2061-10-13"} {
		b.Run(text, func(b *testing.B) {
			benchmarkSolarEclipseMarshal(b, text)
		})
	}
}

func BenchmarkSolarEclipseMarshalOrdinary(b *testing.B) {
	benchmarkSolarEclipseMarshal(b, "2024-04-08")
}
