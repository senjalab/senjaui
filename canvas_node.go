package senjaui

import (
	"image/color"
	"math"

	eimage "github.com/ebitenui/ebitenui/image"
	"github.com/hajimehoshi/ebiten/v2"
)

// AABB represents an Axis-Aligned Bounding Box defined by Min and Max extents.
// It is designed for 2D spatial queries, hit testing, and seamless compatibility with interaction systems like senjalab/interaksi.
type AABB struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// NewAABB constructs an AABB ensuring Min is always <= Max.
func NewAABB(minX, minY, maxX, maxY float64) AABB {
	return AABB{
		MinX: math.Min(minX, maxX),
		MinY: math.Min(minY, maxY),
		MaxX: math.Max(minX, maxX),
		MaxY: math.Max(minY, maxY),
	}
}

// Width returns the horizontal size of the AABB.
func (a AABB) Width() float64 {
	return math.Max(0, a.MaxX-a.MinX)
}

// Height returns the vertical size of the AABB.
func (a AABB) Height() float64 {
	return math.Max(0, a.MaxY-a.MinY)
}

// Center returns the midpoint of the AABB.
func (a AABB) Center() (cx, cy float64) {
	return (a.MinX + a.MaxX) * 0.5, (a.MinY + a.MaxY) * 0.5
}

// Contains tests whether the point (px, py) lies within this bounding box.
func (a AABB) Contains(px, py float64) bool {
	return px >= a.MinX && px <= a.MaxX && py >= a.MinY && py <= a.MaxY
}

// Intersects tests whether this AABB overlaps another AABB.
func (a AABB) Intersects(other AABB) bool {
	return a.MinX <= other.MaxX &&
		a.MaxX >= other.MinX &&
		a.MinY <= other.MaxY &&
		a.MaxY >= other.MinY
}

// Pad returns a new AABB expanded outwards by the given edge insets.
func (a AABB) Pad(top, right, bottom, left float64) AABB {
	return AABB{
		MinX: a.MinX - left,
		MinY: a.MinY - top,
		MaxX: a.MaxX + right,
		MaxY: a.MaxY + bottom,
	}
}

// CanvasNode is an interactive, visual graph node decoupled from game-specific logic.
// It supports 2D canvas transforms, nine-slice background rendering, directed graph relationships
// (e.g., prerequisite trees, skill progression, tech research), and stores custom domain payload T.
type CanvasNode[T any] struct {
	ID string

	// Transform & Bounding Box (Graph canvas coordinates)
	X, Y          float64
	Width, Height float64

	// Visual Asset & Nine-Slice configuration
	Image           *ebiten.Image
	NineSliceConfig NineSliceConfig
	nineSlice       *eimage.NineSlice

	// Visual appearance & modulation
	Tint    color.Color // Optional tint color (e.g., grey for locked, gold for unlocked)
	Opacity float64     // Alpha modulation [0.0, 1.0]
	Hidden  bool        // Whether this node is hidden from graph rendering

	// Directed graph relationships
	InboundIDs  []string // Parent nodes (prerequisites, source nodes)
	OutboundIDs []string // Child nodes (unlocked nodes, target nodes)

	// Custom domain payload (Skill data, Tech research, Dialogue prompt, etc.)
	Data T

	// Event callbacks
	onClick func(node *CanvasNode[T])
	onHover func(node *CanvasNode[T], hovered bool)
}

// CanvasNodeOption is a functional option for configuring a CanvasNode.
type CanvasNodeOption[T any] func(*CanvasNode[T])

// NewCanvasNode constructs a new CanvasNode with default values and functional options.
func NewCanvasNode[T any](id string, x, y, width, height float64, data T, opts ...CanvasNodeOption[T]) *CanvasNode[T] {
	node := &CanvasNode[T]{
		ID:          id,
		X:           x,
		Y:           y,
		Width:       width,
		Height:      height,
		Opacity:     1.0,
		Data:        data,
		InboundIDs:  make([]string, 0),
		OutboundIDs: make([]string, 0),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(node)
		}
	}

	if node.Image != nil && node.nineSlice == nil {
		if node.NineSliceConfig.Image == nil {
			node.NineSliceConfig = NewNineSliceAuto(node.Image)
		}
		node.nineSlice = node.NineSliceConfig.Build()
	}

	return node
}

// WithNodeImage assigns an image and explicit 9-slice margins to the CanvasNode.
func WithNodeImage[T any](img *ebiten.Image, top, right, bottom, left int) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.Image = img
		n.NineSliceConfig = NewNineSliceFromImage(img, top, right, bottom, left)
		n.nineSlice = n.NineSliceConfig.Build()
	}
}

