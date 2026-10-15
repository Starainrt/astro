package catalog

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
)

// ReadCSV 读取 CSV/TSV 星表 / read a CSV or TSV catalog.
func ReadCSV(r io.Reader, columns ColumnMap, options ReadOptions) (*Catalog, error) {
	buffered := bufio.NewReaderSize(r, 64*1024)
	header, headerLine, delimiter, err := readCSVHeader(buffered, options.Delimiter)
	if err != nil {
		return nil, err
	}
	indexes := buildColumnIndexes(header, columns)
	if columns.ID == "" || columns.RA == "" || columns.Dec == "" || indexes.id < 0 || indexes.ra < 0 || indexes.dec < 0 {
		return nil, fmt.Errorf("header line %d: required ID/RA/Dec column missing", headerLine)
	}
	if columns.EpochColumn != "" && indexes.epoch < 0 {
		return nil, fmt.Errorf("header line %d: epoch column %q missing", headerLine, columns.EpochColumn)
	}

	catalog := newCatalog()
	source := &csvLineSource{r: buffered}
	reader := csv.NewReader(source)
	reader.Comma = delimiter
	reader.Comment = '#'
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = true
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			var parseErr *csv.ParseError
			if !errors.As(err, &parseErr) {
				return nil, err
			}
			if options.Strict {
				return nil, fmt.Errorf("line %d: %w", parseErr.Line+headerLine, err)
			}
			catalog.skipped++
			continue
		}
		// 引号内换行是记录跨越的行数，从当前行回推即得记录首行。
		line := source.line + headerLine
		for _, field := range fields {
			line -= strings.Count(field, "\n")
		}
		first := strings.TrimSpace(fields[0])
		if (len(fields) == 1 && first == "") || strings.HasPrefix(first, "#") || strings.HasPrefix(first, "//") ||
			isDelimitedSeparator(fields) || isDelimitedUnits(fields, indexes) {
			continue
		}
		entry, parseErr := csvEntry(fields, indexes, columns, line)
		if parseErr != nil {
			if options.Strict {
				return nil, parseErr
			}
			catalog.skipped++
			continue
		}
		if err := addEntry(catalog, entry, line, options); err != nil {
			return nil, err
		}
	}
	catalog.finalize()
	return catalog, nil
}

// csvLineSource 每次 Read 至多交付一行，line 为已开始交付的行数；csv.Reader 只在需要时才读，因此 line 即其当前所在行。
type csvLineSource struct {
	r       *bufio.Reader
	pending []byte
	err     error
	line    int
	midLine bool
}

func (s *csvLineSource) Read(p []byte) (int, error) {
	if len(s.pending) == 0 {
		if s.err != nil {
			return 0, s.err
		}
		chunk, err := s.r.ReadSlice('\n')
		if err != nil && err != bufio.ErrBufferFull {
			s.err = err
		}
		if len(chunk) == 0 {
			return 0, s.err
		}
		if !s.midLine {
			s.line++
		}
		s.midLine = chunk[len(chunk)-1] != '\n'
		s.pending = chunk
	}
	n := copy(p, s.pending)
	s.pending = s.pending[n:]
	return n, nil
}

// readCSVHeader 跳过注释与空行读出表头；未指定分隔符时按表头里 Tab 与逗号的多少判定。
func readCSVHeader(r *bufio.Reader, delimiter rune) ([]string, int, rune, error) {
	lineNumber := 0
	for {
		line, err := r.ReadString('\n')
		if line == "" && err != nil {
			if err == io.EOF {
				return nil, 0, 0, fmt.Errorf("missing header")
			}
			return nil, 0, 0, err
		}
		lineNumber++
		if lineNumber == 1 {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//") {
			if err != nil {
				return nil, 0, 0, fmt.Errorf("missing header")
			}
			continue
		}
		if delimiter == 0 {
			delimiter = ','
			if strings.Count(line, "\t") > strings.Count(line, ",") {
				delimiter = '\t'
			}
		}
		reader := csv.NewReader(strings.NewReader(line))
		reader.Comma = delimiter
		reader.FieldsPerRecord = -1
		fields, parseErr := reader.Read()
		if parseErr != nil {
			return nil, 0, 0, fmt.Errorf("header line %d: %w", lineNumber, parseErr)
		}
		return fields, lineNumber, delimiter, nil
	}
}

func isDelimitedSeparator(fields []string) bool {
	if len(fields) == 0 {
		return true
	}
	for _, field := range fields {
		value := strings.TrimSpace(field)
		if value == "---" {
			return true
		}
	}
	return false
}

func isDelimitedUnits(fields []string, indexes columnIndexes) bool {
	if indexes.ra < 0 || indexes.dec < 0 || indexes.id < 0 {
		return false
	}
	ra := strings.ToLower(csvField(fields, indexes.ra))
	dec := strings.ToLower(csvField(fields, indexes.dec))
	if ra == "" || dec == "" || (strings.ContainsAny(ra, ".0123456789") && strings.ContainsAny(dec, ".0123456789")) {
		return false
	}
	units := []string{"deg", "rad", "mas", "mag", "km/s", "arcsec", "hour", "hours"}
	for _, value := range append([]string{ra, dec}, strings.ToLower(csvField(fields, indexes.id))) {
		for _, unit := range units {
			if strings.Contains(value, unit) {
				return true
			}
		}
	}
	return false
}

