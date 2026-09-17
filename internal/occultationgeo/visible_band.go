package occultationgeo

import (
	"github.com/starainrt/astro/basic"
	"github.com/starainrt/astro/internal/geodata"
)

func visibleBandPolygons(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
	contours [][]basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
	strongPolarSmoothing bool,
) ([][]geodata.GeoPoint, bool, error) {
	contactLines := occultationContactContourBoundaryLines(contours)
	useContactContours := len(contactLines) > 0
	// Only a single north/south contour pair is a stable global envelope. A
	// contour set split into several branch fragments (common near polar folds)
	// must retain the sampled visible-union path; feeding those fragments to the
	// generic linework selector creates one polygon per numerical branch and is
	// both slower and less faithful than the time-union sweep.
	stableContactEnvelope := useContactContours && len(contours) == 2
	// A station-corrected north/south contact pair already parameterizes the
	// complete time-union boundary. Clip each adjacent time cell to the lunar
	// horizon and merge those cells directly. This is the physical definition
	// of the visible band; it avoids asking a planar polygonizer to choose among
	// folded polar faces and keeps partial/total bands on the same residual
	// model. Footprints remain independent witnesses/timeline data.
	var (
		fallback             [][]geodata.GeoPoint
		swept                [][]geodata.GeoPoint
		sweepErr             error
		sweepComputed        bool
		visibleUnion         [][]geodata.GeoPoint
		visibleUnionComputed bool
	)
	computeSweep := func() ([][]geodata.GeoPoint, error) {
		if !sweepComputed {
			swept, sweepErr = footprintSweepPolygons(footprints, northern, southern)
			sweepComputed = true
		}
		return swept, sweepErr
	}
	computeVisibleUnion := func() [][]geodata.GeoPoint {
		if !visibleUnionComputed {
			visibleUnion = footprintVisibleUnionPolygons(footprints)
			visibleUnionComputed = true
		}
		return visibleUnion
	}
	if !useContactContours && len(footprints) > 0 && len(curves) > 0 {
		// Footprints are already clipped to the lunar horizon at each sampled
		// instant. With no supplied contact contour there is no alternative
		// analytic boundary to select, so the continuous footprint sweep is the
		// authoritative sampled time-union rather than a legacy fallback.
		swept, err := computeSweep()
		if err != nil {
			return nil, false, err
		}
		if len(swept) > 0 {
			return cleanupOccultationAuthoritativeBandPolygons(swept), true, nil
		}
	}
	if len(curves) == 0 && len(footprints) > 0 {
		// Without rise/set curves there is no closed phase boundary to select a
		// face from. Return the horizon-visible footprint sweep directly, even
		// when contact contours are present; feeding an open contour network to
		// the polygonizer can select a complement face and drop the greatest
		// point from the static band.
		fallback, err := computeSweep()
		if err != nil {
			return nil, false, err
		}
		// The sweep has already passed the ordinary topology cleanup; deleting
		// another span here can remove intermediate samples around a polar fold
		// and recreate a long straight chord. Densify the cleaned ring instead.
		fallback = densifyOccultationPolygons(fallback, 50)
		for index := range fallback {
			fallback[index] = smoothOccultationHairpins(fallback[index], 180, 25, 16, 50)
			fallback[index] = removeOccultationSharpCorners(fallback[index], 20, 30)
		}
		fallback = mergeStaticFootprintRepairs(fallback, footprints)
		// If the sparse open sweep is split at horizon transitions, add only the
		// already-closed instantaneous footprints. This restores a closed face
		// around greatest without running a boolean union over every sampled
		// footprint (which is prohibitively expensive for historical events).
		closed := footprintClosedPolygons(footprints)
		if len(closed) > 0 {
			input := append(append([][]geodata.GeoPoint(nil), fallback...), closed...)
			if merged, mergeErr := geodata.UnionPolygons(input); mergeErr == nil && len(merged) > 0 {
				return cleanupFootprintSweepPolygons(merged, northern, southern), false, nil
			}
		}
		return fallback, false, nil
	}
	if len(footprints) > 0 && !stableContactEnvelope {
		if candidate, authoritative, handled := occultationEarlyVisibleBandCandidate(
			footprints, northern, southern, contours, curves,
			useContactContours, strongPolarSmoothing, computeSweep, computeVisibleUnion,
		); handled {
			return candidate, authoritative, nil
		}
	}
	if useContactContours && len(contours) > 2 {
		// Split contact contours describe several valid numerical branches. A
		// planar face selector cannot reliably decide their complement, while the
		// horizon-clipped sweep is a deterministic union of the same instantaneous
		// states. Keep that union authoritative when it remains one component:
		// mutually discontinuous contour fragments cannot witness its boundary, so
		// the single-component construction contract is the applicable check here.
		if swept, err := computeSweep(); err == nil && len(swept) == 1 {
			return cleanupOccultationAuthoritativeBandPolygons(swept), true, nil
		}
	}
	if !useContactContours {
		var err error
		fallback, err = computeSweep()
		if err != nil {
			return nil, false, err
		}
	}
	baseBoundaryLines := occultationFallbackPolygonBoundaryLines(fallback)
	// Greatest is a diagnostic meridian, not an exterior edge of the static
	// visible band. Feeding it to the polygonizer lets a polar fold be selected
	// as a legitimate face boundary, which creates a sharp polar-end kink. Keep
	// it in curve coverage probes and in
	// the exported line features, but exclude it from the authoritative fill
	// topology whenever continuous contact contours are available.
	boundaryCurves := curves
	if useContactContours {
		boundaryCurves = occultationStaticBandCurves(curves)
	}
	// Horizon connectors are part of the authoritative visible boundary, not
	// merely a last-resort repair. When the contact contours and rise/set
	// curves are both open at a moonrise/moonset transition, omitting this
	// short physical arc lets polygonization succeed on a different face and
	// leaves the exported phase curve outside (or inside) the filled band.
	initialConnectorLines := occultationHorizonConnectorBoundaryLines(
		HorizonConnectorSegments(footprints, boundaryCurves, northern, southern),
	)
	if useContactContours {
		baseBoundaryLines = contactLines
	}
	visibleFill, coveragePaths := occultationVisibleFillAndCoverage(footprints)
	contourFill, _ := occultationContourFillAndCoverage(useContactContours, northern, southern, nil)
	// Phase-cycle construction and its spherical witness checks are expensive,
	// especially for minute-sampled curves. Several fallback branches need
	// the same candidate, so build and validate it once per band generation.
	var (
		preferredPhaseBandRaw       [][]geodata.GeoPoint
		preferredPhaseBandRawOK     bool
		preferredPhaseBandComputed  bool
		preferredPhaseBandCandidate [][]geodata.GeoPoint
		preferredPhaseBandAccepted  bool
	)
	computePreferredPhaseBand := func() ([][]geodata.GeoPoint, bool) {
		if !preferredPhaseBandComputed {
			preferredPhaseBandRaw, preferredPhaseBandRawOK = occultationPreferredPhaseBand(
				useContactContours, footprints, curves, northern, southern, contourFill,
			)
			if preferredPhaseBandRawOK {
				physicalBoundary := preferredPhaseBandRaw
				candidate := cleanupOccultationAuthoritativeBandPolygons(preferredPhaseBandRaw)
				candidate = preserveOccultationPhaseBoundaryEnvelope(candidate, physicalBoundary)
				if occultationPreferredPhaseBandAccepted(candidate, contourFill, visibleFill) {
					preferredPhaseBandCandidate = candidate
					preferredPhaseBandAccepted = true
				}
			}
			preferredPhaseBandComputed = true
		}
		return preferredPhaseBandRaw, preferredPhaseBandRawOK
	}
	computeAcceptedPreferredPhaseBand := func() ([][]geodata.GeoPoint, bool) {
		computePreferredPhaseBand()
		return preferredPhaseBandCandidate, preferredPhaseBandAccepted
	}
	var (
		contactSweepLines         [][]geodata.GeoPoint
		contactSweepLinesComputed bool
	)
	computeContactSweepLines := func() [][]geodata.GeoPoint {
		if !contactSweepLinesComputed {
			contactSweepLines = ContactSweepBoundaryLines(footprints)
			contactSweepLinesComputed = true
		}
		return contactSweepLines
	}
	boundaryLines := occultationVisibleBoundaryLinesFromBase(baseBoundaryLines, boundaryCurves, initialConnectorLines)

	if preferred, ok := occultationPreferredContactBandCandidate(
		useContactContours, computePreferredPhaseBand, computeAcceptedPreferredPhaseBand,
		computeVisibleUnion, visibleFill, contourFill, coveragePaths, curves,
	); ok {
		return preferred, true, nil
	}
	if fast, ok := occultationAcceptedFastBandCandidate(
		useContactContours, strongPolarSmoothing, boundaryLines,
		visibleFill, contourFill, coveragePaths, curves, computePreferredPhaseBand,
	); ok {
		return fast, true, nil
	}
	if combined, ok := occultationContactSweepPhaseCandidate(
		useContactContours, footprints, computeSweep, computePreferredPhaseBand,
		visibleFill, contourFill, coveragePaths, curves,
	); ok {
		return combined, true, nil
	}
	if candidate, authoritative, ok := occultationBoundedContactSweepCandidate(
		useContactContours, computeSweep, contours, strongPolarSmoothing,
	); ok {
		return candidate, authoritative, nil
	}
	if len(visibleFill) == 0 && len(contourFill) == 0 {
		return occultationNoVisibleBandFill(
			useContactContours, computeSweep, fallback, strongPolarSmoothing,
		)
	}
	selectionFill, coveragePaths, curveCoveragePaths, visibleFillCoveragePaths :=
		occultationVisibleBandLineworkInputs(
			useContactContours, strongPolarSmoothing,
			visibleFill, contourFill, coveragePaths, curves,
		)
	visible, lineworkErr := occultationRetryVisibleBandLinework(
		occultationVisibleBandLineworkOptions{
			useContactContours:       useContactContours,
			footprints:               footprints,
			northern:                 northern,
			southern:                 southern,
			curves:                   curves,
			boundaryCurves:           boundaryCurves,
			boundaryLines:            boundaryLines,
			baseBoundaryLines:        baseBoundaryLines,
			selectionFill:            selectionFill,
			coveragePaths:            coveragePaths,
			curveCoveragePaths:       curveCoveragePaths,
			visibleFill:              visibleFill,
			visibleFillCoveragePaths: visibleFillCoveragePaths,
			initialConnectorLines:    initialConnectorLines,
			contactLines:             contactLines,
			computeSweep:             computeSweep,
			computeContactSweepLines: computeContactSweepLines,
		},
	)
	visible = occultationMergePreferredVisibleBand(
		visible, useContactContours, footprints, curves, northern, southern,
		contourFill, visibleFill, visibleFillCoveragePaths,
	)
	if lineworkErr == nil {
		lineworkErr = validateOccultationVisibleBandWitnesses(
			visible, useContactContours, visibleFill, contourFill, visibleFillCoveragePaths,
		)
	}
	if lineworkErr != nil && useContactContours && len(visibleFill) > 0 {
		if stableVisible, stableErr := occultationRetryStableContactBoundary(
			occultationVisibleBandLineworkOptions{
				useContactContours:       useContactContours,
				boundaryCurves:           boundaryCurves,
				visibleFill:              visibleFill,
				visibleFillCoveragePaths: visibleFillCoveragePaths,
				initialConnectorLines:    initialConnectorLines,
				computeSweep:             computeSweep,
				computeContactSweepLines: computeContactSweepLines,
			},
		); stableErr == nil {
			visible, lineworkErr = stableVisible, nil
		}
	}
	if lineworkErr != nil {
		if useContactContours {
			// A direct open-footprint sweep is a coverage-preserving fallback only.
			// It does not share the phase/contact cycle used by the authoritative
			// boundary, so never let it replace a successful contour polygonization.
			if direct, directOK := DirectVisibleBandPolygons(footprints); directOK {
				return cleanupOccultationVisibleBandPolygons(direct, strongPolarSmoothing), false, nil
			}
			if fallback, err := computeSweep(); err == nil {
				if authoritative, ok := authoritativeFallbackBandPolygons(fallback, contours, strongPolarSmoothing); ok {
					return authoritative, true, nil
				}
				return cleanupOccultationVisibleBandPolygons(fallback, strongPolarSmoothing), false, nil
			}
			if len(visibleFill) > 0 {
				return cleanupOccultationVisibleBandPolygons(visibleFill, strongPolarSmoothing), false, nil
			}
			return nil, false, lineworkErr
		}
		return cleanupOccultationVisibleBandPolygons(fallback, strongPolarSmoothing), false, nil
	}
	if lineworkErr == nil && useContactContours && len(visibleFill) > 0 {
		// A folded phase cycle can be topologically valid yet select an inner
		// branch at a polar horizon. The instantaneous visible footprints provide
		// an independent witness for the actual outer envelope. Promote their
		// union only when it substantially covers the phase cycle while the phase
		// cycle does not cover the union; this keeps ordinary smooth cycles on the
		// cheaper contact/rise-set path.
		if footprintUnion, unionErr := geodata.UnionPolygons(visibleFill); unionErr == nil && len(footprintUnion) > 0 {
			phaseMiss := geodata.SphericalPolygonsPathMissDistanceKM(visible, footprintUnion, true)
			unionMiss := geodata.SphericalPolygonsPathMissDistanceKM(footprintUnion, visible, true)
			if phaseMiss > 20 && unionMiss <= 50 {
				visible = footprintUnion
			}
		}
	}
	if !useContactContours {
		// The phase-boundary polygonizer supplies the smooth global outline, but a
		// polar fold can leave a very narrow face out of its selected cycle. Merge
		// the horizon-clipped instantaneous faces back into that result: every
		// added point is already Moon-altitude-visible, so this repairs coverage
		// without restoring the below-horizon part of the fallback sweep.
		combinedVisible := append([][]geodata.GeoPoint(nil), visible...)
		combinedVisible = append(combinedVisible, visibleFill...)
		if augmented, unionErr := geodata.UnionPolygons(combinedVisible); unionErr == nil {
			visible = augmented
		}
	}
	if useContactContours {
		return roundOccultationAuthoritativeBandJunctions(
			cleanupOccultationAuthoritativeBandPolygons(visible),
		), true, nil
	}
	return cleanupOccultationVisibleBandPolygons(visible, strongPolarSmoothing), true, nil
}

