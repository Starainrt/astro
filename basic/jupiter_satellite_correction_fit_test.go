package basic

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

func TestJupiterGalileanCorrectionFit(t *testing.T) {
	if os.Getenv("ASTRO_REGENERATE_GALILEAN_CORRECTION") == "1" {
		jupiterGalileanFitRegenerate(t)
		return
	}
	jupiterGalileanFitValidate(t)
}

type jupiterGalileanFitBatch struct {
	Satellite int
	File      string
	Rows      int
}

type jupiterGalileanFitCurve struct {
	Label     string
	Satellite int
	Samples   []struct{ JD, X, Y, Z float64 }
}

type jupiterGalileanFitRow struct {
	et       float64
	elements jupiterGalileanElements
	target   [3]float64
}

type jupiterGalileanFitBasis struct {
	span    jupiterGalileanCorrectionSpan
	periods [5][]float64
	offsets [6]int
}

type jupiterGalileanFitCandidate struct {
	Span   jupiterGalileanCorrectionSpan
	Count  int
	Rows   int
	RMS    float64
	Max    float64
	Models [5]jupiterGalileanElementCorrection
}

var jupiterGalileanFitSpans = [2]jupiterGalileanCorrectionSpan{
	{StartJD: 2415020.5, EndJD: 2488069.5, CenterYear: 2000, HalfYears: 100},
	{StartJD: 2305458.5, EndJD: 2670689.5, CenterYear: 2100, HalfYears: 500},
}

func jupiterGalileanFitValidate(t *testing.T) {
	t.Helper()
	dataBytes := 4 * (5*3 + 4) * 8
	for index, count := range jupiterGalileanCorrectionCounts {
		if count != 0 && count != 10 && count != 20 && count != 40 {
			t.Fatalf("satellite %d invalid period count %d", index+1, count)
		}
		span := jupiterGalileanCorrectionSpans[index]
		if span != jupiterGalileanFitSpans[0] && span != jupiterGalileanFitSpans[1] {
			t.Fatalf("satellite %d invalid span: %+v", index+1, span)
		}
		if jupiterGalileanCorrectionRows[index] <= 0 || jupiterGalileanCorrectionRows[index] > 898736 {
			t.Fatalf("satellite %d fitted rows=%d outside training set", index+1, jupiterGalileanCorrectionRows[index])
		}
		basis := jupiterGalileanFitMakeBasis(index, count, span)
		for element, model := range jupiterGalileanCorrections[index] {
			if len(model.Terms) != len(basis.periods[element]) {
				t.Fatalf("satellite %d element %d period count mismatch", index+1, element)
			}
			for term, period := range basis.periods[element] {
				got := model.Terms[term]
				if got.Period != period || math.IsNaN(got.Sin) || math.IsInf(got.Sin, 0) || math.IsNaN(got.Cos) || math.IsInf(got.Cos, 0) {
					t.Fatalf("satellite %d element %d term %d invalid", index+1, element, term)
				}
			}
			for _, coefficient := range model.Polynomial {
				if math.IsNaN(coefficient) || math.IsInf(coefficient, 0) {
					t.Fatalf("satellite %d element %d non-finite polynomial", index+1, element)
				}
			}
			dataBytes += len(model.Terms) * 3 * 8
		}
	}
	t.Logf("correction periods=%v data=%d bytes / 8192", jupiterGalileanCorrectionCounts, dataBytes)
	if dataBytes > 8192 {
		t.Fatalf("coefficient table size %d exceeds 8192 bytes", dataBytes)
	}
}

