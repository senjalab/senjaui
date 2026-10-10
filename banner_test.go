package senjaui

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestBanner_HierarchyAndReparenting(t *testing.T) {
	parent := NewBanner(
		WithID("parent"),
		WithPosition(100, 100),
		WithSize(400, 300),
	)

	child1 := NewBanner(
		WithID("child1"),
		WithPosition(10, 20),
		WithSize(100, 80),
	)

	child2 := NewBanner(
		WithID("child2"),
		WithPosition(50, 50),
		WithSize(120, 90),
	)

	parent.AddChild(child1)
	parent.AddChild(child2)

	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}
	if child1.Parent != parent || child2.Parent != parent {
		t.Fatalf("expected children's Parent to point to parent")
	}

	// Remove child1
	parent.RemoveChild(child1)
	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child after removal, got %d", len(parent.Children))
	}
	if child1.Parent != nil {
		t.Fatalf("expected removed child's Parent to be nil")
	}

	// Reparent child2 to a new parent
	newParent := NewBanner(WithID("new-parent"), WithSize(500, 500))
	newParent.AddChild(child2)
	if len(parent.Children) != 0 {
		t.Fatalf("expected old parent to have 0 children after reparenting, got %d", len(parent.Children))
	}
	if child2.Parent != newParent {
		t.Fatalf("expected child2 Parent to be newParent")
	}
}

func TestBanner_WorldBoundsCalculation(t *testing.T) {
	// Root banner positioned at (100, 100) with size 400x300, TopLeft anchor & pivot
	parent := NewBanner(
		WithID("parent"),
		WithPosition(100, 100),
		WithSize(400, 300),
		WithPadding(20, 20, 20, 20), // 20px padding on all sides
		WithAnchor(AnchorTopLeft),
		WithPivot(PivotTopLeft),
	)

	wx, wy, ww, wh := parent.GetWorldBounds()
	if wx != 100 || wy != 100 || ww != 400 || wh != 300 {
		t.Fatalf("expected parent bounds (100, 100, 400, 300), got (%v, %v, %v, %v)", wx, wy, ww, wh)
	}

	// Child anchored TopLeft inside parent's padded area
	// Parent padded area: X=100+20=120, Y=100+20=120
	childTL := NewBanner(
		WithID("childTL"),
		WithPosition(10, 15),
		WithSize(80, 50),
		WithAnchor(AnchorTopLeft),
		WithPivot(PivotTopLeft),
	)
	parent.AddChild(childTL)

	cx, cy, cw, ch := childTL.GetWorldBounds()
	expectedX := 100.0 + 20.0 + 10.0 // 130
	expectedY := 100.0 + 20.0 + 15.0 // 135
	if cx != expectedX || cy != expectedY || cw != 80 || ch != 50 {
		t.Fatalf("expected child bounds (%v, %v, 80, 50), got (%v, %v, %v, %v)", expectedX, expectedY, cx, cy, cw, ch)
	}

	// Child anchored Center inside parent
	// Parent content width = 400 - 40 = 360, height = 300 - 40 = 260
	// Center anchor X = 120 + 0.5 * 360 = 300
	// Center anchor Y = 120 + 0.5 * 260 = 250
	// With PivotCenter and size 100x60:
	// Top-left = (300 - 50, 250 - 30) = (250, 220)
	childCenter := NewBanner(
		WithID("childCenter"),
		WithPosition(0, 0),
		WithSize(100, 60),
		WithAnchor(AnchorCenter),
		WithPivot(PivotCenter),
	)
	parent.AddChild(childCenter)

	cx, cy, cw, ch = childCenter.GetWorldBounds()
	if cx != 250 || cy != 220 || cw != 100 || ch != 60 {
		t.Fatalf("expected centered child bounds (250, 220, 100, 60), got (%v, %v, %v, %v)", cx, cy, cw, ch)
	}
}

