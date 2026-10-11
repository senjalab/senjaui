package senjaui

import (
	"image"
	"image/color"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
)

// Banner represents a hierarchical UI node wrapping an EbitenUI container and widget.
// It manages 2D transform propagation, 9-slice background rendering, animations,
// and event callbacks with parent-child bubbling.
type Banner struct {
	id string

	// Transform
	X       float64
	Y       float64
	Width   float64
	Height  float64
	Padding Insets
	Anchor  Anchor
	Pivot   Pivot

	// Hierarchy
	Parent   *Banner
	Children []*Banner

	// EbitenUI Bridge
	container       *widget.Container
	nineSliceConfig NineSliceConfig
	nineSlice       *eimage.NineSlice

	// World Bounds Cache (avoids per-frame heap allocations)
	worldX float64
	worldY float64
	worldW float64
	worldH float64

	// State & Visibility
	isOpen  bool
	visible bool
	opacity float64
	scale   float64

	// Animation
	animType     AnimType
	animDuration float64
	animProgress float64
	animDir      int // +1 opening, -1 closing, 0 idle
	easeFn       EaseFunc
	animStepFn   AnimationStepFunc
	slideOffsetX float64
	slideOffsetY float64

	// Event Callbacks
	onOpen         func()
	onClose        func()
	onContentClick func(childID string)

	// User content widget & container customization
	contentWidget       widget.PreferredSizeLocateableWidget
	customLayouter      widget.Layouter
	customContainerOpts []widget.ContainerOpt
}

// NewBanner constructs a new hierarchical Banner using functional options.
func NewBanner(opts ...BannerOption) *Banner {
	b := &Banner{
		id:           "banner",
		Width:        200,
		Height:       100,
		Anchor:       AnchorTopLeft,
		Pivot:        PivotTopLeft,
		scale:        1.0,
		opacity:      1.0,
		isOpen:       true,
		visible:      true,
		animType:     AnimNone,
		animDuration: 0.25,
		easeFn:       EaseOutQuad,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(b)
		}
	}

	b.initContainer()
	b.recomputeLayout()

	return b
}

func (b *Banner) initContainer() {
	if b.nineSlice == nil {
		b.nineSlice = b.nineSliceConfig.Build()
	}

	containerOpts := []widget.ContainerOpt{
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.TrackHover(true),
			widget.WidgetOpts.MouseButtonClickedHandler(func(args *widget.WidgetMouseButtonClickedEventArgs) {
				b.TriggerContentClick(b.id)
			}),
		),
	}

	if b.customLayouter != nil {
		containerOpts = append(containerOpts, widget.ContainerOpts.Layout(b.customLayouter))
	}

	if b.nineSlice != nil {
		containerOpts = append(containerOpts, widget.ContainerOpts.BackgroundImage(b.nineSlice))
	}

	containerOpts = append(containerOpts, b.customContainerOpts...)

	b.container = widget.NewContainer(containerOpts...)

	if !b.visible || !b.isOpen {
		b.container.GetWidget().SetVisibility(widget.Visibility_Hide)
	}

	if b.contentWidget != nil {
		b.container.AddChild(b.contentWidget)
	}
}

// ID returns the banner identifier.
func (b *Banner) ID() string {
	return b.id
}

// SetID sets the banner identifier.
func (b *Banner) SetID(id string) {
	b.id = id
}

// AddChild attaches a child Banner to this Banner node.
// It establishes the parent-child hierarchy, links the underlying EbitenUI container,
// and automatically propagates transform calculations.
func (b *Banner) AddChild(child *Banner) {
	if child == nil || child == b {
		return
	}

	if child.Parent != nil {
		child.Parent.RemoveChild(child)
	}

	child.Parent = b
	b.Children = append(b.Children, child)

	if b.container != nil && child.container != nil {
		b.container.AddChild(child.container)
	}

	child.recomputeLayout()
}

// RemoveChild removes a child Banner from this node's hierarchy.
func (b *Banner) RemoveChild(child *Banner) {
	if child == nil {
		return
	}

	for i, c := range b.Children {
		if c == child {
			b.Children = append(b.Children[:i], b.Children[i+1:]...)
			child.Parent = nil
			if b.container != nil && child.container != nil {
				b.container.RemoveChild(child.container)
			}
			child.recomputeLayout()
			return
		}
	}
}

// SetPosition sets the local coordinates of the banner and updates the hierarchy.
func (b *Banner) SetPosition(x, y float64) {
	b.X = x
	b.Y = y
	b.recomputeLayout()
}

