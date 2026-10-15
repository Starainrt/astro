package basic

import "testing"

func TestGetMoonLoopsNonPositive(t *testing.T) {
	for _, loop := range []int{0, -1} {
		if got := GetMoonLoops(2000, loop); len(got) != 0 {
			t.Fatalf("GetMoonLoops(2000, %d) = %v, want empty", loop, got)
		}
	}
	if got := GetMoonLoops(2000, 15); len(got) != 15 {
		t.Fatalf("GetMoonLoops(2000, 15) len = %d, want 15", len(got))
	}
}
