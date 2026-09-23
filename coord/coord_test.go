package coord

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/starainrt/astro/basic"
)

func assertClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s mismatch: got %.15f want %.15f", name, got, want)
	}
}

func TestEclipticEquatorialWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 10, 30, 45, 0, time.FixedZone("CST", 8*3600))
	jde := basic.Date2JD(date.UTC())
	lon := 139.686111
	lat := 4.875278

	got := EclipticToEquatorial(date, lon, lat)
	wantRA, wantDec := basic.LoBoToRaDec(jde, lon, lat)
	assertClose(t, "ra", got.RA, wantRA, 1e-12)
	assertClose(t, "dec", got.Dec, wantDec, 1e-12)

	back := EquatorialToEcliptic(date, got.RA, got.Dec)
	assertClose(t, "lon", back.Lon, lon, 1e-10)
	assertClose(t, "lat", back.Lat, lat, 1e-10)
}

func TestTimeAndPrecessionWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	to := time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC)
	jde := basic.Date2JD(date.UTC())

	assertClose(t, "mean sidereal time", MeanSiderealTime(date), basic.MeanSiderealTime(basic.UTC2UT1(jde)), 1e-12)
	assertClose(t, "apparent sidereal time", ApparentSiderealTime(date), basic.ApparentSiderealTime(basic.UTC2UT1(jde)), 1e-12)
	assertClose(t, "obliquity", EclipticObliquity(date, true), basic.EclipticObliquity(jde, true), 1e-12)

	gotLon, gotObl := Nutation2000B(date)
	wantLon, wantObl := basic.Nutation2000B(jde)
	assertClose(t, "nutation longitude", gotLon, wantLon, 1e-12)
	assertClose(t, "nutation obliquity", gotObl, wantObl, 1e-12)

	got := Precess(date, to, 101.28715533, -16.71611586)
	wantRA, wantDec := basic.Precess(101.28715533, -16.71611586, jde, basic.Date2JD(to.UTC()))
	assertClose(t, "precess ra", got.RA, wantRA, 1e-12)
	assertClose(t, "precess dec", got.Dec, wantDec, 1e-12)
}

func TestHorizontalAndTopocentricWrappers(t *testing.T) {
	date := time.Date(2026, 4, 27, 2, 30, 45, 0, time.UTC)
	jde := basic.Date2JD(date.UTC())
	ra := 101.28715533
	dec := -16.71611586
	observerLon := 115.0
	observerLat := 40.0

	hz := EquatorialToHorizontal(date, ra, dec, observerLon, observerLat)
	wantAltitude := basic.StarHeight(jde, ra, dec, observerLon, observerLat, 0)
	assertClose(t, "altitude", hz.Altitude, wantAltitude, 1e-12)
	assertClose(t, "zenith", hz.Zenith, 90-wantAltitude, 1e-12)
	assertClose(t, "azimuth", hz.Azimuth, basic.StarAzimuth(jde, ra, dec, observerLon, observerLat, 0), 1e-12)
	assertClose(t, "hour angle", hz.HourAngle, basic.StarHourAngle(jde, ra, observerLon, 0), 1e-12)
	assertClose(t, "hour angle func", HourAngle(date, ra, observerLon), hz.HourAngle, 1e-12)

	top := TopocentricEquatorial(date, ra, dec, observerLon, observerLat, 0.00257, 53)
	wantRA, wantDec := basic.TopocentricRaDec(ra, dec, observerLat, observerLon, jde, 0.00257, 53)
	assertClose(t, "topocentric ra", top.RA, wantRA, 1e-12)
	assertClose(t, "topocentric dec", top.Dec, wantDec, 1e-12)

	ecl := TopocentricEcliptic(date, 139.686111, 4.875278, observerLon, observerLat, 0.00257, 53)
	assertClose(t, "topocentric lon", ecl.Lon, basic.TopocentricLo(139.686111, 4.875278, observerLat, observerLon, jde, 0.00257, 53), 1e-12)
	assertClose(t, "topocentric lat", ecl.Lat, basic.TopocentricBo(139.686111, 4.875278, observerLat, observerLon, jde, 0.00257, 53), 1e-12)
}

func TestAngularSeparationWrapper(t *testing.T) {
	got := AngularSeparation(101.28715533, -16.71611586, 95.9879578, -52.6956611)
	want := basic.StarAngularSeparation(101.28715533, -16.71611586, 95.9879578, -52.6956611)
	assertClose(t, "angular separation", got, want, 1e-12)
}