// SetSize sets the dimensions of the banner and updates child layouts.
func (b *Banner) SetSize(w, h float64) {
	b.Width = w
	b.Height = h
	b.recomputeLayout()
}

// SetPadding sets edge padding inside this container.
func (b *Banner) SetPadding(top, right, bottom, left float64) {
	b.Padding = NewInsets(top, right, bottom, left)
	b.recomputeLayout()
}

// SetAnchor sets the child alignment point inside parent content space.
func (b *Banner) SetAnchor(a Anchor) {
	b.Anchor = a
	b.recomputeLayout()
}

// SetPivot sets the reference origin point inside the banner.
func (b *Banner) SetPivot(p Pivot) {
	b.Pivot = p
	b.recomputeLayout()
}

// GetWorldBounds computes the recursive world bounding box (x, y, w, h) from parents,
// taking into account hierarchical offsets, anchor alignment, pivot offsets, and animation scales.
func (b *Banner) GetWorldBounds() (x, y, w, h float64) {
	effW := b.Width * b.scale
	effH := b.Height * b.scale

	if b.Parent == nil {
		wx := b.X - b.Pivot.X*effW + b.slideOffsetX
		wy := b.Y - b.Pivot.Y*effH + b.slideOffsetY
		return wx, wy, effW, effH
	}

	pX, pY, pW, pH := b.Parent.GetWorldBounds()
	contentX := pX + b.Parent.Padding.Left
	contentY := pY + b.Parent.Padding.Top
	contentW := pW - (b.Parent.Padding.Left + b.Parent.Padding.Right)
	contentH := pH - (b.Parent.Padding.Top + b.Parent.Padding.Bottom)

	anchorX := contentX + b.Anchor.X*contentW
	anchorY := contentY + b.Anchor.Y*contentH

	wx := anchorX + b.X - b.Pivot.X*effW + b.slideOffsetX
	wy := anchorY + b.Y - b.Pivot.Y*effH + b.slideOffsetY
	return wx, wy, effW, effH
}

// GetBaseWorldBounds computes recursive unscaled world bounds without transient animation offsets.
func (b *Banner) GetBaseWorldBounds() (x, y, w, h float64) {
	if b.Parent == nil {
		wx := b.X - b.Pivot.X*b.Width
		wy := b.Y - b.Pivot.Y*b.Height
		return wx, wy, b.Width, b.Height
	}

	pX, pY, pW, pH := b.Parent.GetBaseWorldBounds()
	contentX := pX + b.Parent.Padding.Left
	contentY := pY + b.Parent.Padding.Top
	contentW := pW - (b.Parent.Padding.Left + b.Parent.Padding.Right)
	contentH := pH - (b.Parent.Padding.Top + b.Parent.Padding.Bottom)

	anchorX := contentX + b.Anchor.X*contentW
	anchorY := contentY + b.Anchor.Y*contentH

	wx := anchorX + b.X - b.Pivot.X*b.Width
	wy := anchorY + b.Y - b.Pivot.Y*b.Height
	return wx, wy, b.Width, b.Height
}

// recomputeLayout recursively updates world positions and container bounds without heap allocations.
func (b *Banner) recomputeLayout() {
	wx, wy, ww, wh := b.GetWorldBounds()
	b.worldX = wx
	b.worldY = wy
	b.worldW = ww
	b.worldH = wh

	if b.container != nil {
		rect := image.Rect(int(wx), int(wy), int(wx+ww), int(wy+wh))
		b.container.SetLocation(rect)
		b.container.RequestRelayout()
	}

	for _, child := range b.Children {
		child.recomputeLayout()
	}
}

// RequestRelayout marks this banner container and all its child containers dirty so their layouts re-synchronize.
func (b *Banner) RequestRelayout() {
	if b.container != nil {
		b.container.RequestRelayout()
	}
	for _, child := range b.Children {
		child.RequestRelayout()
	}
}

// GetContainer returns the underlying EbitenUI container.
func (b *Banner) GetContainer() *widget.Container {
	return b.container
}

// GetWidget returns the EbitenUI widget of the underlying container.
func (b *Banner) GetWidget() *widget.Widget {
	if b.container != nil {
		return b.container.GetWidget()
	}
	return nil
}

