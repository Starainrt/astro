package catalog

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
)

// ReadHipparcos2007 读取 Hipparcos 定宽星表 / read the Hipparcos fixed-width catalog.
func ReadHipparcos2007(r io.Reader, options ReadOptions) (*Catalog, error) {
	catalog := newCatalog()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if isIgnorableFixedLine(line) {
			continue
		}
		entry, err := hipparcosEntry(line, lineNumber)
		if err != nil {
			if options.Strict {
				return nil, err
			}
			catalog.skipped++
			continue
		}
		if err := addEntry(catalog, entry, lineNumber, options); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	catalog.finalize()
	return catalog, nil
}

func hipparcosEntry(line []byte, lineNumber int) (Entry, error) {
	hip, present, err := fixedInt(line, 0, 6)
	if err != nil || !present || hip <= 0 {
		return Entry{}, fixedError(lineNumber, "invalid HIP")
	}
	ra, present, err := fixedFloat(line, 15, 28)
	if err != nil || !present {
		return Entry{}, fixedError(lineNumber, "invalid RArad")
	}
	dec, present, err := fixedFloat(line, 29, 42)
	if err != nil || !present {
		return Entry{}, fixedError(lineNumber, "invalid DErad")
	}
	parallax, _, err := fixedFloat(line, 43, 50)
	if err != nil {
		return Entry{}, fixedError(lineNumber, "invalid Plx")
	}
	pmRA, _, err := fixedFloat(line, 51, 59)
	if err != nil {
		return Entry{}, fixedError(lineNumber, "invalid pmRA")
	}
	pmDec, _, err := fixedFloat(line, 60, 68)
	if err != nil {
		return Entry{}, fixedError(lineNumber, "invalid pmDE")
	}
	magnitude, magnitudePresent, err := fixedFloat(line, 129, 136)
	if err != nil {
		return Entry{}, fixedError(lineNumber, "invalid Hpmag")
	}
	bv, bvPresent, err := fixedFloat(line, 152, 158)
	if err != nil {
		return Entry{}, fixedError(lineNumber, "invalid B-V")
	}
	id := fmt.Sprintf("HIP %d", hip)
	// 非正视差按无视差处理；非有限值原样保留，交给 Validate 拒收。
	if finite(parallax) && parallax <= 0 {
		parallax = 0
	}
	if !magnitudePresent {
		magnitude = math.NaN()
	}
	if !bvPresent {
		bv = math.NaN()
	}
	return Entry{
		ID: id,
		Coordinate: parseCoordinate(id, angularFromRadians(ra), angularFromRadians(dec), pmRA, pmDec, parallax, 0,
			julianYearTime(1991.25), basic.CoordinateFrameICRS),
		Magnitude: magnitude,
		Band:      BandHp,
		BPRP:      math.NaN(),
		BV:        bv,
	}, nil
}

// ReadBSC5 读取 BSC5 定宽星表 / read the BSC5 fixed-width catalog.
func ReadBSC5(r io.Reader, options ReadOptions) (*Catalog, error) {
	catalog := newCatalog()
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Bytes()
		if isIgnorableFixedLine(line) {
			continue
		}
		entry, present, err := bsc5Entry(line, lineNumber)
		if err != nil {
			if options.Strict {
				return nil, err
			}
			catalog.skipped++
			continue
		}
		if !present {
			catalog.skipped++
			continue
		}
		if err := addEntry(catalog, entry, lineNumber, options); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	catalog.finalize()
	return catalog, nil
}

func bsc5Entry(line []byte, lineNumber int) (Entry, bool, error) {
	hr, present, err := fixedInt(line, 0, 4)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid HR")
	}
	if !present || hr <= 0 {
		return Entry{}, false, nil
	}
	name, _ := fixedField(line, 4, 14)
	hd, _, err := fixedInt(line, 25, 31)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid HD")
	}
	sao, _, err := fixedInt(line, 31, 37)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid SAO")
	}
	raHour, raHourPresent, err := fixedInt(line, 75, 77)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid RA hour")
	}
	raMinute, raMinutePresent, err := fixedInt(line, 77, 79)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid RA minute")
	}
	raSecond, raSecondPresent, err := fixedFloat(line, 79, 83)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid RA second")
	}
	decSign, decSignPresent := fixedField(line, 83, 84)
	decDegree, decDegreePresent, err := fixedInt(line, 84, 86)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid Dec degree")
	}
	decMinute, decMinutePresent, err := fixedInt(line, 86, 88)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid Dec minute")
	}
	decSecond, decSecondPresent, err := fixedInt(line, 88, 90)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid Dec second")
	}
	if !raHourPresent || !raMinutePresent || !raSecondPresent || !decSignPresent || !decDegreePresent || !decMinutePresent || !decSecondPresent {
		return Entry{}, false, nil
	}
	ra := (float64(raHour) + float64(raMinute)/60 + raSecond/3600) * 15
	dec := float64(decDegree) + float64(decMinute)/60 + float64(decSecond)/3600
	if strings.HasPrefix(decSign, "-") {
		dec = -dec
	}
	vmag, vmPresent, err := fixedFloat(line, 102, 107)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid Vmag")
	}
	bv, bvPresent, err := fixedFloat(line, 109, 114)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid B-V")
	}
	pmRA, _, err := fixedFloat(line, 148, 154)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid pmRA")
	}
	pmDec, _, err := fixedFloat(line, 154, 160)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid pmDE")
	}
	parallax, _, err := fixedFloat(line, 161, 166)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid parallax")
	}
	radial, _, err := fixedInt(line, 166, 170)
	if err != nil {
		return Entry{}, false, fixedError(lineNumber, "invalid radial velocity")
	}
	id := fmt.Sprintf("HR %d", hr)
	aliases := make([]string, 0, 3)
	if name != "" {
		aliases = append(aliases, strings.Join(strings.Fields(name), " "))
	}
	if hd > 0 {
		aliases = append(aliases, fmt.Sprintf("HD %d", hd))
	}
	if sao > 0 {
		aliases = append(aliases, fmt.Sprintf("SAO %d", sao))
	}
	if !vmPresent {
		vmag = math.NaN()
	}
	if !bvPresent {
		bv = math.NaN()
	}
	// 非正视差按无视差处理；非有限值原样保留，交给 Validate 拒收。
	if finite(parallax) && parallax <= 0 {
		parallax = 0
	}
	return Entry{
		ID: id, Aliases: aliases,
		Coordinate: parseCoordinate(id, ra, dec, pmRA*1000, pmDec*1000, parallax*1000, float64(radial), time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), basic.CoordinateFrameJ2000),
		Magnitude:  vmag, Band: BandV, BPRP: math.NaN(), BV: bv,
	}, true, nil
}

func isIgnorableFixedLine(line []byte) bool {
	trimmed := strings.TrimSpace(string(line))
	return trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//")
}

func fixedError(line int, reason string) error {
	return fmt.Errorf("record %d: %s", line, reason)
}
