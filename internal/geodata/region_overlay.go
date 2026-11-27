package geodata

// SubtractPolygon 在固定经度图幅中扣除内环表示的小侧 / subtract the inner ring's minor side in a fixed longitude chart.
func SubtractPolygon(outer, inner []GeoPoint) ([]Region, error) {
	subject, excluded := canonicalPolygonRegions(outer), canonicalPolygonRegions(inner)
	if len(subject) == 0 || len(excluded) == 0 {
		return subject, nil
	}
	return overlayCanonicalRegions(subject, excluded, regionSubtract)
}

// IntersectPolygon 在固定经度图幅中求两环小侧的交集 / intersect the minor sides of two rings in a fixed longitude chart.
func IntersectPolygon(first, second []GeoPoint) ([]Region, error) {
	subject, clip := canonicalPolygonRegions(first), canonicalPolygonRegions(second)
	if len(subject) == 0 || len(clip) == 0 {
		return nil, nil
	}
	return overlayCanonicalRegions(subject, clip, regionIntersect)
}

func canonicalPolygonRegions(ring []GeoPoint) []Region {
	var regions []Region
	for _, piece := range PolygonFragments(ring, ClipView{Projection: ProjectionEquirectangular}) {
		if stablePolygonArea(piece) < 0 {
			for first, last := 0, len(piece)-1; first < last; first, last = first+1, last-1 {
				piece[first], piece[last] = piece[last], piece[first]
			}
		}
		regions = append(regions, Region{Outer: piece})
	}
	return regions
}

// UnionRegionSets 合并带孔的已切缝区域集合 / union sets of seam-split regions with holes.
func UnionRegionSets(sets [][]Region) ([]Region, error) {
	switch len(sets) {
	case 0:
		return nil, nil
	case 1:
		return sets[0], nil
	}
	middle := len(sets) / 2
	left, err := UnionRegionSets(sets[:middle])
	if err != nil {
		return nil, err
	}
	right, err := UnionRegionSets(sets[middle:])
	if err != nil {
		return nil, err
	}
	if len(left) == 0 {
		return right, nil
	}
	if len(right) == 0 {
		return left, nil
	}
	return polygonUnionMergeRegions(left, right)
}

// RegionOctants 将已切缝区域裁到球面八分区 / clips a seam-split region to spherical octants.
func RegionOctants(region Region) ([]Region, error) {
	var result []Region
	for west := -180.0; west < 180; west += 90 {
		for south := -90.0; south < 90; south += 90 {
			clip := Region{Outer: []GeoPoint{{west, south}, {west + 90, south}, {west + 90, south + 90}, {west, south + 90}}}
			pieces, err := overlayCanonicalRegions([]Region{region}, []Region{clip}, regionIntersect)
			if err != nil {
				return nil, err
			}
			result = append(result, pieces...)
		}
	}
	return result, nil
}
