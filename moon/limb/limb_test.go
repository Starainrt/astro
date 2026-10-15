package limb

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"testing"
)

// 内置表的 sha256；换表时须同步更新。
const kaguyaTableSHA256 = "7d49dc75497c404b78b5b259129d0f5ab17b56522128478591ea33c8cb97d072"

func TestEmbeddedTableChecksum(t *testing.T) {
	sum := sha256.Sum256(kaguyaTable)
	if got := hex.EncodeToString(sum[:]); got != kaguyaTableSHA256 {
		t.Fatalf("embedded table sha256 = %s, want %s", got, kaguyaTableSHA256)
	}
}

func TestLimbRadiusMatchesGenerator(t *testing.T) {
	cases := []struct{ lon, lat, angle, radiusKM float64 }{
		{0, 0, 0, 1737.432000000},
		{3.3, -2.7, 123.45, 1735.688760000},
		{-8.6, 7.9, 359.97, 1738.466256000},
		{5.5, 1.25, 270.05, 1738.604000000},
		{-1.0, -6.0, 45.0, 1737.796000000},
	}
	profile := Default()
	for _, c := range cases {
		got, ok := profile.LimbRadiusKM(c.lon, c.lat, c.angle)
		if !ok {
			t.Fatalf("LimbRadiusKM(%v, %v, %v) not ok", c.lon, c.lat, c.angle)
		}
		if math.Abs(got-c.radiusKM) > 1e-6 {
			t.Errorf("LimbRadiusKM(%v, %v, %v) = %.9f, want %.9f", c.lon, c.lat, c.angle, got, c.radiusKM)
		}
	}
}

func TestLimbRadiusWrapsAndClamps(t *testing.T) {
	profile := Default()
	base, _ := profile.LimbRadiusKM(2, 3, 10)
	for _, angle := range []float64{370, -350} {
		if got, _ := profile.LimbRadiusKM(2, 3, angle); math.Abs(got-base) > 1e-9 {
			t.Errorf("angle %v = %.9f, want %.9f", angle, got, base)
		}
	}
	edge, _ := profile.LimbRadiusKM(9, -8, 10)
	if got, _ := profile.LimbRadiusKM(12, -11, 10); got != edge {
		t.Errorf("out-of-range libration = %.9f, want clamped %.9f", got, edge)
	}
	if _, ok := profile.LimbRadiusKM(math.NaN(), 0, 0); ok {
		t.Error("NaN libration must fail")
	}
}

func TestParseRejectsBadInput(t *testing.T) {
	if _, err := Parse([]byte("nope")); err == nil {
		t.Error("short input must fail")
	}
	broken := append([]byte(nil), kaguyaTable[:headerSize+16]...)
	if _, err := Parse(broken); err == nil {
		t.Error("truncated payload must fail")
	}
	huge := append([]byte(nil), kaguyaTable...)
	binary.LittleEndian.PutUint32(huge[4:], 1<<30)
	binary.LittleEndian.PutUint32(huge[8:], 1<<30)
	if _, err := Parse(huge); err == nil {
		t.Error("oversized header must fail before allocating")
	}
}

func TestParseRejectsNonFiniteHeader(t *testing.T) {
	cases := []struct {
		name           string
		angleStep, ref float64
	}{
		{name: "NaN angle step", angleStep: math.NaN(), ref: 1737.4},
		{name: "Inf angle step", angleStep: math.Inf(1), ref: 1737.4},
		{name: "NaN reference radius", angleStep: 180, ref: math.NaN()},
		{name: "zero reference radius", angleStep: 180, ref: 0},
	}
	for _, c := range cases {
		if _, err := Parse(syntheticTable(c.angleStep, c.ref)); err == nil {
			t.Errorf("%s: Parse err = nil, want an error", c.name)
		}
	}
	if _, err := Parse(syntheticTable(180, 1737.4)); err != nil {
		t.Fatalf("valid header: Parse err = %v", err)
	}
}

// syntheticTable 最小合法 2×2×2 LLB1 表，仅表头两项可变。
func syntheticTable(angleStep, referenceRadiusKM float64) []byte {
	var buf bytes.Buffer
	buf.WriteString("LLB1")
	for _, count := range []uint32{2, 2, 2} {
		binary.Write(&buf, binary.LittleEndian, count)
	}
	for _, value := range []float64{0, 1, 0, 1, angleStep, referenceRadiusKM, 1} {
		binary.Write(&buf, binary.LittleEndian, math.Float64bits(value))
	}
	writer := zlib.NewWriter(&buf)
	writer.Write(make([]byte, 2*8))
	writer.Close()
	return buf.Bytes()
}