func occultationPreferredContactBandCandidate(
	useContactContours bool,
	computePreferred func() ([][]geodata.GeoPoint, bool),
	computeAcceptedPreferred func() ([][]geodata.GeoPoint, bool),
	computeVisibleUnion func() [][]geodata.GeoPoint,
	visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool) {
	if !useContactContours {
		return nil, false
	}
	// The static display contract combines every horizon-visible footprint with
	// the exported start/end phase envelope. Try that complete physical union
	// before a generic polygonizer can select a folded complement face.
	if preferred, ok := computePreferred(); ok {
		if visibleUnion := computeVisibleUnion(); len(visibleUnion) > 0 {
			combined := append([][]geodata.GeoPoint(nil), visibleUnion...)
			combined = append(combined, preferred...)
			if merged, err := geodata.UnionPolygons(combined); err == nil && len(merged) > 0 {
				merged = cleanupOccultationAuthoritativeBandPolygons(merged)
				merged = preserveOccultationPhaseBoundaryEnvelope(merged, preferred)
				if occultationFastPolarBandAccepted(
					merged, visibleFill, contourFill, coveragePaths, curves,
				) {
					return roundOccultationAuthoritativeBandJunctions(merged), true
				}
			}
		}
	}
	// A complete start/end phase cycle is itself the physical outer boundary of
	// a horizon-clipped band. The acceptance gate has already checked its
	// contour and footprint witnesses.
	if preferred, ok := computeAcceptedPreferred(); ok {
		return roundOccultationAuthoritativeBandJunctions(preferred), true
	}
	return nil, false
}

