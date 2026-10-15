// Package limbtable decodes lunar limb profile tables without embedding a dataset.
package limbtable

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

const HeaderSize = 4 + 3*4 + 7*8
const maxCells = 1 << 24

// Profile is a lunar limb profile table.
type Profile struct {
	lonCount, latCount, angleCount int
	lonStart, lonStep              float64
	latStart, latStep              float64
	angleStep                      float64
	referenceRadiusKM              float64
	minHeight, maxHeight           float64
	heights                        []float32
}

// Parse decodes an LLB1 profile table.
func Parse(data []byte) (*Profile, error) {
	if len(data) < HeaderSize || string(data[:4]) != "LLB1" {
		return nil, errors.New("limbtable: not an LLB1 table")
	}
	le := binary.LittleEndian
	counts := [3]int{}
	for index := range counts {
		counts[index] = int(le.Uint32(data[4+4*index:]))
	}
	values := [7]float64{}
	for index := range values {
		values[index] = math.Float64frombits(le.Uint64(data[16+8*index:]))
	}
	profile := &Profile{
		lonCount: counts[0], latCount: counts[1], angleCount: counts[2],
		lonStart: values[0], lonStep: values[1],
		latStart: values[2], latStep: values[3],
		angleStep: values[4], referenceRadiusKM: values[5],
	}
	quantum := values[6]
	finite := true
	for _, value := range values {
		finite = finite && isFinite(value)
	}
	if !finite || profile.lonCount < 2 || profile.latCount < 2 || profile.angleCount < 2 ||
		!(profile.lonStep > 0) || !(profile.latStep > 0) || !(quantum > 0) || !(profile.referenceRadiusKM > 0) ||
		math.Abs(profile.angleStep*float64(profile.angleCount)-360) > 1e-9 ||
		profile.lonCount > maxCells/profile.latCount/profile.angleCount {
		return nil, errors.New("limbtable: invalid LLB1 header")
	}

	reader, err := zlib.NewReader(bytes.NewReader(data[HeaderSize:]))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	total := profile.lonCount * profile.latCount * profile.angleCount
	raw := make([]byte, 2*total)
	if _, err := io.ReadFull(reader, raw); err != nil {
		return nil, err
	}
	row := profile.lonCount * profile.angleCount
	levels := make([]int32, total)
	for index := range levels {
		levels[index] = int32(int16(le.Uint16(raw[2*index:])))
	}
	for index := row; index < total; index++ {
		levels[index] += levels[index-row]
	}
	for lat := 0; lat < profile.latCount; lat++ {
		base := lat * row
		for index := base + profile.angleCount; index < base+row; index++ {
			levels[index] += levels[index-profile.angleCount]
		}
	}
	profile.heights = make([]float32, total)
	profile.minHeight, profile.maxHeight = math.Inf(1), math.Inf(-1)
	for index, level := range levels {
		height := float32(float64(level) * quantum)
		profile.heights[index] = height
		profile.minHeight = math.Min(profile.minHeight, float64(height))
		profile.maxHeight = math.Max(profile.maxHeight, float64(height))
	}
	return profile, nil
}

// LimbRadiusKM returns the Moon-centre to limb-point distance in kilometres.
func (profile *Profile) LimbRadiusKM(librationLongitude, librationLatitude, positionAngle float64) (float64, bool) {
	if profile == nil || len(profile.heights) == 0 ||
		!isFinite(librationLongitude) || !isFinite(librationLatitude) || !isFinite(positionAngle) {
		return 0, false
	}
	lon, lonWeight := profile.cell(librationLongitude, profile.lonStart, profile.lonStep, profile.lonCount)
	lat, latWeight := profile.cell(librationLatitude, profile.latStart, profile.latStep, profile.latCount)
	angle := math.Mod(positionAngle/profile.angleStep, float64(profile.angleCount))
	if angle < 0 {
		angle += float64(profile.angleCount)
	}
	angle0 := int(angle)
	if angle0 >= profile.angleCount {
		angle0 = profile.angleCount - 1
	}
	angleWeight := angle - float64(angle0)
	angle1 := (angle0 + 1) % profile.angleCount
	height := 0.0
	for dLat, wLat := range [2]float64{1 - latWeight, latWeight} {
		for dLon, wLon := range [2]float64{1 - lonWeight, lonWeight} {
			base := ((lat+dLat)*profile.lonCount + lon + dLon) * profile.angleCount
			height += wLat * wLon * ((1-angleWeight)*float64(profile.heights[base+angle0]) +
				angleWeight*float64(profile.heights[base+angle1]))
		}
	}
	return profile.referenceRadiusKM + height/1000, true
}

// AngleStepDeg 位置角网格步长（度） / position-angle grid step in degrees.
func (profile *Profile) AngleStepDeg() float64 {
	if profile == nil {
		return 0
	}
	return profile.angleStep
}

// RadiusRangeKM 表内可能取到的最小、最大月缘半径（km） / smallest and largest limb radius.
func (profile *Profile) RadiusRangeKM() (float64, float64) {
	if profile == nil || len(profile.heights) == 0 {
		return 0, 0
	}
	return profile.referenceRadiusKM + profile.minHeight/1000, profile.referenceRadiusKM + profile.maxHeight/1000
}

func (profile *Profile) cell(value, start, step float64, count int) (int, float64) {
	position := (value - start) / step
	if position <= 0 {
		return 0, 0
	}
	if position >= float64(count-1) {
		return count - 2, 1
	}
	index := int(position)
	return index, position - float64(index)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
