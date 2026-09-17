package svg

import (
	"math"
	"strings"
	"testing"
	"time"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/svgmap"
)

// 复用 partial 结果跳过中心带重建时，导出的字节必须与完整重算中心带的旧路径逐字相同。
func TestSolarEclipseMapSkipCentralBandMatchesFullPath(t *testing.T) {
	for _, fixture := range []struct {
		name string
		date time.Time
	}{
		{name: "2024-04-08 total", date: time.Date(2024, time.April, 8, 0, 0, 0, 0, time.UTC)},
		{name: "2010-01-15 annular", date: time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC)},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			options := SolarEclipseMapSVGOptions{
				Width: 900, Height: 620, Location: time.UTC,
				PartialStep: 10 * time.Minute, CentralStep: 10 * time.Minute,
			}
			normalized := normalizeSolarEclipseMapSVGOptions(fixture.date, options)
			got, ok := SolarEclipseMapSVGNASABulletinSplitK(fixture.date, options)
			if !ok {
				t.Fatal("missing map")
			}
			global, ok := eclipsecore.SolarEclipseOnDateNASABulletinSplitK(fixture.date)
			if !ok {
				t.Fatal("missing eclipse")
			}
			partial, ok := eclipsecore.SolarEclipsePartialFootprintsNASABulletinSplitK(fixture.date, eclipsecore.SolarEclipsePartialFootprintOptions{
				Step:              normalized.PartialStep,
				BoundaryPoints:    normalized.BoundaryPoints,
				CentralShadowStep: normalized.CentralShadowStep,
				// 必须与渲染器同一组等时线电平，否则比较的是两次不同的请求。
				GreatestTimeValues: solarEclipseGreatestTimeLevels(global, normalized),
				MagnitudeValues:    normalized.MagnitudeValues,
			})
			if !ok {
				t.Fatal("missing partial footprints")
			}
			// 旧路径：中心路径不跳过中心带，渲染器不读取它的包络。
			central, hasCentral := eclipsecore.SolarEclipseCentralPathNASABulletinSplitK(fixture.date, eclipsecore.SolarEclipsePathOptions{
				Step:            normalized.CentralStep,
				TargetSpacingKM: normalized.TargetSpacingKM,
			})
			local, hasLocal := eclipsecore.GeometricLocalSolarEclipseOnDateNASABulletinSplitK(
				partial.Eclipse.GreatestEclipse,
				partial.Eclipse.GreatestLongitude,
				partial.Eclipse.GreatestLatitude,
				0,
			)
			projection := resolveSolarEclipseMapProjection(partial, central, hasCentral, normalized.Projection)
			geocentric, hasGeocentric := eclipsecore.SolarEclipseGeocentricPanelAt(fixture.date)
			want := renderSolarEclipseMapSVG(partial, central, hasCentral, local, hasLocal,
				geocentric, hasGeocentric, normalized, projection,
				svgmap.GeoPoint{Longitude: partial.Eclipse.GreatestLongitude, Latitude: partial.Eclipse.GreatestLatitude})
			if got != want {
				t.Fatalf("reused central band changed the rendered document (%d vs %d bytes)", len(got), len(want))
			}
		})
	}
}

// 自动投影只按半球守卫选择极区视图：focus 高但跨越赤道的事件必须留在等距圆柱视图。
func TestSolarEclipseMapAutoProjectionKeepsBothHemispheres(t *testing.T) {
	date := time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC)
	options := SolarEclipseMapSVGOptions{Location: time.UTC, PartialStep: 10 * time.Minute}
	auto, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("missing 2026-08-12 map")
	}
	options.Projection = EclipseMapProjectionEquirectangular
	explicit, ok := SolarEclipseMapSVG(date, options)
	if !ok {
		t.Fatal("missing explicit equirectangular map")
	}
	if auto != explicit {
		t.Fatal("auto projection promoted a two-hemisphere event to a polar view")
	}

	// 单半球事件仍必须自动选极区视图。
	antarctic := time.Date(2021, time.December, 4, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipsecore.SolarEclipsePartialFootprintsNASABulletinSplitK(antarctic, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 180,
	})
	if !ok {
		t.Fatal("missing 2021-12-04 partial footprints")
	}
	central, hasCentral := eclipsecore.SolarEclipseCentralPathNASABulletinSplitK(antarctic, eclipsecore.SolarEclipsePathOptions{Step: 2 * time.Minute})
	if got := resolveSolarEclipseMapProjection(partial, central, hasCentral, EclipseMapProjectionAuto); got != svgmap.ProjectionSouthPolar {
		t.Fatalf("2021-12-04 auto projection = %v, want the south polar view", got)
	}
}

