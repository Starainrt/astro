package catalog

import (
	"math"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

// propagateToJ2000 按线性空间运动把记录推到 J2000.0，返回赤经赤纬（度）。
func propagateToJ2000(entry Entry) (float64, float64) {
	c := entry.Coordinate
	j2000 := basic.JD2DateByZone(2451545.0, time.UTC, false)
	years := j2000.Sub(c.Epoch).Hours() / 24 / 365.25
	ra := c.RA * math.Pi / 180
	dec := c.Dec * math.Pi / 180
	p := [3]float64{math.Cos(dec) * math.Cos(ra), math.Cos(dec) * math.Sin(ra), math.Sin(dec)}
	east := [3]float64{-math.Sin(ra), math.Cos(ra), 0}
	north := [3]float64{-math.Sin(dec) * math.Cos(ra), -math.Sin(dec) * math.Sin(ra), math.Cos(dec)}
	masToRad := math.Pi / 180 / 3600 / 1000
	var q [3]float64
	for i := range q {
		q[i] = p[i] + (east[i]*c.ProperMotionRACosDecMasPerYear+north[i]*c.ProperMotionDecMasPerYear)*masToRad*years
	}
	return math.Mod(math.Atan2(q[1], q[0])*180/math.Pi+360, 360), math.Atan2(q[2], math.Hypot(q[0], q[1])) * 180 / math.Pi
}

func separationMas(ra1, dec1, ra2, dec2 float64) float64 {
	dra := (ra1 - ra2) * math.Cos(dec1*math.Pi/180)
	if dra > 180 {
		dra -= 360
	}
	if dra < -180 {
		dra += 360
	}
	return math.Hypot(dra, dec1-dec2) * 3600 * 1000
}

func readFixtureCatalog(t *testing.T, name string, read func(*os.File) (*Catalog, error)) *Catalog {
	t.Helper()
	file, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	catalog, err := read(file)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestCrossFormatPositionsAtJ2000(t *testing.T) {
	stellarium := readFixtureCatalog(t, "stars_0_0v0_18.cat", func(f *os.File) (*Catalog, error) {
		return ReadStellariumStars(f, ReadOptions{Strict: true})
	})
	hip := readFixtureCatalog(t, "hip2_sample.dat", func(f *os.File) (*Catalog, error) {
		return ReadHipparcos2007(f, ReadOptions{Strict: true})
	})
	bsc := readFixtureCatalog(t, "bsc5_sample.dat", func(f *os.File) (*Catalog, error) {
		return ReadBSC5(f, ReadOptions{})
	})
	if bsc.Skipped() != 1 {
		t.Fatalf("BSC5 skipped = %d, want the record without coordinates", bsc.Skipped())
	}
	stars := map[string]string{"HIP 32349": "HR 2491", "HIP 91262": "HR 7001", "HIP 24436": "HR 1713", "HIP 21421": "HR 1457", "HIP 11767": "HR 424"}
	for hipID, hrID := range stars {
		h, ok := hip.Lookup(hipID)
		if !ok {
			t.Fatalf("hip2 missing %s", hipID)
		}
		b, ok := bsc.Lookup(hrID)
		if !ok {
			t.Fatalf("BSC5 missing %s", hrID)
		}
		hra, hdec := propagateToJ2000(h)
		bra, bdec := propagateToJ2000(b)
		// BSC5 只给到 0.1 s 与 1″，实测最大约 0.75″。
		if d := separationMas(hra, hdec, bra, bdec); d > 1000 {
			t.Fatalf("%s hip2 vs BSC5 = %.1f mas", hipID, d)
		}
		if s, ok := stellarium.Lookup(hipID); ok {
			sra, sdec := propagateToJ2000(s)
			// 两者同源于 Hipparcos 2007，实测 ≤ 0.5 mas。
			if d := separationMas(hra, hdec, sra, sdec); d > 2 {
				t.Fatalf("%s hip2 vs Stellarium = %.1f mas", hipID, d)
			}
		}
	}
}

func TestCDSFieldConventions(t *testing.T) {
	hip := readFixtureCatalog(t, "hip2_sample.dat", func(f *os.File) (*Catalog, error) {
		return ReadHipparcos2007(f, ReadOptions{Strict: true})
	})
	sirius, _ := hip.Lookup("HIP 32349")
	if math.Abs(sirius.Coordinate.ParallaxMas-379.21) > 1e-6 || math.Abs(sirius.Coordinate.ProperMotionRACosDecMasPerYear+546.01) > 1e-6 || sirius.Band != "Hp" {
		t.Fatalf("hip2 Sirius = %#v", sirius)
	}
	bsc := readFixtureCatalog(t, "bsc5_sample.dat", func(f *os.File) (*Catalog, error) {
		return ReadBSC5(f, ReadOptions{})
	})
	vega, ok := bsc.Lookup("alp lyr")
	if !ok {
		vega, ok = bsc.Lookup("HD 172167")
	}
	if !ok || math.Abs(vega.Coordinate.ParallaxMas-123) > 1e-6 || vega.Coordinate.RadialVelocityKmPerSecond != -14 || vega.Band != "V" {
		t.Fatalf("BSC5 Vega = %#v", vega)
	}
	if math.Abs(vega.Coordinate.ProperMotionRACosDecMasPerYear-202) > 1e-6 {
		t.Fatalf("BSC5 pmRA must already include cos(dec): %.3f", vega.Coordinate.ProperMotionRACosDecMasPerYear)
	}
}

func TestCommonNames(t *testing.T) {
	base := readFixtureCatalog(t, "stars_0_0v0_18.cat", func(f *os.File) (*Catalog, error) {
		return ReadStellariumStars(f, ReadOptions{Strict: true})
	})
	names := "# comment\n 32349|_(\"Sirius\") 2\n91262|_(\"Vega\") 2\n32349|_(\"天狼\") 1\n24436|Rigel\n"
	named, err := base.WithNames(strings.NewReader(names), ReadOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sirius", "天狼", "VEGA", "Rigel"} {
		if _, ok := named.Lookup(name); !ok {
			t.Fatalf("lookup %q failed", name)
		}
	}
	if entry, _ := named.Lookup("Sirius"); entry.ID != "HIP 32349" || entry.Magnitude > -1 {
		t.Fatalf("Sirius = %s V %.2f", entry.ID, entry.Magnitude)
	}
	if _, ok := base.Lookup("Sirius"); ok {
		t.Fatal("WithNames must not modify the receiver")
	}
	if _, err := base.WithNames(strings.NewReader("32349|_(\"broken\n"), ReadOptions{Strict: true}); err == nil {
		t.Fatal("unterminated name should fail in strict mode")
	}
}
