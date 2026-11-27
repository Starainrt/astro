package satellite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSatelliteTestdataBudgetConsistency(t *testing.T) {
	var total, internal int64
	files := 0
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || !strings.Contains("/"+filepath.ToSlash(path), "/testdata/") {
			return nil
		}
		total += info.Size()
		files++
		if strings.HasPrefix(filepath.ToSlash(path), "internal/") {
			internal += info.Size()
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 || internal == 0 {
		t.Fatal("missing root or internal testdata", files, internal)
	}
	const maximum = 1000000
	if total > maximum {
		t.Fatalf("satellite testdata total=%d bytes (internal=%d), limit=%d", total, internal, maximum)
	}
	t.Logf("%d testdata files, total=%d bytes, internal=%d bytes, limit=%d", files, total, internal, maximum)
}
