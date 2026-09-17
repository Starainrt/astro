// Package occultationgeo provides shared geographic helpers for occultation encoders and renderers.
package occultationgeo

import (
	"fmt"
	"math"

	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

// 掩星边界连续性判定的几何门限 / geometric thresholds for occultation-boundary continuity checks.
const (
	EarthRadiusKM        = 6378.1366
	BoundaryBranchJumpKM = 750.0
	// BoundaryBranchSpeedKMPerSecond 是判定"物理上不可能的支路跳变"的地面速度门限：
	// 掩星边界随月球影子移动，地面速度上限约 1.1 km/s，取 2 km/s 留约两倍余量。
	// BoundaryBranchSpeedKMPerSecond gates "physically impossible" branch jumps: the boundary
	// follows the lunar shadow at up to about 1.1 km/s on the ground, so 2 km/s keeps a
	// factor-of-two margin.
	BoundaryBranchSpeedKMPerSecond = 2.0
	staticCenterCapRadiusKM        = 31.0
	staticCenterCapPoints          = 16
	closedFootprintSweepPoints     = 96
	closedFootprintSweepMaxStepKM  = 2500.0
)

// SampleRange 是一个可在不跨越支路变化时连接的半开样本区间。
// SampleRange is a half-open range of samples that can be joined without crossing a branch change.
type SampleRange struct {
	Start int
	End   int
}

// ContinuousBoundaryRanges 在物理上不可能的支路跳变处拆分边界。
// ContinuousBoundaryRanges splits a boundary at physically impossible branch changes.
// One-sample ranges are retained so callers can represent event endpoints without reconnecting a jump.
func ContinuousBoundaryRanges(points []basic.OccultationPathPoint) []SampleRange {
	return continuousRanges(len(points), func(index int) bool {
		return BoundaryBranchChanged(points[index-1], points[index])
	})
}

// ContinuousPairedBoundaryRanges 在成对边界任一侧换支时拆分区间。
// ContinuousPairedBoundaryRanges splits paired limits when either side changes branch.
func ContinuousPairedBoundaryRanges(
	first, second []basic.OccultationPathPoint,
) []SampleRange {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	return continuousRanges(count, func(index int) bool {
		return BoundaryBranchChanged(first[index-1], first[index]) ||
			BoundaryBranchChanged(second[index-1], second[index])
	})
}

// PairedBoundaryPolygons 仅在连续成对边界之间返回扫掠单元。
// PairedBoundaryPolygons returns sweep cells only across continuous paired
// limit ranges. Instantaneous footprints remain responsible for end caps.
func PairedBoundaryPolygons(
	first, second []basic.OccultationPathPoint,
) [][]geodata.GeoPoint {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	polygons := make([][]geodata.GeoPoint, 0, count)
	// A branch change on only one side creates a long-lived invalid cross
	// section: the changed side has already moved to its new tangent branch
	// while the other side remains on the old branch. Suppress cells until the
	// next branch transition establishes a new paired branch.
	pendingBranch := false
	for index := 1; index < count; index++ {
		firstChanged := BoundaryBranchChanged(first[index-1], first[index])
		secondChanged := BoundaryBranchChanged(second[index-1], second[index])
		if pendingBranch {
			// 抑制以"下一次任一侧换支"为界；此后若不再换支，剩余单元仍是不匹配的支路对。
			if firstChanged || secondChanged {
				pendingBranch = false
			}
			continue
		}
		if firstChanged != secondChanged {
			pendingBranch = true
			continue
		}
		if firstChanged { // both sides changed at the same transition
			continue
		}
		previousFirst, currentFirst := first[index-1], first[index]
		previousSecond, currentSecond := second[index-1], second[index]
		polygons = append(polygons, []geodata.GeoPoint{
			{Longitude: previousFirst.Longitude, Latitude: previousFirst.Latitude},
			{Longitude: currentFirst.Longitude, Latitude: currentFirst.Latitude},
			{Longitude: currentSecond.Longitude, Latitude: currentSecond.Latitude},
			{Longitude: previousSecond.Longitude, Latitude: previousSecond.Latitude},
		})
	}
	return polygons
}

// RemoveTinyPolygonComponents 删除紧凑极区掩带组面时留下的微小数值薄片。
// RemoveTinyPolygonComponents drops numerical slivers left when a compact
// sweep closes at a shared endpoint. A component survives only when it reaches
// 0.01% of the largest component area, so components comparable to the main
// band and genuine disjoint projected branches are always preserved; the
// largest component itself is retained unconditionally.
func RemoveTinyPolygonComponents(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	if len(polygons) < 2 {
		return polygons
	}
	areas := make([]float64, len(polygons))
	maximum, maximumIndex := 0.0, 0
	for index, polygon := range polygons {
		areas[index] = math.Abs(geoRingArea(polygon))
		if areas[index] > maximum {
			maximum, maximumIndex = areas[index], index
		}
	}
	if maximum <= 0 || !finiteGeo(maximum) {
		return polygons
	}
	threshold := maximum * 1e-4
	filtered := make([][]geodata.GeoPoint, 0, len(polygons))
	for index, polygon := range polygons {
		// 最大分量必然达到门槛，显式保留以固定"主带不会被过滤"的契约。
		if index == maximumIndex || areas[index] >= threshold {
			filtered = append(filtered, polygon)
		}
	}
	return filtered
}

// removeOccultationPolarSliverComponents drops detached high-latitude faces
// with only a handful of vertices. These are polygonizer junction slivers,
// not independent occultation regions; merging one into the main face turns
// its closure into the staircase visible at the south polar tip.
func removeOccultationPolarSliverComponents(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	if len(polygons) < 2 {
		return polygons
	}
	filtered := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		polar := len(polygon) < 8
		for _, point := range polygon {
			if math.Abs(point.Latitude) < 70 {
				polar = false
				break
			}
		}
		if !polar {
			filtered = append(filtered, polygon)
		}
	}
	if len(filtered) == 0 {
		return polygons
	}
	return filtered
}

