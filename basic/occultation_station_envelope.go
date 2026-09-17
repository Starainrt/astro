package basic

import (
	"math"
	"time"
)

// occultationStationDensifyContours applies the map's projection metric after
// station correction. The correction can move adjacent geocentric samples by
// tens of kilometres, so a contour that was dense before correction can still
// contain a long Web Mercator chord. A spatial arclength plane selects the
// midpoint branch even where time reverses and a fixed-time solve is singular.
// The refined curve is split again at time folds for the public time contract.
func occultationStationDensifyContours(
	contours [][]OccultationPathPoint,
	cache *occultationRiseSetEvaluationCache,
	location *time.Location,
) [][]OccultationPathPoint {
	if len(contours) == 0 || cache == nil {
		return contours
	}
	const (
		targetSpacingKM = 30.0
		maxDepth        = 10
	)
	result := make([][]OccultationPathPoint, 0, len(contours))
	for _, contour := range contours {
		if len(contour) < 2 {
			continue
		}
		var refine func(OccultationPathPoint, OccultationPathPoint, int, *[]OccultationPathPoint)
		refine = func(start, end OccultationPathPoint, depth int, output *[]OccultationPathPoint) {
			if depth >= maxDepth || occultationStationProjectedSpacingKM(start, end) <= targetSpacingKM {
				*output = append(*output, end)
				return
			}
			startTT, endTT := centerTimeTT(start.Time), centerTimeTT(end.Time)
			chord := [3]float64{math.Remainder(end.Longitude-start.Longitude, 360), end.Latitude - start.Latitude,
				(endTT - startTT) * occultationStationEnvelopeTimeScale}
			length := math.Sqrt(dotSolarEclipse3(chord, chord))
			if length <= 1e-12 {
				*output = append(*output, end)
				return
			}
			predictor := [3]float64{start.Longitude + chord[0]/2, start.Latitude + chord[1]/2, chord[2] / 2}
			for index := range chord {
				chord[index] /= length
			}
			midpoint, _, ok := occultationStationCorrectEnvelopeArc(predictor, chord, startTT, cache,
				occultationStationEnvelopeModel{kind: occultationStationContactEnvelope}, (start.WidthKM+end.WidthKM)/2, location)
			if ok && occultationPathDistanceKM(start, midpoint.point) < occultationPathDistanceKM(start, end) &&
				occultationPathDistanceKM(midpoint.point, end) < occultationPathDistanceKM(start, end) {
				refine(start, midpoint.point, depth+1, output)
				refine(midpoint.point, end, depth+1, output)
				return
			}
			// Leave a failed local solve untouched. Retrying with an unrelated
			// branch would be less accurate than retaining the source sample.
			*output = append(*output, end)
		}
		refined := []OccultationPathPoint{contour[0]}
		for index := 1; index < len(contour); index++ {
			refine(refined[len(refined)-1], contour[index], 0, &refined)
		}
		result = append(result, occultationStationSplitEnvelopeAtTimeFolds(refined)...)
	}
	return result
}

func occultationStationProjectedSpacingKM(first, second OccultationPathPoint) float64 {
	const maxLatitude = 85.05112878
	firstLatitude := math.Max(-maxLatitude, math.Min(maxLatitude, first.Latitude)) * rad
	secondLatitude := math.Max(-maxLatitude, math.Min(maxLatitude, second.Latitude)) * rad
	longitudeDelta := math.Remainder(second.Longitude-first.Longitude, 360) * rad
	firstY := math.Log(math.Tan(math.Pi/4 + firstLatitude/2))
	secondY := math.Log(math.Tan(math.Pi/4 + secondLatitude/2))
	return 6378.1366 * math.Hypot(longitudeDelta, secondY-firstY)
}
