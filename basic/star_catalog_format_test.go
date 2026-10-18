package basic

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

// starCatalogDigestWant 解码后星表的摘要，锁定逐位一致。
const starCatalogDigestWant = "3c3455cdb09e1e46c5d513a6c26743c6fd1753f65754888f97ec585324af39e4"

func starCatalogDigest(stars []StarData) string {
	hash := sha256.New()
	var buf [8]byte
	writeUint := func(v uint64) {
		binary.LittleEndian.PutUint64(buf[:], v)
		hash.Write(buf[:])
	}
	writeString := func(s string) {
		writeUint(uint64(len(s)))
		hash.Write([]byte(s))
	}
	writeUint(uint64(len(stars)))
	for _, star := range stars {
		writeUint(uint64(star.HR))
		writeUint(uint64(star.HD))
		writeUint(uint64(star.HIP))
		writeString(star.Name)
		for _, v := range []float64{star.Ra, star.Dec, star.Mag, star.PmRA, star.PmDec, star.RadVel, star.RotVel, star.Pc} {
			writeUint(math.Float64bits(v))
		}
		for _, s := range []string{star.ChineseName, star.ChineseAlias, star.ChineseRegion, star.ChineseBayerName,
			star.CommonName, star.CommonAliasName, star.Cst, star.CstChinese} {
			writeString(s)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func TestStarCatalogDigestStable(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	if got := starCatalogDigest(cachedStarData); got != starCatalogDigestWant {
		t.Fatalf("star catalog digest = %s, want %s", got, starCatalogDigestWant)
	}
}

func TestStarCatalogNamesGzipMatchesSource(t *testing.T) {
	source, err := os.ReadFile("star_catalog_names.tsv")
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(starCatalogNamesCompressed))
	if err != nil {
		t.Fatal(err)
	}
	decompressed, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, decompressed) {
		t.Fatal("star_catalog_names.tsv.gz is stale; regenerate it from star_catalog_names.tsv")
	}
}

// starNameLegacyCorrections 旧 .dat 详情与单一表的已知差异；新增名称不在其中。
var starNameLegacyCorrections = map[uint16]map[string]bool{
	390:  {"bayer": true, "chinese": true, "cst": true, "cstzh": true},
	565:  {"chinese": true},
	585:  {"chinese": true},
	1140: {"bayer": true, "chinese": true, "cst": true, "cstzh": true},
	1151: {"bayer": true, "chinese": true},
	1180: {"bayer": true, "chinese": true},
	1324: {"chinese": true},
	1387: {"chinese": true},
	1412: {"chinese": true},
	1656: {"chinese": true},
	2585: {"bayer": true, "cst": true, "cstzh": true},
	2590: {"chinese": true},
	2714: {"chinese": true},
	2878: {"chinese": true},
	3314: {"chinese": true},
	5226: {"chinese": true},
	5396: {"chinese": true},
	5533: {"bayer": true, "chinese": true, "cst": true, "cstzh": true},
	6008: {"bayer": true, "chinese": true, "cst": true, "cstzh": true},
	6554: {"chinese": true},
	6636: {"chinese": true},
	7735: {"chinese": true},
	8313: {"chinese": true},
	8556: {"chinese": true},
	8610: {"bayer": true, "chinese": true, "cst": true, "cstzh": true},
	8982: {"chinese": true},
}

func TestStarCatalogNameTableMatchesLegacyDetails(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	for hr, rec := range hr2name {
		for field, value := range map[string]string{
			"chinese": rec.chinese,
			"aliases": strings.Join(rec.aliases, ";"),
			"region":  rec.region,
		} {
			if strings.ContainsAny(value, "[]") {
				t.Fatalf("HR %d %s has bracket qualifier %q; use the region column", hr, field, value)
			}
		}
	}
	reader, err := gzip.NewReader(bytes.NewReader(starCatalogCompressed))
	if err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, detail, _, err := decodeStarCatalogPayload(payload, true)
	if err != nil {
		t.Fatal(err)
	}
	for hr, info := range detail {
		rec, ok := hr2name[hr]
		if !ok {
			t.Fatalf("HR %d has legacy details but no row in star_catalog_names.tsv", hr)
		}
		checks := []struct {
			field string
			old   string
			new   string
		}{
			{"chinese", info[0], rec.chinese},
			{"aliases", info[1], strings.Join(rec.aliases, ";")},
			{"bayer", info[2], rec.bayerZH},
			{"common", info[3], rec.common},
			{"cstzh", info[4], rec.cstZH},
			{"cst", info[5], rec.cst},
		}
		for _, check := range checks {
			if check.old == "" || check.old == check.new {
				continue
			}
			if starNameLegacyCorrections[hr][check.field] {
				continue
			}
			t.Fatalf("HR %d %s: legacy %q, table %q", hr, check.field, check.old, check.new)
		}
	}
}
