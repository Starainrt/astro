package uranus

import (
	"github.com/starainrt/astro/internal/ephem"
)

// Uranus 指定星历实例的计算模块 / computation module bound to an ephemeris instance.
type Uranus struct{ inst *ephem.Instance }

// NewWithInstance 创建指定星历实例的模块，nil 使用全局模型 / create a module; nil uses the global model.
func NewWithInstance(inst *ephem.Instance) *Uranus { return &Uranus{inst: inst} }

var defaultModule Uranus
