package catalog

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
)

// StellariumOptions 控制 Stellarium 附加文件 / Stellarium side-file options.
type StellariumOptions struct {
	MaxMagnitude float64 // 0 表示不限，可为负 / 0 means unlimited; may be negative
	Names        bool    // 读 name.fab 与 extra_name.fab / read name files
	CrossID      bool    // 读 cross-id.cat / read cross identifiers
}

const stellariumMagic uint32 = 0x835F040A

// ReadStellariumStars 读取单个 Stellarium 星表分片 / read one Stellarium star shard.
func ReadStellariumStars(r io.Reader, read ReadOptions) (*Catalog, error) {
	return readStellariumStars(r, StellariumOptions{}, read)
}

func readStellariumStars(r io.Reader, stellarium StellariumOptions, read ReadOptions) (*Catalog, error) {
	reader := bufio.NewReaderSize(r, 64*1024)
	var headerBytes [28]byte
	if _, err := io.ReadFull(reader, headerBytes[:]); err != nil {
		return nil, fmt.Errorf("stellarium header: %w", err)
	}
	if binary.LittleEndian.Uint32(headerBytes[0:4]) != stellariumMagic {
		if binary.BigEndian.Uint32(headerBytes[0:4]) == stellariumMagic {
			return nil, fmt.Errorf("stellarium big-endian files are unsupported")
		}
		return nil, fmt.Errorf("stellarium header: invalid magic")
	}
	typeID := binary.LittleEndian.Uint32(headerBytes[4:8])
	major := binary.LittleEndian.Uint32(headerBytes[8:12])
	minor := binary.LittleEndian.Uint32(headerBytes[12:16])
	level := binary.LittleEndian.Uint32(headerBytes[16:20])
	magMin := int32(binary.LittleEndian.Uint32(headerBytes[20:24]))
	epochJD := math.Float32frombits(binary.LittleEndian.Uint32(headerBytes[24:28]))
	if major > 0 {
		return nil, fmt.Errorf("stellarium header: unsupported major version %d", major)
	}
	if typeID > 2 {
		return nil, fmt.Errorf("stellarium header: unsupported record type %d", typeID)
	}
	if level > 8 {
		return nil, fmt.Errorf("stellarium header: unsupported level %d", level)
	}
	if epochJD != 2457389.0 {
		return nil, fmt.Errorf("stellarium header: unsupported epoch %.7g", epochJD)
	}
	if stellarium.MaxMagnitude != 0 && float64(magMin)/1000 > stellarium.MaxMagnitude {
		catalog := newCatalog()
		catalog.finalize()
		return catalog, nil
	}
	zoneCount := uint64(20)
	for i := uint32(0); i < level; i++ {
		zoneCount *= 4
	}
	zoneCount++
	counts := make([]uint32, zoneCount)
	var countBytes [4]byte
	for i := range counts {
		if _, err := io.ReadFull(reader, countBytes[:]); err != nil {
			return nil, fmt.Errorf("stellarium counts: %w", err)
		}
		counts[i] = binary.LittleEndian.Uint32(countBytes[:])
	}
	catalog := newCatalog()
	recordSize := map[uint32]int{0: 48, 1: 32, 2: 16}[typeID]
	recordNumber := 0
	buffer := make([]byte, recordSize)
	epoch := basic.JD2DateByZone(float64(epochJD), time.UTC, false)
	for _, count := range counts {
		for i := uint32(0); i < count; i++ {
			recordNumber++
			if _, err := io.ReadFull(reader, buffer); err != nil {
				if read.Strict {
					return nil, fmt.Errorf("record %d: truncated record: %w", recordNumber, err)
				}
				catalog.skipped++
				catalog.finalize()
				return catalog, nil
			}
			entry, err := stellariumEntry(buffer, typeID, epoch, recordNumber)
			if err != nil {
				if read.Strict {
					return nil, err
				}
				catalog.skipped++
				continue
			}
			if stellarium.MaxMagnitude != 0 && (!finite(entry.Magnitude) || entry.Magnitude > stellarium.MaxMagnitude) {
				continue
			}
			if err := addEntry(catalog, entry, recordNumber, read); err != nil {
				return nil, err
			}
		}
	}
	var trailing [1]byte
	if n, err := reader.Read(trailing[:]); n != 0 || err != io.EOF {
		if read.Strict {
			if err == nil {
				return nil, fmt.Errorf("stellarium file has trailing bytes")
			}
			return nil, fmt.Errorf("stellarium file length: %w", err)
		}
		catalog.skipped++
	}
	_ = minor
	catalog.finalize()
	return catalog, nil
}

