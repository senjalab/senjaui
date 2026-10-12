package senjaui

import (
	"errors"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	// rasterPixel is a 1x1 white pixel used for drawing GPU-accelerated raster lines without vector packages.
	rasterPixel = func() *ebiten.Image {
		img := ebiten.NewImage(1, 1)
		img.Fill(color.White)
		return img
	}()

	// ErrNodeNotFound is returned when a referenced node ID is missing from the graph.
	ErrNodeNotFound = errors.New("senjaui: canvas node not found in graph")
	// ErrConnectionExists is returned when two nodes are already connected.
	ErrConnectionExists = errors.New("senjaui: connection already exists between nodes")
)

// CanvasGraph manages a directed graph of interactive CanvasNode instances.
// It handles camera panning, zooming, AABB hit-testing, prerequisite validation,
// and pure raster rendering of connections and nodes without any vector libraries.
type CanvasGraph[T any] struct {
	nodes map[string]*CanvasNode[T]
	order []*CanvasNode[T]

	// Camera & Canvas transform
	PanX float64
	PanY float64
	Zoom float64

	// Connection line styling (pure raster)
	LineColor         color.Color
	ActiveLineColor   color.Color
	LineWidth         float64
	ShowDirectionDots bool

	// Interactive state
	hoveredNode  *CanvasNode[T]
	selectedNode *CanvasNode[T]

	// Pre-allocated line draw options to ensure 0 allocations in Draw
	lineDrawOpts ebiten.DrawImageOptions
}

// NewCanvasGraph constructs an empty CanvasGraph ready for node registration.
func NewCanvasGraph[T any]() *CanvasGraph[T] {
	return &CanvasGraph[T]{
		nodes:             make(map[string]*CanvasNode[T]),
		order:             make([]*CanvasNode[T], 0),
		Zoom:              1.0,
		LineColor:         color.RGBA{100, 110, 130, 200},
		ActiveLineColor:   color.RGBA{46, 204, 113, 255},
		LineWidth:         2.0,
		ShowDirectionDots: true,
	}
}

// AddNode registers a CanvasNode into the graph.
func (g *CanvasGraph[T]) AddNode(node *CanvasNode[T]) *CanvasGraph[T] {
	if node == nil {
		return g
	}

	if _, exists := g.nodes[node.ID]; exists {
		// Update existing node reference in map
		g.nodes[node.ID] = node
		return g
	}

	g.nodes[node.ID] = node
	g.order = append(g.order, node)
	return g
}

// GetNode retrieves a node by its ID.
func (g *CanvasGraph[T]) GetNode(id string) (*CanvasNode[T], bool) {
	node, ok := g.nodes[id]
	return node, ok
}

// RemoveNode removes a node and cleans up all inbound and outbound edges referencing it.
func (g *CanvasGraph[T]) RemoveNode(id string) {
	node, exists := g.nodes[id]
	if !exists {
		return
	}

	// Remove from map
	delete(g.nodes, id)

	// Remove from order slice
	for i, n := range g.order {
		if n.ID == id {
			g.order = append(g.order[:i], g.order[i+1:]...)
			break
		}
	}

	// Clean up incoming references on other nodes
	for _, inID := range node.InboundIDs {
		if src, ok := g.nodes[inID]; ok {
			src.RemoveOutbound(id)
		}
	}

	// Clean up outgoing references on other nodes
	for _, outID := range node.OutboundIDs {
		if dst, ok := g.nodes[outID]; ok {
			dst.RemoveInbound(id)
		}
	}

	if g.hoveredNode == node {
		g.hoveredNode = nil
	}
	if g.selectedNode == node {
		g.selectedNode = nil
	}
}

// Connect creates a directed relationship from sourceID to targetID.
// sourceID adds targetID as Outbound; targetID adds sourceID as Inbound.
func (g *CanvasGraph[T]) Connect(sourceID, targetID string) error {
	src, ok1 := g.nodes[sourceID]
	if !ok1 {
		return ErrNodeNotFound
	}

	dst, ok2 := g.nodes[targetID]
	if !ok2 {
		return ErrNodeNotFound
	}

	if src.HasOutbound(targetID) {
		return ErrConnectionExists
	}

	src.AddOutbound(targetID)
	dst.AddInbound(sourceID)
	return nil
}

