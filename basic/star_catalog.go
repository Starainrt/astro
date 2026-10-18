package basic

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// 本文件包含约 9100 颗亮星 / this file contains bright 9100 stars
// 9100颗亮星列表

type InnerStarData struct {
	HR     uint16  //Bright Star Number;[1/9110]+ Harvard Revised Number;亮星编号
	Name   string  //Name, generally Bayer(如天狼星：Alpha CMA) and/or Flamsteed(如天狼星：9 CMA) name
	HD     uint32  //Henry Draper Catalog Number;HD星表编号
	Ra     float64 //Ra J2000;J2000历元赤经
	Dec    float64 //De J2000;J2000历元赤纬
	Mag    float64 //视星等
	PmRA   float64 //赤经投影年自行 cos(赤纬)*d赤经/dt，单位角秒/年
	PmDec  float64 //赤纬年自行，单位角秒/年
	RadVel float64 //径向速度 km/s
	RotVel float64 //自行速度 km/s
	Pc     float64 //秒差距
	HIP    uint32  //HIP星表编号
}
type StarData struct {
	InnerStarData
	ChineseName      string
	ChineseAlias     string
	ChineseBayerName string
	CommonName       string
	CommonAliasName  string
	Cst              string
	CstChinese       string
	ChineseRegion    string // 星名所属的宿/垣限定，用于区分同名星官 / mansion or enclosure qualifier.
}

var hr2name map[uint16]starNameRecord
var hr2hip map[uint16]uint32
var chnidx map[string][]uint16
var engidx map[string][]uint16
var parsedStarData []InnerStarData
var cachedStarData []StarData
var starDataOnce sync.Once
var starDataErr error

// ErrStarNotFound 未找到匹配的恒星 / no matching star.
var ErrStarNotFound = errors.New("star not found")

// ErrStarNameAmbiguous 星名对应多颗恒星 / the star name matches more than one star.
var ErrStarNameAmbiguous = errors.New("ambiguous star name")

func LoadStarData() error {
	starDataOnce.Do(func() {
		starDataErr = initStarData()
	})
	return starDataErr
}

func initStarData() error {
	parsedStarData = initStarCatalogData()
	cachedStarData = make([]StarData, len(parsedStarData))
	for i, parsed := range parsedStarData {
		cachedStarData[i] = fullStarData(parsed)
	}
	chnidx = make(map[string][]uint16, len(hr2name)*2)
	engidx = make(map[string][]uint16, len(hr2name))
	for hr := 1; hr <= len(cachedStarData); hr++ {
		rec, ok := hr2name[uint16(hr)]
		if !ok {
			continue
		}
		registerStarChineseIndex(uint16(hr), rec.chinese)
		for _, alias := range rec.aliases {
			registerStarChineseIndex(uint16(hr), alias)
			if rec.region != "" {
				registerStarChineseIndex(uint16(hr), alias+"["+rec.region+"]")
			}
		}
		if rec.chinese != "" && rec.region != "" {
			registerStarChineseIndex(uint16(hr), rec.chinese+"["+rec.region+"]")
		}
		addStarEnglishIndex(uint16(hr), rec.common)
	}
	return nil
}

func registerStarChineseIndex(hr uint16, name string) {
	if name == "" {
		return
	}
	addStarChineseIndex(hr, name)
	if strings.HasSuffix(name, "星") {
		trimmed := strings.TrimSuffix(name, "星")
		if trimmed != "" {
			addStarChineseIndex(hr, trimmed)
		}
	}
}

func addStarChineseIndex(hr uint16, name string) {
	list := chnidx[name]
	for _, existing := range list {
		if existing == hr {
			return
		}
	}
	chnidx[name] = append(list, hr)
}

func normalizeStarChineseName(name string) string {
	name = strings.TrimSpace(name)
	if strings.HasSuffix(name, "星") {
		name = strings.TrimSuffix(name, "星")
	}
	return name
}

func normalizeStarEnglishName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func addStarEnglishIndex(hr uint16, name string) {
	name = normalizeStarEnglishName(name)
	if name == "" {
		return
	}
	list := engidx[name]
	for _, existing := range list {
		if existing == hr {
			return
		}
	}
	engidx[name] = append(list, hr)
}

func starDataByEnglishAll(name string) ([]StarData, error) {
	if err := LoadStarData(); err != nil {
		return nil, err
	}
	name = normalizeStarEnglishName(name)
	hrs := engidx[name]
	if len(hrs) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrStarNotFound, name)
	}
	out := make([]StarData, 0, len(hrs))
	for _, hr := range hrs {
		if hr == 0 || int(hr) > len(cachedStarData) {
			continue
		}
		out = append(out, cachedStarData[hr-1])
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrStarNotFound, name)
	}
	return out, nil
}

