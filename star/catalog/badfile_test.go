package catalog

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestStellariumBadFiles(t *testing.T) {
	data := mustFixture(t)
	bigEndianMagic := make([]byte, 4)
	binary.BigEndian.PutUint32(bigEndianMagic, stellariumMagic)
	cases := map[string][]byte{
		"truncated":   data[:len(data)-7],
		"trailing":    append(append([]byte(nil), data...), 0),
		"bad magic":   append([]byte{1, 2, 3, 4}, data[4:]...),
		"short head":  data[:20],
		"big endian":  append(bigEndianMagic, data[4:]...),
		"bad version": append(append(append([]byte(nil), data[:8]...), 1, 0, 0, 0), data[12:]...),
	}
	for name, input := range cases {
		if _, err := ReadStellariumStars(bytes.NewReader(input), ReadOptions{Strict: true}); err == nil {
			t.Fatalf("%s: strict read should fail", name)
		}
	}
	loose, err := ReadStellariumStars(bytes.NewReader(cases["truncated"]), ReadOptions{})
	if err != nil || loose.Skipped() != 1 || loose.Len() == 0 {
		t.Fatalf("loose truncated read = %d entries, %d skipped, %v", loose.Len(), loose.Skipped(), err)
	}
}

func TestStellariumMaxMagnitude(t *testing.T) {
	data := mustFixture(t)
	dir := fstest.MapFS{"stars_0_0v0_18.cat": &fstest.MapFile{Data: data}}
	all, err := ReadStellariumDir(dir, StellariumOptions{}, ReadOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	bright, err := ReadStellariumDir(dir, StellariumOptions{MaxMagnitude: 0.5}, ReadOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	if bright.Len() == 0 || bright.Len() >= all.Len() {
		t.Fatalf("MaxMagnitude 0.5 kept %d of %d", bright.Len(), all.Len())
	}
	for _, entry := range bright.Entries() {
		if entry.Magnitude > 0.5 {
			t.Fatalf("%s V %.2f exceeds limit", entry.ID, entry.Magnitude)
		}
	}
	// 文件头 mag_min = -2.00，阈值更亮时整片跳过。
	none, err := ReadStellariumDir(dir, StellariumOptions{MaxMagnitude: -3}, ReadOptions{Strict: true})
	if err != nil || none.Len() != 0 {
		t.Fatalf("whole shard should be skipped: %d, %v", none.Len(), err)
	}
}

func TestCrossIDBadLength(t *testing.T) {
	data := mustFixture(t)
	crossID, err := os.ReadFile("testdata/cross-id.cat")
	if err != nil {
		t.Fatal(err)
	}
	dir := fstest.MapFS{
		"stars_0_0v0_18.cat": &fstest.MapFile{Data: data},
		"cross-id.cat":       &fstest.MapFile{Data: crossID[:len(crossID)-3]},
	}
	if _, err := ReadStellariumDir(dir, StellariumOptions{CrossID: true}, ReadOptions{Strict: true}); err == nil {
		t.Fatal("cross-id length not a multiple of 21 should fail in strict mode")
	}
}

func TestCDSBadLines(t *testing.T) {
	if _, err := ReadHipparcos2007(strings.NewReader("  1234 garbage\n"), ReadOptions{Strict: true}); err == nil {
		t.Fatal("short hip2 line should fail in strict mode")
	}
	loose, err := ReadHipparcos2007(strings.NewReader("  1234 garbage\n"), ReadOptions{})
	if err != nil || loose.Len() != 0 || loose.Skipped() != 1 {
		t.Fatalf("loose hip2 = %d entries, %d skipped, %v", loose.Len(), loose.Skipped(), err)
	}
}

// hipParallaxLine 造一条最小可解析的 hip2 定宽记录，parallax 为 43..50 列的原始文本。
func hipParallaxLine(parallax string) string {
	line := []byte(strings.Repeat(" ", 160))
	put := func(value string, from, to int) { copy(line[from:to], value) }
	put("    42", 0, 6)
	put("  1.00000000", 15, 28)
	put(" -0.20000000", 29, 42)
	put(parallax, 43, 50)
	put(" -1.00000", 51, 59)
	put(" -2.00000", 60, 68)
	put(" -1.4600", 129, 136)
	put(" 0.000", 152, 158)
	return string(line)
}

func TestHipparcosNonFiniteParallax(t *testing.T) {
	for _, parallax := range []string{"   +Inf", "   -Inf"} {
		input := strings.NewReader(hipParallaxLine(parallax) + "\n")
		if _, err := ReadHipparcos2007(input, ReadOptions{Strict: true}); err == nil {
			t.Fatalf("parallax %q: strict mode should reject a non-finite parallax", parallax)
		}
		loose, err := ReadHipparcos2007(strings.NewReader(hipParallaxLine(parallax)+"\n"), ReadOptions{})
		if err != nil || loose.Len() != 0 || loose.Skipped() != 1 {
			t.Fatalf("parallax %q: loose = %d entries, %d skipped, %v", parallax, loose.Len(), loose.Skipped(), err)
		}
	}
	// 有限的负视差是“按无视差处理”，不能一起被拒。
	fine, err := ReadHipparcos2007(strings.NewReader(hipParallaxLine("   -1.0")+"\n"), ReadOptions{Strict: true})
	if err != nil || fine.Len() != 1 {
		t.Fatalf("finite negative parallax should load: len=%d err=%v", fine.Len(), err)
	}
	if entry, ok := fine.Lookup("HIP 42"); !ok || entry.Coordinate.ParallaxMas != 0 {
		t.Fatalf("finite negative parallax should become 0: %v %v", entry.Coordinate.ParallaxMas, ok)
	}
}