// ConstrainPolygonsWithin 修复子掩带对父掩带的小数值突破，大幅差异保持不变以便诊断。
// ConstrainPolygonsWithin repairs a small numerical breach of a child band
// against its parent band. Finite-disk inner-contact sweeps can differ from
// the outer sweep by a few samples at a branch junction; when the breach is
// small, replacing those samples with the nearest parent boundary vertex
// preserves the child curve while restoring the physical containment
// invariant. Large breaches are left untouched so this helper cannot hide a
// wrong face selection.
func ConstrainPolygonsWithin(
	parent, child [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	if len(parent) == 0 || len(child) == 0 {
		return child
	}
	initialMissDistance := geodata.SphericalPolygonsPathMissDistanceKM(parent, child, true)
	if initialMissDistance <= 0 {
		return child
	}
	const maximumRepairDistanceKM = 100.0
	const maximumResidualMissDistanceKM = 10.0
	if initialMissDistance > maximumRepairDistanceKM {
		return child
	}
	result := make([][]geodata.GeoPoint, len(child))
	// 父带不变，逐点包含索引只建一次；每遍重建会让 6 遍细化退化成 O(passes×父带边数)。
	parentIndex := geodata.NewSphericalPolygonIndex(parent)
	for index, source := range child {
		if len(source) < 4 {
			if len(openFootprintRing(source)) >= 3 && math.Abs(geoRingArea(source)) > 1e-12 {
				result[index] = append([]geodata.GeoPoint(nil), source...)
			}
			continue
		}
		ring := append([]geodata.GeoPoint(nil), source...)
		closed := geodata.SameGeoPoint(ring[0], ring[len(ring)-1])
		limit := len(ring)
		if closed {
			limit--
		}
		containment := parentIndex.ContainsPoints(ring[:limit])
		for pointIndex := 0; pointIndex < limit; pointIndex++ {
			point := ring[pointIndex]
			if containment[pointIndex] {
				continue
			}
			nearest, distance := nearestPolygonBoundaryPoint(parent, point)
			if distance <= maximumRepairDistanceKM {
				ring[pointIndex] = nearest
			}
		}
		if closed {
			ring[len(ring)-1] = ring[0]
		}
		result[index] = ring
	}
	result = usableOccultationPolygons(result)
	if geodata.SphericalPolygonsPathMissDistanceKM(parent, result, true) <= 0 {
		return result
	}
	// Vertex-only repair cannot see a child edge whose endpoints are both
	// inside the parent while its great-circle midpoint crosses outside. Split
	// the repaired ring at the same projected spacing used by output geometry,
	// then apply the local vertex snap to those newly exposed edge probes.
	densified := densifyOccultationPolygons(result, 10)
	densifiedMiss := initialMissDistance
	for pass := 0; pass < 6; pass++ {
		changed := false
		for index, source := range densified {
			ring := append([]geodata.GeoPoint(nil), source...)
			closed := len(ring) > 1 && geodata.SameGeoPoint(ring[0], ring[len(ring)-1])
			limit := len(ring)
			if closed {
				limit--
			}
			containment := parentIndex.ContainsPoints(ring[:limit])
			for pointIndex := 0; pointIndex < limit; pointIndex++ {
				point := ring[pointIndex]
				if containment[pointIndex] {
					continue
				}
				nearest, distance := nearestPolygonBoundaryPoint(parent, point)
				if distance <= maximumRepairDistanceKM {
					ring[pointIndex] = nearest
					changed = true
				}
			}
			if closed {
				ring[len(ring)-1] = ring[0]
			}
			for {
				cleaned := removeDirectProjectedSharpCorners(ring, 20, 30)
				cleaned = removeOccultationSharpCorners(cleaned, 20, 30)
				if len(cleaned) == len(ring) {
					break
				}
				ring = cleaned
			}
			if closed && len(ring) > 1 && !geodata.SameGeoPoint(ring[0], ring[len(ring)-1]) {
				ring = append(ring, ring[0])
			}
			densified[index] = ring
		}
		densified = usableOccultationPolygons(densified)
		densifiedMiss = geodata.SphericalPolygonsPathMissDistanceKM(parent, densified, true)
		if densifiedMiss <= maximumResidualMissDistanceKM {
			return densified
		}
		if !changed {
			break
		}
	}
	// 只有残差实质变小（>10%）才采用细化结果；亚公里级改善不值得改变输出点数，
	// 其余情况按"大突破保持不变"的契约返回 child。
	if len(densified) > 0 && densifiedMiss < initialMissDistance*0.9 {
		return densified
	}
	return child
}

func nearestPolygonVertex(
	polygons [][]geodata.GeoPoint,
	point geodata.GeoPoint,
) (geodata.GeoPoint, float64) {
	nearest := geodata.GeoPoint{}
	distance := math.Inf(1)
	for _, polygon := range polygons {
		for _, candidate := range polygon {
			value := geoDistanceKM(point, candidate)
			if value < distance {
				nearest, distance = candidate, value
			}
		}
	}
	return nearest, distance
}

func nearestPolygonBoundaryPoint(
	polygons [][]geodata.GeoPoint,
	point geodata.GeoPoint,
) (geodata.GeoPoint, float64) {
	nearest := geodata.GeoPoint{}
	distance := math.Inf(1)
	for _, polygon := range polygons {
		if len(polygon) < 2 {
			continue
		}
		limit := len(polygon)
		if limit > 1 && geodata.SameGeoPoint(polygon[0], polygon[limit-1]) {
			limit--
		}
		for index := 0; index < limit; index++ {
			start := polygon[index]
			end := polygon[(index+1)%limit]
			latitude := point.Latitude * math.Pi / 180
			scaleX := math.Cos(latitude)
			startX := math.Remainder(start.Longitude-point.Longitude, 360) * scaleX
			startY := start.Latitude - point.Latitude
			endX := math.Remainder(end.Longitude-point.Longitude, 360) * scaleX
			endY := end.Latitude - point.Latitude
			deltaX, deltaY := endX-startX, endY-startY
			fraction := 0.0
			if lengthSquared := deltaX*deltaX + deltaY*deltaY; lengthSquared > 0 {
				fraction = math.Max(0, math.Min(1,
					-(startX*deltaX+startY*deltaY)/lengthSquared,
				))
			}
			candidate := interpolateOccultationGeoPoint(start, end, fraction)
			value := geoDistanceKM(point, candidate)
			if value < distance {
				nearest, distance = candidate, value
			}
		}
	}
	if math.IsInf(distance, 1) {
		return nearestPolygonVertex(polygons, point)
	}
	return nearest, distance
}

// FootprintSweepPolygons 构造瞬时月掩足迹的静态扫掠并集。
// FootprintSweepPolygons builds the static union of instantaneous occultation
// footprints. Open contact-cone arcs are swept between adjacent samples so
// their per-instant horizon closures do not survive as staircase edges.
// Legacy footprints without Boundaries retain the polygon-union behavior.
func FootprintSweepPolygons(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
) ([][]geodata.GeoPoint, error) {
	return footprintSweepPolygons(footprints, northern, southern)
}

// ContactSweepBoundaryLines 返回由瞬时接触弧的时间扫掠导出的连续边界线网。
// ContactSweepBoundaryLines returns continuous boundary linework derived from
// open contact-cone arcs. Closed instantaneous footprint rings are deliberately
// excluded so callers can use these lines as static-band boundary candidates.
func ContactSweepBoundaryLines(
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	if !footprintBoundariesAvailable(footprints) {
		return nil
	}
	polygons, err := footprintOpenSweepPolygons(footprints)
	if err != nil {
		polygons, err = footprintOpenSweepPolygonsWithoutTransitions(footprints)
	}
	if err != nil || len(polygons) == 0 {
		// A branch change can make the ribbon union fail at a single numerical
		// intersection even though its endpoint tracks remain valid. Expose those
		// tracks as diagnostic boundary lines; they are also the caps used by the
		// bounded endpoint fallback and therefore keep line/fill audits consistent.
		samples := make([]geodata.OpenBoundarySweepSample, 0, len(footprints))
		for _, footprint := range footprints {
			if footprint.Closed {
				samples = append(samples, geodata.OpenBoundarySweepSample{Closed: true})
				continue
			}
			boundaries := footprintGeoBoundaries(footprint)
			if len(boundaries) == 0 {
				continue
			}
			samples = append(samples, geodata.OpenBoundarySweepSample{Boundaries: boundaries})
		}
		outlines, outlineErr := geodata.OpenBoundaryEndpointOutlines(samples)
		if outlineErr != nil {
			return nil
		}
		polygons = outlines
		// Keep the actual open contact arcs as diagnostic linework as well. The
		// endpoint outline alone omits the intermediate limb bulges and can make a
		// valid time-union edge appear a few kilometres detached from its source
		// boundary after spherical union interpolation.
		for _, footprint := range footprints {
			for _, boundary := range footprintGeoBoundaries(footprint) {
				if len(boundary) >= 2 {
					polygons = append(polygons, boundary)
				}
			}
		}
	}
	lines := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		if len(polygon) < 3 {
			continue
		}
		line := append([]geodata.GeoPoint(nil), polygon...)
		if !geodata.SameGeoPoint(line[0], line[len(line)-1]) {
			line = append(line, line[0])
		}
		// Sparse polar footprint sweeps can carry a one-sample numerical return
		// at a branch junction. Remove that local kink before this line is used
		// as a fallback boundary; authoritative contact contours remain untouched.
		line = removeOccultationPolarKinks(line)
		line = removeOccultationSharpCorners(line, 20, 30)
		if len(line) > 1 && !geodata.SameGeoPoint(line[0], line[len(line)-1]) {
			line = append(line, line[0])
		}
		lines = append(lines, line)
	}
	return lines
}

