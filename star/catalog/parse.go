package catalog

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
)

// GaiaDR3Columns 返回 Gaia DR3 归档导出的列映射 / column map for Gaia DR3 archive exports.
func GaiaDR3Columns() ColumnMap {
	return ColumnMap{
		ID: "source_id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec",
		Parallax: "parallax", RadialVelocity: "radial_velocity", Magnitude: "phot_g_mean_mag",
		BPRP: "bp_rp", Band: BandG, Epoch: julianYearTime(2016.0), Frame: basic.CoordinateFrameICRS,
		IDPrefix: "Gaia DR3 ", PMRAIncludesCosDec: true,
	}
}

// Hipparcos2007Columns 返回 VizieR 导出的 Hipparcos 2007 新归算 CSV 列映射，角度已换算为度；原始 hip2.dat 用 ReadHipparcos2007 /
// column map for VizieR CSV exports of Hipparcos 2007 in degrees; use ReadHipparcos2007 for raw hip2.dat.
func Hipparcos2007Columns() ColumnMap {
	return ColumnMap{
		ID: "HIP", RA: "RArad", Dec: "DErad", PMRA: "pmRA", PMDec: "pmDE",
		Parallax: "Plx", Magnitude: "Hpmag", BV: "B-V", Band: BandHp,
		Epoch: julianYearTime(1991.25), Frame: basic.CoordinateFrameICRS, IDPrefix: "HIP ",
		PMRAIncludesCosDec: true, AnglesInRadians: false,
	}
}

func julianYearTime(year float64) time.Time {
	return basic.JD2DateByZone(2451545.0+(year-2000.0)*365.25, time.UTC, false)
}

func parseMissing(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "nan", "--", "null":
		return true
	default:
		return false
	}
}

func parseFloat(value string) (float64, bool, error) {
	if parseMissing(value) {
		return 0, false, nil
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, true, err
	}
	return parsed, true, nil
}

func parseInt(value string) (int64, bool, error) {
	if parseMissing(value) {
		return 0, false, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, true, err
	}
	return parsed, true, nil
}

func parseUint(value string) (uint64, bool, error) {
	if parseMissing(value) {
		return 0, false, nil
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, true, err
	}
	return parsed, true, nil
}

func normalizedFrame(frame basic.CoordinateFrame) basic.CoordinateFrame {
	if frame == "" {
		return basic.CoordinateFrameICRS
	}
	return frame
}

func addEntry(c *Catalog, entry Entry, record int, options ReadOptions) error {
	if entry.ID == "" {
		return recordFailure(c, options, record, "missing ID")
	}
	if err := entry.Coordinate.Validate(); err != nil {
		return recordFailure(c, options, record, err.Error())
	}
	if options.Filter != nil && !options.Filter(entry) {
		return nil
	}
	c.add(entry)
	return nil
}

func recordFailure(c *Catalog, options ReadOptions, record int, reason interface{}) error {
	if options.Strict {
		return fmt.Errorf("record %d: %v", record, reason)
	}
	c.skipped++
	return nil
}

func parseCoordinate(id string, ra, dec, pmRA, pmDec, parallax, radial float64, epoch time.Time, frame basic.CoordinateFrame) basic.StarCoordinate {
	if ra == 360 || ra == -0 {
		ra = 0
	} else {
		ra = normalizeRA(ra)
	}
	// 负视差是测量噪声，按无视差处理；非有限值原样保留，交给 Validate 拒收。
	if finite(parallax) && parallax < 0 {
		parallax = 0
	}
	return basic.StarCoordinate{
		ID:                             id,
		RA:                             ra,
		Dec:                            dec,
		Epoch:                          epoch,
		Frame:                          normalizedFrame(frame),
		ProperMotionRACosDecMasPerYear: pmRA,
		ProperMotionDecMasPerYear:      pmDec,
		ParallaxMas:                    parallax,
		RadialVelocityKmPerSecond:      radial,
	}
}

func fixedFloat(line []byte, start, end int) (float64, bool, error) {
	value, ok := fixedField(line, start, end)
	if !ok {
		return 0, false, nil
	}
	return parseFloat(value)
}

func fixedInt(line []byte, start, end int) (int64, bool, error) {
	value, ok := fixedField(line, start, end)
	if !ok {
		return 0, false, nil
	}
	return parseInt(value)
}

func fixedField(line []byte, start, end int) (string, bool) {
	if start < 0 || end < start || start >= len(line) || end > len(line) {
		return "", false
	}
	return strings.TrimSpace(string(line[start:end])), true
}

func checkFinite(value float64) error {
	if !finite(value) {
		return fmt.Errorf("non-finite value %v", value)
	}
	return nil
}

func angularFromRadians(value float64) float64 {
	return value * 180 / math.Pi
}
