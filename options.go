package senjaui

import (
	"image/color"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// BannerOption represents a functional option for configuring a Banner.
type BannerOption func(*Banner)

// WithID sets the node identifier.
func WithID(id string) BannerOption {
	return func(b *Banner) {
		b.id = id
	}
}

// WithPosition sets the local (X, Y) coordinates of the banner.
func WithPosition(x, y float64) BannerOption {
	return func(b *Banner) {
		b.X = x
		b.Y = y
	}
}

// WithSize sets the width and height of the banner.
func WithSize(w, h float64) BannerOption {
	return func(b *Banner) {
		b.Width = w
		b.Height = h
	}
}

// WithPadding sets inner content margins (top, right, bottom, left).
func WithPadding(top, right, bottom, left float64) BannerOption {
	return func(b *Banner) {
		b.Padding = NewInsets(top, right, bottom, left)
	}
}

// WithPaddingAll sets identical inner content margins on all sides.
func WithPaddingAll(val float64) BannerOption {
	return func(b *Banner) {
		b.Padding = NewInsetsAll(val)
	}
}

// WithPaddingHV sets horizontal (left/right) and vertical (top/bottom) inner margins.
func WithPaddingHV(h, v float64) BannerOption {
	return func(b *Banner) {
		b.Padding = NewInsetsHV(h, v)
	}
}

// WithAnchor sets the alignment point in parent coordinates.
func WithAnchor(anchor Anchor) BannerOption {
	return func(b *Banner) {
		b.Anchor = anchor
	}
}

// WithPivot sets the local origin point of the banner.
func WithPivot(pivot Pivot) BannerOption {
	return func(b *Banner) {
		b.Pivot = pivot
	}
}

// WithNineSliceConfig provides a NineSliceConfig to build the 9-slice background.
func WithNineSliceConfig(cfg NineSliceConfig) BannerOption {
	return func(b *Banner) {
		b.nineSliceConfig = cfg
	}
}

// WithImage configures the banner with an ebiten.Image and explicit edge margins.
func WithImage(img *ebiten.Image, top, right, bottom, left int) BannerOption {
	return func(b *Banner) {
		b.nineSliceConfig = NewNineSliceFromImage(img, top, right, bottom, left)
	}
}

// WithImageAuto configures the banner with an ebiten.Image and automatically detected margins.
func WithImageAuto(img *ebiten.Image) BannerOption {
	return func(b *Banner) {
		b.nineSliceConfig = NewNineSliceAuto(img)
	}
}

// WithImageFile loads an image from file and configures the banner's nine-slice.
func WithImageFile(path string, top, right, bottom, left int) BannerOption {
	return func(b *Banner) {
		img, err := LoadImageFromFile(path)
		if err == nil {
			b.nineSliceConfig = NewNineSliceFromImage(img, top, right, bottom, left)
		}
	}
}

// WithNineSlice directly sets an EbitenUI NineSlice background.
func WithNineSlice(ns *eimage.NineSlice) BannerOption {
	return func(b *Banner) {
		b.nineSlice = ns
	}
}

// WithBackgroundColor creates a solid color 9-slice background.
func WithBackgroundColor(c color.Color) BannerOption {
	return func(b *Banner) {
		b.nineSliceConfig = NewNineSliceColorConfig(c)
	}
}

// WithBorder creates a bordered 9-slice background.
func WithBorder(bodyColor, borderColor color.Color, borderWidth int) BannerOption {
	return func(b *Banner) {
		b.nineSliceConfig = NewNineSliceBorderedConfig(bodyColor, borderColor, borderWidth)
	}
}

// WithAnimation sets transition animation type and duration in seconds.
func WithAnimation(animType AnimType, duration float64) BannerOption {
	return func(b *Banner) {
		b.animType = animType
		b.animDuration = duration
	}
}

// WithEase sets the easing function for animations.
func WithEase(ease EaseFunc) BannerOption {
	return func(b *Banner) {
		b.easeFn = ease
	}
}

// WithAnimationStep sets a custom animation interpolation step hook.
func WithAnimationStep(fn AnimationStepFunc) BannerOption {
	return func(b *Banner) {
		b.animStepFn = fn
		b.animType = AnimCustom
	}
}

// WithOnOpen registers a callback invoked when the banner opens.
func WithOnOpen(fn func()) BannerOption {
	return func(b *Banner) {
		b.onOpen = fn
	}
}

// WithOnClose registers a callback invoked when the banner closes.
func WithOnClose(fn func()) BannerOption {
	return func(b *Banner) {
		b.onClose = fn
	}
}

// WithOnContentClick registers a callback invoked when this node or any child is clicked.
func WithOnContentClick(fn func(childID string)) BannerOption {
	return func(b *Banner) {
		b.onContentClick = fn
	}
}

// WithInitiallyOpen sets the initial open/closed state.
func WithInitiallyOpen(open bool) BannerOption {
	return func(b *Banner) {
		b.isOpen = open
		b.visible = open
		if !open {
			b.scale = 0.0
			b.opacity = 0.0
		}
	}
}

// WithChildren attaches one or more child Banners to this node upon initialization.
func WithChildren(children ...*Banner) BannerOption {
	return func(b *Banner) {
		for _, child := range children {
			b.AddChild(child)
		}
	}
}

// WithWidget places an EbitenUI widget inside this banner container.
func WithWidget(w widget.PreferredSizeLocateableWidget) BannerOption {
	return func(b *Banner) {
		b.contentWidget = w
	}
}

// WithContainerLayout sets a custom Layouter for the banner's container.
func WithContainerLayout(layout widget.Layouter) BannerOption {
	return func(b *Banner) {
		b.customLayouter = layout
	}
}

// WithContainerOpts appends additional EbitenUI ContainerOpt options.
func WithContainerOpts(opts ...widget.ContainerOpt) BannerOption {
	return func(b *Banner) {
		b.customContainerOpts = append(b.customContainerOpts, opts...)
	}
}