func footprintSweepPolygons(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
) ([][]geodata.GeoPoint, error) {
	var (
		staticRepairs [][]geodata.GeoPoint
		closedSweep   [][]geodata.GeoPoint
		legacy        [][]geodata.GeoPoint
		polygons      [][]geodata.GeoPoint
	)
	usedOpenSweep := false
	if footprintBoundariesAvailable(footprints) {
		sweep, err := footprintOpenSweepPolygons(footprints)
		if err != nil {
			sweep = nil
		}
		if len(sweep) > 0 && footprintSweepCoversSamples(sweep, footprints) {
			// A covering open sweep already contains the closed instantaneous
			// footprints between its two horizon-limited runs. Adding those faces and
			// their secondary sweep again introduces coincident edges and can make the
			// spherical union select a sampled scallop or reject an otherwise closed
			// continuous ring. Keep source footprints as witnesses; only independent
			// interior repair faces still need to participate in the output union.
			hasInterior := false
			for _, footprint := range footprints {
				hasInterior = hasInterior || len(footprint.InteriorPolygons) > 0
			}
			if hasInterior {
				staticRepairs = footprintStaticInteriorPolygons(footprints)
				polygons = append(polygons, staticRepairs...)
			}
			polygons = append(polygons, sweep...)
			usedOpenSweep = true
		} else {
			staticRepairs = footprintStaticInteriorPolygons(footprints)
			closedSweep = footprintClosedSweepPolygons(footprints)
			legacy = append(footprintPolygons(footprints), staticRepairs...)
			legacy = append(legacy, closedSweep...)
			polygons = legacy
			if len(closedSweep) == 0 {
				polygons = append(polygons, PairedBoundaryPolygons(northern, southern)...)
			}
		}
	} else {
		staticRepairs = footprintStaticInteriorPolygons(footprints)
		closedSweep = footprintClosedSweepPolygons(footprints)
		legacy = append(footprintPolygons(footprints), staticRepairs...)
		legacy = append(legacy, closedSweep...)
		polygons = legacy
		if len(closedSweep) == 0 {
			polygons = append(polygons, PairedBoundaryPolygons(northern, southern)...)
		}
	}
	if len(polygons) == 0 {
		return nil, fmt.Errorf("occultation footprint sweep has no usable polygons")
	}
	polygons = usableOccultationPolygons(polygons)
	if len(polygons) == 0 {
		return nil, fmt.Errorf("occultation footprint sweep has no non-degenerate polygons")
	}
	if usedOpenSweep && len(polygons) == 1 {
		return cleanupFootprintSweepPolygons(closedOccultationSweepFaces(polygons), northern, southern), nil
	}
	merged, err := geodata.UnionPolygons(polygons)
	if err != nil && usedOpenSweep {
		merged, err = occultationRetryOpenSweepUnion(polygons, footprints, northern, southern, err)
	}
	if err != nil {
		if fallback := closedOccultationSweepFaces(polygons); len(fallback) > 0 {
			return cleanupFootprintSweepPolygons(fallback, northern, southern), nil
		}
		return nil, fmt.Errorf("merge instantaneous footprints: %w", err)
	}
	return cleanupFootprintSweepPolygons(merged, northern, southern), nil
}

