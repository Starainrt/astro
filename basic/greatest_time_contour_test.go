package basic

import (
	"testing"
	"time"
)

// 对齐网格是导出行为（等时线按整刻度取值）的契约，取值本身写成字面量以免与实际生成函数互相自证。
func TestGreatestTimeContourAlignedLevelsGridContracts(t *testing.T) {
	tt := func(hour, minute int) float64 {
		return UTC2TT(Date2JD(time.Date(2024, 1, 1, hour, minute, 0, 0, time.UTC)))
	}
	cases := []struct {
		name       string
		start, end float64
		step       time.Duration
		maxLevels  int
		want       []string
	}{
		{
			name: "起点在刻度之间时从下一个整刻度开始", start: tt(0, 7), end: tt(2, 7),
			step: 30 * time.Minute, maxLevels: 64,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z", "2024-01-01T02:00:00Z"},
		},
		{
			name: "起点恰在刻度上时不重复生成该刻度", start: tt(0, 30), end: tt(2, 0),
			step: 30 * time.Minute, maxLevels: 64,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z", "2024-01-01T02:00:00Z"},
		},
		{
			name: "终点恰在刻度上时包含终点", start: tt(0, 7), end: tt(1, 30),
			step: 30 * time.Minute, maxLevels: 64,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z"},
		},
		{
			name: "上限截断", start: tt(0, 7), end: tt(5, 7),
			step: 30 * time.Minute, maxLevels: 3,
			want: []string{"2024-01-01T00:30:00Z", "2024-01-01T01:00:00Z", "2024-01-01T01:30:00Z"},
		},
		{
			name: "窗口内没有整刻度时为空", start: tt(0, 1), end: tt(0, 20),
			step: time.Hour, maxLevels: 64,
			want: nil,
		},
		{
			name: "非正步长或倒置窗口为空", start: tt(0, 7), end: tt(2, 7),
			step: 0, maxLevels: 64, want: nil,
		},
	}
	for _, testCase := range cases {
		levels := greatestTimeContourAlignedLevels(testCase.start, testCase.end, testCase.step, testCase.maxLevels)
		if len(levels) != len(testCase.want) {
			t.Fatalf("%s: %d levels, want %d", testCase.name, len(levels), len(testCase.want))
		}
		for index, level := range levels {
			got := level.at.UTC().Format(time.RFC3339)
			if got != testCase.want[index] {
				t.Fatalf("%s: level %d at %s, want %s", testCase.name, index, got, testCase.want[index])
			}
			// 标注时刻与求根时刻必须成对：TT 往返 1 ULP 内一致。
			if roundTrip := greatestTimeContourUTCToTT(level.at); roundTrip != level.tt {
				t.Fatalf("%s: level %d tt %.9f != %.9f", testCase.name, index, level.tt, roundTrip)
			}
		}
	}
}