// Disconnect breaks the directed connection between sourceID and targetID.
func (g *CanvasGraph[T]) Disconnect(sourceID, targetID string) {
	if src, ok := g.nodes[sourceID]; ok {
		src.RemoveOutbound(targetID)
	}
	if dst, ok := g.nodes[targetID]; ok {
		dst.RemoveInbound(sourceID)
	}
}

// Nodes returns all registered nodes in insertion order.
func (g *CanvasGraph[T]) Nodes() []*CanvasNode[T] {
	return g.order
}

// Count returns the number of nodes in the graph.
func (g *CanvasGraph[T]) Count() int {
	return len(g.order)
}

// Clear removes all nodes and edges from the graph.
func (g *CanvasGraph[T]) Clear() {
	g.nodes = make(map[string]*CanvasNode[T])
	g.order = g.order[:0]
	g.hoveredNode = nil
	g.selectedNode = nil
}

// ScreenToWorld transforms screen pixel coordinates to world graph canvas coordinates.
func (g *CanvasGraph[T]) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	if g.Zoom == 0 {
		return screenX - g.PanX, screenY - g.PanY
	}
	worldX = (screenX - g.PanX) / g.Zoom
	worldY = (screenY - g.PanY) / g.Zoom
	return worldX, worldY
}

// WorldToScreen transforms world graph canvas coordinates to screen pixel coordinates.
func (g *CanvasGraph[T]) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	screenX = worldX*g.Zoom + g.PanX
	screenY = worldY*g.Zoom + g.PanY
	return screenX, screenY
}

// NodeAt performs an AABB hit test to find which node contains the screen coordinate.
// It iterates in reverse render order so top-most nodes are picked first.
func (g *CanvasGraph[T]) NodeAt(screenX, screenY float64) *CanvasNode[T] {
	worldX, worldY := g.ScreenToWorld(screenX, screenY)

	for i := len(g.order) - 1; i >= 0; i-- {
		n := g.order[i]
		if !n.Hidden && n.Contains(worldX, worldY) {
			return n
		}
	}
	return nil
}

// PrerequisitesMet checks whether all inbound prerequisite nodes of nodeID are present in unlockedIDs.
// Root nodes (nodes with 0 inbound IDs) automatically return true.
func (g *CanvasGraph[T]) PrerequisitesMet(nodeID string, unlockedIDs map[string]bool) bool {
	node, ok := g.nodes[nodeID]
	if !ok {
		return false
	}

	if node.IsRoot() {
		return true
	}

	for _, preID := range node.InboundIDs {
		if !unlockedIDs[preID] {
			return false
		}
	}
	return true
}

// GetPrerequisites returns the list of prerequisite node pointers for a node.
func (g *CanvasGraph[T]) GetPrerequisites(nodeID string) []*CanvasNode[T] {
	node, ok := g.nodes[nodeID]
	if !ok {
		return nil
	}

	result := make([]*CanvasNode[T], 0, len(node.InboundIDs))
	for _, inID := range node.InboundIDs {
		if pre, exists := g.nodes[inID]; exists {
			result = append(result, pre)
		}
	}
	return result
}

// GetDependents returns the list of dependent (child) node pointers for a node.
func (g *CanvasGraph[T]) GetDependents(nodeID string) []*CanvasNode[T] {
	node, ok := g.nodes[nodeID]
	if !ok {
		return nil
	}

	result := make([]*CanvasNode[T], 0, len(node.OutboundIDs))
	for _, outID := range node.OutboundIDs {
		if child, exists := g.nodes[outID]; exists {
			result = append(result, child)
		}
	}
	return result
}

// HoveredNode returns the currently hovered node, if any.
func (g *CanvasGraph[T]) HoveredNode() *CanvasNode[T] {
	return g.hoveredNode
}

// SelectedNode returns the currently selected node, if any.
func (g *CanvasGraph[T]) SelectedNode() *CanvasNode[T] {
	return g.selectedNode
}

// SelectNode sets the selected node by ID.
func (g *CanvasGraph[T]) SelectNode(id string) *CanvasNode[T] {
	if n, ok := g.nodes[id]; ok {
		g.selectedNode = n
		return n
	}
	g.selectedNode = nil
	return nil
}

