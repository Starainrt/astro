package basic

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
)

// starCatalogMagic star_catalog.dat 为 gzip 压缩、按列存储的 uvarint 整数格式。
const starCatalogMagic = "STRCAT02"

const starCatalogVersion = 2

const starCatalogDetailFields = 6

//go:embed star_catalog.dat
var starCatalogCompressed []byte

func initStarCatalogData() []InnerStarData {
	reader, err := gzip.NewReader(bytes.NewReader(starCatalogCompressed))
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	payload, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	// STRCAT02 的详情列不参与加载，星名以 star_catalog_names.tsv 为准。
	stars, _, hip, err := decodeStarCatalogPayload(payload, false)
	if err != nil {
		panic(err)
	}
	names, err := decodeStarNameTable(starCatalogNamesCompressed, len(stars))
	if err != nil {
		panic(err)
	}
	hr2name = names
	hr2hip = hip
	return stars
}

// starCatalogReader 顺序读取 uvarint 载荷。
type starCatalogReader struct {
	data []byte
	pos  int
	err  error
}

func (r *starCatalogReader) uvarint() uint64 {
	if r.err != nil {
		return 0
	}
	value, n := binary.Uvarint(r.data[r.pos:])
	if n <= 0 {
		r.err = errors.New("truncated star catalog")
		return 0
	}
	r.pos += n
	return value
}

func (r *starCatalogReader) bytes(n uint64) []byte {
	if r.err != nil {
		return nil
	}
	if n > uint64(len(r.data)-r.pos) {
		r.err = errors.New("truncated star catalog")
		return nil
	}
	value := r.data[r.pos : r.pos+int(n)]
	r.pos += int(n)
	return value
}

// 解码公式须与编码端一致，运算顺序不可调换，否则结果不再逐位相同。

// starCatalogScaled 解码符号-幅值定点数，保留 -0。
func starCatalogScaled(v uint64, scale float64) float64 {
	value := float64(v>>1) / scale
	if v&1 != 0 {
		return math.Copysign(value, -1)
	}
	return value
}

func starCatalogRa(v uint64) float64 {
	ra := float64(v / 36000)
	minute := (v / 600) % 60
	sec := float64(v%600) / 10
	ra += float64(minute)/60 + sec/3600
	return ra * 15
}

func starCatalogDec(v uint64) float64 {
	total := v >> 1
	dec := float64(total / 3600)
	dec += float64((total/60)%60)/60 + float64(total%60)/3600
	if v&1 != 0 {
		dec = -dec
	}
	return dec
}

func starCatalogMag(v uint64) float64 {
	if v == 0 {
		return 9999.9999
	}
	return starCatalogScaled(v-1, 100)
}

func starCatalogPc(v uint64) float64 {
	if rc := starCatalogScaled(v, 1000); rc != 0 {
		return 1 / rc
	}
	return 0
}

func starCatalogZigzag(v uint64) int64 {
	return int64(v>>1) ^ -int64(v&1)
}

func decodeStarCatalogPayload(payload []byte, includeDetail bool) ([]InnerStarData, map[uint16][]string, map[uint16]uint32, error) {
	if len(payload) < len(starCatalogMagic)+1 || string(payload[:len(starCatalogMagic)]) != starCatalogMagic {
		return nil, nil, nil, errors.New("invalid star catalog magic")
	}
	if version := payload[len(starCatalogMagic)]; version != starCatalogVersion {
		return nil, nil, nil, fmt.Errorf("unsupported star catalog version %d", version)
	}
	r := &starCatalogReader{data: payload, pos: len(starCatalogMagic) + 1}
	count := r.uvarint()
	if r.err == nil && count > math.MaxUint16 {
		return nil, nil, nil, fmt.Errorf("star catalog count %d out of range", count)
	}
	stars := make([]InnerStarData, count)
	for i := range stars {
		stars[i].HR = uint16(i + 1)
		stars[i].Name = string(r.bytes(r.uvarint()))
	}
	hd := int64(0)
	for i := range stars {
		hd += starCatalogZigzag(r.uvarint())
		stars[i].HD = uint32(hd)
	}
	for i := range stars {
		stars[i].Ra = starCatalogRa(r.uvarint())
	}
	for i := range stars {
		stars[i].Dec = starCatalogDec(r.uvarint())
	}
	for i := range stars {
		stars[i].Mag = starCatalogMag(r.uvarint())
	}
	for i := range stars {
		stars[i].PmRA = starCatalogScaled(r.uvarint(), 1000)
	}
	for i := range stars {
		stars[i].PmDec = starCatalogScaled(r.uvarint(), 1000)
	}
	for i := range stars {
		stars[i].RadVel = starCatalogScaled(r.uvarint(), 1)
	}
	for i := range stars {
		stars[i].RotVel = starCatalogScaled(r.uvarint(), 1)
	}
	for i := range stars {
		stars[i].Pc = starCatalogPc(r.uvarint())
	}

	stringCount := r.uvarint()
	if r.err == nil && stringCount > uint64(len(payload)) {
		return nil, nil, nil, fmt.Errorf("star catalog string count %d out of range", stringCount)
	}
	detail, err := decodeStarCatalogDetail(r, count, stringCount, includeDetail)
	if err != nil {
		return nil, nil, nil, err
	}
	hip := make(map[uint16]uint32)
	prevHIP := int64(0)
	for i := range stars {
		value := r.uvarint()
		if value == 0 {
			continue
		}
		prevHIP += starCatalogZigzag(value - 1)
		hip[uint16(i+1)] = uint32(prevHIP)
	}
	if r.err != nil {
		return nil, nil, nil, r.err
	}
	if r.pos != len(payload) {
		return nil, nil, nil, fmt.Errorf("%d trailing bytes in star catalog", len(payload)-r.pos)
	}
	return stars, detail, hip, nil
}

func decodeStarCatalogDetail(r *starCatalogReader, count, stringCount uint64, includeDetail bool) (map[uint16][]string, error) {
	if !includeDetail {
		for i := uint64(1); i <= stringCount; i++ {
			r.bytes(r.uvarint())
		}
		for k := 0; k < starCatalogDetailFields; k++ {
			for i := 0; i < int(count); i++ {
				r.uvarint()
			}
		}
		return nil, nil
	}
	stringsTable := make([]string, stringCount+1)
	for i := 1; i < len(stringsTable); i++ {
		stringsTable[i] = string(r.bytes(r.uvarint()))
	}
	indexes := make([][starCatalogDetailFields]uint64, count)
	for k := 0; k < starCatalogDetailFields; k++ {
		for i := range indexes {
			indexes[i][k] = r.uvarint()
		}
	}
	detail := make(map[uint16][]string)
	for i, record := range indexes {
		hasValue := false
		values := make([]string, starCatalogDetailFields)
		for k, index := range record {
			if index >= uint64(len(stringsTable)) {
				return nil, fmt.Errorf("detail string index out of range: hr=%d index=%d", i+1, index)
			}
			values[k] = stringsTable[index]
			hasValue = hasValue || index != 0
		}
		if hasValue {
			detail[uint16(i+1)] = values
		}
	}
	return detail, nil
}