func TestBanner_NestedGrandchildTransformPropagation(t *testing.T) {
	root := NewBanner(WithPosition(50, 50), WithSize(500, 500))
	branch := NewBanner(WithPosition(50, 50), WithSize(200, 200))
	leaf := NewBanner(WithPosition(10, 10), WithSize(50, 50))

	root.AddChild(branch)
	branch.AddChild(leaf)

	lx, ly, lw, lh := leaf.GetWorldBounds()
	// root (50, 50) + branch (50, 50) + leaf (10, 10) = (110, 110)
	if lx != 110 || ly != 110 || lw != 50 || lh != 50 {
		t.Fatalf("expected grandchild bounds (110, 110, 50, 50), got (%v, %v, %v, %v)", lx, ly, lw, lh)
	}

	// Move root: grandchild must automatically propagate
	root.SetPosition(100, 200)
	lx, ly, _, _ = leaf.GetWorldBounds()
	if lx != 160 || ly != 260 {
		t.Fatalf("expected grandchild updated bounds (160, 260), got (%v, %v)", lx, ly)
	}
}

func TestBanner_DimensionUpdateAutomaticallyRecomputesChildren(t *testing.T) {
	parent := NewBanner(
		WithSize(400, 200),
		WithAnchor(AnchorTopLeft),
		WithPivot(PivotTopLeft),
	)

	// Child anchored to right side (AnchorTopRight, PivotTopRight)
	child := NewBanner(
		WithPosition(-20, 10),
		WithSize(100, 50),
		WithAnchor(AnchorTopRight),
		WithPivot(PivotTopRight),
	)
	parent.AddChild(child)

	// In initial size (400x200):
	// anchorX = 400, pos = -20, pivot = 1.0 (right edge):
	// wx = 400 + (-20) - 1.0 * 100 = 280
	cx, _, _, _ := child.GetWorldBounds()
	if cx != 280 {
		t.Fatalf("expected child X to be 280, got %v", cx)
	}

	// Now resize parent width to 600:
	parent.SetSize(600, 200)

	// Child must automatically adjust:
	// anchorX = 600, pos = -20, pivot = 1.0:
	// wx = 600 + (-20) - 100 = 480
	cx, _, _, _ = child.GetWorldBounds()
	if cx != 480 {
		t.Fatalf("expected child X to be automatically updated to 480, got %v", cx)
	}

	// Verify EbitenUI container location was also synchronized
	cRect := child.GetContainer().GetWidget().Rect
	if cRect.Min.X != 480 {
		t.Fatalf("expected underlying EbitenUI container Rect.Min.X to be 480, got %v", cRect.Min.X)
	}
}

func TestBanner_EventCallbacksAndBubbling(t *testing.T) {
	parent := NewBanner(WithID("notification-banner"), WithSize(400, 200))
	cardA := NewBanner(WithID("card-accept"), WithSize(120, 60))
	cardB := NewBanner(WithID("card-decline"), WithSize(120, 60))

	parent.AddChild(cardA)
	parent.AddChild(cardB)

	var parentReceivedChildID string
	var cardAReceivedID string

	parent.OnContentClick(func(childID string) {
		parentReceivedChildID = childID
	})

	cardA.OnContentClick(func(childID string) {
		cardAReceivedID = childID
	})

	// Click on cardA
	cardA.TriggerContentClick(cardA.ID())

	if cardAReceivedID != "card-accept" {
		t.Fatalf("expected cardA callback to receive 'card-accept', got %q", cardAReceivedID)
	}
	if parentReceivedChildID != "card-accept" {
		t.Fatalf("expected parent to bubble 'card-accept', got %q", parentReceivedChildID)
	}

	// Click on cardB
	cardB.TriggerContentClick(cardB.ID())
	if parentReceivedChildID != "card-decline" {
		t.Fatalf("expected parent to bubble 'card-decline', got %q", parentReceivedChildID)
	}
}

func TestBanner_OpenCloseCallbacksAndTransitions(t *testing.T) {
	banner := NewBanner(
		WithInitiallyOpen(false),
		WithAnimation(AnimScaleIn, 0.2),
		WithEase(EaseLinear),
	)

	var opened, closed bool
	banner.OnOpen(func() { opened = true })
	banner.OnClose(func() { closed = true })

	if banner.IsOpen() {
		t.Fatalf("expected banner to be initially closed")
	}

	banner.Open()
	if !opened {
		t.Fatalf("expected OnOpen callback to be called")
	}

	// Step forward half duration (0.1s out of 0.2s)
	banner.Update(0.1)
	if math.Abs(banner.Scale()-0.5) > 0.01 {
		t.Fatalf("expected scale ~0.5 at halfway, got %v", banner.Scale())
	}

	// Step to end of opening
	banner.Update(0.15)
	if math.Abs(banner.Scale()-1.0) > 0.001 {
		t.Fatalf("expected scale 1.0 at completion, got %v", banner.Scale())
	}

	// Now close
	banner.Close()
	banner.Update(0.25)
	if banner.IsOpen() {
		t.Fatalf("expected banner to be closed after full duration")
	}
	if !closed {
		t.Fatalf("expected OnClose callback to be called")
	}
}

