package senjaui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// SkillData is a sample custom domain payload modeling a game skill tree.
type SkillData struct {
	Name        string
	Tier        int
	Cost        int
	Unlocked    bool
	Description string
}

func TestCanvasNode_AABBAndSpatialQueries(t *testing.T) {
	node := NewCanvasNode(
		"fireball",
		100, 150, 80, 60,
		SkillData{Name: "Fireball", Tier: 1, Cost: 50},
	)

	// Test AABB coordinates
	aabb := node.AABB()
	if aabb.MinX != 100 || aabb.MinY != 150 || aabb.MaxX != 180 || aabb.MaxY != 210 {
		t.Fatalf("unexpected AABB extents: %+v", aabb)
	}

	if aabb.Width() != 80 || aabb.Height() != 60 {
		t.Fatalf("unexpected AABB dimensions: %vx%v", aabb.Width(), aabb.Height())
	}

	cx, cy := aabb.Center()
	if cx != 140 || cy != 180 {
		t.Fatalf("unexpected AABB center: (%v, %v)", cx, cy)
	}

	// Test Contains (inclusive)
	if !node.Contains(100, 150) || !node.Contains(180, 210) || !node.Contains(140, 180) {
		t.Fatalf("expected node to contain inside points")
	}
	if node.Contains(99, 150) || node.Contains(100, 149) || node.Contains(181, 210) {
		t.Fatalf("expected node not to contain outside points")
	}

	// Test Intersects
	overlapping := NewAABB(120, 170, 200, 250)
	if !node.Intersects(overlapping) {
		t.Fatalf("expected overlapping AABB to intersect")
	}

	disjoint := NewAABB(200, 300, 250, 350)
	if node.Intersects(disjoint) {
		t.Fatalf("expected disjoint AABB not to intersect")
	}
}

func TestCanvasNode_NineSliceIntegration(t *testing.T) {
	img, _ := MustLoadNineSliceImageFile("assets/ui_banner.png")

	node := NewCanvasNode(
		"strike",
		50, 50, 120, 80,
		SkillData{Name: "Strike", Tier: 1},
		WithNodeImage[SkillData](img, 6, 6, 6, 6),
		WithNodeTint[SkillData](color.RGBA{255, 200, 100, 255}),
	)

	if node.Image != img {
		t.Fatalf("expected node.Image to match provided img")
	}

	ns := node.GetNineSlice()
	if ns == nil {
		t.Fatalf("expected GetNineSlice to return valid NineSlice")
	}

	// Test ToBanner bridge
	banner := node.ToBanner()
	if banner == nil {
		t.Fatalf("expected ToBanner to return non-nil Banner")
	}
	if banner.ID() != "strike" {
		t.Fatalf("expected Banner ID to match node ID, got %q", banner.ID())
	}
	if banner.Width != 120 || banner.Height != 80 {
		t.Fatalf("expected Banner dimensions 120x80, got %vx%v", banner.Width, banner.Height)
	}
}

func TestCanvasNode_GraphRelationships(t *testing.T) {
	root := NewCanvasNode("root", 0, 0, 50, 50, "Root Skill")
	tier2A := NewCanvasNode("tier2A", 100, 0, 50, 50, "Fire Stream")
	tier2B := NewCanvasNode("tier2B", 100, 100, 50, 50, "Ice Shard")
	tier3 := NewCanvasNode("tier3", 200, 50, 50, 50, "Elemental Storm")

	if !root.IsRoot() {
		t.Fatalf("expected root node to be root")
	}

	// Build edges
	root.AddOutbound("tier2A", "tier2B")
	tier2A.AddInbound("root")
	tier2B.AddInbound("root")

	tier2A.AddOutbound("tier3")
	tier2B.AddOutbound("tier3")
	tier3.AddInbound("tier2A", "tier2B")

	if root.IsLeaf() || !tier3.IsLeaf() {
		t.Fatalf("unexpected leaf classification")
	}

	if len(root.OutboundIDs) != 2 || len(tier3.InboundIDs) != 2 {
		t.Fatalf("unexpected edge counts")
	}

	// Test removal
	tier3.RemoveInbound("tier2A")
	if len(tier3.InboundIDs) != 1 || tier3.HasInbound("tier2A") {
		t.Fatalf("expected tier2A to be removed from tier3 inbounds")
	}
}

