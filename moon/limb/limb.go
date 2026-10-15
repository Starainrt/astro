// Package limb provides the embedded Kaguya LALT lunar limb profile.
package limb

import (
	_ "embed"
	"sync"

	"github.com/starainrt/astro/moon/limbtable"
)

//go:embed kaguya_lalt.bin
var kaguyaTable []byte

// Profile is the decoded Kaguya lunar limb profile.
type Profile = limbtable.Profile

var (
	defaultOnce    sync.Once
	defaultProfile *Profile
)

// Default returns the embedded Kaguya LALT profile, decoded on first use.
func Default() *Profile {
	defaultOnce.Do(func() {
		profile, err := Parse(kaguyaTable)
		if err != nil {
			panic(err)
		}
		defaultProfile = profile
	})
	return defaultProfile
}

// Parse decodes an LLB1 profile table.
func Parse(data []byte) (*Profile, error) { return limbtable.Parse(data) }

const headerSize = limbtable.HeaderSize