type columnIndexes struct {
	id, ra, dec, pmRA, pmDec, parallax, radial, magnitude, bprp, bv, epoch int
}

func buildColumnIndexes(header []string, columns ColumnMap) columnIndexes {
	lookup := make(map[string]int, len(header))
	for i, name := range header {
		lookup[normalizeColumn(name)] = i
	}
	indexes := columnIndexes{
		id: -1, ra: -1, dec: -1, pmRA: -1, pmDec: -1, parallax: -1, radial: -1, magnitude: -1, bprp: -1, bv: -1, epoch: -1,
	}
	set := func(name string) int {
		if name == "" {
			return -1
		}
		index, ok := lookup[normalizeColumn(name)]
		if !ok {
			return -1
		}
		return index
	}
	indexes.id, indexes.ra, indexes.dec = set(columns.ID), set(columns.RA), set(columns.Dec)
	indexes.pmRA, indexes.pmDec, indexes.parallax = set(columns.PMRA), set(columns.PMDec), set(columns.Parallax)
	indexes.radial, indexes.magnitude, indexes.bprp = set(columns.RadialVelocity), set(columns.Magnitude), set(columns.BPRP)
	indexes.bv, indexes.epoch = set(columns.BV), set(columns.EpochColumn)
	return indexes
}

func normalizeColumn(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func csvField(fields []string, index int) string {
	if index < 0 || index >= len(fields) {
		return ""
	}
	return strings.TrimSpace(fields[index])
}

func csvEntry(fields []string, indexes columnIndexes, columns ColumnMap, line int) (Entry, error) {
	if columns.ID == "" || columns.RA == "" || columns.Dec == "" || indexes.id < 0 || indexes.ra < 0 || indexes.dec < 0 {
		return Entry{}, fmt.Errorf("line %d: required ID/RA/Dec column missing", line)
	}
	id := strings.TrimSpace(csvField(fields, indexes.id))
	if id == "" {
		return Entry{}, fmt.Errorf("line %d: missing ID", line)
	}
	if columns.IDPrefix != "" {
		id = columns.IDPrefix + id
	}
	ra, raPresent, err := parseFloat(csvField(fields, indexes.ra))
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid RA: %w", line, err)
	}
	if !raPresent {
		return Entry{}, fmt.Errorf("line %d: missing RA", line)
	}
	dec, decPresent, err := parseFloat(csvField(fields, indexes.dec))
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid Dec: %w", line, err)
	}
	if !decPresent {
		return Entry{}, fmt.Errorf("line %d: missing Dec", line)
	}
	if columns.AnglesInRadians {
		ra = angularFromRadians(ra)
		dec = angularFromRadians(dec)
	}
	pmRA, _, err := optionalFloat(fields, indexes.pmRA)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid PMRA: %w", line, err)
	}
	pmDec, _, err := optionalFloat(fields, indexes.pmDec)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid PMDec: %w", line, err)
	}
	if !columns.PMRAIncludesCosDec {
		pmRA *= math.Cos(dec * math.Pi / 180)
	}
	parallax, _, err := optionalFloat(fields, indexes.parallax)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid parallax: %w", line, err)
	}
	radial, _, err := optionalFloat(fields, indexes.radial)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid radial velocity: %w", line, err)
	}
	magnitude, magnitudePresent, err := optionalFloat(fields, indexes.magnitude)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid magnitude: %w", line, err)
	}
	bprp, bprpPresent, err := optionalFloat(fields, indexes.bprp)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid BP-RP: %w", line, err)
	}
	bv, bvPresent, err := optionalFloat(fields, indexes.bv)
	if err != nil {
		return Entry{}, fmt.Errorf("line %d: invalid B-V: %w", line, err)
	}
	epoch := columns.Epoch
	if columns.EpochColumn != "" {
		year, present, epochErr := optionalFloat(fields, indexes.epoch)
		if epochErr != nil {
			return Entry{}, fmt.Errorf("line %d: invalid epoch: %w", line, epochErr)
		}
		if present {
			epoch = julianYearTime(year)
		}
	}
	if epoch.IsZero() {
		return Entry{}, fmt.Errorf("line %d: epoch is required", line)
	}
	if !magnitudePresent {
		magnitude = math.NaN()
	}
	if !bprpPresent {
		bprp = math.NaN()
	}
	if !bvPresent {
		bv = math.NaN()
	}
	return Entry{
		ID: id, Coordinate: parseCoordinate(id, ra, dec, pmRA, pmDec, parallax, radial, epoch, columns.Frame),
		Magnitude: magnitude, Band: columns.Band, BPRP: bprp, BV: bv,
	}, nil
}

func optionalFloat(fields []string, index int) (float64, bool, error) {
	if index < 0 {
		return 0, false, nil
	}
	return parseFloat(csvField(fields, index))
}