// 站心黄道纬度必须留在 [-90,90]，且与"站心赤道坐标再转黄道"这条独立路径一致。
// 旧实现把黄经的分母复用到纬度的 Atan2：黄经落在 90°–270° 时纬度会被切到对顶象限
// （物理 +4.7° 报成约 -176°），且大视差目标（月球）量值也偏。
func TestTopocentricEclipticLatitudeStaysPhysical(t *testing.T) {
	date := time.Date(2026, 1, 15, 4, 0, 0, 0, time.UTC)
	const (
		observerLon = 115.0
		observerLat = 40.0
		height      = 53.0
		distanceAU  = 0.00257
	)
	cases := []struct{ lon, lat float64 }{
		{254.978443, -5.093420}, // 黄经落在 90°–270°，旧实现翻象限
		{109.000000, 4.681960},  // 黄经落在 0°–90°，作对照
	}
	for _, tc := range cases {
		got := TopocentricEcliptic(date, tc.lon, tc.lat, observerLon, observerLat, distanceAU, height)
		if got.Lat < -90 || got.Lat > 90 {
			t.Fatalf("黄经 %.6f: 站心黄纬 %.6f 越出 [-90,90]", tc.lon, got.Lat)
		}
		eq := EclipticToEquatorial(date, tc.lon, tc.lat)
		top := TopocentricEquatorial(date, eq.RA, eq.Dec, observerLon, observerLat, distanceAU, height)
		want := EquatorialToEcliptic(date, top.RA, top.Dec)
		assertClose(t, "topocentric lon", got.Lon, want.Lon, 1e-9)
		assertClose(t, "topocentric lat", got.Lat, want.Lat, 1e-9)
		if math.Abs(got.Lat-tc.lat) > 1.5 {
			t.Fatalf("黄经 %.6f: 站心黄纬 %.6f 与地心黄纬 %.6f 的差超过月球视差量级", tc.lon, got.Lat, tc.lat)
		}
	}
}

// TopocentricEcliptic 对同一时刻只求一次儒略日：参考实现按旧口径重复求值，逐位对照。
func TestTopocentricEclipticMatchesDuplicatedJDE(t *testing.T) {
	type sample struct {
		label      string
		date       time.Time
		lon, lat   float64
		obsLon     float64
		obsLat     float64
		distanceAU float64
		height     float64
	}
	zones := []*time.Location{
		time.UTC,
		time.FixedZone("CST", 8*3600),
		time.FixedZone("EST", -5*3600),
		time.FixedZone("LMT", -7*3600-52*60-58),
	}
	sites := []struct {
		name     string
		lon, lat float64
	}{
		{"shanghai", 121.4737, 31.2304},
		{"sydney", 151.2093, -33.8688},
		{"north-pole", 0, 89.9999},
		{"south-pole", 0, -89.9999},
		{"dateline-west", -179.99, 12},
		{"dateline-east", 179.99, -12},
		{"equator", 0, 0},
	}
	years := []int{-500, 1000, 1582, 2025, 2100, 3000, 4000}
	hours := []int{0, 5, 12, 20, 23}
	cases := make([]sample, 0, len(sites)*len(zones)*len(years)*len(hours)+200)
	for _, site := range sites {
		for _, zone := range zones {
			for _, year := range years {
				for _, hour := range hours {
					cases = append(cases, sample{
						label:      site.name + "/" + zone.String(),
						date:       time.Date(year, 3, 17, hour, 43, 21, 123456789, zone),
						lon:        139.686111,
						lat:        4.875278,
						obsLon:     site.lon,
						obsLat:     site.lat,
						distanceAU: 0.00257,
						height:     53,
					})
				}
			}
		}
	}
	rng := rand.New(rand.NewSource(20260915))
	for i := 0; i < 200; i++ {
		zone := time.FixedZone("random", (rng.Intn(97)-48)*1800)
		cases = append(cases, sample{
			label: "random",
			date: time.Date(
				rng.Intn(8000)-2000, time.Month(1+rng.Intn(12)), 1+rng.Intn(28),
				rng.Intn(24), rng.Intn(60), rng.Intn(60), rng.Intn(1000000000), zone,
			),
			lon:        rng.Float64()*360 - 180,
			lat:        rng.Float64()*180 - 90,
			obsLon:     rng.Float64()*360 - 180,
			obsLat:     rng.Float64()*179.8 - 89.9,
			distanceAU: 0.0001 + rng.Float64()*40,
			height:     rng.Float64() * 5000,
		})
	}

	for _, tc := range cases {
		got := TopocentricEcliptic(tc.date, tc.lon, tc.lat, tc.obsLon, tc.obsLat, tc.distanceAU, tc.height)
		wantLon := basic.TopocentricLo(tc.lon, tc.lat, tc.obsLat, tc.obsLon, jdUTC(tc.date), tc.distanceAU, tc.height)
		wantLat := basic.TopocentricBo(tc.lon, tc.lat, tc.obsLat, tc.obsLon, jdUTC(tc.date), tc.distanceAU, tc.height)
		if got.Lon != wantLon || got.Lat != wantLat {
			t.Fatalf("%s %s: got (%.17g, %.17g) want (%.17g, %.17g)", tc.label, tc.date.Format(time.RFC3339Nano), got.Lon, got.Lat, wantLon, wantLat)
		}
	}
}