// occultationRetryOpenSweepUnion 在时间扫掠的球面并集失败后依次重试去掉过渡端帽、
// 只用配对限带、退回传统瞬时面；全部失败才把原始错误交回调用方。
func occultationRetryOpenSweepUnion(
	polygons [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
	original error,
) ([][]geodata.GeoPoint, error) {
	bareSweep, bareErr := footprintOpenSweepPolygonsWithoutTransitions(footprints)
	if bareErr == nil && len(bareSweep) > 0 && footprintSweepCoversSamples(bareSweep, footprints) {
		input := footprintClosedPolygons(footprints)
		input = append(input, footprintStaticInteriorPolygons(footprints)...)
		input = append(input, footprintClosedSweepPolygons(footprints)...)
		input = append(input, bareSweep...)
		if merged, err := geodata.UnionPolygons(input); err == nil {
			return merged, nil
		}
	}
	paired := PairedBoundaryPolygons(northern, southern)
	if len(paired) > 0 {
		if merged, err := geodata.UnionPolygons(paired); err == nil {
			return merged, nil
		}
	}
	legacy := footprintPolygons(footprints)
	legacy = append(legacy, footprintStaticInteriorPolygons(footprints)...)
	legacy = append(legacy, footprintClosedSweepPolygons(footprints)...)
	legacy = append(legacy, paired...)
	if len(legacy) > 0 {
		if merged, err := geodata.UnionPolygons(legacy); err == nil {
			return merged, nil
		}
	}
	return nil, original
}

func closedOccultationSweepFaces(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	result := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		open := openFootprintRing(polygon)
		if len(open) < 3 || math.Abs(geoRingArea(open)) <= 1e-12 {
			continue
		}
		closed := append([]geodata.GeoPoint(nil), open...)
		closed = append(closed, open[0])
		result = append(result, closed)
	}
	return result
}

