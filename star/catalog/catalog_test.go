package catalog

import (
	"math"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/starainrt/astro/basic"
)

func TestReadCSVAndFilters(t *testing.T) {
	input := "source_id,ra,dec,pmra,pmdec,parallax,radial_velocity,phot_g_mean_mag,bp_rp\n" +
		"1,360,0,1,2,-3,,5,1\n" +
		"2,1,91,1,2,3,0,6,1\n"
	catalog, err := ReadCSV(strings.NewReader(input), GaiaDR3Columns(), ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Len() != 1 || catalog.Skipped() != 1 {
		t.Fatalf("len/skipped = %d/%d", catalog.Len(), catalog.Skipped())
	}
	entry, ok := catalog.Lookup("gaia   dr3 1")
	if !ok || entry.Coordinate.RA != 0 || entry.Coordinate.ParallaxMas != 0 || entry.Coordinate.RadialVelocityKmPerSecond != 0 {
		t.Fatalf("unexpected entry: %#v", entry)
	}
	if !math.IsNaN(entry.BV) {
		t.Fatalf("missing B-V = %v", entry.BV)
	}
	if value, ok := VMagnitude(entry); !ok || math.Abs(value-5.21414) > 1e-4 {
		t.Fatalf("V magnitude = %v, %v", value, ok)
	}
}

func TestReadCSVStreamsQuotedNewlinesAndComments(t *testing.T) {
	input := "\ufeff# exported\n\nid,ra,dec,name\n" +
		"a,10,20,\"two\nlines\"\n" +
		"# note\n// note\n" +
		"b,350,-20,plain\n"
	columns := ColumnMap{ID: "id", RA: "ra", Dec: "dec", Epoch: julianYearTime(2000), Frame: basic.CoordinateFrameICRS}
	catalog, err := ReadCSV(strings.NewReader(input), columns, ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Len() != 2 || catalog.Skipped() != 0 {
		t.Fatalf("len/skipped = %d/%d", catalog.Len(), catalog.Skipped())
	}
	_, err = ReadCSV(strings.NewReader("id,ra,dec\na,1,2\nb,x\"y,3\n"), columns, ReadOptions{Strict: true})
	if err == nil || !strings.Contains(err.Error(), "line 3") {
		t.Fatalf("strict parse error = %v", err)
	}
}

func TestReadCSVReportsRecordStartLine(t *testing.T) {
	columns := ColumnMap{ID: "id", RA: "ra", Dec: "dec", Epoch: julianYearTime(2000), Frame: basic.CoordinateFrameICRS}
	cases := []struct {
		input string
		want  string
	}{
		{"id,ra,dec,name\na,1,2,\"x\ny\nz\"\n,3,4,q\n", "line 5:"},
		{"# c\n\nid,ra,dec,name\r\na,1,2,\"x\r\ny\"\r\n# note\r\n\r\nb,bad,4,q\r\n", "line 8:"},
		{"id,ra,dec,name\n\"m\nn\",1,2,p\n\"k\nj\",bad,3,q\n", "line 4:"},
		{"id,ra,dec,name\na,1,2," + strings.Repeat("w", 200000) + "\nb,1,2,\"s\nt\"\n,1,2,q\n", "line 5:"},
	}
	for _, c := range cases {
		_, err := ReadCSV(strings.NewReader(c.input), columns, ReadOptions{Strict: true})
		if err == nil || !strings.HasPrefix(err.Error(), c.want) {
			t.Errorf("err = %v, want prefix %q", err, c.want)
		}
	}
}

func TestCSVStrictAndCosineConvention(t *testing.T) {
	columns := ColumnMap{ID: "id", RA: "ra", Dec: "dec", PMRA: "pmra", PMDec: "pmdec", Epoch: GaiaDR3Columns().Epoch, Frame: basic.CoordinateFrameICRS}
	withoutCos, err := ReadCSV(strings.NewReader("id,ra,dec,pmra,pmdec\npolar,0,80,10,2\n"), columns, ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	columns.PMRAIncludesCosDec = true
	withCos, err := ReadCSV(strings.NewReader("id,ra,dec,pmra,pmdec\npolar,0,80,1.7364818,2\n"), columns, ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(withoutCos.Entries()[0].Coordinate.ProperMotionRACosDecMasPerYear-withCos.Entries()[0].Coordinate.ProperMotionRACosDecMasPerYear) > 1e-5 {
		t.Fatal("pmRA cosine convention differs")
	}
	_, err = ReadCSV(strings.NewReader("id,ra,dec\nx,0,91\n"), columns, ReadOptions{Strict: true})
	if err == nil || !strings.Contains(err.Error(), "record 2") {
		t.Fatalf("strict error = %v", err)
	}
}

func TestCatalogSearchAndMerge(t *testing.T) {
	base := newCatalog()
	base.add(Entry{ID: "A", Aliases: []string{"Alpha"}, Coordinate: basic.StarCoordinate{ID: "A", RA: 359.9, Dec: 0, Epoch: GaiaDR3Columns().Epoch, Frame: basic.CoordinateFrameICRS}})
	base.finalize()
	other := newCatalog()
	other.add(Entry{ID: "A", Aliases: []string{"Beta"}, Coordinate: base.entries[0].Coordinate})
	other.finalize()
	merged := base.Merge(other)
	if _, ok := merged.Lookup("beta"); !ok {
		t.Fatal("merged alias missing")
	}
	if got := merged.Near(0, 0, 1); len(got) != 1 {
		t.Fatalf("near across zero = %d", len(got))
	}
	if got := merged.Band(-1, 1); len(got) != 1 {
		t.Fatalf("band = %d", len(got))
	}
}

func TestMergeKeepsComponents(t *testing.T) {
	coordinate := basic.StarCoordinate{ID: "HIP 1", RA: 10, Dec: 0, Epoch: GaiaDR3Columns().Epoch, Frame: basic.CoordinateFrameICRS}
	first := newCatalog()
	first.add(Entry{ID: "HIP 1", Component: "A", Magnitude: 1, Coordinate: coordinate})
	first.finalize()
	second := newCatalog()
	second.add(Entry{ID: "HIP 1", Component: "B", Magnitude: 8, Coordinate: coordinate})
	second.finalize()
	merged := first.Merge(second)
	if merged.Len() != 2 {
		t.Fatalf("merged len = %d", merged.Len())
	}
	if entry, ok := merged.Lookup("HIP 1"); !ok || entry.Component != "A" {
		t.Fatalf("lookup component = %q", entry.Component)
	}
}

func TestReadStellariumFixture(t *testing.T) {
	file, err := os.Open("testdata/stars_0_0v0_18.cat")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	catalog, err := ReadStellariumStars(file, ReadOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string][4]float64{
		"HIP 32349": {101.284621, -16.721552, -546.06, -1223.15},
		"HIP 91262": {279.235881, 38.784961, 0, 0},
		"HIP 24436": {78.634473, -8.201636, 0, 0},
	}
	for id, want := range checks {
		entry, ok := catalog.Lookup(id)
		if !ok {
			t.Fatalf("missing %s", id)
		}
		got := entry.Coordinate
		if math.Abs(got.RA-want[0]) > 1e-6 || math.Abs(got.Dec-want[1]) > 1e-6 {
			t.Fatalf("%s position = %.9f %.9f", id, got.RA, got.Dec)
		}
		if want[2] != 0 && math.Abs(got.ProperMotionRACosDecMasPerYear-want[2]) > 0.02 {
			t.Fatalf("%s pmRA = %.6f", id, got.ProperMotionRACosDecMasPerYear)
		}
	}
}

func TestCrossIDLengthAndStellariumDir(t *testing.T) {
	fixture := func(name string) []byte {
		data, err := os.ReadFile("testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	catalog, err := ReadStellariumDir(fstest.MapFS{
		"stars_0_0v0_18.cat": &fstest.MapFile{Data: fixture("stars_0_0v0_18.cat")},
		"name.fab":           &fstest.MapFile{Data: fixture("name.fab")},
		"extra_name.fab":     &fstest.MapFile{Data: fixture("extra_name.fab")},
		"cross-id.cat":       &fstest.MapFile{Data: fixture("cross-id.cat")},
	}, StellariumOptions{Names: true, CrossID: true}, ReadOptions{Strict: true})
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := catalog.Lookup("HIP 32349")
	if !ok || !contains(entry.Aliases, "SAO 151881") || !contains(entry.Aliases, "α CMa") {
		t.Fatalf("aliases missing: %#v", entry.Aliases)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func mustFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/stars_0_0v0_18.cat")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestReadCSVRejectsNonFiniteMotion(t *testing.T) {
	header := "source_id,ra,dec,pmra,pmdec,parallax,radial_velocity,phot_g_mean_mag,bp_rp\n"
	for _, row := range []string{"1,10,0,1,2,3,+Inf,5,1\n", "1,10,0,1,2,3,-Inf,5,1\n", "1,10,0,1,2,+Inf,0,5,1\n", "1,10,0,1,2,3,1e6,5,1\n"} {
		if _, err := ReadCSV(strings.NewReader(header+row), GaiaDR3Columns(), ReadOptions{Strict: true}); err == nil {
			t.Errorf("strict read accepted %q", row)
		}
		catalog, err := ReadCSV(strings.NewReader(header+row), GaiaDR3Columns(), ReadOptions{})
		if err != nil || catalog.Len() != 0 || catalog.Skipped() != 1 {
			t.Errorf("lenient read of %q: len %d skipped %d err %v, want the record skipped", row, catalog.Len(), catalog.Skipped(), err)
		}
	}
}

func TestReadCSVDropsDuplicateIDs(t *testing.T) {
	input := "source_id,ra,dec,pmra,pmdec,parallax,radial_velocity,phot_g_mean_mag,bp_rp\n" +
		"1,10,5,0,0,1,,5,1\n" +
		"1,20,6,0,0,1,,6,1\n" +
		"2,30,7,0,0,1,,7,1\n"
	catalog, err := ReadCSV(strings.NewReader(input), GaiaDR3Columns(), ReadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Len() != 2 || catalog.Skipped() != 1 {
		t.Fatalf("len/skipped = %d/%d", catalog.Len(), catalog.Skipped())
	}
	entry, ok := catalog.Lookup("gaia dr3 1")
	if !ok || entry.Coordinate.RA != 10 {
		t.Fatalf("lookup kept %#v", entry)
	}
	if near := catalog.Near(20, 6, 0.5); len(near) != 0 {
		t.Fatalf("dropped duplicate still reachable through Near: %#v", near)
	}
}

func TestCSVNonFiniteParallax(t *testing.T) {
	columns := ColumnMap{ID: "id", RA: "ra", Dec: "dec", Parallax: "plx",
		Epoch: julianYearTime(2000), Frame: basic.CoordinateFrameICRS}
	for _, parallax := range []string{"+Inf", "-Inf"} {
		input := "id,ra,dec,plx\nx,10,20," + parallax + "\n"
		if _, err := ReadCSV(strings.NewReader(input), columns, ReadOptions{Strict: true}); err == nil {
			t.Fatalf("parallax %q: strict mode should reject a non-finite parallax", parallax)
		}
		loose, err := ReadCSV(strings.NewReader(input), columns, ReadOptions{})
		if err != nil || loose.Len() != 0 || loose.Skipped() != 1 {
			t.Fatalf("parallax %q: loose = %d entries, %d skipped, %v", parallax, loose.Len(), loose.Skipped(), err)
		}
	}
	// 有限的负视差按无视差处理，仍然入库。
	fine, err := ReadCSV(strings.NewReader("id,ra,dec,plx\nx,10,20,-1.5\n"), columns, ReadOptions{Strict: true})
	if err != nil || fine.Len() != 1 {
		t.Fatalf("finite negative parallax should load: len=%d err=%v", fine.Len(), err)
	}
	if entry, ok := fine.Lookup("x"); !ok || entry.Coordinate.ParallaxMas != 0 {
		t.Fatalf("finite negative parallax should become 0: %v %v", entry.Coordinate.ParallaxMas, ok)
	}
}