func occultationAcceptedFastBandCandidate(
	useContactContours, strongPolarSmoothing bool,
	boundaryLines, visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
	computePreferred func() ([][]geodata.GeoPoint, bool),
) ([][]geodata.GeoPoint, bool) {
	visible, ok := occultationFastPolarBand(
		useContactContours, strongPolarSmoothing, boundaryLines,
		visibleFill, contourFill, coveragePaths, curves,
	)
	if !ok {
		return nil, false
	}
	visible = cleanupOccultationAuthoritativeBandPolygons(visible)
	if occultationFastPolarBandAccepted(visible, visibleFill, contourFill, coveragePaths, curves) {
		return visible, true
	}
	if !useContactContours {
		return nil, false
	}
	// A fast polar face can choose an inner branch. Union it with the explicit
	// phase cycle and accept only when the same physical witnesses still pass.
	preferred, preferredOK := computePreferred()
	if !preferredOK {
		return nil, false
	}
	combined := append([][]geodata.GeoPoint(nil), visible...)
	combined = append(combined, preferred...)
	merged, err := geodata.UnionPolygons(combined)
	if err != nil {
		return nil, false
	}
	merged = cleanupOccultationAuthoritativeBandPolygons(merged)
	merged = preserveOccultationPhaseBoundaryEnvelope(merged, preferred)
	if !occultationFastPolarBandAccepted(merged, visibleFill, contourFill, coveragePaths, curves) {
		return nil, false
	}
	return merged, true
}

func occultationContactSweepPhaseCandidate(
	useContactContours bool,
	footprints []basic.OccultationFootprint,
	computeSweep func() ([][]geodata.GeoPoint, error),
	computePreferred func() ([][]geodata.GeoPoint, bool),
	visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool) {
	if !useContactContours {
		return nil, false
	}
	// The normal sweep is cheaper; the direct sweep is retained for ribbons
	// whose cross-arcs need its denser union. Both must be joined to the same
	// accepted phase family and pass identical witnesses.
	if swept, err := computeSweep(); err == nil {
		if merged, ok := occultationMergeSweepWithPreferredPhase(
			swept, computePreferred, visibleFill, contourFill, coveragePaths, curves,
		); ok {
			return roundOccultationAuthoritativeBandJunctions(merged), true
		}
	}
	if direct, ok := DirectVisibleBandPolygons(footprints); ok {
		if merged, ok := occultationMergeSweepWithPreferredPhase(
			direct, computePreferred, visibleFill, contourFill, coveragePaths, curves,
		); ok {
			return roundOccultationAuthoritativeBandJunctions(merged), true
		}
	}
	return nil, false
}

func occultationMergeSweepWithPreferredPhase(
	sweep [][]geodata.GeoPoint,
	computePreferred func() ([][]geodata.GeoPoint, bool),
	visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool) {
	if len(sweep) == 0 {
		return nil, false
	}
	preferred, ok := computePreferred()
	if !ok {
		return nil, false
	}
	combined := append([][]geodata.GeoPoint(nil), sweep...)
	combined = append(combined, preferred...)
	merged, err := geodata.UnionPolygons(combined)
	if err != nil {
		return nil, false
	}
	merged = cleanupOccultationAuthoritativeBandPolygons(merged)
	merged = preserveOccultationPhaseBoundaryEnvelope(merged, preferred)
	if !occultationFastPolarBandAccepted(merged, visibleFill, contourFill, coveragePaths, curves) {
		return nil, false
	}
	return merged, true
}

