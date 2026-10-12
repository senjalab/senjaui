package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/ebitenui/ebitenui"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/senjalab/senjaui"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Skill represents the custom domain payload for a skill tree node.
type Skill struct {
	Name        string
	Category    string
	Cost        int
	Description string
	Unlocked    bool
}

// SkillTreeGame manages the interactive skill tree canvas and HUD.
type SkillTreeGame struct {
	graph       *senjaui.CanvasGraph[Skill]
	ui          *ebitenui.UI
	hudBanner   *senjaui.Banner
	detailCard  *senjaui.Banner
	fontFace    text.Face
	skillPoints int
	statusMsg   string
	selected    *senjaui.CanvasNode[Skill]

	// Dragging / panning state
	isDragging bool
	dragStartX float64
	dragStartY float64
	camStartX  float64
	camStartY  float64
}

func resolveAssetPath(name string) string {
	candidates := []string{
		filepath.Join("assets", name),
		filepath.Join("..", "assets", name),
		filepath.Join("..", "..", "assets", name),
		name,
		filepath.Join("..", name),
		filepath.Join("..", "..", name),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join("assets", name)
}

func NewSkillTreeGame() *SkillTreeGame {
	g := &SkillTreeGame{
		fontFace:    text.Face(text.NewGoXFace(basicfont.Face7x13)),
		skillPoints: 5,
		statusMsg:   "Welcome to the Skill Tree! Click available (green) nodes to unlock.",
	}

	g.setupGraph()
	g.setupUI()
	return g
}

func (g *SkillTreeGame) setupGraph() {
	g.graph = senjaui.NewCanvasGraph[Skill]()
	g.graph.LineWidth = 3.0
	g.graph.LineColor = color.RGBA{70, 80, 105, 200}
	g.graph.ActiveLineColor = color.RGBA{46, 204, 113, 255}

	// Load 9-slice image asset
	assetPath := resolveAssetPath("ui_banner.png")
	nodeImg, nodeCfg, err := senjaui.LoadNineSliceImageFile(assetPath)
	if err != nil {
		log.Printf("Warning: failed to load %q (%v), generating procedural fallback", assetPath, err)
		nodeImg = senjaui.CreateProceduralNineSliceImage(32, 5, color.RGBA{22, 28, 42, 240}, color.RGBA{212, 175, 55, 255}, 3)
		nodeCfg = senjaui.NewNineSliceConfig(nodeImg, 6, 6, 6, 6)
	}

	createNode := func(id, name, cat string, cost int, desc string, x, y float64) *senjaui.CanvasNode[Skill] {
		data := Skill{
			Name:        name,
			Category:    cat,
			Cost:        cost,
			Description: desc,
			Unlocked:    false,
		}
		return senjaui.NewCanvasNode(
			id, x, y, 110, 65, data,
			senjaui.WithNodeImage[Skill](nodeImg, nodeCfg.Top, nodeCfg.Right, nodeCfg.Bottom, nodeCfg.Left),
		)
	}

	// Node Hierarchy (Skill Tree Layout)
	// Root Tier
	nRoot := createNode("iron_will", "Iron Will", "Passive", 0, "Base fortitude: +15% Max HP.", 10, 270)
	nRoot.Data.Unlocked = true // Starts unlocked

	// Tier 1
	nStrike := createNode("power_strike", "Power Strike", "Offense", 1, "Heavy melee blow dealing 150% dmg.", 280, 170)
	nShield := createNode("shield_block", "Shield Block", "Defense", 1, "Negates 40% incoming damage.", 280, 370)

	// Tier 2
	nWhirlwind := createNode("whirlwind", "Whirlwind", "Offense", 2, "360-degree blade tempest.", 460, 120)
	nCleave := createNode("cleave", "Cleave", "Offense", 2, "Attacks hit up to 3 adjacent foes.", 460, 220)
	nFortress := createNode("fortress", "Fortress", "Defense", 2, "Immunity to stuns and knockback.", 460, 370)

	// Tier 3 (Ultimate)
	nAvatar := createNode("titan_avatar", "Titan Avatar", "Ultimate", 3, "Unleashes primordial titan energy.", 640, 270)

	// Add all nodes
	g.graph.AddNode(nRoot).
		AddNode(nStrike).
		AddNode(nShield).
		AddNode(nWhirlwind).
		AddNode(nCleave).
		AddNode(nFortress).
		AddNode(nAvatar)

	// Connect directed prerequisites
	_ = g.graph.Connect("iron_will", "power_strike")
	_ = g.graph.Connect("iron_will", "shield_block")

	_ = g.graph.Connect("power_strike", "whirlwind")
	_ = g.graph.Connect("power_strike", "cleave")
	_ = g.graph.Connect("shield_block", "fortress")

	_ = g.graph.Connect("cleave", "titan_avatar")
	_ = g.graph.Connect("fortress", "titan_avatar")

	g.selected = nRoot
	g.refreshNodeVisuals()
}

func (g *SkillTreeGame) setupUI() {
	// Top HUD Banner
	g.hudBanner = senjaui.NewBanner(
		senjaui.WithID("hud-banner"),
		senjaui.WithPosition(float64(screenWidth)/2.0, 35),
		senjaui.WithSize(760, 50),
		senjaui.WithAnchor(senjaui.AnchorTopCenter),
		senjaui.WithPivot(senjaui.PivotTopCenter),
		senjaui.WithBorder(color.RGBA{18, 22, 34, 230}, color.RGBA{212, 175, 55, 255}, 2),
		senjaui.WithPaddingHV(16, 8),
	)

	// Bottom detail card
	g.detailCard = senjaui.NewBanner(
		senjaui.WithID("detail-card"),
		senjaui.WithPosition(float64(screenWidth)/2.0, float64(screenHeight)-20),
		senjaui.WithSize(760, 75),
		senjaui.WithAnchor(senjaui.AnchorBottomCenter),
		senjaui.WithPivot(senjaui.PivotBottomCenter),
		senjaui.WithBorder(color.RGBA{15, 20, 30, 240}, color.RGBA{100, 150, 220, 255}, 2),
		senjaui.WithPaddingHV(16, 8),
	)

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.BackgroundImage(senjaui.NewNineSliceColorConfig(color.Transparent).Build()),
	)
	rootContainer.AddChild(g.hudBanner.GetContainer())
	rootContainer.AddChild(g.detailCard.GetContainer())

	g.ui = &ebitenui.UI{Container: rootContainer}
}

