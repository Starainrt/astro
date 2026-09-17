package geojson

import (
	"fmt"
	"math"
	"time"

	"github.com/starainrt/astro/basic"

	eclipsecore "github.com/starainrt/astro/eclipse"
	"github.com/starainrt/astro/internal/geodata"
	"github.com/starainrt/astro/internal/solarclosure"
)

const (
	// 区域角色永远是面，退化则缺省；物理边界角色永远只承载曲线。
	solarCentralShadowFootprintRole = "central-shadow-footprint"
	solarCentralShadowBoundaryRole  = "central-shadow-boundary"
	// 半影用与采样序列同名的区域角色，物理边界另立角色。
	// 细于该长度的足迹是首末接触处的残片，不导出。
	solarCentralShadowMinimumRegionLengthKM = 1.0
	// 面积下限比导出侧的零面积下限（1e-12 平方度）高 1000 倍，退化的环直接缺省。
	solarCentralShadowMinimumRegionAreaDegrees2 = 1e-9

	solarPartialFootprintRole         = "partial-footprint"
	solarPartialFootprintBoundaryRole = "partial-footprint-boundary"
)

// appendSolarHorizonClosedShadowFootprint 输出被地平线切断的瞬时中心影足迹：区域加物理边界 / horizon-cut footprint as a region plus its physical boundary.
func appendSolarHorizonClosedShadowFootprint(
	features []feature,
	footprint eclipsecore.SolarEclipsePartialFootprint,
	properties map[string]interface{},
) ([]feature, error) {
	curve, err := solarShadowFootprintCurve(footprint)
	if err != nil {
		return nil, fmt.Errorf("geojson: solar %s at %s: %w",
			solarCentralShadowFootprintRole, formatTime(footprint.Time), err)
	}
	if len(curve) < 2 {
		return features, nil
	}
	ring, boundary := solarShadowFootprintHorizonRing(
		footprint.Time, footprint.Boundaries, footprint.HorizonEnds, curve,
	)
	if solarShadowRegionDegenerate(curve, ring) {
		return features, nil
	}
	value, err := multiPolygonGeometry([][]geodata.GeoPoint{ring})
	if err != nil {
		return nil, fmt.Errorf("geojson: solar %s at %s: %w",
			solarCentralShadowFootprintRole, formatTime(footprint.Time), err)
	}
	regionProperties := cloneProperties(properties)
	regionProperties["time"] = formatTime(footprint.Time)
	regionProperties["source_boundary_closed"] = false
	regionProperties["geometry_role"] = "horizon-closed-region"
	regionProperties["closure"] = solarHorizonClosureProperties(
		footprint.Time, solarHorizonClosureExact(footprint.Boundaries, footprint.HorizonEnds),
	)
	regionProperties["interp_signature"] = solarShadowFootprintSignature(
		footprint.Boundaries, false, eclipsecore.SolarEclipseShadowUmbra,
	)
	features = append(features, newFeature(
		solarEclipseEvent, solarCentralShadowFootprintRole, value, regionProperties,
	))

	boundaryValue, err := geoMultiLineGeometry(boundary, false)
	if err != nil {
		return nil, fmt.Errorf("geojson: solar %s at %s: %w",
			solarCentralShadowBoundaryRole, formatTime(footprint.Time), err)
	}
	boundaryProperties := cloneProperties(properties)
	boundaryProperties["time"] = formatTime(footprint.Time)
	boundaryProperties["source_boundary_closed"] = false
	boundaryProperties["geometry_role"] = "open-boundary"
	boundaryProperties["interp_signature"] = solarShadowFootprintSignature(
		footprint.Boundaries, false, eclipsecore.SolarEclipseShadowUmbra,
	)
	return append(features, newFeature(
		solarEclipseEvent, solarCentralShadowBoundaryRole, boundaryValue, boundaryProperties,
	)), nil
}