func fullStarData(star InnerStarData) StarData {
	star.HIP = hr2hip[star.HR]
	rec, ok := hr2name[star.HR]
	if !ok {
		return StarData{InnerStarData: star}
	}
	return StarData{
		InnerStarData:    star,
		ChineseName:      rec.chinese,
		ChineseAlias:     strings.Join(rec.aliases, ";"),
		ChineseRegion:    rec.region,
		ChineseBayerName: rec.bayerZH,
		CommonName:       rec.common,
		Cst:              rec.cst,
		CstChinese:       rec.cstZH,
	}
}

// StarDataByChinese 按中文名或别名查唯一恒星；同名多星返回 ErrStarNameAmbiguous / returns the unique star, or ErrStarNameAmbiguous for duplicate names.
func StarDataByChinese(name string) (StarData, error) {
	matches, err := StarDataByChineseAll(name)
	if err != nil {
		return StarData{}, err
	}
	if len(matches) != 1 {
		return StarData{}, fmt.Errorf("%w: %q has %d candidates", ErrStarNameAmbiguous, name, len(matches))
	}
	return matches[0], nil
}

// StarDataByChineseAll 返回中文名或别名的全部候选，按 HR 升序 / all candidates by HR order.
func StarDataByChineseAll(name string) ([]StarData, error) {
	if err := LoadStarData(); err != nil {
		return nil, err
	}
	name = normalizeStarChineseName(name)
	hrs := chnidx[name]
	if len(hrs) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrStarNotFound, name)
	}
	out := make([]StarData, 0, len(hrs))
	for _, hr := range hrs {
		if hr == 0 || int(hr) > len(cachedStarData) {
			continue
		}
		out = append(out, cachedStarData[hr-1])
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrStarNotFound, name)
	}
	return out, nil
}

// StarDataByName 按中文名、中文别名或英文常用名查唯一恒星；同名多星返回 ErrStarNameAmbiguous / returns the unique star by Chinese name, alias or English common name, or ErrStarNameAmbiguous for duplicate names.
func StarDataByName(name string) (StarData, error) {
	matches, err := StarDataByNameAll(name)
	if err != nil {
		return StarData{}, err
	}
	if len(matches) != 1 {
		return StarData{}, fmt.Errorf("%w: %q has %d candidates", ErrStarNameAmbiguous, name, len(matches))
	}
	return matches[0], nil
}

