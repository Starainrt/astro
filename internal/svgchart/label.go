package svgchart

// LabelBox 是文本或图形占位的保守包围盒。/ LabelBox is a conservative bounding box for a reserved label or graphic.
type LabelBox struct {
	X, Y, Width, Height float64
}

// LabelPlacement 是一个候选位次：文本锚点与 text-anchor 取值。
// LabelPlacement is one candidate slot: the text anchor point and its text-anchor value.
type LabelPlacement struct {
	X, Y   float64
	Anchor string
}

// LabelTable 记录已占位的矩形，后来的标签按候选顺序避让。
// LabelTable records reserved rectangles; later labels avoid them in candidate order.
type LabelTable struct {
	boxes []LabelBox
}

// labelClearance 是落位判定要求的间隙：包围盒各向外放 1 px，写出的坐标取整后两个盒子也不会贴成相交。
const labelClearance = 1.0

// Reserve 占位一个矩形；非正宽高直接忽略。
// Reserve claims a rectangle; non-positive extents are ignored.
func (table *LabelTable) Reserve(x, y, width, height float64) {
	if width <= 0 || height <= 0 {
		return
	}
	table.boxes = append(table.boxes, LabelBox{X: x, Y: y, Width: width, Height: height})
}

// LabelTextBox 返回文本的保守包围盒，anchor 取 start、middle 或 end。
// LabelTextBox returns a text's conservative box for a start, middle, or end anchor.
func LabelTextBox(x, baseline, fontSize float64, text, anchor string) LabelBox {
	if fontSize <= 0 {
		fontSize = 12
	}
	width := EstimatedTextWidth(text, fontSize)
	left := x
	switch anchor {
	case "middle":
		left = x - width/2
	case "end":
		left = x - width
	}
	above, below := EstimatedTextExtents(fontSize)
	return LabelBox{X: left, Y: baseline - above, Width: width, Height: above + below}
}

// ReserveText 占位一段文本并返回它的包围盒。
// ReserveText claims a text box and returns it.
func (table *LabelTable) ReserveText(x, baseline, fontSize float64, text, anchor string) LabelBox {
	box := LabelTextBox(x, baseline, fontSize, text, anchor)
	table.Reserve(box.X, box.Y, box.Width, box.Height)
	return box
}

// Overlaps 报告矩形是否压到已占位矩形；两个盒子之间至少留 labelClearance。
// Overlaps reports whether a rectangle collides with any reserved box, keeping labelClearance between them.
func (table *LabelTable) Overlaps(x, y, width, height float64) bool {
	for _, box := range table.boxes {
		if box.X+box.Width+labelClearance <= x || x+width+labelClearance <= box.X {
			continue
		}
		if box.Y+box.Height+labelClearance <= y || y+height+labelClearance <= box.Y {
			continue
		}
		return true
	}
	return false
}

// OverlapsText 报告文本是否压到已占位矩形。
// OverlapsText reports whether a text collides with any reserved box.
func (table *LabelTable) OverlapsText(x, baseline, fontSize float64, text, anchor string) bool {
	box := LabelTextBox(x, baseline, fontSize, text, anchor)
	return table.Overlaps(box.X, box.Y, box.Width, box.Height)
}

// Place 按候选顺序放一段文本：返回首个不压叠的位次并占位；全部压叠时返回 false 且不占位。
// Place takes the first candidate that keeps the text clear, claims it, and reports false when none does.
func (table *LabelTable) Place(text string, fontSize float64, candidates []LabelPlacement) (LabelPlacement, bool) {
	for _, candidate := range candidates {
		box := LabelTextBox(candidate.X, candidate.Y, fontSize, text, candidate.Anchor)
		if table.Overlaps(box.X, box.Y, box.Width, box.Height) {
			continue
		}
		table.Reserve(box.X, box.Y, box.Width, box.Height)
		return candidate, true
	}
	return LabelPlacement{}, false
}

// MirrorLabelAnchor 返回镜像后的 text-anchor：start 与 end 互换，middle 不变。
// MirrorLabelAnchor flips start and end, and keeps middle.
func MirrorLabelAnchor(anchor string) string {
	switch anchor {
	case "start":
		return "end"
	case "end":
		return "start"
	default:
		return "middle"
	}
}

// LabelCandidates 生成绕锚点的候选位次：原位、上下左右、四角，再按镜像 anchor 重复一遍。
// LabelCandidates lists candidate slots around an anchor: in place, then the four sides, then the
// four corners, repeated for the mirrored anchor.
func LabelCandidates(x, y float64, anchor string, radius float64) []LabelPlacement {
	candidates := make([]LabelPlacement, 0, 18)
	for _, current := range []string{anchor, MirrorLabelAnchor(anchor)} {
		candidates = append(candidates,
			LabelPlacement{X: x, Y: y, Anchor: current},
			LabelPlacement{X: x, Y: y - radius, Anchor: current},
			LabelPlacement{X: x, Y: y + radius, Anchor: current},
			LabelPlacement{X: x + radius, Y: y, Anchor: current},
			LabelPlacement{X: x - radius, Y: y, Anchor: current},
			LabelPlacement{X: x + radius, Y: y - radius, Anchor: current},
			LabelPlacement{X: x - radius, Y: y - radius, Anchor: current},
			LabelPlacement{X: x + radius, Y: y + radius, Anchor: current},
			LabelPlacement{X: x - radius, Y: y + radius, Anchor: current},
		)
	}
	return candidates
}