func jupiterGalileanFitMakeBasis(index, count int, span jupiterGalileanCorrectionSpan) jupiterGalileanFitBasis {
	basis := jupiterGalileanFitBasis{span: span}
	spanDays := span.EndJD - span.StartJD
	maxPeriod := 1000 * 365.25
	if span.HalfYears == 100 {
		maxPeriod = spanDays / 3
	}
	for element, block := range []int{1, 2, 2, 3, 3} {
		terms := append([]jupiterGalileanL1Term(nil), jupiterGalileanL1LongPeriodTerms[index][block]...)
		terms = append(terms, jupiterGalileanL1CrossPeriodTerms[index][block]...)
		sort.SliceStable(terms, func(left, right int) bool { return math.Abs(terms[left].Amp) > math.Abs(terms[right].Amp) })
		for _, term := range terms {
			period := math.Abs(term.Period)
			if period <= 0.5 || period >= maxPeriod {
				continue
			}
			duplicate := false
			for _, selected := range basis.periods[element] {
				if span.HalfYears == 100 && math.Abs(1/period-1/selected)*spanDays < 2 || span.HalfYears == 500 && math.Abs(period-selected)/math.Max(period, selected) < 1e-6 {
					duplicate = true
					break
				}
			}
			if duplicate || len(basis.periods[element]) >= count {
				continue
			}
			basis.periods[element] = append(basis.periods[element], period)
		}
		basis.offsets[element+1] = basis.offsets[element] + 3 + 2*len(basis.periods[element])
	}
	return basis
}

func (basis jupiterGalileanFitBasis) values(et float64, values []float64) {
	tau := basis.span.normalizedTime(et)
	for element, periods := range basis.periods {
		offset := basis.offsets[element]
		values[offset], values[offset+1], values[offset+2] = 1, tau, 2*tau*tau-1
		for term, period := range periods {
			values[offset+3+2*term], values[offset+4+2*term] = math.Sincos(2 * math.Pi * et / period)
		}
	}
}

func jupiterGalileanFitPureElements(et float64, index int, includeSolar bool) jupiterGalileanElements {
	longPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1LongPeriodTerms[index], et+jupiterGalileanLongPeriodShift, et, includeSolar, index)
	crossPeriod := jupiterGalileanEvaluateSeries(jupiterGalileanL1CrossPeriodTerms[index], et, et, false, index)
	elements := jupiterGalileanElements{
		A: longPeriod.A + crossPeriod.A, L: longPeriod.L + crossPeriod.L + jupiterGalileanBaseMeanLongitudes[index]*et,
		K: longPeriod.K + crossPeriod.K, H: longPeriod.H + crossPeriod.H,
		Q: longPeriod.Q + crossPeriod.Q, P: longPeriod.P + crossPeriod.P,
	}
	elements.L = math.Atan2(math.Sin(elements.L), math.Cos(elements.L))
	if elements.L < 0 {
		elements.L += 2 * math.Pi
	}
	return elements
}

func jupiterGalileanFitPureState(et float64, index int, includeSolar bool) JupiterGalileanState {
	pv := jupiterGalileanElementsToPV(jupiterGalileanMu[index], jupiterGalileanFitPureElements(et, index, includeSolar))
	cosNode, sinNode := math.Cos(jupiterGalileanFrameNode), math.Sin(jupiterGalileanFrameNode)
	cosTilt, sinTilt := math.Cos(jupiterGalileanFrameTilt), math.Sin(jupiterGalileanFrameTilt)
	return JupiterGalileanState{
		X:  pv[0]*cosNode - pv[1]*sinNode*cosTilt + pv[2]*sinTilt*sinNode,
		Y:  pv[0]*sinNode + pv[1]*cosNode*cosTilt - pv[2]*sinTilt*cosNode,
		Z:  pv[1]*sinTilt + pv[2]*cosTilt,
		VX: pv[3]*cosNode - pv[4]*sinNode*cosTilt + pv[5]*sinTilt*sinNode,
		VY: pv[3]*sinNode + pv[4]*cosNode*cosTilt - pv[5]*sinTilt*cosNode,
		VZ: pv[4]*sinTilt + pv[5]*cosTilt,
	}
}

func jupiterGalileanFitTarget(position [3]float64) [3]float64 {
	cosNode, sinNode := math.Cos(jupiterGalileanFrameNode), math.Sin(jupiterGalileanFrameNode)
	cosTilt, sinTilt := math.Cos(jupiterGalileanFrameTilt), math.Sin(jupiterGalileanFrameTilt)
	xValue := position[0]*cosNode + position[1]*sinNode
	yValue := -position[0]*sinNode + position[1]*cosNode
	return [3]float64{xValue / astronomicalUnitKM, (yValue*cosTilt + position[2]*sinTilt) / astronomicalUnitKM, (-yValue*sinTilt + position[2]*cosTilt) / astronomicalUnitKM}
}

