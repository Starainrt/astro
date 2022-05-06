package basic

import (
	"testing"
)

func Test_ParseStar(t *testing.T) {
	//dat := []byte(`2491  9Alp CMaBD-16 1591  48915151881 257I   5423           064044.6-163444064508.9-164258227.22-08.88-1.46   0.00 -0.05 -0.03   A1Vm               -0.553-1.205 +.375-008SBO    13 10.3  11.2AB   4*`)
	LoadStarData()
	for _, v := range stardat {
		_, err := parseStarData(v)
		if err != nil {
			t.Fatal(err)
		}
	}

}