func solarShadowFootprintCurve(
	footprint eclipsecore.SolarEclipsePartialFootprint,
) ([]geodata.GeoPoint, error) {
	return solarShadowFootprintCurveFromSegments(footprint.Boundaries)
}

func solarShadowFootprintCurveFromSegments(
	boundaries [][]eclipsecore.SolarEclipsePathPoint,
) ([]geodata.GeoPoint, error) {
	segments := make([][]geodata.GeoPoint, 0, len(boundaries))
	for _, source := range boundaries {
		segment := make([]geodata.GeoPoint, len(source))
		for index, point := range source {
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return nil, err
			}
			segment[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		segments = append(segments, segment)
	}
	return geodata.JoinPolylineSegments(segments), nil
}

// solarShadowRegionDegenerate 判定影区是否退化：物理边界过短或闭合面积近零，导出侧会拒绝。
func solarShadowRegionDegenerate(curve, ring []geodata.GeoPoint) bool {
	if len(curve) < 2 || solarShadowFootprintCurveLengthKM(curve) < solarCentralShadowMinimumRegionLengthKM {
		return true
	}
	points := openRing(ring)
	if len(points) < 3 {
		return true
	}
	return math.Abs(solarShadowRegionAreaDegrees2(points)) < solarCentralShadowMinimumRegionAreaDegrees2
}

// solarShadowRegionAreaDegrees2 用展开经度算鞋带面积，跨换日线的环不会得到假的近零面积。
func solarShadowRegionAreaDegrees2(ring []geodata.GeoPoint) float64 {
	if len(ring) < 3 {
		return 0
	}
	unwrapped := make([]geodata.GeoPoint, len(ring))
	unwrapped[0] = ring[0]
	for index := 1; index < len(ring); index++ {
		longitude := ring[index].Longitude
		for longitude-unwrapped[index-1].Longitude > 180 {
			longitude -= 360
		}
		for longitude-unwrapped[index-1].Longitude < -180 {
			longitude += 360
		}
		unwrapped[index] = geodata.GeoPoint{Longitude: longitude, Latitude: ring[index].Latitude}
	}
	area := 0.0
	for index, point := range unwrapped {
		next := unwrapped[(index+1)%len(unwrapped)]
		area += point.Longitude*next.Latitude - next.Longitude*point.Latitude
	}
	return area / 2
}

func solarShadowFootprintCurveLengthKM(curve []geodata.GeoPoint) float64 {
	length := 0.0
	for index := 1; index < len(curve); index++ {
		length += solarCentralBandGeoPointDistanceKM(curve[index-1], curve[index])
	}
	return length
}

func solarShadowFootprintHorizonRing(
	value time.Time,
	boundaries [][]eclipsecore.SolarEclipsePathPoint,
	horizonEnds []eclipsecore.SolarEclipsePathPoint,
	curve []geodata.GeoPoint,
) ([]geodata.GeoPoint, []geodata.GeoPoint) {
	return solarclosure.HorizonRing(solarclosure.Footprint{
		Boundaries:  solarClosureSegments(boundaries),
		HorizonEnds: solarClosureEnds(horizonEnds),
		Subsolar:    solarSubsolarPoint(value),
	}, curve)
}

func solarShadowFootprintHorizonEnds(
	boundaries [][]eclipsecore.SolarEclipsePathPoint,
	horizonEnds []eclipsecore.SolarEclipsePathPoint,
) []geodata.GeoPoint {
	return solarclosure.HorizonEnds(solarclosure.Footprint{
		Boundaries:  solarClosureSegments(boundaries),
		HorizonEnds: solarClosureEnds(horizonEnds),
	})
}

// solarClosureSegments 按原顺序转换边界分段，不校验坐标。
func solarClosureSegments(boundaries [][]eclipsecore.SolarEclipsePathPoint) [][]geodata.GeoPoint {
	segments := make([][]geodata.GeoPoint, len(boundaries))
	for index, source := range boundaries {
		segments[index] = solarCentralBandGeoPoints(source)
	}
	return segments
}

// solarClosureEnds 转换地平擦地点；点数不是 2 或坐标非法时返回 nil，闭合退回近似弧。
func solarClosureEnds(horizonEnds []eclipsecore.SolarEclipsePathPoint) []geodata.GeoPoint {
	if len(horizonEnds) != 2 {
		return nil
	}
	points := make([]geodata.GeoPoint, 2)
	for index, end := range horizonEnds {
		if err := validateCoordinate(end.Longitude, end.Latitude); err != nil {
			return nil
		}
		points[index] = geodata.GeoPoint{Longitude: end.Longitude, Latitude: end.Latitude}
	}
	return points
}

// solarClosureFootprint 转换瞬时足迹；边界坐标非法时报错。
func solarClosureFootprint(
	footprint eclipsecore.SolarEclipsePartialFootprint,
) (solarclosure.Footprint, error) {
	segments := make([][]geodata.GeoPoint, len(footprint.Boundaries))
	for index, source := range footprint.Boundaries {
		segment := make([]geodata.GeoPoint, len(source))
		for pointIndex, point := range source {
			if err := validateCoordinate(point.Longitude, point.Latitude); err != nil {
				return solarclosure.Footprint{}, err
			}
			segment[pointIndex] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
		}
		segments[index] = segment
	}
	return solarclosure.Footprint{
		Boundaries:  segments,
		HorizonEnds: solarClosureEnds(footprint.HorizonEnds),
		Subsolar:    solarSubsolarPoint(footprint.Time),
		Closed:      footprint.Closed,
	}, nil
}

// solarHorizonClosureExact 报告能否用两个精确擦地点闭合区域。
func solarHorizonClosureExact(
	boundaries [][]eclipsecore.SolarEclipsePathPoint,
	horizonEnds []eclipsecore.SolarEclipsePathPoint,
) bool {
	return len(solarShadowFootprintHorizonEnds(boundaries, horizonEnds)) == 2
}

// solarHorizonClosureProperties 声明式闭合弧；exact 为假表示缺精确擦地点、按采样端点近似闭合。
func solarHorizonClosureProperties(value time.Time, exact bool) map[string]interface{} {
	subsolar := solarSubsolarPoint(value)
	return map[string]interface{}{
		"kind":     "horizon",
		"exact":    exact,
		"time":     formatTime(value),
		"subsolar": []float64{subsolar.Longitude, subsolar.Latitude},
	}
}

// solarShadowSegmentClosed 与 basic 层同口径：分段首末点相距不到 1 mm 才算真正闭合。
func solarShadowSegmentClosed(segment []eclipsecore.SolarEclipsePathPoint) bool {
	if len(segment) <= 2 {
		return false
	}
	return solarCentralBandPathDistanceKM(segment[0], segment[len(segment)-1]) < 1e-6
}

// solarShadowFootprintSignature 由物理边界算出与 basic 层同口径的插值签名 / interpolation signature from the physical boundary.
func solarShadowFootprintSignature(
	boundaries [][]eclipsecore.SolarEclipsePathPoint,
	closed bool,
	kind eclipsecore.SolarEclipseShadowKind,
) string {
	topology := basic.SolarEclipseShadowTopology{
		Kind:     basic.SolarEclipseShadowKind(kind),
		Segments: len(boundaries),
		Closed:   closed,
	}
	winding := 0.0
	for _, segment := range boundaries {
		topology.Vertices += len(segment)
		for index := 1; index < len(segment); index++ {
			winding += math.Remainder(segment[index].Longitude-segment[index-1].Longitude, 360)
		}
		if solarShadowSegmentClosed(segment) {
			winding += math.Remainder(segment[0].Longitude-segment[len(segment)-1].Longitude, 360)
		}
	}
	if topology.Segments == 1 && math.Abs(winding) >= 180 {
		topology.EnclosesPole = true
	}
	return topology.Signature()
}