func jupiterGalileanFitReadRows(t *testing.T, index int, batches []jupiterGalileanFitBatch, span jupiterGalileanCorrectionSpan) []jupiterGalileanFitRow {
	t.Helper()
	rows := make([]jupiterGalileanFitRow, 0, 898736)
	for _, batch := range batches {
		if batch.Satellite != index+1 {
			continue
		}
		file, err := os.Open(filepath.Join("/tmp/astro-a1g", batch.File))
		if err != nil {
			t.Fatal(err)
		}
		vectors := make([][7]float64, batch.Rows)
		readErr := binary.Read(file, binary.LittleEndian, vectors)
		var extra [1]byte
		_, extraErr := file.Read(extra[:])
		file.Close()
		if readErr != nil || extraErr != io.EOF {
			t.Fatalf("invalid training batch %s: read=%v trailing=%v", batch.File, readErr, extraErr)
		}
		for _, vector := range vectors {
			if vector[0] < span.StartJD || vector[0] > span.EndJD {
				continue
			}
			et := vector[0] - jupiterGalileanReferenceJD
			rows = append(rows, jupiterGalileanFitRow{
				et: et, elements: jupiterGalileanFitPureElements(et, index, jupiterGalileanUseSolarLongPeriod(vector[0])),
				target: jupiterGalileanFitTarget([3]float64{vector[1], vector[2], vector[3]}),
			})
		}
	}
	if len(rows) == 0 || span.HalfYears == 500 && len(rows) != 898736 {
		t.Fatalf("satellite %d invalid training rows=%d for span %+v", index+1, len(rows), span)
	}
	return rows
}

func jupiterGalileanFitApply(elements *jupiterGalileanElements, correction [5]float64) {
	elements.L += correction[0]
	elements.K += correction[1]
	elements.H += correction[2]
	elements.Q += correction[3]
	elements.P += correction[4]
}

func jupiterGalileanFitNormal(t *testing.T, index int, rows []jupiterGalileanFitRow, basis jupiterGalileanFitBasis, coefficients []float64) ([]float64, []float64) {
	t.Helper()
	dimension := basis.offsets[5]
	normal, rhs := make([]float64, dimension*dimension), make([]float64, dimension)
	values := make([]float64, dimension)
	const differenceStep = 1e-9
	for rowIndex, row := range rows {
		basis.values(row.et, values)
		elements := row.elements
		if coefficients != nil {
			var correction [5]float64
			for element := range correction {
				for column := basis.offsets[element]; column < basis.offsets[element+1]; column++ {
					correction[element] += values[column] * coefficients[column]
				}
			}
			jupiterGalileanFitApply(&elements, correction)
		}
		position := jupiterGalileanElementsToPV(jupiterGalileanMu[index], elements)
		residual := [3]float64{row.target[0] - position[0], row.target[1] - position[1], row.target[2] - position[2]}
		var derivatives [5][3]float64
		for element := range derivatives {
			left, right := elements, elements
			var delta [5]float64
			delta[element] = differenceStep
			jupiterGalileanFitApply(&right, delta)
			delta[element] = -differenceStep
			jupiterGalileanFitApply(&left, delta)
			leftPV := jupiterGalileanElementsToPV(jupiterGalileanMu[index], left)
			rightPV := jupiterGalileanElementsToPV(jupiterGalileanMu[index], right)
			for axis := range derivatives[element] {
				derivatives[element][axis] = (rightPV[axis] - leftPV[axis]) / (2 * differenceStep)
			}
		}
		for element, derivative := range derivatives {
			projected := derivative[0]*residual[0] + derivative[1]*residual[1] + derivative[2]*residual[2]
			for column := basis.offsets[element]; column < basis.offsets[element+1]; column++ {
				rhs[column] += values[column] * projected
			}
			for other := 0; other <= element; other++ {
				dot := derivative[0]*derivatives[other][0] + derivative[1]*derivatives[other][1] + derivative[2]*derivatives[other][2]
				for column := basis.offsets[element]; column < basis.offsets[element+1]; column++ {
					limit := basis.offsets[other+1]
					if other == element {
						limit = column + 1
					}
					rowValues := normal[column*dimension : (column+1)*dimension]
					factor := dot * values[column]
					for previous := basis.offsets[other]; previous < limit; previous++ {
						rowValues[previous] += factor * values[previous]
					}
				}
			}
		}
		if (rowIndex+1)%200000 == 0 {
			t.Logf("satellite %d parameters=%d iteration rows=%d/%d", index+1, dimension, rowIndex+1, len(rows))
		}
	}
	return normal, rhs
}

