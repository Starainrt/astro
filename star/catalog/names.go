package catalog

import "io"

// WithNames 返回并入名称文件后的新星表；每行 `HIP 号或 Gaia 号|名称`，
// 名称可写作 `α_And` 或 `_("Sirius") 2`。原星表不变。
//
// WithNames returns a new catalog with names merged from an `id|name` file.
// Names may be written as `α_And` or `_("Sirius") 2`. The receiver is unchanged.
func (c *Catalog) WithNames(r io.Reader, read ReadOptions) (*Catalog, error) {
	named := newCatalog()
	if c != nil {
		for _, entry := range c.entries {
			named.add(cloneEntry(entry))
		}
		named.skipped = c.skipped
	}
	if err := applyNameLines(r, named, "names", read); err != nil {
		return nil, err
	}
	named.finalize()
	return named, nil
}
