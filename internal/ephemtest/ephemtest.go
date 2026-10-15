// Package ephemtest 测试辅助：固定星历模型、按模型选择基线文件。
package ephemtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/starainrt/astro/internal/ephem"
)

// Pin 在本测试内固定模型，结束后恢复 / pins the model for the duration of a test.
func Pin(tb testing.TB, m ephem.Model) {
	prev := ephem.Active()
	ephem.SetActive(m)
	tb.Cleanup(func() { ephem.SetActive(prev) })
}

// PinLegacy 只针对旧级数实现本身的测试使用 / for tests of the legacy series implementation itself.
func PinLegacy(tb testing.TB) { Pin(tb, ephem.Legacy) }

// Golden 当前模型对应的基线文件：旧模型用原路径，新模型用同目录 semianalytic/ 下的同名文件。
func Golden(path string) string {
	if ephem.Active() != ephem.Semianalytic {
		return path
	}
	return filepath.Join(filepath.Dir(path), "semianalytic", filepath.Base(path))
}

// Update 环境变量 ASTRO_UPDATE_GOLDEN=1 时重写当前模型的基线 / whether to rewrite the active model's baselines.
func Update() bool { return os.Getenv("ASTRO_UPDATE_GOLDEN") == "1" }