// AddWidget adds an arbitrary EbitenUI widget inside this banner container.
func (b *Banner) AddWidget(w widget.PreferredSizeLocateableWidget) {
	b.contentWidget = w
	if b.container != nil && w != nil {
		b.container.AddChild(w)
	}
}

// SetNineSliceConfig updates the NineSlice background from configuration.
func (b *Banner) SetNineSliceConfig(cfg NineSliceConfig) {
	b.nineSliceConfig = cfg
	b.nineSlice = cfg.Build()
	if b.container != nil {
		b.container.SetBackgroundImage(b.nineSlice)
	}
}

// SetNineSlice directly applies an existing EbitenUI NineSlice.
func (b *Banner) SetNineSlice(ns *eimage.NineSlice) {
	b.nineSlice = ns
	if b.container != nil {
		b.container.SetBackgroundImage(ns)
	}
}

// SetImage dynamically updates the banner's nine-slice texture with explicit border margins.
// The nine-slice will adapt automatically without breaking any hierarchy or child transforms.
func (b *Banner) SetImage(img *ebiten.Image, top, right, bottom, left int) {
	b.SetNineSliceConfig(NewNineSliceFromImage(img, top, right, bottom, left))
}

// SetImageAuto dynamically updates the banner's nine-slice texture, automatically computing optimal border margins.
func (b *Banner) SetImageAuto(img *ebiten.Image) {
	b.SetNineSliceConfig(NewNineSliceAuto(img))
}

// SetImageFile dynamically loads an image file from disk and updates the banner's nine-slice texture.
func (b *Banner) SetImageFile(path string, top, right, bottom, left int) error {
	img, err := LoadImageFromFile(path)
	if err != nil {
		return err
	}
	b.SetImage(img, top, right, bottom, left)
	return nil
}

// SetBackgroundColor sets a solid color background.
func (b *Banner) SetBackgroundColor(c color.Color) {
	b.SetNineSliceConfig(NewNineSliceColorConfig(c))
}

// SetBorderColor sets a bordered background.
func (b *Banner) SetBorderColor(bodyColor, borderColor color.Color, borderWidth int) {
	b.SetNineSliceConfig(NewNineSliceBorderedConfig(bodyColor, borderColor, borderWidth))
}

// IsOpen returns whether the banner is currently open or opening.
func (b *Banner) IsOpen() bool {
	return b.isOpen
}

// IsVisible returns whether the banner is visible on screen.
func (b *Banner) IsVisible() bool {
	return b.visible
}

// Scale returns the current animation scale.
func (b *Banner) Scale() float64 {
	return b.scale
}

// Opacity returns the current animation opacity in [0.0, 1.0].
func (b *Banner) Opacity() float64 {
	return b.opacity
}

// OnOpen registers a callback invoked when the banner opens.
func (b *Banner) OnOpen(fn func()) {
	b.onOpen = fn
}

// OnClose registers a callback invoked when the banner closes.
func (b *Banner) OnClose(fn func()) {
	b.onClose = fn
}

// OnContentClick registers a callback invoked when this banner or any child is clicked.
// childID identifies which node or card was clicked.
func (b *Banner) OnContentClick(fn func(childID string)) {
	b.onContentClick = fn
}

// TriggerContentClick invokes the OnContentClick callback and bubbles up to the parent banner.
func (b *Banner) TriggerContentClick(childID string) {
	if b.onContentClick != nil {
		b.onContentClick(childID)
	}

	if b.Parent != nil {
		b.Parent.TriggerContentClick(childID)
	}
}

// SetAnimation configures transition animation parameters.
func (b *Banner) SetAnimation(animType AnimType, duration float64, ease EaseFunc) {
	b.animType = animType
	b.animDuration = duration
	if ease != nil {
		b.easeFn = ease
	}
}

// SetAnimationStep registers a custom interpolation step callback.
func (b *Banner) SetAnimationStep(fn AnimationStepFunc) {
	b.animStepFn = fn
	b.animType = AnimCustom
}

// Open opens the banner, triggering OnOpen and running the opening transition.
func (b *Banner) Open() {
	if b.isOpen && b.animDir >= 0 {
		return
	}

	b.isOpen = true
	b.visible = true
	if b.container != nil {
		b.container.GetWidget().SetVisibility(widget.Visibility_Show)
	}

	if b.animType == AnimNone || b.animDuration <= 0 {
		b.animProgress = 1.0
		b.animDir = 0
		b.scale = 1.0
		b.opacity = 1.0
		b.slideOffsetX = 0
		b.slideOffsetY = 0
		b.recomputeLayout()
		if b.onOpen != nil {
			b.onOpen()
		}
		return
	}

	b.animDir = 1
	b.animProgress = 0.0
	b.applyAnimationStep(0.0)
	b.recomputeLayout()

	if b.onOpen != nil {
		b.onOpen()
	}
}

