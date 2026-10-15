package basic

import (
	"errors"
	"math"
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
}

var hr2detail map[uint16][]string
var hr2hip map[uint16]uint32
var chnidx map[string]uint16
var parsedStarData []InnerStarData
var cachedStarData []StarData
var starDataOnce sync.Once
var starDataErr error

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
	chnidx = make(map[string]uint16, len(hr2detail)*2)
	for hr := 1; hr <= len(cachedStarData); hr++ {
		info, ok := hr2detail[uint16(hr)]
		if !ok {
			continue
		}
		registerStarChineseIndex(uint16(hr), info[0])
		registerStarChineseIndex(uint16(hr), info[1])
	}
	return nil
}

func registerStarChineseIndex(hr uint16, name string) {
	if name == "" {
		return
	}
	chnidx[name] = hr
	if strings.HasSuffix(name, "星") {
		trimmed := strings.TrimSuffix(name, "星")
		if trimmed != "" {
			chnidx[trimmed] = hr
		}
	}
}

func fullStarData(star InnerStarData) StarData {
	star.HIP = hr2hip[star.HR]
	if info, ok := hr2detail[star.HR]; ok {
		return StarData{
			InnerStarData:    star,
			ChineseName:      info[0],
			ChineseAlias:     info[1],
			ChineseBayerName: info[2],
			CommonName:       info[3],
			CommonAliasName:  "",
			Cst:              info[5],
			CstChinese:       info[4],
		}
	}

	return StarData{InnerStarData: star}
}

func StarDataByChinese(name string) (StarData, error) {
	if err := LoadStarData(); err != nil {
		return StarData{}, err
	}
	if strings.HasSuffix(name, "星") {
		name = strings.TrimSuffix(name, "星")
	}
	hr, ok := chnidx[name]
	if !ok || hr == 0 || int(hr) > len(cachedStarData) {
		return StarData{}, errors.New("no such star")
	}
	return cachedStarData[hr-1], nil
}

func StarDataByHR(hr int) (StarData, error) {
	if err := LoadStarData(); err != nil {
		return StarData{}, err
	}
	if hr <= 0 || hr > len(cachedStarData) {
		return StarData{}, errors.New("no such star")
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
