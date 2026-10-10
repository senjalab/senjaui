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

// Skin represents a raster 9-slice image preset (no vectors).
type Skin struct {
	Name   string
	Image  *ebiten.Image
	Top    int
	Right  int
	Bottom int
	Left   int
}

// Game implements ebiten.Game and holds interactive UI state.
type Game struct {
	ui           *ebitenui.UI
	banner       *senjaui.Banner
	cardAccept   *senjaui.Banner
	cardDismiss  *senjaui.Banner
	fontFace     text.Face
	statusText   string
	animStyle    string
	sizeMode     int
	lastClickLog string

	// Dynamic skins array showcasing runtime image swapping
	skins       []Skin
	currentSkin int
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

func NewGame() *Game {
	g := &Game{
		fontFace:     text.Face(text.NewGoXFace(basicfont.Face7x13)),
		statusText:   "Ready: Using assets/ui_banner.png. Press [C] or button to switch images dynamically.",
		animStyle:    "Scale & Fade",
		lastClickLog: "No events yet",
	}

	g.loadSkins()
	g.setupUI()
	return g
}

func (g *Game) loadSkins() {
	// 1. Primary skin: Load user's provided assets/ui_banner.png asset
	assetPath := resolveAssetPath("ui_banner.png")
	userImg, userCfg, err := senjaui.LoadNineSliceImageFile(assetPath)
	if err != nil {
		log.Printf("Warning: failed to load %q (%v), generating fallback raster banner", assetPath, err)
		userImg = senjaui.CreateProceduralNineSliceImage(32, 6, color.RGBA{18, 22, 32, 240}, color.RGBA{212, 175, 55, 255}, 3)
		userCfg = senjaui.NewNineSliceConfig(userImg, 6, 6, 6, 6)
	}

	g.skins = []Skin{
		{
			Name:   "assets/ui_banner.png (Original Asset)",
			Image:  userImg,
			Top:    userCfg.Top,
			Right:  userCfg.Right,
			Bottom: userCfg.Bottom,
			Left:   userCfg.Left,
		},
		{
			Name:   "Cyber Neon (Raster 32x32)",
			Image:  senjaui.CreateProceduralNineSliceImage(32, 4, color.RGBA{14, 24, 40, 245}, color.RGBA{0, 220, 255, 255}, 3),
			Top:    6,
			Right:  6,
			Bottom: 6,
			Left:   6,
		},
		{
			Name:   "Ancient Gold (Raster 32x32)",
			Image:  senjaui.CreateProceduralNineSliceImage(32, 5, color.RGBA{35, 25, 15, 245}, color.RGBA{240, 180, 40, 255}, 3),
			Top:    6,
			Right:  6,
			Bottom: 6,
			Left:   6,
		},
		{
			Name:   "Molten Obsidian (Raster 32x32)",
			Image:  senjaui.CreateProceduralNineSliceImage(32, 4, color.RGBA{38, 14, 18, 245}, color.RGBA{255, 75, 60, 255}, 3),
			Top:    6,
			Right:  6,
			Bottom: 6,
			Left:   6,
		},
	}
}

func (g *Game) setupUI() {
	activeSkin := g.skins[g.currentSkin]

	// 1. Build Parent Notification Banner using active skin (ui_banner.png)
	g.banner = senjaui.NewBanner(
		senjaui.WithID("parent-notification-banner"),
		senjaui.WithPosition(float64(screenWidth)/2.0, float64(screenHeight)/2.0),
		senjaui.WithSize(560, 330),
		senjaui.WithAnchor(senjaui.AnchorCenter),
		senjaui.WithPivot(senjaui.PivotCenter),
		senjaui.WithPadding(24, 24, 24, 24),
		senjaui.WithImage(activeSkin.Image, activeSkin.Top, activeSkin.Right, activeSkin.Bottom, activeSkin.Left),
		senjaui.WithAnimation(senjaui.AnimScaleAndFade, 0.35),
		senjaui.WithEase(senjaui.EaseOutBack),
	)

	// Event Callbacks on Parent Banner
	g.banner.OnOpen(func() {
		g.statusText = "Event: Parent Notification Banner opened!"
	})

	g.banner.OnClose(func() {
		g.statusText = "Event: Parent Notification Banner closed!"
	})

	// Hierarchical event bubbling: Any click on nested cards routes here!
	g.banner.OnContentClick(func(childID string) {
		g.lastClickLog = fmt.Sprintf("Parent Banner caught click from: %q", childID)
		g.statusText = fmt.Sprintf("Action handled: %s", childID)
	})

	// Header container with title, description, and interactive skin swap button
	headerContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(6),
		)),
	)

	titleLabel := senjaui.CreateLabel("[QUEST AVAILABLE] The Ruins of Senja", &g.fontFace, color.RGBA{255, 220, 100, 255})
	descLabel := senjaui.CreateLabel("Ancient beacon detected. Sliced from assets/ui_banner.png without any vectors.", &g.fontFace, color.RGBA{210, 220, 235, 255})

	btnSwapSkin := senjaui.CreateStyledButton(
		"SWAP BANNER IMAGE (DYNAMIC 9-SLICE)",
		&g.fontFace,
		color.RGBA{45, 55, 80, 255},
		color.RGBA{65, 80, 115, 255},
		color.RGBA{30, 40, 60, 255},
		color.RGBA{255, 230, 120, 255},
		func() {
			g.cycleSkin()
		},
	)

	headerContainer.AddChild(titleLabel)
	headerContainer.AddChild(descLabel)
	headerContainer.AddChild(btnSwapSkin)
	g.banner.AddWidget(headerContainer)

	// 2. Build Nested Card 1: Quest Acceptance Card (Raster 9-slice, bottom-left)
	cardAcceptImg := senjaui.CreateProceduralNineSliceImage(
		24, 4,
		color.RGBA{15, 35, 25, 245},
		color.RGBA{46, 204, 113, 255},
		2,
	)
	g.cardAccept = senjaui.NewBanner(
		senjaui.WithID("card-accept"),
		senjaui.WithPosition(24, -20),
		senjaui.WithSize(230, 160),
		senjaui.WithAnchor(senjaui.AnchorBottomLeft),
		senjaui.WithPivot(senjaui.PivotBottomLeft),
		senjaui.WithPaddingAll(12),
		senjaui.WithImage(cardAcceptImg, 6, 6, 6, 6),
		senjaui.WithContainerLayout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
	)

	card1Title := senjaui.CreateLabel("MISSION ACCEPTANCE", &g.fontFace, color.RGBA{46, 204, 113, 255})
	card1Desc := senjaui.CreateLabel("Deploy vanguard squad.\nRewards: +850 XP\nBonus: Rare Relic", &g.fontFace, color.RGBA{220, 240, 230, 255})
	btnAccept := senjaui.CreateStyledButton(
		"ACCEPT QUEST",
		&g.fontFace,
		color.RGBA{39, 174, 96, 255},
		color.RGBA{46, 204, 113, 255},
		color.RGBA{30, 132, 73, 255},
		color.White,
		func() {
			g.cardAccept.TriggerContentClick("card-accept:btn-accept")
		},
	)
	g.cardAccept.AddWidget(card1Title)
	g.cardAccept.AddWidget(card1Desc)
	g.cardAccept.AddWidget(btnAccept)

	// 3. Build Nested Card 2: Dismiss Card (Raster 9-slice, bottom-right)
	cardDismissImg := senjaui.CreateProceduralNineSliceImage(
		24, 4,
		color.RGBA{38, 18, 22, 245},
		color.RGBA{231, 76, 60, 255},
		2,
	)
	g.cardDismiss = senjaui.NewBanner(
		senjaui.WithID("card-dismiss"),
		senjaui.WithPosition(-24, -20),
		senjaui.WithSize(230, 160),
		senjaui.WithAnchor(senjaui.AnchorBottomRight),
		senjaui.WithPivot(senjaui.PivotBottomRight),
		senjaui.WithPaddingAll(12),
		senjaui.WithImage(cardDismissImg, 6, 6, 6, 6),
		senjaui.WithContainerLayout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
	)

	card2Title := senjaui.CreateLabel("TACTICAL RETREAT", &g.fontFace, color.RGBA{231, 76, 60, 255})
	card2Desc := senjaui.CreateLabel("Hold perimeter line.\nRisk: None\nCost: 0 Resources", &g.fontFace, color.RGBA{240, 220, 220, 255})
	btnDismiss := senjaui.CreateStyledButton(
		"DECLINE MISSION",
		&g.fontFace,
		color.RGBA{192, 57, 43, 255},
		color.RGBA{231, 76, 60, 255},
		color.RGBA{146, 43, 33, 255},
		color.White,
		func() {
			g.cardDismiss.TriggerContentClick("card-dismiss:btn-decline")
		},
	)
	g.cardDismiss.AddWidget(card2Title)
	g.cardDismiss.AddWidget(card2Desc)
	g.cardDismiss.AddWidget(btnDismiss)

	// 4. Establish hierarchy: Add both cards as children to Parent Banner!
	g.banner.AddChild(g.cardAccept)
	g.banner.AddChild(g.cardDismiss)

	// 5. Connect to EbitenUI canvas
	g.ui = senjaui.NewUI(g.banner)
}