func TestBanner_CustomAnimationStep(t *testing.T) {
	var stepCount int
	var lastT float64

	banner := NewBanner(
		WithInitiallyOpen(false),
		WithAnimation(AnimCustom, 0.1),
		WithAnimationStep(func(b *Banner, t float64) {
			stepCount++
			lastT = t
		}),
	)

	banner.Open()
	banner.Update(0.05)
	banner.Update(0.05)

	if stepCount == 0 {
		t.Fatalf("expected custom animation step to be invoked")
	}
	if math.Abs(lastT-1.0) > 0.01 {
		t.Fatalf("expected final t ~1.0, got %v", lastT)
	}
}

func TestBanner_NineSliceBridging(t *testing.T) {
	// Color nine slice
	bColor := NewBanner(WithBackgroundColor(color.RGBA{20, 30, 40, 255}))
	if bColor.GetContainer() == nil {
		t.Fatalf("expected EbitenUI container to be initialized")
	}

	// Bordered nine slice
	bBorder := NewBanner(WithBorder(color.Black, color.White, 2))
	if bBorder.GetContainer() == nil {
		t.Fatalf("expected container to be initialized")
	}

	// Procedural 9-slice image
	img := CreateProceduralNineSliceImage(24, 4, color.RGBA{100, 100, 100, 255}, color.RGBA{255, 200, 50, 255}, 2)
	cfg := NewNineSliceConfig(img, 4, 4, 4, 4)
	bImg := NewBanner(WithNineSliceConfig(cfg))
	if bImg.GetContainer() == nil {
		t.Fatalf("expected container to be initialized")
	}
}

func TestBanner_ZeroHeapAllocationsDuringUpdate(t *testing.T) {
	parent := NewBanner(
		WithPosition(50, 50),
		WithSize(400, 300),
		WithAnimation(AnimScaleAndFade, 0.5),
	)
	card1 := NewBanner(WithPosition(10, 10), WithSize(100, 80))
	card2 := NewBanner(WithPosition(120, 10), WithSize(100, 80))
	parent.AddChild(card1)
	parent.AddChild(card2)

	parent.Open()

	// Warm up
	parent.Update(0.016)

	// Measure heap allocations during active per-frame update
	allocs := testing.AllocsPerRun(100, func() {
		parent.Update(0.016)
	})

	if allocs > 0 {
		t.Fatalf("expected 0 heap allocations per update frame, got %v allocs/run", allocs)
	}
}

func BenchmarkBanner_Update(b *testing.B) {
	parent := NewBanner(
		WithPosition(100, 100),
		WithSize(600, 400),
		WithAnimation(AnimScaleAndFade, 1.0),
	)
	for i := 0; i < 5; i++ {
		child := NewBanner(WithPosition(float64(i*50), 20), WithSize(80, 80))
		parent.AddChild(child)
	}
	parent.Open()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parent.Update(0.016)
	}
}

func BenchmarkBanner_GetWorldBounds(b *testing.B) {
	root := NewBanner(WithPosition(10, 10), WithSize(500, 500))
	mid := NewBanner(WithPosition(20, 20), WithSize(300, 300))
	leaf := NewBanner(WithPosition(30, 30), WithSize(100, 100))
	root.AddChild(mid)
	mid.AddChild(leaf)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = leaf.GetWorldBounds()
	}
}

func TestBanner_DirectDrawCall(t *testing.T) {
	screen := ebiten.NewImage(800, 600)
	parent := NewBanner(
		WithPosition(100, 100),
		WithSize(300, 200),
		WithBorder(color.RGBA{30, 30, 30, 200}, color.RGBA{200, 200, 200, 255}, 2),
	)
	child := NewBanner(
		WithPosition(20, 20),
		WithSize(100, 50),
		WithBackgroundColor(color.RGBA{100, 150, 200, 255}),
	)
	parent.AddChild(child)

	// Draw directly onto ebiten Image
	parent.Draw(screen)
}