func jupiterGalileanFitSolve(normal, rhs []float64, maximum, basis jupiterGalileanFitBasis) ([]float64, error) {
	dimension, fullDimension := basis.offsets[5], maximum.offsets[5]
	indices := make([]int, 0, dimension)
	for element := range basis.periods {
		for column := 0; column < basis.offsets[element+1]-basis.offsets[element]; column++ {
			indices = append(indices, maximum.offsets[element]+column)
		}
	}
	scale, solution, lower := make([]float64, dimension), make([]float64, dimension), make([]float64, dimension*dimension)
	for row, source := range indices {
		scale[row] = math.Sqrt(normal[source*fullDimension+source])
		if scale[row] == 0 || math.IsNaN(scale[row]) {
			return nil, fmt.Errorf("zero or invalid normal diagonal %d", row)
		}
		solution[row] = rhs[source] / scale[row]
	}
	for row, sourceRow := range indices {
		for column := 0; column <= row; column++ {
			sourceColumn := indices[column]
			value := normal[sourceRow*fullDimension+sourceColumn] / (scale[row] * scale[column])
			for previous := 0; previous < column; previous++ {
				value -= lower[row*dimension+previous] * lower[column*dimension+previous]
			}
			if row == column {
				if value <= 0 || math.IsNaN(value) {
					return nil, fmt.Errorf("normal matrix not positive definite at %d: %.17g", row, value)
				}
				lower[row*dimension+column] = math.Sqrt(value)
			} else {
				lower[row*dimension+column] = value / lower[column*dimension+column]
			}
		}
	}
	for row := 0; row < dimension; row++ {
		for column := 0; column < row; column++ {
			solution[row] -= lower[row*dimension+column] * solution[column]
		}
		solution[row] /= lower[row*dimension+row]
	}
	for row := dimension - 1; row >= 0; row-- {
		for column := row + 1; column < dimension; column++ {
			solution[row] -= lower[column*dimension+row] * solution[column]
		}
		solution[row] /= lower[row*dimension+row]
	}
	for column := range solution {
		solution[column] /= scale[column]
	}
	return solution, nil
}

func jupiterGalileanFitScore(index int, curves []jupiterGalileanFitCurve, basis jupiterGalileanFitBasis, coefficients []float64) (float64, float64) {
	squares, maximum, count := 0.0, 0.0, 0
	values := make([]float64, basis.offsets[5])
	for _, curve := range curves {
		if curve.Satellite != index+1 || curve.Label != "holdout-modern" {
			continue
		}
		for _, sample := range curve.Samples {
			et := sample.JD - jupiterGalileanReferenceJD
			elements := jupiterGalileanFitPureElements(et, index, jupiterGalileanUseSolarLongPeriod(sample.JD))
			basis.values(et, values)
			var correction [5]float64
			for element := range correction {
				for column := basis.offsets[element]; column < basis.offsets[element+1]; column++ {
					correction[element] += values[column] * coefficients[column]
				}
			}
			jupiterGalileanFitApply(&elements, correction)
			position := jupiterGalileanElementsToPV(jupiterGalileanMu[index], elements)
			target := jupiterGalileanFitTarget([3]float64{sample.X, sample.Y, sample.Z})
			residual := 0.0
			for axis := range target {
				residual += math.Pow((position[axis]-target[axis])*astronomicalUnitKM, 2)
			}
			squares += residual
			maximum = math.Max(maximum, math.Sqrt(residual))
			count++
		}
	}
	if count == 0 {
		return math.Inf(1), math.Inf(1)
	}
	return math.Sqrt(squares / float64(count)), maximum
}

