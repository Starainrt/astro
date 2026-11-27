//go:build astro_lowmem
// +build astro_lowmem

package basic

// lowmem 下两张记忆表各约 6 KB，未命中只多算一次，结果不变。
const (
	nutationMemoBits = 8
	siderealMemoBits = 8
)