func cleanupFootprintSweepPolygons(
	polygons [][]geodata.GeoPoint,
	northern, southern []basic.OccultationPathPoint,
) [][]geodata.GeoPoint {
	if len(northern) > 0 || len(southern) > 0 {
		polygons = RemoveTinyPolygonComponents(polygons)
	}
	for index := range polygons {
		polygons[index] = removeOccultationHairpins(polygons[index], 35, 25, 12)
		polygons[index] = removeOccultationHairpins(polygons[index], 100, 25, 32)
		polygons[index] = removeOccultationSharpCorners(polygons[index], 20, 30)
	}
	return polygons
}

// VisibleBandPolygons 返回瞬时可见接触区域的连续时间并集。
// VisibleBandPolygons returns the union of the instantaneous visible
// footprints when samples are available. That union is the geographic area
// where the occultation occurs at any time while the Moon is above the local
// horizon; rise/set phase curves are diagnostic/display boundaries, not the
// outer edge of this time-union. The contour/linework construction remains a
// fallback for callers that do not provide footprints.
func VisibleBandPolygons(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	polygons, authoritative, err := visibleBandPolygons(footprints, northern, southern, nil, curves, false)
	return normalizeOccultationBandOutput(polygons), authoritative, err
}

// VisibleTotalBandPolygons 是 VisibleBandPolygons 的全掩带变体。
// VisibleTotalBandPolygons is the total-occultation variant of
// VisibleBandPolygons. Inner-contact total bands can retain compact numerical
// polar returns after linework polygonization, so they enable the stronger
// smoothing pass that would be too aggressive for partial-band slivers.
func VisibleTotalBandPolygons(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	polygons, authoritative, err := visibleBandPolygons(footprints, northern, southern, nil, curves, true)
	if authoritative {
		polygons = roundOccultationTotalBandJunctions(polygons)
	}
	return normalizeOccultationBandOutput(polygons), authoritative, err
}

