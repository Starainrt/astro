package basic

import ()

// MoonHorizonWithInstance 用指定星历实例求值；inst 为 nil 时用进程当前模型 / evaluation with an explicit instance; nil uses the process model.
func MoonHorizonWithInstance(jdUTC float64, samples int, inst *Instance) [][2]float64 {
	if !finite(jdUTC) {
		return nil
	}
	return MoonStateAtWithInstance(jdUTC, inst).MoonHorizon(samples)
}