// WithNodeImageAuto assigns an image and auto-detects optimal 9-slice edge margins.
func WithNodeImageAuto[T any](img *ebiten.Image) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.Image = img
		n.NineSliceConfig = NewNineSliceAuto(img)
		n.nineSlice = n.NineSliceConfig.Build()
	}
}

// WithNodeNineSliceConfig sets a custom NineSliceConfig for the node.
func WithNodeNineSliceConfig[T any](cfg NineSliceConfig) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.NineSliceConfig = cfg
		n.Image = cfg.Image
		n.nineSlice = cfg.Build()
	}
}

// WithNodeInbound pre-populates prerequisite/incoming node IDs.
func WithNodeInbound[T any](ids ...string) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.AddInbound(ids...)
	}
}

// WithNodeOutbound pre-populates unlocked/outgoing node IDs.
func WithNodeOutbound[T any](ids ...string) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.AddOutbound(ids...)
	}
}

// WithNodeTint sets an initial visual tint color.
func WithNodeTint[T any](c color.Color) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.Tint = c
	}
}

// WithNodeClick sets a click handler on this node.
func WithNodeClick[T any](fn func(node *CanvasNode[T])) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.onClick = fn
	}
}

// WithNodeHover sets a hover state change handler on this node.
func WithNodeHover[T any](fn func(node *CanvasNode[T], hovered bool)) CanvasNodeOption[T] {
	return func(n *CanvasNode[T]) {
		n.onHover = fn
	}
}

// AABB returns the axis-aligned bounding box of this node in canvas coordinates.
func (n *CanvasNode[T]) AABB() AABB {
	return NewAABB(n.X, n.Y, n.X+n.Width, n.Y+n.Height)
}

// Bounds returns the min and max coordinates (minX, minY, maxX, maxY).
func (n *CanvasNode[T]) Bounds() (minX, minY, maxX, maxY float64) {
	return n.X, n.Y, n.X + n.Width, n.Y + n.Height
}

// Contains tests whether the given point lies inside this node's AABB.
func (n *CanvasNode[T]) Contains(px, py float64) bool {
	return px >= n.X && px <= n.X+n.Width && py >= n.Y && py <= n.Y+n.Height
}

// Center returns the midpoint coordinate of the node.
func (n *CanvasNode[T]) Center() (cx, cy float64) {
	return n.X + n.Width*0.5, n.Y + n.Height*0.5
}

// Intersects tests whether this node's AABB overlaps another AABB.
func (n *CanvasNode[T]) Intersects(other AABB) bool {
	return n.AABB().Intersects(other)
}

// SetPosition updates the node's canvas position.
func (n *CanvasNode[T]) SetPosition(x, y float64) {
	n.X = x
	n.Y = y
}

// SetSize updates the node's dimensions.
func (n *CanvasNode[T]) SetSize(w, h float64) {
	n.Width = w
	n.Height = h
}

// SetImage updates the visual image asset and 9-slice margins.
func (n *CanvasNode[T]) SetImage(img *ebiten.Image, top, right, bottom, left int) {
	n.Image = img
	n.NineSliceConfig = NewNineSliceFromImage(img, top, right, bottom, left)
	n.nineSlice = n.NineSliceConfig.Build()
}

// SetImageAuto updates the visual asset and automatically detects 9-slice margins.
func (n *CanvasNode[T]) SetImageAuto(img *ebiten.Image) {
	n.Image = img
	n.NineSliceConfig = NewNineSliceAuto(img)
	n.nineSlice = n.NineSliceConfig.Build()
}

// SetNineSliceConfig updates the NineSliceConfig and rebuilds the NineSlice.
func (n *CanvasNode[T]) SetNineSliceConfig(cfg NineSliceConfig) {
	n.NineSliceConfig = cfg
	n.Image = cfg.Image
	n.nineSlice = cfg.Build()
}

// GetNineSlice returns the compiled NineSlice or builds it if necessary.
func (n *CanvasNode[T]) GetNineSlice() *eimage.NineSlice {
	if n.nineSlice == nil && (n.Image != nil || n.NineSliceConfig.FillColor != nil) {
		if n.NineSliceConfig.Image == nil && n.Image != nil {
			n.NineSliceConfig = NewNineSliceAuto(n.Image)
		}
		n.nineSlice = n.NineSliceConfig.Build()
	}
	return n.nineSlice
}

