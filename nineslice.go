package senjaui

import (
	"image"
	"image/color"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

// CenterFillStrategy specifies how the interior of a nine-slice box is rendered.
type CenterFillStrategy int

const (
	// CenterFillStretch stretches the center segment of the 9-slice image.
	CenterFillStretch CenterFillStrategy = iota
	// CenterFillTransparent keeps the center completely transparent (border-only).
	CenterFillTransparent
	// CenterFillColor fills the center rectangle with a designated solid color.
	CenterFillColor
	// CenterFillTile repeats the center texture in both X and Y dimensions.
	CenterFillTile
)

// NineSliceConfig defines configuration parameters for building an EbitenUI NineSlice.
// It defines edge margins (Top, Right, Bottom, Left), center fill strategy, and *ebiten.Image.
type NineSliceConfig struct {
	Image              *ebiten.Image
	Top                int
	Right              int
	Bottom             int
	Left               int
	CenterFillStrategy CenterFillStrategy
	FillColor          color.Color
	BorderColor        color.Color
}

// NewNineSliceConfig creates a NineSliceConfig with given image and border margins.
func NewNineSliceConfig(img *ebiten.Image, top, right, bottom, left int) NineSliceConfig {
	return NineSliceConfig{
		Image:              img,
		Top:                top,
		Right:              right,
		Bottom:             bottom,
		Left:               left,
		CenterFillStrategy: CenterFillStretch,
	}
}

// NewNineSliceColorConfig creates a simple solid color nine-slice configuration.
func NewNineSliceColorConfig(fill color.Color) NineSliceConfig {
	return NineSliceConfig{
		FillColor:          fill,
		CenterFillStrategy: CenterFillColor,
	}
}

// NewNineSliceBorderedConfig creates a bordered color nine-slice configuration.
func NewNineSliceBorderedConfig(fill color.Color, borderColor color.Color, borderWidth int) NineSliceConfig {
	return NineSliceConfig{
		FillColor:          fill,
		BorderColor:        borderColor,
		Top:                borderWidth,
		Right:              borderWidth,
		Bottom:             borderWidth,
		Left:               borderWidth,
		CenterFillStrategy: CenterFillColor,
	}
}

// NewNineSliceAdvancedConfig creates an advanced colored nine-slice configuration with asymmetric borders.
func NewNineSliceAdvancedConfig(fill color.Color, top, right, bottom, left int, borderColor color.Color) NineSliceConfig {
	return NineSliceConfig{
		FillColor:          fill,
		BorderColor:        borderColor,
		Top:                top,
		Right:              right,
		Bottom:             bottom,
		Left:               left,
		CenterFillStrategy: CenterFillColor,
	}
}

// Build compiles the NineSliceConfig into an *eimage.NineSlice ready for EbitenUI.
func (cfg NineSliceConfig) Build() *eimage.NineSlice {
	if cfg.Image == nil {
		return cfg.buildColorNineSlice()
	}

	return cfg.buildImageNineSlice()
}

// ToNineSlice is an alias for Build() to provide an ergonomic bridge.
func (cfg NineSliceConfig) ToNineSlice() *eimage.NineSlice {
	return cfg.Build()
}

func (cfg NineSliceConfig) buildColorNineSlice() *eimage.NineSlice {
	if cfg.CenterFillStrategy == CenterFillTransparent {
		if cfg.BorderColor != nil && (cfg.Top > 0 || cfg.Right > 0 || cfg.Bottom > 0 || cfg.Left > 0) {
			b := eimage.NewBorder(cfg.Top, cfg.Right, cfg.Bottom, cfg.Left, cfg.BorderColor)
			return eimage.NewAdvancedNineSliceColor(color.Transparent, b)
		}
		return eimage.NewNineSliceColor(color.Transparent)
	}

	if cfg.FillColor != nil {
		if cfg.BorderColor != nil && (cfg.Top > 0 || cfg.Right > 0 || cfg.Bottom > 0 || cfg.Left > 0) {
			b := eimage.NewBorder(cfg.Top, cfg.Right, cfg.Bottom, cfg.Left, cfg.BorderColor)
			return eimage.NewAdvancedNineSliceColor(cfg.FillColor, b)
		}
		return eimage.NewNineSliceColor(cfg.FillColor)
	}

	if cfg.BorderColor != nil && (cfg.Top > 0 || cfg.Right > 0 || cfg.Bottom > 0 || cfg.Left > 0) {
		b := eimage.NewBorder(cfg.Top, cfg.Right, cfg.Bottom, cfg.Left, cfg.BorderColor)
		return eimage.NewAdvancedNineSliceColor(color.Transparent, b)
	}

	return eimage.NewNineSliceColor(color.Transparent)
}

func (cfg NineSliceConfig) buildImageNineSlice() *eimage.NineSlice {
	bounds := cfg.Image.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	left := max(0, cfg.Left)
	right := max(0, cfg.Right)
	centerW := max(0, imgW-left-right)

	top := max(0, cfg.Top)
	bottom := max(0, cfg.Bottom)
	centerH := max(0, imgH-top-bottom)

	srcImg := cfg.Image

	// Handle center fill customization if requested
	if cfg.CenterFillStrategy == CenterFillTransparent || (cfg.CenterFillStrategy == CenterFillColor && cfg.FillColor != nil) {
		modImg := ebiten.NewImage(imgW, imgH)
		modImg.DrawImage(cfg.Image, nil)

		if centerW > 0 && centerH > 0 {
			centerRect := image.Rect(left, top, left+centerW, top+centerH)
			if cfg.CenterFillStrategy == CenterFillTransparent {
				sub := modImg.SubImage(centerRect).(*ebiten.Image)
				sub.Clear()
			} else if cfg.CenterFillStrategy == CenterFillColor && cfg.FillColor != nil {
				sub := modImg.SubImage(centerRect).(*ebiten.Image)
				sub.Fill(cfg.FillColor)
			}
		}
		srcImg = modImg
	}

	if cfg.BorderColor != nil {
		b := eimage.NewBorder(top, right, bottom, left, cfg.BorderColor)
		return eimage.NewAdvancedNineSliceImage(srcImg, b)
	}

	return eimage.NewNineSlice(srcImg, [3]int{left, centerW, right}, [3]int{top, centerH, bottom})
}

// CreateProceduralNineSliceImage creates an in-memory 9-slice image with custom borders and corners.
// This allows applications to use nine-slice graphics without bundling external asset files.
func CreateProceduralNineSliceImage(size int, cornerRadius int, fill, border color.Color, borderWidth int) *ebiten.Image {
	if size < 3 {
		size = 24
	}
	img := ebiten.NewImage(size, size)
	img.Fill(fill)

	// Draw border pixels
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			isBorder := x < borderWidth || x >= size-borderWidth || y < borderWidth || y >= size-borderWidth
			if isBorder {
				img.Set(x, y, border)
			}
		}
	}

	// Bevel / round corners if radius > 0
	if cornerRadius > 0 && cornerRadius < size/2 {
		for r := 0; r < cornerRadius; r++ {
			// Top-Left corner cutout
			for cy := 0; cy < cornerRadius-r; cy++ {
				img.Set(r, cy, color.Transparent)
				img.Set(cy, r, color.Transparent)
				// Top-Right corner cutout
				img.Set(size-1-r, cy, color.Transparent)
				img.Set(size-1-cy, r, color.Transparent)
				// Bottom-Left corner cutout
				img.Set(r, size-1-cy, color.Transparent)
				img.Set(cy, size-1-r, color.Transparent)
				// Bottom-Right corner cutout
				img.Set(size-1-r, size-1-cy, color.Transparent)
				img.Set(size-1-cy, size-1-r, color.Transparent)
			}
		}
	}

	return img
}