// cycleSkin demonstrates dynamically changing the 9-slice image at runtime!
func (g *Game) cycleSkin() {
	g.currentSkin = (g.currentSkin + 1) % len(g.skins)
	s := g.skins[g.currentSkin]

	// Dynamically update the banner's nine-slice image:
	// Child relative transforms, padding, and layout remain perfectly preserved!
	g.banner.SetImage(s.Image, s.Top, s.Right, s.Bottom, s.Left)

	g.statusText = fmt.Sprintf("Dynamic Image Swap -> Switched to: %s", s.Name)
	g.banner.TriggerContentClick("action:swap-skin")
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0

	// 1. Advance UI animations (0 heap allocations)
	g.banner.Update(dt)

	// 2. Update EbitenUI input and events
	g.ui.Update()

	// 3. Interactive keyboard controls:
	// Space: Toggle banner open/close
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.banner.Toggle()
	}

	// C or S: Dynamically swap the nine-slice image at runtime!
	if inpututil.IsKeyJustPressed(ebiten.KeyC) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.cycleSkin()
	}

	// 1: Scale-In
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.banner.SetAnimation(senjaui.AnimScaleIn, 0.35, senjaui.EaseOutBack)
		g.animStyle = "Scale-In (EaseOutBack)"
		g.banner.Open()
	}
	// 2: Fade-In
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.banner.SetAnimation(senjaui.AnimFadeIn, 0.3, senjaui.EaseLinear)
		g.animStyle = "Fade-In (EaseLinear)"
		g.banner.Open()
	}
	// 3: Scale & Fade
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.banner.SetAnimation(senjaui.AnimScaleAndFade, 0.35, senjaui.EaseOutQuad)
		g.animStyle = "Scale & Fade (EaseOutQuad)"
		g.banner.Open()
	}
	// 4: Slide-In from Top
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.banner.SetAnimation(senjaui.AnimSlideInTop, 0.4, senjaui.EaseOutBounce)
		g.animStyle = "Slide-In Top (EaseOutBounce)"
		g.banner.Open()
	}

	// R: Cycle parent banner size to demonstrate automatic child transform propagation
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.sizeMode = (g.sizeMode + 1) % 3
		switch g.sizeMode {
		case 0:
			g.banner.SetSize(560, 330)
		case 1:
			g.banner.SetSize(620, 360)
		case 2:
			g.banner.SetSize(520, 310)
		}
		g.statusText = fmt.Sprintf("Parent resized to: %.0fx%.0f (children automatically realigned!)", g.banner.Width, g.banner.Height)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Dark background
	screen.Fill(color.RGBA{12, 14, 20, 255})

	// Render EbitenUI canvas (Parent banner & nested cards)
	g.ui.Draw(screen)

	// Render HUD information overlay
	hudY := 20.0
	g.drawText(screen, "=== SENJAUI DYNAMIC NINE-SLICE BANNER (NO VECTORS) ===", 20, hudY, color.RGBA{255, 215, 0, 255})
	hudY += 20
	currentSkinName := g.skins[g.currentSkin].Name
	g.drawText(screen, fmt.Sprintf("Active Texture: %s | Press [C] or button to Swap Image", currentSkinName), 20, hudY, color.RGBA{100, 255, 200, 255})
	hudY += 20
	g.drawText(screen, fmt.Sprintf("Controls: [Space] Open/Close | [C] Swap Image | [1-4] Anim: %s | [R] Resize", g.animStyle), 20, hudY, color.RGBA{180, 190, 210, 255})
	hudY += 20
	g.drawText(screen, g.statusText, 20, hudY, color.RGBA{100, 220, 255, 255})
	hudY += 20
	g.drawText(screen, fmt.Sprintf("Event Bubbling Log: %s", g.lastClickLog), 20, hudY, color.RGBA{255, 180, 100, 255})

	// Render card coordinate debug info at bottom
	if g.banner.IsOpen() {
		bx, by, bw, bh := g.banner.GetWorldBounds()
		ax, ay, aw, ah := g.cardAccept.GetWorldBounds()
		dx, dy, dw, dh := g.cardDismiss.GetWorldBounds()
		botY := float64(screenHeight - 35)
		g.drawText(screen, fmt.Sprintf("Parent: (%.0f, %.0f, %.0f, %.0f) | CardAccept: (%.0f, %.0f, %.0f, %.0f) | CardDismiss: (%.0f, %.0f, %.0f, %.0f)",
			bx, by, bw, bh, ax, ay, aw, ah, dx, dy, dw, dh), 20, botY, color.RGBA{150, 150, 170, 255})
	}
}

func (g *Game) drawText(dst *ebiten.Image, msg string, x, y float64, col color.Color) {
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(col)
	text.Draw(dst, msg, g.fontFace, opts)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("SenjaUI - Dynamic Nine-Slice Banner & Card System")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Game loop error: %v", err)
	}
}