// Close closes the banner, running the closing transition and triggering OnClose.
func (b *Banner) Close() {
	if !b.isOpen && b.animDir <= 0 {
		return
	}

	if b.animType == AnimNone || b.animDuration <= 0 {
		b.isOpen = false
		b.visible = false
		b.animProgress = 0.0
		b.animDir = 0
		b.scale = 0.0
		b.opacity = 0.0
		b.slideOffsetX = 0
		b.slideOffsetY = 0
		if b.container != nil {
			b.container.GetWidget().SetVisibility(widget.Visibility_Hide)
		}
		b.recomputeLayout()
		if b.onClose != nil {
			b.onClose()
		}
		return
	}

	b.animDir = -1
}

// Toggle flips between open and closed states.
func (b *Banner) Toggle() {
	if b.isOpen && b.animDir >= 0 {
		b.Close()
	} else {
		b.Open()
	}
}

// Update advances animation interpolation steps without allocating heap memory.
func (b *Banner) Update(dt float64) {
	if b.animDir != 0 && b.animDuration > 0 {
		step := dt / b.animDuration
		if b.animDir > 0 {
			b.animProgress += step
			if b.animProgress >= 1.0 {
				b.animProgress = 1.0
				b.animDir = 0
			}
		} else {
			b.animProgress -= step
			if b.animProgress <= 0.0 {
				b.animProgress = 0.0
				b.animDir = 0
				b.isOpen = false
				b.visible = false
				if b.container != nil {
					b.container.GetWidget().SetVisibility(widget.Visibility_Hide)
				}
				if b.onClose != nil {
					b.onClose()
				}
			}
		}

		t := b.animProgress
		if b.easeFn != nil {
			t = b.easeFn(b.animProgress)
		}
		b.applyAnimationStep(t)
		b.recomputeLayout()
	}

	for _, child := range b.Children {
		child.Update(dt)
	}
}

func (b *Banner) applyAnimationStep(t float64) {
	if b.animStepFn != nil {
		b.animStepFn(b, t)
		return
	}

	switch b.animType {
	case AnimScaleIn:
		b.scale = t
		b.opacity = 1.0
		b.slideOffsetX = 0
		b.slideOffsetY = 0
	case AnimFadeIn:
		b.scale = 1.0
		b.opacity = t
		b.slideOffsetX = 0
		b.slideOffsetY = 0
	case AnimScaleAndFade:
		b.scale = 0.5 + 0.5*t
		b.opacity = t
		b.slideOffsetX = 0
		b.slideOffsetY = 0
	case AnimSlideInTop:
		b.scale = 1.0
		b.opacity = t
		b.slideOffsetX = 0
		b.slideOffsetY = -(1.0 - t) * 80.0
	case AnimSlideInBottom:
		b.scale = 1.0
		b.opacity = t
		b.slideOffsetX = 0
		b.slideOffsetY = (1.0 - t) * 80.0
	case AnimSlideInLeft:
		b.scale = 1.0
		b.opacity = t
		b.slideOffsetX = -(1.0 - t) * 80.0
		b.slideOffsetY = 0
	case AnimSlideInRight:
		b.scale = 1.0
		b.opacity = t
		b.slideOffsetX = (1.0 - t) * 80.0
		b.slideOffsetY = 0
	default:
		b.scale = 1.0
		b.opacity = 1.0
		b.slideOffsetX = 0
		b.slideOffsetY = 0
	}
}

// Draw renders the NineSlice background and child banners onto screen with active opacity and geometry.
func (b *Banner) Draw(screen *ebiten.Image) {
	if !b.visible || b.opacity <= 0.001 {
		return
	}

	wx, wy, ww, wh := b.GetWorldBounds()
	if b.nineSlice != nil && ww > 0 && wh > 0 {
		b.nineSlice.Draw(screen, int(ww), int(wh), func(opts *ebiten.DrawImageOptions) {
			if b.opacity < 1.0 {
				opts.ColorScale.ScaleAlpha(float32(b.opacity))
			}
			opts.GeoM.Translate(wx, wy)
		})
	}

	for _, child := range b.Children {
		child.Draw(screen)
	}
}
