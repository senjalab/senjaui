package senjaui

// Insets represents edge insets or margins (Top, Right, Bottom, Left).
type Insets struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// NewInsets constructs an Insets with specific edge values.
func NewInsets(top, right, bottom, left float64) Insets {
	return Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

// NewInsetsAll constructs an Insets with equal margins on all four sides.
func NewInsetsAll(val float64) Insets {
	return Insets{
		Top:    val,
		Right:  val,
		Bottom: val,
		Left:   val,
	}
}

// NewInsetsHV constructs an Insets with horizontal (left/right) and vertical (top/bottom) values.
func NewInsetsHV(h, v float64) Insets {
	return Insets{
		Top:    v,
		Right:  h,
		Bottom: v,
		Left:   h,
	}
}

// Horizontal returns the combined left and right insets.
func (i Insets) Horizontal() float64 {
	return i.Left + i.Right
}

// Vertical returns the combined top and bottom insets.
func (i Insets) Vertical() float64 {
	return i.Top + i.Bottom
}

// Anchor specifies the alignment point of a child within its parent's content rectangle.
// X and Y are normalized in [0.0, 1.0], where (0, 0) is top-left and (1, 1) is bottom-right.
type Anchor struct {
	X float64
	Y float64
}

// Pivot specifies the reference origin point within the banner itself.
// X and Y are normalized in [0.0, 1.0], where (0, 0) is top-left, (0.5, 0.5) is center, and (1, 1) is bottom-right.
type Pivot struct {
	X float64
	Y float64
}

// Standard Anchor Presets.
var (
	AnchorTopLeft      = Anchor{X: 0.0, Y: 0.0}
	AnchorTopCenter    = Anchor{X: 0.5, Y: 0.0}
	AnchorTopRight     = Anchor{X: 1.0, Y: 0.0}
	AnchorMiddleLeft   = Anchor{X: 0.0, Y: 0.5}
	AnchorCenter       = Anchor{X: 0.5, Y: 0.5}
	AnchorMiddleRight  = Anchor{X: 1.0, Y: 0.5}
	AnchorBottomLeft   = Anchor{X: 0.0, Y: 1.0}
	AnchorBottomCenter = Anchor{X: 0.5, Y: 1.0}
	AnchorBottomRight  = Anchor{X: 1.0, Y: 1.0}
)

// Standard Pivot Presets.
var (
	PivotTopLeft      = Pivot{X: 0.0, Y: 0.0}
	PivotTopCenter    = Pivot{X: 0.5, Y: 0.0}
	PivotTopRight     = Pivot{X: 1.0, Y: 0.0}
	PivotMiddleLeft   = Pivot{X: 0.0, Y: 0.5}
	PivotCenter       = Pivot{X: 0.5, Y: 0.5}
	PivotMiddleRight  = Pivot{X: 1.0, Y: 0.5}
	PivotBottomLeft   = Pivot{X: 0.0, Y: 1.0}
	PivotBottomCenter = Pivot{X: 0.5, Y: 1.0}
	PivotBottomRight  = Pivot{X: 1.0, Y: 1.0}
)