// VisibleBandPolygonsFromContours 使用与 VisibleBandPolygons 相同的时间并集语义，并保留连续接触包络。
// VisibleBandPolygonsFromContours uses the same time-union semantics as
// VisibleBandPolygons. Continuous contact envelopes and rise/set curves remain
// available as fallback/diagnostic geometry, while supplied footprints define
// the static visible area.
func VisibleBandPolygonsFromContours(
	footprints []basic.OccultationFootprint,
	contours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	polygons, authoritative, err := visibleBandPolygons(footprints, northern, southern, contours, curves, false)
	return normalizeOccultationBandOutput(polygons), authoritative, err
}

// VisibleTotalBandPolygonsFromContours 是基于内接触轮廓的全掩带变体。
// VisibleTotalBandPolygonsFromContours is the inner-contact total-band variant
// of VisibleBandPolygonsFromContours.
func VisibleTotalBandPolygonsFromContours(
	footprints []basic.OccultationFootprint,
	contours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	polygons, authoritative, err := visibleBandPolygons(footprints, northern, southern, contours, curves, true)
	if authoritative {
		polygons = roundOccultationTotalBandJunctions(polygons)
	}
	return normalizeOccultationBandOutput(polygons), authoritative, err
}

// VisibleBandPolygonsFromAnalyticContours 从连续解析接触包络构造静态可见集。
// VisibleBandPolygonsFromAnalyticContours constructs the static visible set
// from the complete analytic boundary network. Contact contours bound the time
// union of F<=0, visibility contours bound the time union of H>=0, and the
// start/end rise-set curves are their F=0,H=0 transitions. Footprints and limit
// strips select the covered faces only; none of their edges can enter the
// returned boundary.
func VisibleBandPolygonsFromAnalyticContours(
	footprints []basic.OccultationFootprint,
	contactContours, visibilityContours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	return visibleBandPolygonsFromAnalyticContours(
		footprints, contactContours, visibilityContours, northern, southern, curves, false,
	)
}

// VisibleStarBandPolygonsFromAnalyticContours 是点光源恒星掩带的解析轮廓构造入口。
// VisibleStarBandPolygonsFromAnalyticContours is the point-source stellar
// variant. Stellar start/end contacts can require a short temporal lunar-
// horizon connector and stricter graph snapping than finite-disk contacts.
func VisibleStarBandPolygonsFromAnalyticContours(
	footprints []basic.OccultationFootprint,
	contactContours, visibilityContours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	return visibleBandPolygonsFromAnalyticContours(
		footprints, contactContours, visibilityContours, northern, southern, curves, true,
	)
}

// VisibleTotalBandPolygonsFromAnalyticContours 是解析可见集的内接触全掩带变体。
// VisibleTotalBandPolygonsFromAnalyticContours is the inner-contact variant of
// VisibleBandPolygonsFromAnalyticContours.
func VisibleTotalBandPolygonsFromAnalyticContours(
	footprints []basic.OccultationFootprint,
	contactContours, visibilityContours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool, error) {
	return visibleBandPolygonsFromAnalyticContours(
		footprints, contactContours, visibilityContours, northern, southern, curves, false,
	)
}