func occultationBoundedContactSweepCandidate(
	useContactContours bool,
	computeSweep func() ([][]geodata.GeoPoint, error),
	contours [][]basic.OccultationPathPoint,
	strongPolarSmoothing bool,
) ([][]geodata.GeoPoint, bool, bool) {
	if !useContactContours {
		return nil, false, false
	}
	swept, err := computeSweep()
	if err != nil || len(swept) == 0 {
		return nil, false, false
	}
	// Bound the expensive linework retry tree with the already computed,
	// coverage-preserving temporal sweep. It is authoritative only when the
	// continuous contact contours independently witness its boundary.
	if authoritative, ok := authoritativeFallbackBandPolygons(
		swept, contours, strongPolarSmoothing,
	); ok {
		return authoritative, true, true
	}
	cleaned := cleanupOccultationVisibleBandPolygons(swept, strongPolarSmoothing)
	if len(cleaned) == 0 {
		return nil, false, false
	}
	return cleaned, false, true
}

func occultationEarlyVisibleBandCandidate(
	footprints []basic.OccultationFootprint,
	northern, southern []basic.OccultationPathPoint,
	contours [][]basic.OccultationPathPoint,
	curves []basic.OccultationRiseSetCurve,
	useContactContours, strongPolarSmoothing bool,
	computeSweep func() ([][]geodata.GeoPoint, error),
	computeVisibleUnion func() [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool, bool) {
	// A split contour set can still use the bounded endpoint fallback when
	// north/south branches remain paired. Odd or incomplete branches continue
	// through the phase/linework selector for a smooth envelope.
	if useContactContours && len(contours) > 2 && len(contours)%2 == 0 {
		contourFill, _ := occultationContourFillAndCoverage(true, northern, southern, nil)
		if fallback, ok := footprintEndpointContourFallback(footprints, contourFill); ok && len(fallback) == 1 {
			// A continuous sweep is more faithful when it proves that all
			// horizon-visible source faces belong to the same temporal ribbon.
			if swept, sweepErr := computeSweep(); sweepErr == nil && len(swept) == 1 {
				visibleUnion := computeVisibleUnion()
				if len(visibleUnion) > 1 && footprintSweepCoversSamples(swept, footprints) {
					return cleanupOccultationVisibleBandPolygons(swept, strongPolarSmoothing), true, true
				}
			}
			return fallback, true, true
		}
	}
	sweepRequiresLinework := false
	// Prefer one continuous temporal sweep over independently closed
	// instantaneous footprints. The sweep carries the moving contact arc
	// between samples, avoiding a staircase from a single horizon closure.
	if swept, sweepErr := computeSweep(); sweepErr == nil && len(swept) > 0 {
		// The same contact-contour witness used by the later linework fallback can
		// be checked before constructing the expensive horizon-visible union. When
		// it accepts a single sweep, the union cannot change the selected band: its
		// only purpose on that path is to discover a fragmented mask that is already
		// implied by the witnessed contact envelope.
		if useContactContours && len(swept) == 1 && hasBracketedClosedFootprintRun(footprints) {
			if authoritative, authoritativeOK := authoritativeFallbackBandPolygons(
				swept, contours, strongPolarSmoothing,
			); authoritativeOK {
				return authoritative, true, true
			}
		}
		visibleUnion := computeVisibleUnion()
		if len(visibleUnion) > 0 && footprintSweepNeedsHorizonClipping(swept, visibleUnion, footprints) {
			if len(visibleUnion) > 1 && len(swept) == 1 && footprintSweepCoversSamples(swept, footprints) {
				return cleanupOccultationVisibleBandPolygons(swept, strongPolarSmoothing), true, true
			}
			outsideVisibleKM := 0.0
			// A fragmented union already decides the linework path; avoid the
			// expensive edge-distance calculation in that case.
			if !(len(visibleUnion) > 1 && len(swept) == 1) {
				outsideVisibleKM = geodata.SphericalPolygonsPathMissDistanceKM(visibleUnion, swept, true)
			}
			if len(visibleUnion) > 1 && len(swept) == 1 {
				sweepRequiresLinework = len(curves) > 0
			}
			if sweepRequiresLinework && useContactContours {
				if authoritative, authoritativeOK := authoritativeFallbackBandPolygons(
					swept, contours, strongPolarSmoothing,
				); authoritativeOK {
					return authoritative, true, true
				}
			}
			if outsideVisibleKM > 20 && len(curves) > 0 {
				sweepRequiresLinework = true
			}
			if outsideVisibleKM <= 20 && len(curves) > 0 && len(northern) > 0 && len(southern) > 0 {
				contourFill, _ := occultationContourFillAndCoverage(true, northern, southern, nil)
				if phaseBand, ok := occultationPreferredPhaseBand(
					true, footprints, curves, northern, southern, contourFill,
				); ok {
					combined := append([][]geodata.GeoPoint(nil), swept...)
					combined = append(combined, phaseBand...)
					if merged, mergeErr := geodata.UnionPolygons(combined); mergeErr == nil && len(merged) > 0 {
						return cleanupOccultationAuthoritativeBandPolygons(merged), true, true
					}
				}
			}
			// A sweep entering the below-horizon complement cannot be repaired by
			// splitting at classified vertices without manufacturing polar faces.
			visibleUnion = cleanupOccultationFootprintUnionPolygons(visibleUnion, strongPolarSmoothing)
			if len(visibleUnion) == 1 {
				return visibleUnion, true, true
			}
			if !sweepRequiresLinework {
				return cleanupOccultationVisibleBandPolygons(swept, strongPolarSmoothing), true, true
			}
		}
	}
	// Legacy footprints without open boundary arcs retain their exact visible
	// union as a bounded fallback when continuous linework is unnecessary.
	if !sweepRequiresLinework {
		if visibleUnion := computeVisibleUnion(); len(visibleUnion) > 0 {
			return cleanupOccultationFootprintUnionPolygons(
				visibleUnion, strongPolarSmoothing && !useContactContours,
			), true, true
		}
	}
	return nil, false, false
}

func hasBracketedClosedFootprintRun(footprints []basic.OccultationFootprint) bool {
	firstClosed, lastClosed := -1, -1
	for index, footprint := range footprints {
		if !footprint.Closed {
			continue
		}
		if firstClosed < 0 {
			firstClosed = index
		}
		lastClosed = index
	}
	return firstClosed >= 3 && lastClosed > firstClosed && lastClosed+3 < len(footprints)
}

// footprintEndpointContourFallback is the bounded fallback for a contact
// envelope split into incompatible branches. The paired contact fill carries
// the interval interior while the open-footprint sweep provides temporal caps.
// Unlike a union of every instantaneous horizon polygon, this construction
// cannot turn small sampling seams into hundreds of static-band components.
// The sweep is authoritative only when its boundary is independently witnessed
// by the contact contours supplied by the caller.
func footprintEndpointContourFallback(
	footprints []basic.OccultationFootprint,
	contourFill [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool) {
	if len(footprints) == 0 || len(contourFill) == 0 {
		return nil, false
	}
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
	endpointOutlines, err := geodata.OpenBoundaryEndpointOutlines(samples)
	if err != nil || len(endpointOutlines) == 0 {
		return nil, false
	}
	inputs := append([][]geodata.GeoPoint(nil), contourFill...)
	inputs = append(inputs, endpointOutlines...)
	merged, err := geodata.UnionPolygons(inputs)
	if err != nil || len(merged) == 0 {
		return nil, false
	}
	// A horizon transition can make the full visible-source union fail at one
	// numerically open edge. Recover the missing time-union lobe by greedily
	// adding only source faces that (a) have vertices outside the current face
	// and (b) union into that same single connected face. Detached components and
	// the malformed source are skipped, so this cannot turn a local repair into
	// a collection of sampled fragments.
	merged = mergeEndpointFallbackVisibleSources(merged, footprints)
	merged = normalizeOccultationBandOutput(merged)
	if len(merged) == 0 {
		return nil, false
	}
	return densifyOccultationPolygons(merged, 40), true
}

func mergeEndpointFallbackVisibleSources(
	base [][]geodata.GeoPoint,
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	if len(base) != 1 || len(base[0]) < 3 {
		return base
	}
	sources := occultationVisibleFootprintFillOnly(footprints)
	if len(sources) == 0 {
		return base
	}
	merged := base
	accepted := 0
	const maximumAcceptedSources = 48
	// Start at the final temporal samples: endpoint fallback already carries the
	// first/last contact caps, so the missing lobe is normally adjacent to one
	// of the last visible faces. Working backwards avoids adding an unrelated
	// early-time face that can move the opposite boundary by a few kilometres.
	for sourceIndex := len(sources) - 1; sourceIndex >= 0; sourceIndex-- {
		source := sources[sourceIndex]
		if accepted >= maximumAcceptedSources || len(source) < 3 {
			break
		}
		// Skip faces already covered by the current candidate. Sampling vertices
		// keeps this gate bounded while still catching a narrow endpoint lobe.
		outside := false
		touches := false
		step := (len(source) + 31) / 32
		if step < 1 {
			step = 1
		}
		for index := 0; index < len(source); index += step {
			point := source[index]
			if !geodata.SphericalPolygonsContainPoints(merged, []geodata.GeoPoint{point})[0] {
				outside = true
				_, distance := nearestPolygonBoundaryPoint(merged, point)
				if distance <= 120 {
					touches = true
				}
			}
		}
		if !outside || !touches {
			continue
		}
		input := append([][]geodata.GeoPoint(nil), merged...)
		input = append(input, source)
		candidate, err := geodata.UnionPolygons(input)
		if err != nil || len(candidate) != 1 {
			continue
		}
		merged = candidate
		accepted++
	}
	return merged
}

type occultationVisibleBandLineworkOptions struct {
	useContactContours       bool
	footprints               []basic.OccultationFootprint
	northern                 []basic.OccultationPathPoint
	southern                 []basic.OccultationPathPoint
	curves                   []basic.OccultationRiseSetCurve
	boundaryCurves           []basic.OccultationRiseSetCurve
	boundaryLines            [][]geodata.GeoPoint
	baseBoundaryLines        [][]geodata.GeoPoint
	selectionFill            [][]geodata.GeoPoint
	coveragePaths            [][]geodata.GeoPoint
	curveCoveragePaths       [][]geodata.GeoPoint
	visibleFill              [][]geodata.GeoPoint
	visibleFillCoveragePaths [][]geodata.GeoPoint
	initialConnectorLines    [][]geodata.GeoPoint
	contactLines             [][]geodata.GeoPoint
	computeSweep             func() ([][]geodata.GeoPoint, error)
	computeContactSweepLines func() [][]geodata.GeoPoint
}

// occultationRetryVisibleBandLinework owns the bounded candidate sequence for
// events that have no accepted direct phase cycle. Keeping the retries in one
// helper makes their order explicit and keeps visibleBandPolygons focused on
// selecting between the fast physical constructions and this slow path.
func occultationRetryVisibleBandLinework(
	options occultationVisibleBandLineworkOptions,
) ([][]geodata.GeoPoint, error) {
	useContactContours := options.useContactContours
	footprints := options.footprints
	northern, southern := options.northern, options.southern
	curves, boundaryCurves := options.curves, options.boundaryCurves
	boundaryLines := options.boundaryLines
	baseBoundaryLines := options.baseBoundaryLines
	selectionFill := options.selectionFill
	coveragePaths := options.coveragePaths
	curveCoveragePaths := options.curveCoveragePaths
	visibleFill := options.visibleFill
	visibleFillCoveragePaths := options.visibleFillCoveragePaths
	initialConnectorLines := options.initialConnectorLines
	contactLines := options.contactLines
	computeContactSweepLines := options.computeContactSweepLines

	visible, lineworkErr := geodata.VisibleLineworkPolygons(
		boundaryLines, selectionFill, coveragePaths, 75,
	)
	if lineworkErr != nil && useContactContours && len(visibleFill) > 0 {
		if retryVisible, retryErr := geodata.VisibleLineworkPolygons(
			boundaryLines, visibleFill, visibleFillCoveragePaths, 75,
		); retryErr == nil {
			visible, lineworkErr = retryVisible, nil
		}
	}
	if lineworkErr != nil && useContactContours && len(visibleFill) > 0 {
		if stableVisible, stableErr := occultationRetryStableContactBoundary(options); stableErr == nil {
			visible, lineworkErr = stableVisible, nil
		}
	}
	if lineworkErr != nil {
		// A polar rise/set curve can contain a fold with two physical branches.
		// Retry a bounded set of branch-specific boundary candidates.
		for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
			baseBoundaryLines, boundaryCurves, initialConnectorLines,
		) {
			visible, lineworkErr = geodata.VisibleLineworkPolygons(
				candidate, selectionFill, coveragePaths, 75,
			)
			if lineworkErr == nil {
				break
			}
		}
	}
	if lineworkErr != nil && len(curveCoveragePaths) > 0 {
		// Polar footprint probes can land on a numerically ambiguous junction.
		// Retry with only curve-side probes while retaining the source audit.
		curveCoverage := limitOccultationCoveragePaths(curveCoveragePaths, 2048)
		visible, lineworkErr = geodata.VisibleLineworkPolygons(
			boundaryLines, selectionFill, curveCoverage, 75,
		)
		if lineworkErr != nil {
			for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
				baseBoundaryLines, boundaryCurves, initialConnectorLines,
			) {
				visible, lineworkErr = geodata.VisibleLineworkPolygons(
					candidate, selectionFill, curveCoverage, 75,
				)
				if lineworkErr == nil {
					break
				}
			}
		}
	}
	if lineworkErr != nil && len(initialConnectorLines) == 0 {
		connectorLines := occultationHorizonConnectorBoundaryLines(
			HorizonConnectorSegments(footprints, curves, northern, southern),
		)
		if len(connectorLines) > 0 {
			connectorBoundaryLines := append([][]geodata.GeoPoint(nil), boundaryLines...)
			connectorBoundaryLines = append(connectorBoundaryLines, connectorLines...)
			visible, lineworkErr = geodata.VisibleLineworkPolygons(
				connectorBoundaryLines, selectionFill, coveragePaths, 75,
			)
			if lineworkErr != nil {
				for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
					baseBoundaryLines, boundaryCurves, connectorLines,
				) {
					visible, lineworkErr = geodata.VisibleLineworkPolygons(
						candidate, selectionFill, coveragePaths, 75,
					)
					if lineworkErr == nil {
						break
					}
				}
			}
			if lineworkErr != nil && len(curveCoveragePaths) > 0 {
				curveCoverage := limitOccultationCoveragePaths(curveCoveragePaths, 2048)
				visible, lineworkErr = geodata.VisibleLineworkPolygons(
					connectorBoundaryLines, selectionFill, curveCoverage, 75,
				)
				if lineworkErr != nil {
					for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
						baseBoundaryLines, boundaryCurves, connectorLines,
					) {
						visible, lineworkErr = geodata.VisibleLineworkPolygons(
							candidate, selectionFill, curveCoverage, 75,
						)
						if lineworkErr == nil {
							break
						}
					}
				}
			}
		}
	}
	if lineworkErr != nil && useContactContours {
		if sweepLines := computeContactSweepLines(); len(sweepLines) > 0 {
			sweepBaseBoundaryLines := append(append([][]geodata.GeoPoint(nil), contactLines...), sweepLines...)
			sweepBoundaryLines := occultationVisibleBoundaryLinesFromBase(
				sweepBaseBoundaryLines, boundaryCurves, initialConnectorLines,
			)
			visible, lineworkErr = geodata.VisibleLineworkPolygons(
				sweepBoundaryLines, selectionFill, coveragePaths, 75,
			)
			if lineworkErr != nil {
				for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
					sweepBaseBoundaryLines, boundaryCurves, initialConnectorLines,
				) {
					visible, lineworkErr = geodata.VisibleLineworkPolygons(
						candidate, selectionFill, coveragePaths, 75,
					)
					if lineworkErr == nil {
						break
					}
				}
			}
			if lineworkErr != nil && len(curveCoveragePaths) > 0 {
				curveCoverage := limitOccultationCoveragePaths(curveCoveragePaths, 2048)
				visible, lineworkErr = geodata.VisibleLineworkPolygons(
					sweepBoundaryLines, selectionFill, curveCoverage, 75,
				)
				if lineworkErr != nil {
					for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
						sweepBaseBoundaryLines, boundaryCurves, initialConnectorLines,
					) {
						visible, lineworkErr = geodata.VisibleLineworkPolygons(
							candidate, selectionFill, curveCoverage, 75,
						)
						if lineworkErr == nil {
							break
						}
					}
				}
			}
			if lineworkErr != nil && len(initialConnectorLines) == 0 {
				connectorLines := occultationHorizonConnectorBoundaryLines(
					HorizonConnectorSegments(footprints, curves, northern, southern),
				)
				if len(connectorLines) > 0 {
					connectorBoundaryLines := append([][]geodata.GeoPoint(nil), sweepBoundaryLines...)
					connectorBoundaryLines = append(connectorBoundaryLines, connectorLines...)
					visible, lineworkErr = geodata.VisibleLineworkPolygons(
						connectorBoundaryLines, selectionFill, coveragePaths, 75,
					)
					if lineworkErr != nil {
						for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
							sweepBaseBoundaryLines, boundaryCurves, connectorLines,
						) {
							visible, lineworkErr = geodata.VisibleLineworkPolygons(
								candidate, selectionFill, coveragePaths, 75,
							)
							if lineworkErr == nil {
								break
							}
						}
					}
					if lineworkErr != nil && len(curveCoveragePaths) > 0 {
						curveCoverage := limitOccultationCoveragePaths(curveCoveragePaths, 2048)
						visible, lineworkErr = geodata.VisibleLineworkPolygons(
							connectorBoundaryLines, selectionFill, curveCoverage, 75,
						)
						if lineworkErr != nil {
							for _, candidate := range occultationCurveBoundaryAlternativesFromBase(
								sweepBaseBoundaryLines, boundaryCurves, connectorLines,
							) {
								visible, lineworkErr = geodata.VisibleLineworkPolygons(
									candidate, selectionFill, curveCoverage, 75,
								)
								if lineworkErr == nil {
									break
								}
							}
						}
					}
				}
			}
		}
	}
	return visible, lineworkErr
}

