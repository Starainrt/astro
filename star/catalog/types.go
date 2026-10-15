package catalog

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/starainrt/astro/basic"
)

// Entry 一条外挂星表记录 / one external catalog record.
type Entry struct {
	ID         string
	Aliases    []string
	Coordinate basic.StarCoordinate
	Magnitude  float64
	Band       Band
	BPRP       float64
	BV         float64
	Component  string
}

// Band 星等波段 / photometric band.
type Band string

const (
	// BandV Johnson V 波段 / Johnson V band.
	BandV Band = "V"
	// BandG Gaia G 波段 / Gaia G band.
	BandG Band = "G"
	// BandHp Hipparcos Hp 波段 / Hipparcos Hp band.
	BandHp Band = "Hp"
)

// ColumnMap 描述分隔文本的列与单位 / delimited-text columns and units.
type ColumnMap struct {
	ID, RA, Dec, PMRA, PMDec, Parallax, RadialVelocity, Magnitude, BPRP, BV string
	PMRAIncludesCosDec                                                      bool
	Band                                                                    Band
	Epoch                                                                   time.Time
	EpochColumn                                                             string
	Frame                                                                   basic.CoordinateFrame
	IDPrefix                                                                string
	AnglesInRadians                                                         bool
}

// ReadOptions 控制读取严格性和入选过滤 / read strictness and admission filtering.
type ReadOptions struct {
	Strict    bool
	Delimiter rune
	Filter    func(Entry) bool
}

// Catalog 是不可变的外挂星表容器 / immutable external-star catalog.
type Catalog struct {
	entries []Entry
	lookup  map[string]int
	// bands 按 0.5° 赤纬分带，带内按 RA 升序存放记录下标。
	bands   [decBandCount][]int
	skipped int
}

func newCatalog() *Catalog {
	return &Catalog{lookup: make(map[string]int)}
}

func (c *Catalog) add(entry Entry) {
	entry.ID = strings.TrimSpace(entry.ID)
	entry.Aliases = uniqueStrings(entry.Aliases)
	c.entries = append(c.entries, entry)
}

// dropDuplicates 同主 ID 同分量只保留首条，其余计入跳过数。
func (c *Catalog) dropDuplicates() {
	seen := make(map[string]struct{}, len(c.entries))
	kept := c.entries[:0]
	for _, entry := range c.entries {
		if normalizeKey(entry.ID) != "" {
			key := entryKey(entry)
			if _, exists := seen[key]; exists {
				c.skipped++
				continue
			}
			seen[key] = struct{}{}
		}
		kept = append(kept, entry)
	}
	c.entries = kept
}

func (c *Catalog) finalize() {
	c.dropDuplicates()
	c.lookup = make(map[string]int, len(c.entries)*2)
	c.bands = [decBandCount][]int{}
	for i := range c.entries {
		entry := &c.entries[i]
		entry.Aliases = uniqueStrings(entry.Aliases)
		if key := normalizeKey(entry.ID); key != "" {
			if _, exists := c.lookup[key]; !exists {
				c.lookup[key] = i
			}
		}
		for _, alias := range entry.Aliases {
			key := normalizeKey(alias)
			if key != "" {
				if _, exists := c.lookup[key]; !exists {
					c.lookup[key] = i
				}
			}
		}
		band := decBand(entry.Coordinate.Dec)
		c.bands[band] = append(c.bands[band], i)
	}
	for band := range c.bands {
		if len(c.bands[band]) < 2 {
			continue
		}
		sort.SliceStable(c.bands[band], func(i, j int) bool {
			return c.entries[c.bands[band][i]].Coordinate.RA < c.entries[c.bands[band][j]].Coordinate.RA
		})
	}
}

// Len 返回记录数 / record count.
func (c *Catalog) Len() int {
	if c == nil {
		return 0
	}
	return len(c.entries)
}

// Entries 返回记录快照 / record snapshot.
func (c *Catalog) Entries() []Entry {
	if c == nil {
		return nil
	}
	entries := make([]Entry, len(c.entries))
	copy(entries, c.entries)
	for i := range entries {
		entries[i].Aliases = append([]string(nil), entries[i].Aliases...)
	}
	return entries
}

// Lookup 按主 ID 或别名精确查找，同 ID 多分量时返回最先读入者 / exact lookup by ID or alias; first-read component wins.
func (c *Catalog) Lookup(id string) (Entry, bool) {
	if c == nil {
		return Entry{}, false
	}
	i, ok := c.lookup[normalizeKey(id)]
	if !ok {
		return Entry{}, false
	}
	entry := c.entries[i]
	entry.Aliases = append([]string(nil), entry.Aliases...)
	return entry, true
}

// Skipped 返回宽松模式跳过的记录数 / number of skipped records.
func (c *Catalog) Skipped() int {
	if c == nil {
		return 0
	}
	return c.skipped
}

// Merge 合并两个星表，后者覆盖同主 ID 同分量记录 / merge catalogs; right-hand records win per ID and component.
func (c *Catalog) Merge(other *Catalog) *Catalog {
	merged := newCatalog()
	if c != nil {
		for _, entry := range c.entries {
			merged.add(cloneEntry(entry))
		}
		merged.skipped += c.skipped
	}
	if other != nil {
		positions := make(map[string]int, len(merged.entries))
		for i := range merged.entries {
			positions[entryKey(merged.entries[i])] = i
		}
		for _, entry := range other.entries {
			key := entryKey(entry)
			if i, ok := positions[key]; ok {
				aliases := append(merged.entries[i].Aliases, entry.Aliases...)
				merged.entries[i] = cloneEntry(entry)
				merged.entries[i].Aliases = uniqueStrings(aliases)
			} else {
				positions[key] = len(merged.entries)
				merged.add(cloneEntry(entry))
			}
		}
		merged.skipped += other.skipped
	}
	merged.finalize()
	return merged
}

// entryKey 以主 ID 与分量区分记录，双星各分量共用主 ID。
func entryKey(entry Entry) string {
	return normalizeKey(entry.ID) + "\x00" + entry.Component
}

func cloneEntry(entry Entry) Entry {
	entry.Aliases = append([]string(nil), entry.Aliases...)
	return entry
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
		if value == "" {
			continue
		}
		key := normalizeKey(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

const decBandCount = 360

func decBand(dec float64) int {
	band := int(math.Floor((dec + 90) * 2))
	if band < 0 {
		return 0
	}
	if band >= decBandCount {
		return decBandCount - 1
	}
	return band
}
