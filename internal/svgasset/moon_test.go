package svgasset

import (
	"strings"
	"testing"
)

func TestMoonFaceSymbol(t *testing.T) {
	if symbol := MoonFaceSymbol(); !strings.Contains(symbol, `<symbol id="le-moon"`) {
		t.Fatal("MoonFaceSymbol() does not contain the lunar symbol")
	}
}