// refreshNodeVisuals dynamically updates node tints based on unlock status and prerequisite availability.
func (g *SkillTreeGame) refreshNodeVisuals() {
	unlockedMap := make(map[string]bool)
	for _, node := range g.graph.Nodes() {
		if node.Data.Unlocked {
			unlockedMap[node.ID] = true
		}
	}

	for _, node := range g.graph.Nodes() {
		if node.Data.Unlocked {
			// Unlocked: Golden/Bronze Glow
			node.Tint = color.RGBA{255, 215, 80, 255}
			node.Opacity = 1.0
		} else if g.graph.PrerequisitesMet(node.ID, unlockedMap) {
			// Ready to unlock: Bright Emerald
			node.Tint = color.RGBA{46, 204, 113, 255}
			node.Opacity = 0.95
		} else {
			// Locked: Dimmed Steel
			node.Tint = color.RGBA{100, 110, 130, 200}
			node.Opacity = 0.65
		}
	}
}

func (g *SkillTreeGame) unlockNode(node *senjaui.CanvasNode[Skill]) {
	if node.Data.Unlocked {
		g.statusMsg = fmt.Sprintf("%s is already unlocked!", node.Data.Name)
		return
	}

	unlockedMap := make(map[string]bool)
	for _, n := range g.graph.Nodes() {
		if n.Data.Unlocked {
			unlockedMap[n.ID] = true
		}
	}

	if !g.graph.PrerequisitesMet(node.ID, unlockedMap) {
		g.statusMsg = fmt.Sprintf("Cannot unlock %s: prerequisites not satisfied!", node.Data.Name)
		return
	}

	if g.skillPoints < node.Data.Cost {
		g.statusMsg = fmt.Sprintf("Not enough Skill Points! Need %d, have %d.", node.Data.Cost, g.skillPoints)
		return
	}

	g.skillPoints -= node.Data.Cost
	node.Data.Unlocked = true
	g.statusMsg = fmt.Sprintf("Unlocked %s (-%d SP)!", node.Data.Name, node.Data.Cost)
	g.refreshNodeVisuals()
}

