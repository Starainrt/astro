package satellite

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

type elementFixture struct {
	Label    string     `json:"label"`
	Source   string     `json:"source"`
	Format   string     `json:"format"`
	Raw      string     `json:"raw"`
	Expected []Elements `json:"expected"`
}

func readElementFixtures(t *testing.T) []elementFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/celestrak_elements.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []elementFixture
	if err = json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func TestElementsAgeSynthetic(t *testing.T) {
	epoch := time.Date(2026, 11, 13, 21, 59, 13, 395264000, time.UTC)
	e := Elements{Epoch: epoch}
	for _, delta := range []time.Duration{-48 * time.Hour, 0, 72*time.Hour + 123*time.Nanosecond} {
		if got := e.Age(epoch.Add(delta).In(time.FixedZone("east", 8*3600))); got != delta {
			t.Fatalf("Age=%v want %v", got, delta)
		}
	}
}
