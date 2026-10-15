package basic

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"testing"
)

// starCatalogDigestWant 解码后星表的摘要，锁定逐位一致。
const starCatalogDigestWant = "6a538c97a35dfc9fd2ecad2268ec9a15c3efa1bf4f9ee1cff752448c9c5a7a53"

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
		for _, s := range []string{star.ChineseName, star.ChineseAlias, star.ChineseBayerName, star.CommonName,
			star.CommonAliasName, star.Cst, star.CstChinese} {
			writeString(s)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func TestStarCatalogDigestMatchesLegacyParser(t *testing.T) {
	if err := LoadStarData(); err != nil {
		t.Fatal(err)
	}
	if got := starCatalogDigest(cachedStarData); got != starCatalogDigestWant {
		t.Fatalf("star catalog digest = %s, want %s", got, starCatalogDigestWant)
	}
}