func jupiterGalileanFitRegenerate(t *testing.T) {
	t.Helper()
	var batches []jupiterGalileanFitBatch
	var curves []jupiterGalileanFitCurve
	for filename, destination := range map[string]interface{}{
		"/tmp/astro-a1g/training.json":                     &batches,
		"../jupiter/testdata/galilean_jup365_holdout.json": &curves,
	} {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, destination); err != nil {
			t.Fatal(err)
		}
	}
	var candidates [4][]jupiterGalileanFitCandidate
	for index := range candidates {
		started := time.Now()
		for spanIndex, span := range jupiterGalileanFitSpans {
			rows := jupiterGalileanFitReadRows(t, index, batches, span)
			maximum := jupiterGalileanFitMakeBasis(index, 40, span)
			t.Logf("satellite %d S%d training=%d maximum parameters=%d preprocessing=%v", index+1, spanIndex+1, len(rows), maximum.offsets[5], time.Since(started))
			normal, rhs := jupiterGalileanFitNormal(t, index, rows, maximum, nil)
			for _, count := range []int{0, 10, 20, 40} {
				basis := jupiterGalileanFitMakeBasis(index, count, span)
				coefficients, err := jupiterGalileanFitSolve(normal, rhs, maximum, basis)
				if err != nil {
					t.Fatalf("satellite %d S%d N=%d first iteration: %v", index+1, spanIndex+1, count, err)
				}
				secondNormal, secondRHS := jupiterGalileanFitNormal(t, index, rows, basis, coefficients)
				delta, err := jupiterGalileanFitSolve(secondNormal, secondRHS, basis, basis)
				if err != nil {
					t.Fatalf("satellite %d S%d N=%d second iteration: %v", index+1, spanIndex+1, count, err)
				}
				for column := range coefficients {
					coefficients[column] += delta[column]
				}
				rms, maximumError := jupiterGalileanFitScore(index, curves, basis, coefficients)
				t.Logf("satellite %d S%d N=%d modern holdout RMS=%.6f km maximum=%.6f km elapsed=%v", index+1, spanIndex+1, count, rms, maximumError, time.Since(started))
				candidate := jupiterGalileanFitCandidate{Span: span, Count: count, Rows: len(rows), RMS: rms, Max: maximumError}
				for element, periods := range basis.periods {
					offset := basis.offsets[element]
					model := jupiterGalileanElementCorrection{Polynomial: [3]float64{coefficients[offset], coefficients[offset+1], coefficients[offset+2]}}
					for term, period := range periods {
						model.Terms = append(model.Terms, jupiterGalileanCorrectionTerm{Period: period, Sin: coefficients[offset+3+2*term], Cos: coefficients[offset+4+2*term]})
					}
					candidate.Models[element] = model
				}
				candidates[index] = append(candidates[index], candidate)
			}
		}
		sort.SliceStable(candidates[index], func(left, right int) bool { return candidates[index][left].RMS < candidates[index][right].RMS })
		result, err := json.MarshalIndent(candidates, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("/tmp/astro-a1g/fit-candidates.json", result, 0644); err != nil {
			t.Fatal(err)
		}
	}
	selected := jupiterGalileanFitSelect(t, candidates)
	jupiterGalileanFitWrite(t, selected)
}

func (candidate jupiterGalileanFitCandidate) bytes() int {
	dataBytes := (5*3 + 4) * 8
	for _, model := range candidate.Models {
		dataBytes += len(model.Terms) * 24
	}
	return dataBytes
}

func (candidate jupiterGalileanFitCandidate) acceptable(index int) bool {
	return candidate.RMS <= [4]float64{20, 20, 35, 20}[index] && candidate.Max <= [4]float64{45, 45, 70, 45}[index]
}