// 包络守卫必须校验真正绘制的三条线：中心线按 2 km，限界按食带宽的四分之一。
func TestSolarEclipseCentralEnvelopeGuardChecksDrawnLines(t *testing.T) {
	date := time.Date(2010, time.January, 15, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipsecore.SolarEclipsePartialFootprints(date, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 180,
	})
	if !ok {
		t.Fatal("missing partial footprints")
	}
	central, ok := eclipsecore.SolarEclipseCentralPath(date, eclipsecore.SolarEclipsePathOptions{
		Step: 2 * time.Minute, SkipCentralBand: true,
	})
	if !ok {
		t.Fatal("missing central path")
	}
	if !solarEclipseCentralEnvelopeCoversPath(partial.CentralBandSegments, central) {
		t.Fatal("authoritative envelope was rejected for its own eclipse")
	}
	strayed := central
	strayed.SouthernLimit = solarEclipseShiftedLatitudes(central.SouthernLimit, -9)
	if solarEclipseCentralEnvelopeCoversPath(partial.CentralBandSegments, strayed) {
		t.Fatal("guard accepted a southern limit about 1000 km outside the envelope")
	}
	// 中心线单独南移约 2200 km：限界仍在容差内，只有中心线检查能拒绝它。
	strayedAxis := central
	strayedAxis.CenterLine = solarEclipseShiftedLatitudes(central.CenterLine, -20)
	if solarEclipseCentralEnvelopeCoversPath(partial.CentralBandSegments, strayedAxis) {
		t.Fatal("guard accepted a center line far outside the envelope")
	}

	// 回退到弦带时必须标注来源，消费方可区分权威包络与粗弦带。
	var builder strings.Builder
	frame := svgmap.Frame{Width: 3600, Height: 1800, Projection: svgmap.ProjectionEquirectangular}
	writeSolarEclipseCentralPath(&builder, strayed, partial.CentralBandFootprints, partial.CentralBandSegments, frame, partial.Eclipse.Type)
	if !strings.Contains(builder.String(), `<g class="central-eclipse-band" data-source="paired-limit-chords"`) {
		t.Fatal("chord-band fallback is not marked with its source")
	}
	if strings.Contains(builder.String(), "besselian-critical-envelope") {
		t.Fatal("straying limits still filled with the authoritative envelope")
	}
}

// 采样扫掠回退同样要标注来源。
func TestSolarEclipseSampledSweepFallbackIsMarked(t *testing.T) {
	date := time.Date(2043, time.April, 9, 0, 0, 0, 0, time.UTC)
	partial, ok := eclipsecore.SolarEclipsePartialFootprints(date, eclipsecore.SolarEclipsePartialFootprintOptions{
		Step: 10 * time.Minute, BoundaryPoints: 96, CentralShadowStep: 10 * time.Minute,
	})
	if !ok || len(partial.CentralShadowFootprints) == 0 {
		t.Fatal("missing 2043 central-shadow footprints")
	}
	var builder strings.Builder
	frame := svgmap.Frame{Width: 3600, Height: 1800, Projection: svgmap.ProjectionEquirectangular}
	writeSolarEclipseCentralShadowSweep(&builder, partial.CentralShadowFootprints, frame, partial.Eclipse.Type)
	if !strings.Contains(builder.String(), `<g class="solar-central-shadow-sweep" data-source="sampled-open-sweep"`) {
		t.Fatal("sampled central-shadow sweep is not marked with its source")
	}
}

// 升落阶段与方向进入 class 属性前必须转义。
func TestSolarEclipseRiseSetCurveClassNamesAreEscaped(t *testing.T) {
	curves := []eclipsecore.SolarEclipseRiseSetCurve{{
		Phase:     eclipsecore.RiseSetPhase(`start" onload="alert(1)`),
		Direction: eclipsecore.RiseSetDirection(`rise`),
		Segments: [][]eclipsecore.SolarEclipsePathPoint{{
			{Longitude: 0, Latitude: 0}, {Longitude: 1, Latitude: 1},
		}},
	}}
	var builder strings.Builder
	frame := svgmap.Frame{Width: 360, Height: 180, Projection: svgmap.ProjectionEquirectangular}
	writeSolarEclipseRiseSetCurves(&builder, curves, frame)
	if strings.Contains(builder.String(), `start" onload=`) {
		t.Fatal("rise/set phase was written into the class attribute unescaped")
	}
	if err := validateEclipseMapXML(builder.String()); err != nil {
		t.Fatalf("escaped rise/set curve is not valid XML: %v", err)
	}
}

func solarEclipseShiftedLatitudes(
	points []eclipsecore.SolarEclipsePathPoint,
	delta float64,
) []eclipsecore.SolarEclipsePathPoint {
	shifted := make([]eclipsecore.SolarEclipsePathPoint, len(points))
	copy(shifted, points)
	for index := range shifted {
		shifted[index].Latitude = math.Max(-90, math.Min(90, shifted[index].Latitude+delta))
	}
	return shifted
}