// StarDataByNameAll 返回中文名、别名或英文常用名的全部候选，按 HR 升序 / all candidates by HR order.
func StarDataByNameAll(name string) ([]StarData, error) {
	if err := LoadStarData(); err != nil {
		return nil, err
	}
	chinese, chineseErr := StarDataByChineseAll(name)
	english, englishErr := starDataByEnglishAll(name)
	if chineseErr != nil && !errors.Is(chineseErr, ErrStarNotFound) {
		return nil, chineseErr
	}
	if englishErr != nil && !errors.Is(englishErr, ErrStarNotFound) {
		return nil, englishErr
	}
	matches := append(append([]StarData{}, chinese...), english...)
	sort.Slice(matches, func(i, j int) bool { return matches[i].HR < matches[j].HR })
	out := matches[:0]
	for _, star := range matches {
		if len(out) == 0 || out[len(out)-1].HR != star.HR {
			out = append(out, star)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%w: %q", ErrStarNotFound, name)
	}
	return out, nil
}

// StarDataByNameIn 按中文名、别名或英文常用名与宿/垣限定查唯一恒星 / looks up the unique star within a mansion or enclosure.
func StarDataByNameIn(name, region string) (StarData, error) {
	region = strings.TrimSpace(region)
	if region == "" {
		return StarDataByName(name)
	}
	matches, err := StarDataByNameAll(name)
	if err != nil {
		return StarData{}, err
	}
	var found []StarData
	for _, star := range matches {
		if star.ChineseRegion == region {
			found = append(found, star)
		}
	}
	switch len(found) {
	case 0:
		return StarData{}, fmt.Errorf("%w: %q in %q", ErrStarNotFound, name, region)
	case 1:
		return found[0], nil
	default:
		return StarData{}, fmt.Errorf("%w: %q in %q has %d candidates", ErrStarNameAmbiguous, name, region, len(found))
	}
}

// StarDataByChineseIn 按中文名或别名与宿/垣限定查唯一恒星 / looks up the unique star within a mansion or enclosure.
func StarDataByChineseIn(name, region string) (StarData, error) {
	region = strings.TrimSpace(region)
	if region == "" {
		return StarDataByChinese(name)
	}
	matches, err := StarDataByChineseAll(name)
	if err != nil {
		return StarData{}, err
	}
	var found []StarData
	for _, star := range matches {
		if star.ChineseRegion == region {
			found = append(found, star)
		}
	}
	switch len(found) {
	case 0:
		return StarData{}, fmt.Errorf("%w: %q in %q", ErrStarNotFound, name, region)
	case 1:
		return found[0], nil
	default:
		return StarData{}, fmt.Errorf("%w: %q in %q has %d candidates", ErrStarNameAmbiguous, name, region, len(found))
	}
}

func StarDataByHR(hr int) (StarData, error) {
	if err := LoadStarData(); err != nil {
		return StarData{}, err
	}
	if hr <= 0 || hr > len(cachedStarData) {
		return StarData{}, fmt.Errorf("%w: HR %d", ErrStarNotFound, hr)
	}
	return cachedStarData[hr-1], nil
}

// RaDecByJde 按 TT 儒略日归算恒星位置，单位为度 / star position in degrees for a TT Julian date.
// Pc>0 时按三维空间运动推进（自行 + RadVel），否则退回只推进角分量的二维。
// With Pc>0 the position advances in three dimensions (proper motion plus radial velocity); otherwise only the angular components advance.
func (s InnerStarData) RaDecByJde(jde float64) (float64, float64) {
	if s.distanceAU() > 0 {
		return s.raDecByJde3D(jde)
	}
	return s.raDecByJde2D(jde)
}

// RaDecByDate 按给定时刻归算恒星位置，单位为度 / star position in degrees at the given instant.
// 传入时刻按 UTC 解释并换算成 TT 儒略日，与 RaDecByJde 同口径。
// The instant is read as UTC and converted to a TT Julian date, the same convention as RaDecByJde.
func (s StarData) RaDecByDate(date time.Time) (float64, float64) {
	// 星表历元与岁差都按 TT 度量，自行归算的历元差必须用 TT 儒略日。
	jde := UTC2TT(Date2JD(date.UTC()))
	return s.RaDecByJde(jde)
}

func (s InnerStarData) distanceAU() float64 {
	if s.Pc <= 0 {
		return 0
	}
	// 秒差距到天文单位：648000/π 为定义值。
	return s.Pc * 648000 / math.Pi
}

// raDecByJde2D 只推进赤经赤纬两个角分量，等价于把恒星当作无穷远。
func (s InnerStarData) raDecByJde2D(jde float64) (float64, float64) {
	// 自行按儒略年计：J2000 是儒略历元，年长 365.25 日而非回归年。
	// pmRA 是投影自行 cos(Dec)*dRA/dt，不是 dRA/dt。
	year := (jde - 2451545.0) / 365.25
	dec := s.Dec + year*s.PmDec/3600
	cosDec := math.Cos(s.Dec * math.Pi / 180)
	ra := s.Ra
	if math.Abs(cosDec) > 1e-12 {
		ra += year * s.PmRA / (3600 * cosDec)
	}
	return Precess(ra, dec, 2451545.0, jde)
}

// raDecByJde3D 按三维匀速直线运动推进位置矢量，赤经赤纬只是其方向。
func (s InnerStarData) raDecByJde3D(jde float64) (float64, float64) {
	year := (jde - 2451545.0) / 365.25
	raRad := s.Ra * math.Pi / 180
	decRad := s.Dec * math.Pi / 180
	sinDec, cosDec := math.Sin(decRad), math.Cos(decRad)
	distance := s.distanceAU()
	// 切向速度 AU/年 = 角速度(rad/年) × 距离；径向速度 km/s 换算为 AU/年。
	pmRA := s.PmRA * math.Pi / (180 * 3600) * distance
	pmDec := s.PmDec * math.Pi / (180 * 3600) * distance
	radial := s.RadVel * 365.25 * 86400 / 149597870.7
	cosRA, sinRA := math.Cos(raRad), math.Sin(raRad)
	position := [3]float64{
		distance*cosDec*cosRA + year*(pmDec*(-sinDec*cosRA)-pmRA*sinRA+radial*cosDec*cosRA),
		distance*cosDec*sinRA + year*(pmDec*(-sinDec*sinRA)+pmRA*cosRA+radial*cosDec*sinRA),
		distance*sinDec + year*(pmDec*cosDec+radial*sinDec),
	}
	norm := math.Sqrt(position[0]*position[0] + position[1]*position[1] + position[2]*position[2])
	ra := math.Atan2(position[1], position[0]) * 180 / math.Pi
	if ra < 0 {
		ra += 360
	}
	dec := math.Asin(position[2]/norm) * 180 / math.Pi
	return Precess(ra, dec, 2451545.0, jde)
}
