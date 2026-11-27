//go:build !astro_lowmem
// +build !astro_lowmem

// 被测契约：legacy.bin 与源码字面量逐位一致；重新生成需设环境变量（不注册 flag，免得改动基准二进制的布局）。
package planet

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"testing"
)

//go:generate env ASTRO_REGENERATE_LEGACY_TABLE=1 go test -run=TestLegacyTableBlobMatchesLiterals -count=1
const legacyGenerateEnv = "ASTRO_REGENERATE_LEGACY_TABLE"

func encodeLegacyTables(raw [][]float64, cir [][][]float64) []byte {
	var out bytes.Buffer
	out.WriteString("LGC1")
	var tmp [binary.MaxVarintLen64]byte
	uvar := func(n int) { out.Write(tmp[:binary.PutUvarint(tmp[:], uint64(n))]) }
	floats := func(v []float64) {
		uvar(len(v))
		for _, f := range v {
			binary.LittleEndian.PutUint64(tmp[:8], math.Float64bits(f))
			out.Write(tmp[:8])
		}
	}
	uvar(len(raw))
	for _, v := range raw {
		floats(v)
	}
	uvar(len(cir))
	for _, c := range cir {
		uvar(len(c))
		for _, v := range c {
			floats(v)
		}
	}
	return out.Bytes()
}

func TestLegacyTableBlobMatchesLiterals(t *testing.T) {
	want := encodeLegacyTables(planetRawData, moonCir)
	if os.Getenv(legacyGenerateEnv) == "1" {
		if err := os.WriteFile("legacy.bin", want, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile("legacy.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("legacy.bin is stale: %d bytes vs %d; regenerate with %s=1 go test -run=%s",
			len(got), len(want), legacyGenerateEnv, t.Name())
	}
	raw, cir, err := decodeLegacyTables(string(got))
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != len(planetRawData) || len(cir) != len(moonCir) {
		t.Fatalf("decoded shape mismatch")
	}
	for i := range raw {
		assertSameFloats(t, raw[i], planetRawData[i])
	}
	for i := range cir {
		if len(cir[i]) != len(moonCir[i]) {
			t.Fatalf("moon coord %d: %d series vs %d", i, len(cir[i]), len(moonCir[i]))
		}
		for j := range cir[i] {
			assertSameFloats(t, cir[i][j], moonCir[i][j])
		}
	}
}

func assertSameFloats(t *testing.T, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length %d vs %d", len(got), len(want))
	}
	for k := range got {
		if math.Float64bits(got[k]) != math.Float64bits(want[k]) {
			t.Fatalf("index %d: %v vs %v", k, got[k], want[k])
		}
	}
}

func TestDecodeLegacyTablesRejectsBadInput(t *testing.T) {
	blob := string(encodeLegacyTables(planetRawData, moonCir))
	for _, s := range []string{"", "LGC0", "LGC1", blob[:len(blob)/2], blob[:len(blob)-1], blob + "x"} {
		if _, _, err := decodeLegacyTables(s); err == nil {
			t.Errorf("accepted bad blob of %d bytes", len(s))
		}
	}
}