func jupiterGalileanFitSelect(t *testing.T, candidates [4][]jupiterGalileanFitCandidate) [4]jupiterGalileanFitCandidate {
	t.Helper()
	var selected [4]jupiterGalileanFitCandidate
	var ranks [4]int
	totalBytes := 0
	for index := range selected {
		selected[index] = candidates[index][0]
		if !selected[index].acceptable(index) {
			t.Fatalf("satellite %d best S%.0f N=%d RMS=%.6f maximum=%.6f exceeds modern limits", index+1, selected[index].Span.HalfYears, selected[index].Count, selected[index].RMS, selected[index].Max)
		}
		totalBytes += selected[index].bytes()
	}
	for totalBytes > 8192 {
		bestIndex, bestRank, bestDelta := -1, -1, math.Inf(1)
		for index, current := range selected {
			for rank := ranks[index] + 1; rank < len(candidates[index]); rank++ {
				alternative := candidates[index][rank]
				if !alternative.acceptable(index) || alternative.bytes() >= current.bytes() {
					continue
				}
				if delta := alternative.RMS - current.RMS; delta < bestDelta {
					bestIndex, bestRank, bestDelta = index, rank, delta
				}
				break
			}
		}
		if bestIndex < 0 {
			t.Fatalf("coefficient table %d bytes exceeds 8192 with no acceptable downgrade", totalBytes)
		}
		previous := selected[bestIndex]
		selected[bestIndex] = candidates[bestIndex][bestRank]
		ranks[bestIndex] = bestRank
		totalBytes += selected[bestIndex].bytes() - previous.bytes()
		t.Logf("budget downgrade satellite %d N=%d->%d RMS increment=%.6f km bytes=%d", bestIndex+1, previous.Count, selected[bestIndex].Count, bestDelta, totalBytes)
	}
	for index, candidate := range selected {
		t.Logf("selected satellite %d span=%.1f..%.1f N=%d rows=%d RMS=%.6f maximum=%.6f table=%d bytes", index+1, candidate.Span.StartJD, candidate.Span.EndJD, candidate.Count, candidate.Rows, candidate.RMS, candidate.Max, totalBytes)
	}
	return selected
}