func visibleBandPolygonsFromAnalyticContours(
	footprints []basic.OccultationFootprint,
	contactContours, visibilityContours [][]basic.OccultationPathPoint,
	northern, southern []basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
	pointSource bool,
) ([][]geodata.GeoPoint, bool, error) {
	// The analytic boundary is preferred because it preserves the continuous
	// contact envelope.  Some grazing/polar tracks still produce a valid set of
	// instantaneous visible footprints while their phase graph has no accepted
	// closed face.  Keep that physical time-union as a bounded fallback instead
	// of turning a real event into a serialization error.
	fallbackVisible := func() ([][]geodata.GeoPoint, bool) {
		if len(footprints) == 0 {
			return nil, false
		}
		fallback, fallbackErr := footprintSweepPolygons(footprints, northern, southern)
		if fallbackErr != nil || len(fallback) == 0 {
			return nil, false
		}
		fallback = normalizeOccultationBandOutput(fallback)
		fallback = densifyOccultationPolygons(fallback, 30)
		return fallback, len(fallback) > 0
	}
	if len(curves) == 0 && len(footprints) > 0 {
		// Analytic contact contours alone do not form a closed visible boundary
		// when rise/set computation is disabled. Combine the open sweep with only
		// closed instantaneous footprints; a full sparse union is both expensive
		// and unnecessary for this compatibility path.
		if fallback, fallbackErr := footprintSweepPolygons(footprints, northern, southern); fallbackErr == nil {
			closed := footprintClosedPolygons(footprints)
			input := append(append([][]geodata.GeoPoint(nil), fallback...), closed...)
			if merged, mergeErr := geodata.UnionPolygons(input); mergeErr == nil && len(merged) > 0 {
				return cleanupFootprintSweepPolygons(merged, northern, southern), false, nil
			}
		}
	}
	contactLines := occultationContactContourBoundaryLines(contactContours)
	if len(contactLines) == 0 {
		return nil, false, fmt.Errorf("analytic occultation boundary has no contact temporal envelope")
	}
	boundaryCurves := occultationStaticBandCurves(curves)
	lineworkCurves := boundaryCurves
	if pointSource && len(visibilityContours) == 0 {
		// Without a separate H=0 temporal envelope, a greatest-rise/set arc can
		// become part of the outer visible-set boundary in a short horizon wedge.
		// Include all six phase curves in the face graph; merging selected faces
		// removes any portions that are truly internal.
		lineworkCurves = curves
	}
	connectors := HorizonConnectorSegments(footprints, boundaryCurves, northern, southern)
	if pointSource {
		connectors = StarHorizonConnectorSegments(footprints, boundaryCurves, northern, southern)
	}
	connectorLines := occultationHorizonConnectorBoundaryLines(connectors)
	boundaryLines := occultationVisibleBoundaryLinesFromBase(
		contactLines, lineworkCurves,
		occultationContactContourBoundaryLines(visibilityContours),
	)
	boundaryLines = append(boundaryLines, connectorLines...)
	var footprintFill [][]geodata.GeoPoint
	fillReady := false
	getFootprintFill := func() [][]geodata.GeoPoint {
		if !fillReady {
			footprintFill = occultationVisibleFootprintFillOnly(footprints)
			fillReady = true
		}
		return footprintFill
	}
	var selectionFill [][]geodata.GeoPoint
	if pointSource || len(visibilityContours) == 0 {
		// Geocentric limit strips can extend beyond the station-corrected
		// stellar envelope or omit a finite-disk horizon extremum. Use visible
		// footprints as witnesses; the analytic linework supplies every edge.
		selectionFill = getFootprintFill()
	}
	if len(selectionFill) == 0 {
		selectionFill = occultationLimitVisibleFillPolygons(northern, southern)
	}
	if len(selectionFill) == 0 {
		selectionFill = getFootprintFill()
	}
	if len(selectionFill) == 0 {
		return nil, false, fmt.Errorf("analytic occultation boundary has no interior selection fill")
	}
	lineworkToleranceKM := 2.0
	if !pointSource && len(visibilityContours) == 0 {
		// Finite-disk compatibility: without a separate H=0 envelope, sparse
		// contact/limit samples need the established one-edge graph tolerance.
		lineworkToleranceKM = 40
	}
	phaseLines := occultationRiseSetBoundaryLines(curves)
	polygonize := func(coverage [][]geodata.GeoPoint) ([][]geodata.GeoPoint, error) {
		if pointSource {
			return geodata.VisibleLineworkPolygonsWithAuditTolerance(
				boundaryLines, selectionFill, coverage, lineworkToleranceKM, 30,
			)
		}
		return geodata.VisibleLineworkPolygons(
			boundaryLines, selectionFill, coverage, lineworkToleranceKM,
		)
	}
	var initialCoverage [][]geodata.GeoPoint
	if pointSource && len(visibilityContours) == 0 {
		initialCoverage = phaseLines
	}
	polygons, err := polygonize(initialCoverage)
	if err != nil && len(phaseLines) > 0 {
		// A multi-branch rise/set network can contain several closed faces with
		// identical physical junctions. If fill-only selection chooses the
		// adjacent face, require every exported phase line as a coverage witness
		// and retry without changing the boundary network or snap tolerance.
		polygons, err = polygonize(phaseLines)
	}
	if err != nil {
		if fallback, ok := fallbackVisible(); ok {
			return fallback, false, nil
		}
		return nil, false, fmt.Errorf("analytic occultation boundary: %w", err)
	}
	polygons = normalizeOccultationBandOutput(polygons)
	if pointSource {
		polygons = RemoveTinyPolygonComponents(polygons)
	}
	polygons = densifyOccultationPolygons(polygons, 30)
	if len(polygons) == 0 {
		if fallback, ok := fallbackVisible(); ok {
			return fallback, false, nil
		}
		return nil, false, fmt.Errorf("analytic occultation boundary produced no polygon")
	}
	if len(phaseLines) > 0 && !geodata.SphericalPolygonsContainPathsWithinKM(polygons, phaseLines, false, 2) {
		// A coarse limit strip reliably identifies the main face, but it need not
		// reach a narrow face that terminates where a rise/set phase meets the
		// contact or visibility envelope. Use instantaneous footprints only to
		// decide which side of each phase curve is physically inside, then repeat
		// face selection against the unchanged analytic boundary network.
		phaseCoverage := occultationCurveCoverageProbes(curves, getFootprintFill())
		if len(phaseCoverage) > 0 {
			// 相位重试必须与首遍走同一个 polygonize 闭包：单独硬编码节点吸附容差会让
			// 恰好依赖有限圆盘 40 km 吸附的场景静默退回 footprint-sweep。
			selected, selectionErr := polygonize(phaseCoverage)
			if selectionErr != nil {
				if fallback, ok := fallbackVisible(); ok {
					return fallback, false, nil
				}
				return nil, false, fmt.Errorf("analytic occultation boundary phase selection: %w", selectionErr)
			}
			selected = normalizeOccultationBandOutput(selected)
			if pointSource {
				selected = RemoveTinyPolygonComponents(selected)
			}
			selected = densifyOccultationPolygons(selected, 30)
			if len(selected) > 0 {
				polygons = selected
			}
		}
	}
	if len(phaseLines) > 0 && !geodata.SphericalPolygonsContainPathsWithinKM(polygons, phaseLines, false, 2) {
		miss := geodata.SphericalPolygonsPathMissDistanceKM(polygons, phaseLines, false)
		// When no visibility contour exists, the event is visible throughout the
		// contact envelope and there is no H=0 transition to close. A small phase
		// residual can remain where the sampled contact envelope meets a rise/set
		// branch; keep the analytic face if that residual is below one rendered
		// edge. Events with visibility contours retain the strict 2 km invariant.
		if !pointSource && len(visibilityContours) == 0 && miss <= 40 {
			return polygons, true, nil
		}
		if fallback, ok := fallbackVisible(); ok {
			return fallback, false, nil
		}
		return nil, false, fmt.Errorf(
			"analytic occultation boundary misses a rise/set phase by %.1f km",
			miss,
		)
	}
	return polygons, true, nil
}

func occultationRiseSetBoundaryLines(
	curves []basic.OccultationRiseSetCurve,
) [][]geodata.GeoPoint {
	lines := make([][]geodata.GeoPoint, 0, len(curves)*2)
	for _, curve := range curves {
		lines = append(lines, occultationCurveBoundaryLines(curve)...)
	}
	return lines
}

func normalizeOccultationBandOutput(polygons [][]geodata.GeoPoint) [][]geodata.GeoPoint {
	if len(polygons) == 0 {
		return nil
	}
	result := make([][]geodata.GeoPoint, 0, len(polygons))
	for _, polygon := range polygons {
		open := openFootprintRing(polygon)
		if len(open) < 3 || math.Abs(geoRingArea(open)) <= 1e-12 {
			continue
		}
		result = append(result, polygon)
	}
	return result
}