func stellariumEntry(record []byte, typeID uint32, epoch time.Time, recordNumber int) (Entry, error) {
	switch typeID {
	case 0:
		return stellariumStar1(record, epoch, recordNumber)
	case 1:
		return stellariumStar2(record, epoch, recordNumber)
	case 2:
		return stellariumStar3(record, epoch, recordNumber)
	default:
		return Entry{}, fmt.Errorf("record %d: unsupported record type %d", recordNumber, typeID)
	}
}

func stellariumStar1(record []byte, epoch time.Time, recordNumber int) (Entry, error) {
	gaiaID := int64(binary.LittleEndian.Uint64(record[0:8]))
	x := [3]float64{
		float64(int32(binary.LittleEndian.Uint32(record[8:12]))) / 2e9,
		float64(int32(binary.LittleEndian.Uint32(record[12:16]))) / 2e9,
		float64(int32(binary.LittleEndian.Uint32(record[16:20]))) / 2e9,
	}
	dx := [3]float64{
		float64(int32(binary.LittleEndian.Uint32(record[20:24]))) / 1000,
		float64(int32(binary.LittleEndian.Uint32(record[24:28]))) / 1000,
		float64(int32(binary.LittleEndian.Uint32(record[28:32]))) / 1000,
	}
	norm := math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
	if !finite(norm) || norm == 0 {
		return Entry{}, fmt.Errorf("record %d: zero position vector", recordNumber)
	}
	ra := normalizeRA(math.Atan2(x[1], x[0]) * 180 / math.Pi)
	dec := math.Asin(clamp(x[2]/norm, -1, 1)) * 180 / math.Pi
	alpha := ra * math.Pi / 180
	delta := dec * math.Pi / 180
	pmRA := -math.Sin(alpha)*dx[0] + math.Cos(alpha)*dx[1]
	pmDec := -math.Sin(delta)*math.Cos(alpha)*dx[0] - math.Sin(delta)*math.Sin(alpha)*dx[1] + math.Cos(delta)*dx[2]
	radialPM := math.Cos(delta)*math.Cos(alpha)*dx[0] + math.Cos(delta)*math.Sin(alpha)*dx[1] + math.Sin(delta)*dx[2]
	if !finite(radialPM) || math.Abs(radialPM) > 1 {
		return Entry{}, fmt.Errorf("record %d: non-tangential proper motion", recordNumber)
	}
	bv := float64(int16(binary.LittleEndian.Uint16(record[32:34]))) / 1000
	vmag := float64(int16(binary.LittleEndian.Uint16(record[34:36]))) / 1000
	parallax := float64(binary.LittleEndian.Uint16(record[36:38])) * 0.02
	radial := float64(int16(binary.LittleEndian.Uint16(record[40:42]))) / 10
	hipPacked := uint32(record[45]) | uint32(record[46])<<8 | uint32(record[47])<<16
	hip := hipPacked >> 5
	componentNumber := hipPacked & 31
	component := componentName(componentNumber)
	id := ""
	aliases := make([]string, 0, 1)
	if hip > 0 {
		id = fmt.Sprintf("HIP %d", hip)
		if gaiaID > 0 {
			aliases = append(aliases, fmt.Sprintf("Gaia DR3 %d", gaiaID))
		}
	} else if gaiaID > 0 {
		id = fmt.Sprintf("Gaia DR3 %d", gaiaID)
	}
	if id == "" {
		return Entry{}, fmt.Errorf("record %d: missing HIP and Gaia ID", recordNumber)
	}
	return Entry{ID: id, Aliases: aliases, Component: component, Coordinate: parseCoordinate(id, ra, dec, pmRA, pmDec, parallax, radial, epoch, basic.CoordinateFrameICRS), Magnitude: vmag, Band: BandV, BPRP: math.NaN(), BV: bv}, nil
}