func TestCanvasGraph_ProgressionAndSkillTree(t *testing.T) {
	graph := NewCanvasGraph[SkillData]()

	nStrike := NewCanvasNode("strike", 50, 100, 70, 50, SkillData{Name: "Strike", Cost: 0})
	nSlash := NewCanvasNode("slash", 160, 60, 70, 50, SkillData{Name: "Slash", Cost: 100})
	nBash := NewCanvasNode("bash", 160, 140, 70, 50, SkillData{Name: "Bash", Cost: 100})
	nWhirl := NewCanvasNode("whirlwind", 280, 100, 70, 50, SkillData{Name: "Whirlwind", Cost: 250})

	graph.AddNode(nStrike).
		AddNode(nSlash).
		AddNode(nBash).
		AddNode(nWhirl)

	// Connect nodes in directed graph
	if err := graph.Connect("strike", "slash"); err != nil {
		t.Fatalf("failed to connect strike -> slash: %v", err)
	}
	if err := graph.Connect("strike", "bash"); err != nil {
		t.Fatalf("failed to connect strike -> bash: %v", err)
	}
	if err := graph.Connect("slash", "whirlwind"); err != nil {
		t.Fatalf("failed to connect slash -> whirlwind: %v", err)
	}
	if err := graph.Connect("bash", "whirlwind"); err != nil {
		t.Fatalf("failed to connect bash -> whirlwind: %v", err)
	}

	// Check graph counts
	if graph.Count() != 4 {
		t.Fatalf("expected 4 nodes, got %d", graph.Count())
	}

	// Check prerequisites progression
	unlocked := map[string]bool{}

	// Strike is root -> prerequisites met initially
	if !graph.PrerequisitesMet("strike", unlocked) {
		t.Fatalf("expected strike prerequisites to be met (root)")
	}
	// Slash & Bash require Strike
	if graph.PrerequisitesMet("slash", unlocked) {
		t.Fatalf("expected slash prerequisites not to be met before unlocking strike")
	}

	// Unlock Strike
	unlocked["strike"] = true
	if !graph.PrerequisitesMet("slash", unlocked) || !graph.PrerequisitesMet("bash", unlocked) {
		t.Fatalf("expected slash and bash prerequisites to be met after unlocking strike")
	}
	if graph.PrerequisitesMet("whirlwind", unlocked) {
		t.Fatalf("expected whirlwind prerequisites not to be met before slash and bash")
	}

	// Unlock Slash and Bash -> now Whirlwind is ready
	unlocked["slash"] = true
	unlocked["bash"] = true
	if !graph.PrerequisitesMet("whirlwind", unlocked) {
		t.Fatalf("expected whirlwind prerequisites to be met after unlocking prerequisites")
	}
}

func TestCanvasGraph_SpatialHitTestingAndInteraction(t *testing.T) {
	graph := NewCanvasGraph[string]()
	nodeA := NewCanvasNode("nodeA", 100, 100, 80, 60, "Data A")
	nodeB := NewCanvasNode("nodeB", 250, 100, 80, 60, "Data B")

	var clickedID string
	nodeA.OnClick(func(n *CanvasNode[string]) {
		clickedID = n.ID
	})

	var hoveredID string
	var hoverState bool
	nodeA.OnHover(func(n *CanvasNode[string], hovered bool) {
		hoveredID = n.ID
		hoverState = hovered
	})

	graph.AddNode(nodeA).AddNode(nodeB)

	// Hit test without pan/zoom
	hit := graph.NodeAt(120, 120)
	if hit != nodeA {
		t.Fatalf("expected hit on nodeA, got %v", hit)
	}

	hitMiss := graph.NodeAt(50, 50)
	if hitMiss != nil {
		t.Fatalf("expected miss, got %v", hitMiss)
	}

	// Test interaction Update (hover)
	graph.Update(120, 120, false)
	if hoveredID != "nodeA" || !hoverState {
		t.Fatalf("expected nodeA to receive hover enter")
	}
	if graph.HoveredNode() != nodeA {
		t.Fatalf("expected graph.HoveredNode() to be nodeA")
	}

	// Hover out
	graph.Update(10, 10, false)
	if hoveredID != "nodeA" || hoverState {
		t.Fatalf("expected nodeA to receive hover exit")
	}

	// Test click
	clickedNode := graph.Update(120, 120, true)
	if clickedNode != nodeA || clickedID != "nodeA" {
		t.Fatalf("expected nodeA to be clicked, got clickedID=%q", clickedID)
	}
	if graph.SelectedNode() != nodeA {
		t.Fatalf("expected selectedNode to be nodeA")
	}

	// Test camera Pan & Zoom
	graph.Pan(50, 30) // Pan canvas by +50, +30
	graph.SetZoom(2.0)

	// With zoom=2.0 and pan=(50, 30):
	// NodeA X=100 -> ScreenX = 100 * 2.0 + 50 = 250
	// NodeA Y=100 -> ScreenY = 100 * 2.0 + 30 = 230
	hitZoomed := graph.NodeAt(255, 235)
	if hitZoomed != nodeA {
		t.Fatalf("expected hit on nodeA with camera pan & zoom, got %v", hitZoomed)
	}
}

func TestCanvasGraph_RasterDraw(t *testing.T) {
	img, _ := MustLoadNineSliceImageFile("assets/ui_banner.png")

	graph := NewCanvasGraph[int]()
	node1 := NewCanvasNode("1", 50, 50, 60, 40, 1, WithNodeImage[int](img, 6, 6, 6, 6))
	node2 := NewCanvasNode("2", 180, 100, 60, 40, 2, WithNodeImage[int](img, 6, 6, 6, 6))
	graph.AddNode(node1).AddNode(node2)
	graph.Connect("1", "2")

	screen := ebiten.NewImage(400, 300)

	// Verify Draw executes cleanly without panic
	graph.Draw(screen)
}