func (g *SkillTreeGame) Update() error {
	cx, cy := ebiten.CursorPosition()
	cursorX, cursorY := float64(cx), float64(cy)
	justClicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)

	// Pan navigation with Right-Click or Middle-Click drag
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		g.isDragging = true
		g.dragStartX = cursorX
		g.dragStartY = cursorY
		g.camStartX = g.graph.PanX
		g.camStartY = g.graph.PanY
	}
	if g.isDragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			g.graph.PanX = g.camStartX + (cursorX - g.dragStartX)
			g.graph.PanY = g.camStartY + (cursorY - g.dragStartY)
		} else {
			g.isDragging = false
		}
	}

	// Keyboard pan controls
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.graph.Pan(4, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.graph.Pan(-4, 0)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.graph.Pan(0, 4)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		g.graph.Pan(0, -4)
	}

	// Zoom controls
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.graph.SetZoom(g.graph.Zoom + 0.1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.graph.SetZoom(g.graph.Zoom - 0.1)
	}

	// Reset pan/zoom
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.graph.PanX = 0
		g.graph.PanY = 0
		g.graph.SetZoom(1.0)
	}

	// Add bonus skill points for testing
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.skillPoints += 3
		g.statusMsg = fmt.Sprintf("Added +3 SP! Current SP: %d", g.skillPoints)
		g.refreshNodeVisuals()
	}

	// CanvasGraph AABB hit testing & click dispatching
	clickedNode := g.graph.Update(cursorX, cursorY, justClicked)
	if clickedNode != nil {
		g.selected = clickedNode
		g.unlockNode(clickedNode)
	}

	g.ui.Update()
	return nil
}

func (g *SkillTreeGame) Draw(screen *ebiten.Image) {
	// Deep obsidian galaxy background
	screen.Fill(color.RGBA{10, 12, 18, 255})

	// 1. Draw CanvasGraph (Pure raster connections and nine-slice nodes, NO VECTORS!)
	g.graph.Draw(screen)

	// 2. Draw node labels centered on each node
	for _, node := range g.graph.Nodes() {
		if node.Hidden {
			continue
		}
		// Calculate screen center of node
		ncx, ncy := node.Center()
		scx, scy := g.graph.WorldToScreen(ncx, ncy)

		// Label title
		var titleCol color.Color = color.White
		if node.Data.Unlocked {
			titleCol = color.RGBA{255, 230, 120, 255}
		} else if node.Tint != nil {
			titleCol = color.RGBA{220, 220, 220, 255}
		}

		g.drawCenteredText(screen, node.Data.Name, scx, scy-8, titleCol)

		// Status / Cost tag
		costStr := fmt.Sprintf("%d SP", node.Data.Cost)
		if node.Data.Unlocked {
			costStr = "UNLOCKED"
		}
		g.drawCenteredText(screen, costStr, scx, scy+8, color.RGBA{180, 190, 210, 255})
	}

	// 3. Draw EbitenUI HUD banners
	g.ui.Draw(screen)

	// 4. Render HUD Text Overlays
	// Top Header
	g.drawText(screen, "SKILL PROGRESSION TREE", 40, 22, color.RGBA{255, 215, 0, 255})
	g.drawText(screen, fmt.Sprintf("Skill Points Available: %d  |  [P] +3 SP  [R] Reset View  [Drag] Pan", g.skillPoints), 40, 38, color.RGBA{200, 220, 255, 255})

	// Bottom Detail Card
	if g.selected != nil {
		skill := g.selected.Data
		statusLabel := "LOCKED"
		statusCol := color.RGBA{231, 76, 60, 255}
		if skill.Unlocked {
			statusLabel = "UNLOCKED"
			statusCol = color.RGBA{46, 204, 113, 255}
		}

		g.drawText(screen, fmt.Sprintf("SELECTED: %s (%s) - Cost: %d SP [%s]", skill.Name, skill.Category, skill.Cost, statusLabel), 40, float64(screenHeight)-60, statusCol)
		g.drawText(screen, fmt.Sprintf("Description: %s  |  %s", skill.Description, g.statusMsg), 40, float64(screenHeight)-42, color.RGBA{220, 230, 240, 255})
	}
}

func (g *SkillTreeGame) drawText(dst *ebiten.Image, msg string, x, y float64, col color.Color) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(col)
	text.Draw(dst, msg, g.fontFace, opts)
}

func (g *SkillTreeGame) drawCenteredText(dst *ebiten.Image, msg string, cx, cy float64, col color.Color) {
	w, h := text.Measure(msg, g.fontFace, 0)
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(cx-w*0.5, cy-h*0.5)
	opts.ColorScale.ScaleWithColor(col)
	text.Draw(dst, msg, g.fontFace, opts)
}

func (g *SkillTreeGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SenjaUI - Interactive Skill Tree with CanvasNode (Pure Raster)")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewSkillTreeGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Game error: %v", err)
	}
}