func stellariumStar2(record []byte, epoch time.Time, recordNumber int) (Entry, error) {
	gaiaID := int64(binary.LittleEndian.Uint64(record[0:8]))
	ra := float64(int32(binary.LittleEndian.Uint32(record[8:12]))) / 3600000
	dec := float64(int32(binary.LittleEndian.Uint32(record[12:16]))) / 3600000
	pmRA := float64(int32(binary.LittleEndian.Uint32(record[16:20]))) / 1000
	pmDec := float64(int32(binary.LittleEndian.Uint32(record[20:24]))) / 1000
	pmRA *= math.Cos(dec * math.Pi / 180)
	bv := float64(int16(binary.LittleEndian.Uint16(record[24:26]))) / 1000
	vmag := float64(int16(binary.LittleEndian.Uint16(record[26:28]))) / 1000
	parallax := float64(binary.LittleEndian.Uint16(record[28:30])) / 100
	if gaiaID <= 0 {
		return Entry{}, fmt.Errorf("record %d: missing Gaia ID", recordNumber)
	}
	id := fmt.Sprintf("Gaia DR3 %d", gaiaID)
	return Entry{ID: id, Coordinate: parseCoordinate(id, ra, dec, pmRA, pmDec, parallax, 0, epoch, basic.CoordinateFrameICRS), Magnitude: vmag, Band: BandV, BPRP: math.NaN(), BV: bv}, nil
}

func stellariumStar3(record []byte, epoch time.Time, recordNumber int) (Entry, error) {
	raValue := uint32(record[8]) | uint32(record[9])<<8 | uint32(record[10])<<16
	decValue := uint32(record[11]) | uint32(record[12])<<8 | uint32(record[13])<<16
	ra := float64(raValue) * 0.1 / 3600
	dec := (float64(decValue) - 90*36000) * 0.1 / 3600
	vmag := (float64(record[15])*20 + 16000) / 1000
	bv := float64(record[14])*0.025 - 1
	gaiaID := int64(binary.LittleEndian.Uint64(record[0:8]))
	if gaiaID <= 0 {
		return Entry{}, fmt.Errorf("record %d: missing Gaia ID", recordNumber)
	}
	id := fmt.Sprintf("Gaia DR3 %d", gaiaID)
	return Entry{ID: id, Coordinate: parseCoordinate(id, ra, dec, 0, 0, 0, 0, epoch, basic.CoordinateFrameICRS), Magnitude: vmag, Band: BandV, BPRP: math.NaN(), BV: bv}, nil
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func componentName(value uint32) string {
	if value == 0 {
		return ""
	}
	if value <= 26 {
		return string(rune('A' + value - 1))
	}
	return strconv.FormatUint(uint64(value), 10)
}

// ReadStellariumDir 读取 Stellarium 星表目录 / read a Stellarium catalog directory.
func ReadStellariumDir(dir fs.FS, stellarium StellariumOptions, read ReadOptions) (*Catalog, error) {
	entries, err := fs.ReadDir(dir, ".")
	if err != nil {
		return nil, err
	}
	shards := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && isStellariumShard(entry.Name()) {
			shards = append(shards, entry.Name())
		}
	}
	sort.Strings(shards)
	merged := newCatalog()
	for _, name := range shards {
		file, openErr := dir.Open(name)
		if openErr != nil {
			return nil, openErr
		}
		part, readErr := readStellariumStars(file, stellarium, read)
		closeErr := file.Close()
		if readErr != nil {
			return nil, fmt.Errorf("%s: %w", name, readErr)
		}
		if closeErr != nil {
			return nil, closeErr
		}
		merged = merged.Merge(part)
	}
	if stellarium.Names {
		for _, name := range []string{"name.fab", "extra_name.fab"} {
			if err := addFabAliases(dir, merged, name, read); err != nil {
				return nil, err
			}
		}
	}
	if stellarium.CrossID {
		if err := addCrossIDAliases(dir, merged, "cross-id.cat", read); err != nil {
			return nil, err
		}
	}
	merged.finalize()
	return merged, nil
}

func isStellariumShard(name string) bool {
	if !strings.HasPrefix(name, "stars_") || !strings.HasSuffix(name, ".cat") {
		return false
	}
	rest := strings.TrimPrefix(strings.TrimSuffix(name, ".cat"), "stars_")
	if rest == "" || rest[0] < '0' || rest[0] > '9' {
		return false
	}
	return strings.Contains(rest, "_")
}