func occultationRetryStableContactBoundary(
	options occultationVisibleBandLineworkOptions,
) ([][]geodata.GeoPoint, error) {
	stableBoundary, err := options.computeSweep()
	if err != nil {
		return nil, err
	}
	stableBoundaryLines := occultationVisibleBoundaryLines(
		stableBoundary, options.boundaryCurves, options.initialConnectorLines,
	)
	if options.useContactContours {
		if sweepLines := options.computeContactSweepLines(); len(sweepLines) > 0 {
			stableBoundaryLines = occultationVisibleBoundaryLinesFromBase(
				sweepLines, options.boundaryCurves, options.initialConnectorLines,
			)
		}
	}
	return geodata.VisibleLineworkPolygons(
		stableBoundaryLines, options.visibleFill, options.visibleFillCoveragePaths, 75,
	)
}

// occultationSweepWitnessedByContactContours 报告扫掠边界是否被连续接触包络见证。
func occultationSweepWitnessedByContactContours(
	polygons [][]geodata.GeoPoint,
	contours [][]basic.OccultationPathPoint,
) bool {
	if len(polygons) == 0 || len(contours) == 0 {
		return false
	}
	contactLines := occultationContactContourBoundaryLines(contours)
	if len(contactLines) == 0 {
		return false
	}
	return geodata.SphericalPolygonsContainPathsWithinKM(polygons, contactLines, true, 25)
}