// Update processes cursor hovering and click hit testing.
// Returns the clicked CanvasNode if a node was clicked on this frame.
func (g *CanvasGraph[T]) Update(cursorX, cursorY float64, clicked bool) *CanvasNode[T] {
	hit := g.NodeAt(cursorX, cursorY)

	// Update hover states
	if hit != g.hoveredNode {
		if g.hoveredNode != nil {
			g.hoveredNode.TriggerHover(false)
		}
		g.hoveredNode = hit
		if g.hoveredNode != nil {
			g.hoveredNode.TriggerHover(true)
		}
	}

	// Process click
	if clicked && hit != nil {
		g.selectedNode = hit
		hit.TriggerClick()
		return hit
	}

	return nil
}

// Pan shifts the camera canvas position by (dx, dy).
func (g *CanvasGraph[T]) Pan(dx, dy float64) {
	g.PanX += dx
	g.PanY += dy
}

// SetZoom updates the zoom factor (clamped to [0.1, 5.0]).
func (g *CanvasGraph[T]) SetZoom(zoom float64) {
	g.Zoom = math.Max(0.1, math.Min(zoom, 5.0))
}

// Draw renders the entire graph: first raster connection lines, then nine-slice nodes.
// Zero vector libraries are used; all geometry is rendered via hardware rasterization.
func (g *CanvasGraph[T]) Draw(screen *ebiten.Image) {
	// 1. Draw connection lines between nodes
	g.drawConnections(screen)

	// 2. Draw nodes in insertion order
	for _, node := range g.order {
		node.Draw(screen, g.PanX, g.PanY, g.Zoom)
	}
}

func (g *CanvasGraph[T]) drawConnections(screen *ebiten.Image) {
	lineWidth := g.LineWidth * g.Zoom
	if lineWidth < 1.0 {
		lineWidth = 1.0
	}

	for _, src := range g.order {
		if src.Hidden {
			continue
		}

		srcCX, srcCY := src.Center()
		sx1, sy1 := g.WorldToScreen(srcCX, srcCY)

		for _, outID := range src.OutboundIDs {
			dst, ok := g.nodes[outID]
			if !ok || dst.Hidden {
				continue
			}

			dstCX, dstCY := dst.Center()
			sx2, sy2 := g.WorldToScreen(dstCX, dstCY)

			lineCol := g.LineColor
			g.drawRasterLine(screen, sx1, sy1, sx2, sy2, lineWidth, lineCol)

			// Draw small directional marker at 65% of connection line
			if g.ShowDirectionDots {
				dotX := sx1 + (sx2-sx1)*0.65
				dotY := sy1 + (sy2-sy1)*0.65
				dotSize := lineWidth * 2.5
				g.drawRasterDot(screen, dotX, dotY, dotSize, lineCol)
			}
		}
	}
}

func (g *CanvasGraph[T]) drawRasterLine(screen *ebiten.Image, x1, y1, x2, y2, width float64, col color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	dist := math.Hypot(dx, dy)
	if dist <= 0.001 {
		return
	}

	angle := math.Atan2(dy, dx)

	g.lineDrawOpts.GeoM.Reset()
	g.lineDrawOpts.ColorScale.Reset()

	g.lineDrawOpts.GeoM.Scale(dist, width)
	g.lineDrawOpts.GeoM.Translate(0, -width*0.5)
	g.lineDrawOpts.GeoM.Rotate(angle)
	g.lineDrawOpts.GeoM.Translate(x1, y1)

	g.lineDrawOpts.ColorScale.ScaleWithColor(col)
	screen.DrawImage(rasterPixel, &g.lineDrawOpts)
}

func (g *CanvasGraph[T]) drawRasterDot(screen *ebiten.Image, cx, cy, size float64, col color.Color) {
	g.lineDrawOpts.GeoM.Reset()
	g.lineDrawOpts.ColorScale.Reset()

	g.lineDrawOpts.GeoM.Scale(size, size)
	g.lineDrawOpts.GeoM.Translate(cx-size*0.5, cy-size*0.5)
	g.lineDrawOpts.ColorScale.ScaleWithColor(col)
	screen.DrawImage(rasterPixel, &g.lineDrawOpts)
}