func addFabAliases(dir fs.FS, catalog *Catalog, name string, read ReadOptions) error {
	file, err := dir.Open(name)
	if err != nil {
		if errorsIsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	return applyNameLines(file, catalog, name, read)
}

// applyNameLines 读 `id|名称` 行并把名称并入对应记录的别名。
func applyNameLines(r io.Reader, catalog *Catalog, label string, read ReadOptions) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	index := buildCatalogNumberIndex(catalog)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		id, err := uint64(0), error(nil)
		if len(parts) == 2 {
			id, err = strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 64)
		}
		alias := ""
		if len(parts) == 2 && err == nil {
			alias = nameField(parts[1])
		}
		if alias == "" {
			if read.Strict {
				return fmt.Errorf("%s line %d: expected id|name", label, lineNumber)
			}
			catalog.skipped++
			continue
		}
		for _, i := range index.lookup(id, 0) {
			catalog.entries[i].Aliases = append(catalog.entries[i].Aliases, alias)
		}
	}
	return scanner.Err()
}

// nameField 取名称字段：`_("Sirius") 2` 取引号内文字，否则把 `_` 换成空格。
func nameField(field string) string {
	field = strings.TrimSpace(field)
	if strings.HasPrefix(field, `_("`) {
		end := strings.Index(field[3:], `")`)
		if end < 0 {
			return ""
		}
		return strings.TrimSpace(field[3 : 3+end])
	}
	return strings.TrimSpace(strings.ReplaceAll(field, "_", " "))
}

func addCrossIDAliases(dir fs.FS, catalog *Catalog, name string, read ReadOptions) error {
	file, err := dir.Open(name)
	if err != nil {
		if errorsIsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64*1024)
	index := buildCatalogNumberIndex(catalog)
	var record [21]byte
	recordNumber := 0
	for {
		n, readErr := io.ReadFull(reader, record[:])
		if readErr == io.EOF && n == 0 {
			break
		}
		recordNumber++
		if readErr != nil {
			if read.Strict {
				return fmt.Errorf("%s record %d: length is not 21 bytes", name, recordNumber)
			}
			catalog.skipped++
			break
		}
		id := binary.LittleEndian.Uint64(record[0:8])
		component := uint32(record[8])
		aliases := make([]string, 0, 3)
		if value := int32(binary.LittleEndian.Uint32(record[9:13])); value != 0 {
			aliases = append(aliases, fmt.Sprintf("SAO %d", value))
		}
		if value := int32(binary.LittleEndian.Uint32(record[13:17])); value != 0 {
			aliases = append(aliases, fmt.Sprintf("HD %d", value))
		}
		if value := int32(binary.LittleEndian.Uint32(record[17:21])); value != 0 {
			aliases = append(aliases, fmt.Sprintf("HR %d", value))
		}
		for _, i := range index.lookup(id, component) {
			catalog.entries[i].Aliases = append(catalog.entries[i].Aliases, aliases...)
		}
	}
	return nil
}

type catalogNumberIndex struct {
	hip, gaia     map[uint64][]int
	hipComponents map[string][]int
}

func buildCatalogNumberIndex(catalog *Catalog) catalogNumberIndex {
	index := catalogNumberIndex{hip: make(map[uint64][]int), gaia: make(map[uint64][]int), hipComponents: make(map[string][]int)}
	for i, entry := range catalog.entries {
		if strings.HasPrefix(entry.ID, "HIP ") {
			value, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(entry.ID, "HIP ")), 10, 64)
			if err == nil {
				index.hip[value] = append(index.hip[value], i)
				index.hipComponents[fmt.Sprintf("%d\x00%s", value, entry.Component)] = append(index.hipComponents[fmt.Sprintf("%d\x00%s", value, entry.Component)], i)
			}
		} else if strings.HasPrefix(entry.ID, "Gaia DR3 ") {
			value, err := strconv.ParseUint(strings.TrimSpace(strings.TrimPrefix(entry.ID, "Gaia DR3 ")), 10, 64)
			if err == nil {
				index.gaia[value] = append(index.gaia[value], i)
			}
		}
	}
	return index
}

func (index catalogNumberIndex) lookup(id uint64, component uint32) []int {
	if id <= 120416 {
		if component == 0 {
			return index.hip[id]
		}
		return index.hipComponents[fmt.Sprintf("%d\x00%s", id, componentName(component))]
	}
	if component != 0 {
		return nil
	}
	return index.gaia[id]
}

func errorsIsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