// AddInbound registers one or more prerequisite (parent) node IDs.
func (n *CanvasNode[T]) AddInbound(ids ...string) {
	for _, id := range ids {
		if id != "" && !n.HasInbound(id) {
			n.InboundIDs = append(n.InboundIDs, id)
		}
	}
}

// HasInbound checks if the node has the specified prerequisite ID.
func (n *CanvasNode[T]) HasInbound(id string) bool {
	for _, in := range n.InboundIDs {
		if in == id {
			return true
		}
	}
	return false
}

// RemoveInbound removes a prerequisite node ID.
func (n *CanvasNode[T]) RemoveInbound(id string) {
	for i, in := range n.InboundIDs {
		if in == id {
			n.InboundIDs = append(n.InboundIDs[:i], n.InboundIDs[i+1:]...)
			return
		}
	}
}

// AddOutbound registers one or more unlocked (child) node IDs.
func (n *CanvasNode[T]) AddOutbound(ids ...string) {
	for _, id := range ids {
		if id != "" && !n.HasOutbound(id) {
			n.OutboundIDs = append(n.OutboundIDs, id)
		}
	}
}

// HasOutbound checks if the node has the specified child ID.
func (n *CanvasNode[T]) HasOutbound(id string) bool {
	for _, out := range n.OutboundIDs {
		if out == id {
			return true
		}
	}
	return false
}

// RemoveOutbound removes a child node ID.
func (n *CanvasNode[T]) RemoveOutbound(id string) {
	for i, out := range n.OutboundIDs {
		if out == id {
			n.OutboundIDs = append(n.OutboundIDs[:i], n.OutboundIDs[i+1:]...)
			return
		}
	}
}

// IsRoot returns true if this node has no prerequisites (inbound connections).
func (n *CanvasNode[T]) IsRoot() bool {
	return len(n.InboundIDs) == 0
}

// IsLeaf returns true if this node has no dependents (outbound connections).
func (n *CanvasNode[T]) IsLeaf() bool {
	return len(n.OutboundIDs) == 0
}

// OnClick registers a click event callback on this node.
func (n *CanvasNode[T]) OnClick(fn func(node *CanvasNode[T])) {
	n.onClick = fn
}

// OnHover registers a hover state change callback on this node.
func (n *CanvasNode[T]) OnHover(fn func(node *CanvasNode[T], hovered bool)) {
	n.onHover = fn
}

// TriggerHover programmatically invokes the onHover callback if registered.
func (n *CanvasNode[T]) TriggerHover(hovered bool) {
	if n.onHover != nil {
		n.onHover(n, hovered)
	}
}

// TriggerClick programmatically fires the onClick callback if registered.
func (n *CanvasNode[T]) TriggerClick() {
	if n.onClick != nil {
		n.onClick(n)
	}
}

// Draw renders this node onto the screen taking canvas pan and zoom into account.
func (n *CanvasNode[T]) Draw(screen *ebiten.Image, panX, panY, zoom float64) {
	if n.Hidden || n.Opacity <= 0.001 {
		return
	}

	screenX := n.X*zoom + panX
	screenY := n.Y*zoom + panY
	screenW := n.Width * zoom
	screenH := n.Height * zoom

	ns := n.GetNineSlice()
	if ns != nil && screenW > 0 && screenH > 0 {
		ns.Draw(screen, int(screenW), int(screenH), func(opts *ebiten.DrawImageOptions) {
			if n.Tint != nil {
				opts.ColorScale.ScaleWithColor(n.Tint)
			}
			if n.Opacity < 1.0 {
				opts.ColorScale.ScaleAlpha(float32(n.Opacity))
			}
			opts.GeoM.Translate(screenX, screenY)
		})
	}
}

// ToBanner bridges this CanvasNode into a SenjaUI Banner node.
// This allows embedding a CanvasNode directly into the hierarchical SenjaUI container tree.
func (n *CanvasNode[T]) ToBanner(opts ...BannerOption) *Banner {
	allOpts := []BannerOption{
		WithID(n.ID),
		WithPosition(n.X, n.Y),
		WithSize(n.Width, n.Height),
	}

	if n.nineSlice != nil {
		allOpts = append(allOpts, WithNineSlice(n.nineSlice))
	} else if n.Image != nil {
		allOpts = append(allOpts, WithNineSliceConfig(n.NineSliceConfig))
	}

	allOpts = append(allOpts, opts...)
	return NewBanner(allOpts...)
}
