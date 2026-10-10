# SenjaUI

**SenjaUI** (`github.com/senjalab/senjaui`) is a high-performance, fluent UI container and banner package for Go game developers built on top of [Ebitengine v2](https://ebitengine.org) and [EbitenUI](https://github.com/ebitenui/ebitenui).

It wraps 9-slice sprite rendering into an intuitive, hierarchical container system with parent-child relationships, automatic transform propagation, animation hooks, and event bubbling—with **zero heap allocations during per-frame updates**.

---

## Key Highlights

- **Hierarchical Node Tree**: Parent-child banner containers (`AddChild`, `RemoveChild`) with automatic 2D transform propagation.
- **Nine-Slice Wrapper**: Bridges seamlessly to EbitenUI's `image.NewNineSlice`, `image.NewNineSliceColor`, `image.NewAdvancedNineSliceColor`, and procedural 9-slice generators.
- **Flexible Transforms**: Position `(X, Y)`, Size `(Width, Height)`, Padding `(Top, Right, Bottom, Left)`, and normalized `Anchor` / `Pivot` alignments.
- **Automatic Child Recalculation**: Resizing or repositioning a parent immediately propagates world bounds down the entire hierarchy.
- **Event Callbacks & Bubbling**: `OnOpen`, `OnClose`, and `OnContentClick(func(childID string))` with recursive parent event bubbling.
- **Smooth Animations & Easing**: Built-in scale-in, alpha fade, combined scale-and-fade, slide transitions, easing curves (`EaseOutQuad`, `EaseOutBack`, `EaseOutBounce`, etc.), and custom interpolation steps.
- **Zero Allocations Per Frame**: Zero heap allocations during `Update(dt)` and layout propagation (verified via benchmarks).
- **Self-Contained**: Works with procedural textures and standard fonts out of the box.

---

## Installation

```bash
go get github.com/senjalab/senjaui@latest
```

---

## Quick Start

```go
package main

import (
	"image/color"
	"log"

	"github.com/ebitenui/ebitenui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/senjalab/senjaui"
)

type Game struct {
	ui     *ebitenui.UI
	banner *senjaui.Banner
}

func NewGame() *Game {
	// Create parent notification banner with 9-slice border
	banner := senjaui.NewBanner(
		senjaui.WithID("notification-banner"),
		senjaui.WithPosition(400, 300), // Screen center
		senjaui.WithSize(500, 260),
		senjaui.WithAnchor(senjaui.AnchorCenter),
		senjaui.WithPivot(senjaui.PivotCenter),
		senjaui.WithBorder(color.RGBA{20, 24, 36, 245}, color.RGBA{212, 175, 55, 255}, 2),
		senjaui.WithAnimation(senjaui.AnimScaleAndFade, 0.35),
		senjaui.WithEase(senjaui.EaseOutBack),
	)

	// Nested interactable Card A
	cardAccept := senjaui.NewBanner(
		senjaui.WithID("card-accept"),
		senjaui.WithPosition(20, -20),
		senjaui.WithSize(210, 140),
		senjaui.WithAnchor(senjaui.AnchorBottomLeft),
		senjaui.WithPivot(senjaui.PivotBottomLeft),
		senjaui.WithBorder(color.RGBA{15, 35, 25, 245}, color.RGBA{46, 204, 113, 255}, 2),
	)

	// Nested interactable Card B
	cardDecline := senjaui.NewBanner(
		senjaui.WithID("card-decline"),
		senjaui.WithPosition(-20, -20),
		senjaui.WithSize(210, 140),
		senjaui.WithAnchor(senjaui.AnchorBottomRight),
		senjaui.WithPivot(senjaui.PivotBottomRight),
		senjaui.WithBorder(color.RGBA{35, 18, 22, 245}, color.RGBA{231, 76, 60, 255}, 2),
	)

	banner.AddChild(cardAccept)
	banner.AddChild(cardDecline)

	// Hierarchical event bubbling: parent catches all child clicks
	banner.OnContentClick(func(childID string) {
		println("Parent banner caught click from:", childID)
	})

	ui := senjaui.NewUI(banner)
	return &Game{ui: ui, banner: banner}
}

func (g *Game) Update() error {
	g.banner.Update(1.0 / 60.0)
	g.ui.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.ui.Draw(screen)
}

func (g *Game) Layout(w, h int) (int, int) {
	return 800, 600
}

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("SenjaUI Demo")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
```

---

## Core Concepts

### 1. NineSlice Configuration

`NineSliceConfig` defines margins, fill strategies, and source textures:

```go
cfg := senjaui.NewNineSliceConfig(image, top, right, bottom, left)
cfg.CenterFillStrategy = senjaui.CenterFillStretch // or CenterFillTransparent, CenterFillColor, CenterFillTile
cfg.FillColor = color.RGBA{30, 30, 40, 255}

// Compile directly into EbitenUI NineSlice
ns := cfg.Build()
```

Convenience constructors:
- `NewNineSliceColorConfig(fill color.Color)`
- `NewNineSliceBorderedConfig(fill, border color.Color, borderWidth int)`
- `NewNineSliceAdvancedConfig(fill color.Color, top, right, bottom, left int, borderColor color.Color)`
- `CreateProceduralNineSliceImage(size, cornerRadius int, fill, border color.Color, borderWidth int)`

### 2. Hierarchical Transforms & Anchoring

Children calculate positions relative to parent content bounds:

```
Child World X = Parent World X + Parent Padding Left + (Anchor.X * Content Width) + Child.X - (Pivot.X * Child Width)
```

Presets available:
- **Anchors**: `AnchorTopLeft`, `AnchorTopCenter`, `AnchorTopRight`, `AnchorMiddleLeft`, `AnchorCenter`, `AnchorMiddleRight`, `AnchorBottomLeft`, `AnchorBottomCenter`, `AnchorBottomRight`
- **Pivots**: `PivotTopLeft`, `PivotTopCenter`, `PivotCenter`, `PivotBottomRight`, etc.

When a parent's size or position changes via `banner.SetSize(w, h)` or `banner.SetPosition(x, y)`, child positions update automatically across all descendants.

### 3. Event Hooks & Bubbling

- `banner.OnOpen(func())`
- `banner.OnClose(func())`
- `banner.OnContentClick(func(childID string))`

When a child node or card triggers a click via `child.TriggerContentClick("card-id")`, the event bubbles up through its ancestor chain, allowing the root banner to handle events centrally.

### 4. Animation Interpolation

Smooth transitions on open/close:
- `AnimScaleIn`: Scales smoothly around pivot
- `AnimFadeIn`: Alpha fade
- `AnimScaleAndFade`: Dialogue scale & fade
- `AnimSlideInTop` / `AnimSlideInBottom` / `AnimSlideInLeft` / `AnimSlideInRight`
- `AnimCustom`: Custom callback using `WithAnimationStep(func(b *Banner, t float64))`

Easing options: `EaseLinear`, `EaseOutQuad`, `EaseInOutQuad`, `EaseOutCubic`, `EaseOutBack`, `EaseOutBounce`.

---

## Running the Example

```bash
go run ./examples/notification
```

- **[Space]**: Toggle open / close animation
- **[1 - 4]**: Switch animation style (Scale, Fade, Scale & Fade, Slide-In)
- **[R]**: Cycle parent banner size to demonstrate automatic child realignment
- **Click cards / buttons**: Demonstrates hierarchical event bubbling

---

## Benchmarks

```bash
$ go test -bench=. ./...
BenchmarkBanner_Update-8          41949722    27.69 ns/op    0 B/op    0 allocs/op
BenchmarkBanner_GetWorldBounds-8  97558924    11.61 ns/op    0 B/op    0 allocs/op
```

---

## License

MIT License. Designed and built by SenjaLab.

