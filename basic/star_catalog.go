package basic

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// this file contains bright 9100 stars
// 9100颗亮星列表

type InnerStarData struct {
	HR     uint16  //Bright Star Number;[1/9110]+ Harvard Revised Number;亮星编号
	Name   string  //Name, generally Bayer(如天狼星：Alpha CMA) and/or Flamsteed(如天狼星：9 CMA) name
	HD     uint32  //Henry Draper Catalog Number;HD星表编号
	Ra     float64 //Ra J2000;J2000历元赤经
	Dec    float64 //De J2000;J2000历元赤纬
	Mag    float64 //视星等
	PmRA   float64 //赤经年自行
	PmDec  float64 //赤纬年自行
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

func parseStarData(star []byte) (InnerStarData, error) {
	var err error
	var stardata InnerStarData
	if len(star) < 160 {
		return stardata, errors.New("invalid stardat")
	}
	for i := 0; i < 4; i++ {
		if star[i] == ' ' {
			continue
		}
		stardata.HR = stardata.HR*10 + uint16(star[i]-48)
	}
	stardata.Name = string(bytes.TrimSpace(star[4:14]))
	for i := 25; i < 31; i++ {
		if star[i] == ' ' {
			continue
		}
		stardata.HD = stardata.HD*10 + uint32(star[i]-48)
	}
	stardata.Ra, err = parseRa(star)
	if err != nil {
		return stardata, fmt.Errorf("parse ra failed:%v", err)
	}
	stardata.Dec, err = parseDec(star)
	if err != nil {
		return stardata, fmt.Errorf("parse dec failed:%v", err)
	}
	magOri := string(bytes.TrimSpace(star[102:107]))
	if magOri != "" {
		stardata.Mag, err = strconv.ParseFloat(magOri, 64)
		if err != nil {
			return stardata, fmt.Errorf("parse mag failed:%v", err)
		}
	} else {
		stardata.Mag = 9999.9999
	}
	stardata.PmRA, _ = strconv.ParseFloat(string(bytes.TrimSpace(star[148:154])), 64)
	stardata.PmDec, _ = strconv.ParseFloat(string(bytes.TrimSpace(star[154:160])), 64)
	if len(star) >= 170 {
		stardata.RadVel, _ = strconv.ParseFloat(string(bytes.TrimSpace(star[166:170])), 64)
	}
	if len(star) >= 179 {
		stardata.RotVel, _ = strconv.ParseFloat(string(bytes.TrimSpace(star[176:179])), 64)
	}
	if len(star) > 161 {
		rc, _ := strconv.ParseFloat(string(bytes.TrimSpace(star[162:166])), 64)
		if rc != 0 {
			stardata.Pc = 1 / rc
		}
	}
	return stardata, nil
}

func parseRa(star []byte) (float64, error) {
	var sec float64
	var err error
	ra := float64(0)
	for i := 75; i < 77; i++ {
		if star[i] == ' ' {
			continue
		}
		ra = ra*10 + float64(star[i]-48)
	}
	minute := uint8(0)
	for i := 77; i < 79; i++ {
		if star[i] == ' ' {
			continue
		}
		minute = minute*10 + (star[i] - 48)
	}
	ori := string(bytes.TrimSpace(star[79:83]))
	if ori != "" {
		sec, err = strconv.ParseFloat(ori, 64)
		if err != nil {
			return ra, err
		}
	}
	ra += float64(minute)/60 + sec/3600
	return ra * 15, nil
}

func parseDec(star []byte) (float64, error) {
	var sec float64
	var err error
	underZero := false
	if star[83] == '-' {
		underZero = true
	}
	dec := float64(0)
	for i := 84; i < 86; i++ {
		if star[i] == ' ' {
			continue
		}
		dec = dec*10 + float64(star[i]-48)
	}
	minute := uint8(0)
	for i := 86; i < 88; i++ {
		if star[i] == ' ' {
			continue
		}
		minute = minute*10 + (star[i] - 48)
	}
	ori := string(bytes.TrimSpace(star[88:90]))
	if ori != "" {
		sec, err = strconv.ParseFloat(ori, 64)
		if err != nil {
			return dec, err
		}
	}
	dec += float64(minute)/60 + sec/3600
	if underZero {
		dec = -dec
	}
	return dec, nil
}

var stardat [][]byte
var starhdindex map[string]int
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
	data := initStarCatalogData()
	stardat = bytes.Split(data, []byte("\n"))
	parsedStarData = make([]InnerStarData, len(stardat))
	cachedStarData = make([]StarData, len(stardat))
	for i, row := range stardat {
		parsed, err := parseStarData(row)
		if err != nil {
			return fmt.Errorf("parse star %d failed: %w", i+1, err)
		}
		parsedStarData[i] = parsed
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

func (s InnerStarData) RaDecByJde(jde float64) (float64, float64) {
	//计算自行
	year := ((jde - 2451545.0) / 365.2422)
	return Precess(s.Ra+(year*s.PmRA/3600), s.Dec+(year*s.PmDec/3600), 2451545.0, jde)
}

func (s StarData) RaDecByDate(date time.Time) (float64, float64) {
	jde := Date2JDE(date.UTC())
	return s.RaDecByJde(jde)
}
