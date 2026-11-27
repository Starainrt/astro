//go:build !astro_lowmem
// +build !astro_lowmem

package planet

func planetRawTable() [][]float64 { return planetRawData }
func moonCirTable() [][][]float64 { return moonCir }
