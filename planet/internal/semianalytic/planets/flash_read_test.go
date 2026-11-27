//go:build astro_lowmem
// +build astro_lowmem

package planets

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func sameBitReaderError(got, want error) bool {
	return got == nil && want == nil || got != nil && want != nil && got.Error() == want.Error()
}

func TestFlashBitWindowConsistency(t *testing.T) {
	for length := 0; length <= 24; length++ {
		data := make([]byte, length)
		for index := range data {
			data[index] = byte(index*73 + length*19)
		}
		for position := 0; position <= length*8+9; position++ {
			got := strBits{s: string(data), p: position}
			want := bitReader{b: data, p: position}
			if got.window() != want.window() {
				t.Fatalf("length=%d position=%d: window mismatch", length, position)
			}
		}
	}
}

func TestFlashBitGetConsistency(t *testing.T) {
	for length := 0; length <= 16; length++ {
		data := make([]byte, length)
		for index := range data {
			data[index] = byte(index*59 + length*31)
		}
		for position := 0; position <= length*8+1; position++ {
			for width := 0; width <= 64; width++ {
				got := strBits{s: string(data), p: position}
				want := bitReader{b: data, p: position}
				gotValue, wantValue := got.get(width), want.get(width)
				if gotValue != wantValue || got.p != want.p || !sameBitReaderError(got.err, want.err) {
					t.Fatalf("length=%d position=%d width=%d: got=%d/%d/%v want=%d/%d/%v", length, position, width, gotValue, got.p, got.err, wantValue, want.p, want.err)
				}
			}
		}
	}
}

func TestFlashBitGammaConsistency(t *testing.T) {
	for _, value := range []uint64{0, 1, 2, 3, 31, 127, 65535, 1<<27 - 1, 1<<28 - 1, 1<<29 - 2} {
		encoded := strconv.FormatUint(value+1, 2)
		encoded = strings.Repeat("0", len(encoded)-1) + encoded
		for offset := 0; offset < 8; offset++ {
			stream := strings.Repeat("1", offset) + encoded
			data := make([]byte, (len(stream)+7)/8)
			for index, digit := range stream {
				if digit == '1' {
					data[index/8] |= 1 << uint(7-index%8)
				}
			}
			for length := 0; length <= len(data); length++ {
				got := strBits{s: string(data[:length]), p: offset}
				want := bitReader{b: data[:length], p: offset}
				gotValue, wantValue := got.gamma(), want.gamma()
				if gotValue != wantValue || got.p != want.p || !sameBitReaderError(got.err, want.err) {
					t.Fatalf("value=%d offset=%d length=%d: got=%d/%d/%v want=%d/%d/%v", value, offset, length, gotValue, got.p, got.err, wantValue, want.p, want.err)
				}
				if length == len(data) && gotValue != int(value) {
					t.Fatalf("value=%d offset=%d: got=%d", value, offset, gotValue)
				}
			}
		}
	}
	for _, stream := range []string{"", "\x00", strings.Repeat("\x00", 16)} {
		got := strBits{s: stream}
		want := bitReader{b: []byte(stream)}
		if got.gamma() != want.gamma() || got.p != want.p || !sameBitReaderError(got.err, want.err) {
			t.Fatal("invalid gamma mismatch")
		}
	}
}

func TestFlashBitCoefficientConsistency(t *testing.T) {
	for length := 0; length <= 16; length++ {
		for _, fill := range []byte{0, 0x55, 0x80, 0xff} {
			data := []byte(strings.Repeat(string([]byte{fill}), length))
			for position := 0; position <= length*8+1; position++ {
				for width := 0; width <= 63; width++ {
					got := strBits{s: string(data), p: position}
					want := bitReader{b: data, p: position}
					negative := want.get(1) != 0
					wantValue := float64(want.get(width))
					if negative {
						wantValue = -wantValue
					}
					wantValue *= 0.125
					gotValue := got.coef(width, 0.125)
					if math.Float64bits(gotValue) != math.Float64bits(wantValue) || got.p != want.p || !sameBitReaderError(got.err, want.err) {
						t.Fatalf("length=%d fill=%d position=%d width=%d: got=%g/%d/%v want=%g/%d/%v", length, fill, position, width, gotValue, got.p, got.err, wantValue, want.p, want.err)
					}
				}
			}
		}
	}
}

var flashReadBenchmarkSink [3]float64

func BenchmarkFlashJupiterVaryingTime(b *testing.B) {
	model := FlashDefault()
	if _, err := model.Heliocentric(Jupiter, 2461041.5); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for sample := 0; sample < b.N; sample++ {
		position, err := model.Heliocentric(Jupiter, 2461041.5+float64(sample%210240)*5/1440)
		if err != nil {
			b.Fatal(err)
		}
		flashReadBenchmarkSink = position
	}
}
