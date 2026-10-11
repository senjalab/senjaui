package senjaui

import (
	"image"
	"image/color"

	"github.com/ebitenui/ebitenui"
	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type rootScreenLayouter struct {
	root *Banner
}

func (l *rootScreenLayouter) PreferredSize(widgets []widget.PreferredSizeLocateableWidget) (int, int) {
	if l.root != nil {
		return int(l.root.Width), int(l.root.Height)
	}
	return 0, 0
}

func (l *rootScreenLayouter) Layout(widgets []widget.PreferredSizeLocateableWidget, rect image.Rectangle) {
	if l.root != nil {
		l.root.recomputeLayout()
	}
}

// NewUI creates an *ebitenui.UI wrapping the given root Banner in a full-screen canvas container.
// It sets up automatic layout recalculation and connects EbitenUI input layers.
func NewUI(root *Banner) *ebitenui.UI {
	canvas := widget.NewContainer(
		widget.ContainerOpts.Layout(&rootScreenLayouter{root: root}),
		widget.ContainerOpts.BackgroundImage(eimage.NewNineSliceColor(color.Transparent)),
	)

	if root != nil && root.container != nil {
		canvas.AddChild(root.container)
	}

	return &ebitenui.UI{
		Container: canvas,
	}
}

// CreateCard is a convenience constructor for creating nested, styled UI cards with 9-slice borders.
func CreateCard(id string, w, h float64, fill, border color.Color, borderWidth int, opts ...BannerOption) *Banner {
	allOpts := []BannerOption{
		WithID(id),
		WithSize(w, h),
		WithBorder(fill, border, borderWidth),
	}
	allOpts = append(allOpts, opts...)
	return NewBanner(allOpts...)
}

// CreateStyledButton constructs an EbitenUI button with custom NineSlice states and text.
func CreateStyledButton(label string, face *text.Face, idleColor, hoverColor, pressedColor color.Color, textColor color.Color, onClick func()) *widget.Button {
	idleNS := eimage.NewNineSliceColor(idleColor)
	hoverNS := eimage.NewNineSliceColor(hoverColor)
	pressedNS := eimage.NewNineSliceColor(pressedColor)

	buttonImage := &widget.ButtonImage{
		Idle:    idleNS,
		Hover:   hoverNS,
		Pressed: pressedNS,
	}

	btn := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text(label, face, &widget.ButtonTextColor{
			Idle:    textColor,
			Hover:   textColor,
			Pressed: textColor,
		}),
		widget.ButtonOpts.TextPadding(widget.NewInsetsSimple(8)),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			if onClick != nil {
				onClick()
			}
		}),
	)

	return btn
}

// CreateLabel constructs an EbitenUI text widget with centered text.
func CreateLabel(content string, face *text.Face, col color.Color) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(content, face, col),
		widget.TextOpts.Position(widget.TextPositionCenter, widget.TextPositionCenter),
	)
}

// CreateLeftLabel constructs an EbitenUI text widget aligned to start (left) to keep text within parent bounds.
func CreateLeftLabel(content string, face *text.Face, col color.Color) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(content, face, col),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
	)
}

// CreateLeftLabelWithMaxWidth constructs an EbitenUI text widget aligned to start with a maximum width constraint.
func CreateLeftLabelWithMaxWidth(content string, face *text.Face, col color.Color, maxWidth float64) *widget.Text {
	return widget.NewText(
		widget.TextOpts.Text(content, face, col),
		widget.TextOpts.Position(widget.TextPositionStart, widget.TextPositionCenter),
		widget.TextOpts.MaxWidth(maxWidth),
	)
}