func jupiterGalileanFitWrite(t *testing.T, selected [4]jupiterGalileanFitCandidate) {
	t.Helper()
	var output bytes.Buffer
	fmt.Fprintln(&output, "// Code generated by TestJupiterGalileanCorrectionFit; DO NOT EDIT.\n\npackage basic")
	fmt.Fprintf(&output, "\nvar jupiterGalileanCorrectionCounts = [4]int{%d, %d, %d, %d}\n", selected[0].Count, selected[1].Count, selected[2].Count, selected[3].Count)
	fmt.Fprintf(&output, "var jupiterGalileanCorrectionRows = [4]int{%d, %d, %d, %d}\n", selected[0].Rows, selected[1].Rows, selected[2].Rows, selected[3].Rows)
	fmt.Fprintln(&output, "var jupiterGalileanCorrectionSpans = [4]jupiterGalileanCorrectionSpan{")
	for _, candidate := range selected {
		span := candidate.Span
		fmt.Fprintf(&output, "{StartJD: %.17g, EndJD: %.17g, CenterYear: %.17g, HalfYears: %.17g},\n", span.StartJD, span.EndJD, span.CenterYear, span.HalfYears)
	}
	fmt.Fprintln(&output, "}")
	fmt.Fprintln(&output, "var jupiterGalileanCorrections = [4][5]jupiterGalileanElementCorrection{")
	for _, candidate := range selected {
		fmt.Fprintln(&output, "{")
		for _, model := range candidate.Models {
			fmt.Fprintf(&output, "{Polynomial: [3]float64{%.17g, %.17g, %.17g}, Terms: []jupiterGalileanCorrectionTerm{\n", model.Polynomial[0], model.Polynomial[1], model.Polynomial[2])
			for _, term := range model.Terms {
				fmt.Fprintf(&output, "{Period: %.17g, Sin: %.17g, Cos: %.17g},\n", term.Period, term.Sin, term.Cos)
			}
			fmt.Fprintln(&output, "}},")
		}
		fmt.Fprintln(&output, "},")
	}
	fmt.Fprintln(&output, "}")
	data, err := format.Source(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("jupiter_satellite_correction_data.go", data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestJupiterGalileanCorrectionSyntheticWindow(t *testing.T) {
	for index, span := range jupiterGalileanCorrectionSpans {
		if span != jupiterGalileanFitSpans[0] && span != jupiterGalileanFitSpans[1] {
			t.Fatalf("satellite %d invalid correction span: %+v", index+1, span)
		}
		for _, sample := range []struct {
			jd     float64
			weight float64
		}{
			{span.StartJD - 36525, 0},
			{span.StartJD - 27393.75, (2 - math.Sqrt2) / 4},
			{span.StartJD - 18262.5, 0.5},
			{span.StartJD - 9131.25, (2 + math.Sqrt2) / 4},
			{span.StartJD, 1},
			{2451545, 1},
			{span.EndJD, 1},
			{span.EndJD + 9131.25, (2 + math.Sqrt2) / 4},
			{span.EndJD + 18262.5, 0.5},
			{span.EndJD + 27393.75, (2 - math.Sqrt2) / 4},
			{span.EndJD + 36525, 0},
		} {
			et := sample.jd - jupiterGalileanReferenceJD
			got := span.window(et)
			if math.Abs(got-sample.weight) > 1e-15 {
				t.Fatalf("satellite %d window at %.9f = %.17g want %.17g", index+1, sample.jd, got, sample.weight)
			}
			actual := jupiterGalileanCorrectionDirectForTest(et, index)
			unwindowed := jupiterGalileanCorrectionUnwindowedForTest(et, index)
			for element, value := range actual {
				want := sample.weight * unwindowed[element]
				if !isFinite(value) || math.Abs(value-want) > 1e-15 {
					t.Errorf("satellite %d at %.9f element %d correction %.17g want %.17g", index+1, sample.jd, element, value, want)
				}
			}
		}
		for _, boundary := range []struct {
			jd     float64
			weight float64
		}{
			{span.StartJD - 36525, 0},
			{span.StartJD, 1},
			{span.EndJD, 1},
			{span.EndJD + 36525, 0},
		} {
			leftET := boundary.jd - jupiterGalileanReferenceJD - 0.001/86400
			rightET := boundary.jd - jupiterGalileanReferenceJD + 0.001/86400
			left := jupiterGalileanCorrectionDirectForTest(leftET, index)
			right := jupiterGalileanCorrectionDirectForTest(rightET, index)
			leftUnwindowed := jupiterGalileanCorrectionUnwindowedForTest(leftET, index)
			rightUnwindowed := jupiterGalileanCorrectionUnwindowedForTest(rightET, index)
			for element, value := range left {
				naturalDelta := boundary.weight * (rightUnwindowed[element] - leftUnwindowed[element])
				residual := (right[element] - value) - naturalDelta
				if !isFinite(residual) || math.Abs(residual) >= 1e-12 {
					t.Errorf("satellite %d boundary %.1f element %d splice residual %.12g exceeds 1e-12 (natural delta %.12g)", index+1, boundary.jd, element, residual, naturalDelta)
				}
			}
		}
		years := []int{1400, 2800}
		if span.HalfYears == 100 {
			years = append(years, 1700, 2300)
		}
		for _, year := range years {
			jd := JDCalc(year, 1, 1)
			et := jd - jupiterGalileanReferenceJD
			got := jupiterGalileanSatelliteCorrectedStateAtET(et, index, jupiterGalileanUseSolarLongPeriod(jd))
			want := jupiterGalileanFitPureState(et, index, jupiterGalileanUseSolarLongPeriod(jd))
			if got != want || JupiterGalileanSatelliteState(jd, index+1) != want {
				t.Fatalf("satellite %d year %d state differs from uncorrected state: got %+v want %+v", index+1, year, got, want)
			}
		}
	}
}

func jupiterGalileanCorrectionUnwindowedForTest(et float64, index int) [5]float64 {
	span := jupiterGalileanCorrectionSpans[index]
	year := 2000 + (et+jupiterGalileanReferenceJD-2451545)/365.25
	normalized := (year - span.CenterYear) / span.HalfYears
	if normalized < -1 {
		normalized = -1
	} else if normalized > 1 {
		normalized = 1
	}
	var correction [5]float64
	for element, model := range jupiterGalileanCorrections[index] {
		correction[element] = model.Polynomial[0] + model.Polynomial[1]*normalized + model.Polynomial[2]*(2*normalized*normalized-1)
		for _, term := range model.Terms {
			phase := 2 * math.Pi * et / term.Period
			correction[element] += term.Sin*math.Sin(phase) + term.Cos*math.Cos(phase)
		}
	}
	return correction
}
