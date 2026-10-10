package senjaui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"

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

// LoadImageFromFile decodes an image file from disk and returns an *ebiten.Image.
func LoadImageFromFile(path string) (*ebiten.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("senjaui: failed to read image file %q: %w", path, err)
	}
	return LoadImageFromBytes(data)
}

// MustLoadImageFromFile loads an image from disk and panics if an error occurs.
func MustLoadImageFromFile(path string) *ebiten.Image {
	img, err := LoadImageFromFile(path)
	if err != nil {
		panic(err)
	}
	return img
}

// LoadImageFromBytes decodes an image from byte slice and returns an *ebiten.Image.
func LoadImageFromBytes(data []byte) (*ebiten.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("senjaui: failed to decode image: %w", err)
	}
	return ebiten.NewImageFromImage(img), nil
}

// DetectNineSliceMargins dynamically analyzes a raster image to detect optimal 9-slice edge margins (Top, Right, Bottom, Left).
// It identifies transparent padding and the boundary where decorative borders transition into the interior center fill.
// LoadNineSliceImageFile loads an image file from disk, automatically analyzes its margins,
// and returns both the *ebiten.Image and a configured NineSliceConfig.
func LoadNineSliceImageFile(path string) (*ebiten.Image, NineSliceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NineSliceConfig{}, fmt.Errorf("senjaui: failed to read %q: %w", path, err)
	}
	return LoadNineSliceImageBytes(data)
}

// MustLoadNineSliceImageFile loads an image file and returns an *ebiten.Image and NineSliceConfig, panicking on error.
func MustLoadNineSliceImageFile(path string) (*ebiten.Image, NineSliceConfig) {
	ebImg, cfg, err := LoadNineSliceImageFile(path)
	if err != nil {
		panic(err)
	}
	return ebImg, cfg
}

// LoadNineSliceImageBytes decodes image bytes and returns the *ebiten.Image and optimal NineSliceConfig.
func LoadNineSliceImageBytes(data []byte) (*ebiten.Image, NineSliceConfig, error) {
	stdImg, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, NineSliceConfig{}, fmt.Errorf("senjaui: failed to decode image: %w", err)
	}

	top, right, bottom, left := DetectNineSliceMargins(stdImg)
	ebImg := ebiten.NewImageFromImage(stdImg)
	cfg := NewNineSliceConfig(ebImg, top, right, bottom, left)
	return ebImg, cfg, nil
}

// DetectNineSliceMargins dynamically analyzes a raster image to detect optimal 9-slice edge margins (Top, Right, Bottom, Left).
// It identifies transparent padding and the boundary where decorative borders transition into the interior center fill.
func DetectNineSliceMargins(img image.Image) (top, right, bottom, left int) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return 0, 0, 0, 0
	}

	// Safe fallback in case img is an uninitialized *ebiten.Image outside game loop
	fallbackTop := max(1, h/4)
	fallbackBottom := fallbackTop
	fallbackLeft := max(1, w/4)
	fallbackRight := fallbackLeft

	defer func() {
		if r := recover(); r != nil {
			top, right, bottom, left = fallbackTop, fallbackRight, fallbackBottom, fallbackLeft
		}
	}()

	// 1. Detect content bounding box (skipping transparent outer margins)
	minX, minY, maxX, maxY := w, h, -1, -1
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, _, a := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if maxX < minX || maxY < minY {
		return fallbackTop, fallbackRight, fallbackBottom, fallbackLeft
	}

	// 2. Scan inward from content bounds toward center to find where uniform interior fill begins
	centerX := (minX + maxX) / 2
	centerY := (minY + maxY) / 2
	centerCol := img.At(bounds.Min.X+centerX, bounds.Min.Y+centerY)

	// Top margin
	top = minY + 1
	for y := minY; y <= centerY; y++ {
		if img.At(bounds.Min.X+centerX, bounds.Min.Y+y) == centerCol {
			top = y
			break
		}
	}

	// Bottom margin
	bottom = (h - 1 - maxY) + 1
	for y := maxY; y >= centerY; y-- {
		if img.At(bounds.Min.X+centerX, bounds.Min.Y+y) == centerCol {
			bottom = h - 1 - y
			break
		}
	}

	// Left margin
	left = minX + 1
	for x := minX; x <= centerX; x++ {
		if img.At(bounds.Min.X+x, bounds.Min.Y+centerY) == centerCol {
			left = x
			break
		}
	}

	// Right margin
	right = (w - 1 - maxX) + 1
	for x := maxX; x >= centerX; x-- {
		if img.At(bounds.Min.X+x, bounds.Min.Y+centerY) == centerCol {
			right = w - 1 - x
			break
		}
	}

	// Sanity clamping: ensure margins don't exceed half the image dimensions
	if top+bottom >= h {
		top = h / 4
		bottom = h / 4
	}
	if left+right >= w {
		left = w / 4
		right = w / 4
	}

	return top, right, bottom, left
}

// NewNineSliceAuto creates a NineSliceConfig from an ebiten image with automatically detected edge margins.
func NewNineSliceAuto(img *ebiten.Image) NineSliceConfig {
	if img == nil {
		return NewNineSliceColorConfig(color.Transparent)
	}
	top, right, bottom, left := DetectNineSliceMargins(img)
	return NewNineSliceConfig(img, top, right, bottom, left)
}

// NewNineSliceFromImage creates a NineSliceConfig from an ebiten image with explicit edge margins.
func NewNineSliceFromImage(img *ebiten.Image, top, right, bottom, left int) NineSliceConfig {
	return NewNineSliceConfig(img, top, right, bottom, left)
}