// authoritativeFallbackBandPolygons accepts the instantaneous sweep only when
// its resulting boundary remains tightly attached to the sampled contact
// contours. Some polar Saturn events expose several valid phase branches but
// no single start/end/connector cycle; the sweep is still an authoritative
// geometry in that case because its boundary is independently witnessed by the
// continuous contact envelope. A generous miss would turn an arbitrary face
// into a false authoritative result, so keep this gate deliberately small.
func authoritativeFallbackBandPolygons(
	fallback [][]geodata.GeoPoint,
	contours [][]basic.OccultationPathPoint,
	strongPolarSmoothing bool,
) ([][]geodata.GeoPoint, bool) {
	if !occultationSweepWitnessedByContactContours(fallback, contours) {
		return nil, false
	}
	// The fallback is accepted as authoritative only because its boundary is
	// witnessed by the continuous contact contours. First complete the shared
	// cleanup, then round the finished sweep ring once. Applying this rounder to
	// intermediate linework can preserve a seam which the later union would
	// otherwise remove.
	cleaned := cleanupOccultationAuthoritativeBandPolygons(fallback)
	return roundOccultationAuthoritativeBandJunctions(cleaned), true
}

func occultationContourFillAndCoverage(
	useContactContours bool,
	northern, southern []basic.OccultationPathPoint,
	coveragePaths [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, [][]geodata.GeoPoint) {
	if !useContactContours {
		return nil, coveragePaths
	}
	contourFill := occultationLimitVisibleFillPolygons(northern, southern)
	if len(contourFill) > 0 {
		coveragePaths = append(coveragePaths, contourFill...)
	}
	return contourFill, coveragePaths
}

func occultationNoVisibleBandFill(
	useContactContours bool,
	computeSweep func() ([][]geodata.GeoPoint, error),
	fallback [][]geodata.GeoPoint,
	strongPolarSmoothing bool,
) ([][]geodata.GeoPoint, bool, error) {
	if useContactContours {
		fallback, err := computeSweep()
		if err != nil {
			return nil, false, err
		}
		return cleanupOccultationVisibleBandPolygons(fallback, strongPolarSmoothing), false, nil
	}
	return cleanupOccultationVisibleBandPolygons(fallback, strongPolarSmoothing), false, nil
}

func occultationPreferredPhaseBand(
	useContactContours bool,
	footprints []basic.OccultationFootprint,
	curves []basic.OccultationRiseSetCurve,
	northern, southern []basic.OccultationPathPoint,
	contourFill [][]geodata.GeoPoint,
) ([][]geodata.GeoPoint, bool) {
	if !useContactContours {
		return nil, false
	}
	return occultationPhaseBoundaryPolygons(
		curves,
		HorizonConnectorSegments(footprints, curves, northern, southern),
		contourFill,
	)
}

func occultationMergePreferredVisibleBand(
	visible [][]geodata.GeoPoint,
	useContactContours bool,
	footprints []basic.OccultationFootprint,
	curves []basic.OccultationRiseSetCurve,
	northern, southern []basic.OccultationPathPoint,
	contourFill, visibleFill, visibleFillCoveragePaths [][]geodata.GeoPoint,
) [][]geodata.GeoPoint {
	if len(visible) > 1 {
		visible = removeOccultationPolarSliverComponents(visible)
	}
	if !useContactContours || len(visible) == 0 {
		return visible
	}
	preferredVisible, preferredOK := occultationPreferredPhaseBand(
		useContactContours, footprints, curves, northern, southern, contourFill,
	)
	if !preferredOK || len(preferredVisible) == 0 {
		return visible
	}
	candidate := append([][]geodata.GeoPoint(nil), visible...)
	candidate = append(candidate, preferredVisible...)
	merged, unionErr := geodata.UnionPolygons(candidate)
	if unionErr != nil {
		return visible
	}
	pruned := RemoveTinyPolygonComponents(merged)
	if len(pruned) == 0 {
		return visible
	}
	if witnessErr := validateOccultationVisibleBandWitnesses(
		pruned, useContactContours, visibleFill, contourFill, visibleFillCoveragePaths,
	); witnessErr != nil {
		return visible
	}
	return pruned
}

func occultationFastPolarBand(
	useContactContours, strongPolarSmoothing bool,
	boundaryLines, visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) ([][]geodata.GeoPoint, bool) {
	if !useContactContours || !strongPolarSmoothing {
		return nil, false
	}
	// The fast polar attempt only needs the source fills as witnesses. Building
	// curve-side probes here duplicates the expensive spherical containment pass
	// that the general retry path performs after this candidate is rejected.
	selectionFill := append([][]geodata.GeoPoint(nil), contourFill...)
	selectionFill = append(selectionFill, visibleFill...)
	fastCoveragePaths := coveragePaths
	visible, err := geodata.VisibleLineworkPolygons(
		boundaryLines, selectionFill, fastCoveragePaths, 75,
	)
	if err != nil {
		return nil, false
	}
	return visible, true
}

// occultationFastPolarBandAccepted keeps the cheap polar polygonizer on the
// normal path only when its selected face agrees with both independent
// witnesses: the instantaneous visible footprints and every exported phase
// curve.  A folded horizon can otherwise produce a valid complement that
// looks like a staircase and leaves the purple visibility line outside the
// blue fill.  The miss-distance checks are bounded by the existing probe
// budgets, so rejected faces fall through to the deterministic retries below.
func occultationFastPolarBandAccepted(
	visible, visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) bool {
	if len(visible) == 0 {
		return false
	}
	if validateOccultationVisibleBandWitnesses(
		visible, true, visibleFill, contourFill, coveragePaths,
	) != nil {
		return false
	}
	phasePaths := make([][]geodata.GeoPoint, 0, len(curves)*2)
	for _, curve := range curves {
		for _, line := range occultationCurveBoundaryLines(curve) {
			if len(line) >= 2 {
				phasePaths = append(phasePaths, line)
			}
		}
	}
	if len(phasePaths) == 0 {
		return true
	}
	// A phase curve may lie exactly on the static ring; allow a small numerical
	// tolerance for the spherical edge/midpoint probes but reject a different
	// polygonizer face by a clearly visible (>75 km) displacement.
	return geodata.SphericalPolygonsContainPathsWithinKM(
		visible, phasePaths, true, 10,
	)
}

func occultationPreferredPhaseBandAccepted(
	preferred, contourFill, visibleFill [][]geodata.GeoPoint,
) bool {
	if len(preferred) == 0 || len(contourFill) == 0 {
		return false
	}
	// Open instantaneous footprints can become very long at lunar rise/set.
	// They are valid timeline geometry, but their horizon closure is not the
	// compact band's outer envelope. The phase-cycle constructor has already
	// checked both directions against the continuous contact contour; repeat
	// that bounded check here without allowing the horizon-extended footprint
	// fill to reject the physical compact cycle.
	if !geodata.SphericalPolygonsContainPathsWithinKM(preferred, contourFill, true, 150) {
		return false
	}
	if !geodata.SphericalPolygonsContainPathsWithinKM(contourFill, preferred, true, 200) {
		return false
	}
	// The phase cycle is only an outer closure candidate. When instantaneous
	// visible witnesses extend farther along the event interval, accepting the
	// compact cycle would silently turn the full time-union into a moonrise/
	// moonset-only strip. Require the candidate to cover those witnesses before
	// promoting it to the static band.
	if len(visibleFill) > 0 {
		if !geodata.SphericalPolygonsContainPathsWithinKM(preferred, visibleFill, true, 50) {
			return false
		}
	}
	return true
}

func occultationVisibleBandLineworkInputs(
	useContactContours, strongPolarSmoothing bool,
	visibleFill, contourFill, coveragePaths [][]geodata.GeoPoint,
	curves []basic.OccultationRiseSetCurve,
) (
	[][]geodata.GeoPoint,
	[][]geodata.GeoPoint,
	[][]geodata.GeoPoint,
	[][]geodata.GeoPoint,
) {
	footprintCoveragePaths := append([][]geodata.GeoPoint(nil), coveragePaths...)
	// Contact contours remain the primary static envelope for both partial and
	// total bands, but the footprint polygons still carry real interior witness
	// points. Feed both into face selection so the chosen face stays anchored to
	// the contour while also inheriting the footprint lobe that only the sweep
	// sees.
	selectionFill := visibleFill
	if useContactContours && len(contourFill) > 0 {
		selectionFill = append(append([][]geodata.GeoPoint(nil), contourFill...), visibleFill...)
	}
	if !useContactContours {
		// Legacy callers provide rise/set curves without a continuous contact
		// contour. Their footprint fill already supplies the face witnesses; curve
		// side probes would repeat an O(curve-points * fill-edges) spherical scan
		// without adding a reliable distinction between faces.
		return selectionFill, coveragePaths, nil, footprintCoveragePaths
	}
	curveProbeFill := selectionFill
	if len(contourFill) > 0 {
		curveProbeFill = contourFill
	}
	curveCoveragePaths := occultationCurveCoverageProbes(curves, curveProbeFill)
	coveragePaths = append(coveragePaths, curveCoveragePaths...)
	coveragePaths = limitOccultationCoveragePaths(coveragePaths, 2048)
	visibleFillCoveragePaths := append([][]geodata.GeoPoint(nil), footprintCoveragePaths...)
	visibleFillCoveragePaths = append(visibleFillCoveragePaths, curveCoveragePaths...)
	visibleFillCoveragePaths = limitOccultationCoveragePaths(visibleFillCoveragePaths, 2048)
	return selectionFill, coveragePaths, curveCoveragePaths, visibleFillCoveragePaths
}

func occultationVisibleFillAndCoverage(
	footprints []basic.OccultationFootprint,
) ([][]geodata.GeoPoint, [][]geodata.GeoPoint) {
	visibleFill := make([][]geodata.GeoPoint, 0, len(footprints)*2)
	coveragePaths := make([][]geodata.GeoPoint, 0, len(footprints))
	for _, footprint := range footprints {
		for _, source := range footprint.Polygons {
			if len(source) < 3 {
				continue
			}
			if occultationInteriorPolygon(source, footprint.InteriorPolygons) {
				continue
			}
			visibleSource := clipOccultationPolygonToHorizon(source)
			if len(visibleSource) < 3 {
				continue
			}
			polygon := make([]geodata.GeoPoint, len(visibleSource))
			for index, point := range visibleSource {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			visibleFill = append(visibleFill, polygon)
			if !occultationPolygonNeedsInteriorProbes(visibleSource) {
				continue
			}
			for _, probe := range occultationVisibleFootprintProbes(visibleSource, polygon) {
				if occultationCoverageProbeExists(coveragePaths, probe) {
					continue
				}
				coveragePaths = append(coveragePaths, []geodata.GeoPoint{probe})
			}
		}
	}
	return visibleFill, coveragePaths
}

// occultationVisibleFootprintFillOnly builds the horizon-clipped source faces
// without the interior probe set used by polygon selection. The mask audit
// only needs the faces themselves, so keeping probe generation out of this
// path avoids repeating the expensive polar witness calculations.
func occultationVisibleFootprintFillOnly(
	footprints []basic.OccultationFootprint,
) [][]geodata.GeoPoint {
	visibleFill := make([][]geodata.GeoPoint, 0, len(footprints)*2)
	for _, footprint := range footprints {
		for _, source := range footprint.Polygons {
			if len(source) < 3 || occultationInteriorPolygon(source, footprint.InteriorPolygons) {
				continue
			}
			visibleSource := clipOccultationPolygonToHorizon(source)
			if len(visibleSource) < 3 {
				continue
			}
			polygon := make([]geodata.GeoPoint, len(visibleSource))
			for index, point := range visibleSource {
				polygon[index] = geodata.GeoPoint{Longitude: point.Longitude, Latitude: point.Latitude}
			}
			visibleFill = append(visibleFill, polygon)
		}
	}
	return visibleFill
}

func occultationCoverageProbeExists(
	coveragePaths [][]geodata.GeoPoint,
	probe geodata.GeoPoint,
) bool {
	for _, existing := range coveragePaths {
		if geoDistanceKM(existing[0], probe) < 10 {
			return true
		}
	}
	return false
}
