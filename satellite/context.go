package satellite

import "github.com/starainrt/astro/internal/ephem"

// Satellite 指定星历实例的卫星计算模块 / satellite module bound to an ephemeris instance.
type Satellite struct{ inst *ephem.Instance }

// NewWithInstance 创建指定星历实例的模块，nil使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Satellite { return &Satellite{inst: inst} }

var defaultModule Satellite
