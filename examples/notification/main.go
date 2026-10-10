package main

import (
	"fmt"
	"image/color"
	"log"

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

// Game implements ebiten.Game and holds UI state.
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
}

func NewGame() *Game {
	g := &Game{
		fontFace:     text.Face(text.NewGoXFace(basicfont.Face7x13)),
		statusText:   "Status: Notification Banner Open. Click either card or button below.",
		animStyle:    "Scale & Fade",
		lastClickLog: "No events yet",
	}

	g.setupUI()
	return g
}

func (g *Game) setupUI() {
	// 1. Procedural 9-slice backgrounds for game aesthetic
	bannerImg := senjaui.CreateProceduralNineSliceImage(
		32, 6,
		color.RGBA{18, 22, 32, 240}, // Deep midnight blue background
		color.RGBA{212, 175, 55, 255}, // Elegant gold border
		3,
	)
	bannerCfg := senjaui.NewNineSliceConfig(bannerImg, 8, 8, 8, 8)

	cardAcceptImg := senjaui.CreateProceduralNineSliceImage(
		24, 4,
		color.RGBA{15, 35, 25, 245}, // Dark emerald fill
		color.RGBA{46, 204, 113, 255}, // Bright emerald border
		2,
	)
	cardAcceptCfg := senjaui.NewNineSliceConfig(cardAcceptImg, 6, 6, 6, 6)

	cardDismissImg := senjaui.CreateProceduralNineSliceImage(
		24, 4,
		color.RGBA{38, 18, 22, 245}, // Dark crimson fill
		color.RGBA{231, 76, 60, 255}, // Bright coral red border
		2,
	)
	cardDismissCfg := senjaui.NewNineSliceConfig(cardDismissImg, 6, 6, 6, 6)

	// 2. Build Parent Notification Banner
	// Centered on screen using AnchorCenter and PivotCenter
	g.banner = senjaui.NewBanner(
		senjaui.WithID("parent-notification-banner"),
		senjaui.WithPosition(float64(screenWidth)/2.0, float64(screenHeight)/2.0),
		senjaui.WithSize(560, 320),
		senjaui.WithAnchor(senjaui.AnchorCenter),
		senjaui.WithPivot(senjaui.PivotCenter),
		senjaui.WithPadding(24, 24, 24, 24),
		senjaui.WithNineSliceConfig(bannerCfg),
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
		g.lastClickLog = fmt.Sprintf("Parent Banner received click from childID: %q", childID)
		g.statusText = fmt.Sprintf("Action handled: %s", childID)
	})

	// Add banner header text widgets
	headerContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(6),
		)),
	)

	titleLabel := senjaui.CreateLabel("[QUEST AVAILABLE] The Ruins of Senja", &g.fontFace, color.RGBA{255, 220, 100, 255})
	descLabel := senjaui.CreateLabel("An ancient beacon was detected. Choose your team's deployment:", &g.fontFace, color.RGBA{200, 210, 230, 255})
	headerContainer.AddChild(titleLabel)
	headerContainer.AddChild(descLabel)
	g.banner.AddWidget(headerContainer)

	// 3. Build Nested Card 1: Quest Acceptance Card (Anchored bottom-left)
	g.cardAccept = senjaui.NewBanner(
		senjaui.WithID("card-accept"),
		senjaui.WithPosition(24, -20),
		senjaui.WithSize(230, 170),
		senjaui.WithAnchor(senjaui.AnchorBottomLeft),
		senjaui.WithPivot(senjaui.PivotBottomLeft),
		senjaui.WithPaddingAll(12),
		senjaui.WithNineSliceConfig(cardAcceptCfg),
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

	// 4. Build Nested Card 2: Dismiss / Tactical Retreat Card (Anchored bottom-right)
	g.cardDismiss = senjaui.NewBanner(
		senjaui.WithID("card-dismiss"),
		senjaui.WithPosition(-24, -20),
		senjaui.WithSize(230, 170),
		senjaui.WithAnchor(senjaui.AnchorBottomRight),
		senjaui.WithPivot(senjaui.PivotBottomRight),
		senjaui.WithPaddingAll(12),
		senjaui.WithNineSliceConfig(cardDismissCfg),
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

	// 5. Establish hierarchy: Add both cards to Parent Banner!
	g.banner.AddChild(g.cardAccept)
	g.banner.AddChild(g.cardDismiss)

	// 6. Connect to EbitenUI canvas
	g.ui = senjaui.NewUI(g.banner)
}

func (g *Game) Update() error {
	const dt = 1.0 / 60.0

	// 1. Advance UI animations (0 heap allocations)
	g.banner.Update(dt)

	// 2. Update EbitenUI input and events
	g.ui.Update()

	// 3. Interactive keyboard controls for live demonstration:
	// Space: Toggle banner open/close
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.banner.Toggle()
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
			g.banner.SetSize(560, 320)
		case 1:
			g.banner.SetSize(620, 350)
		case 2:
			g.banner.SetSize(520, 300)
		}
		g.statusText = fmt.Sprintf("Parent resized to: %.0fx%.0f (children automatically realigned!)", g.banner.Width, g.banner.Height)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Dark grid backdrop
	screen.Fill(color.RGBA{12, 14, 20, 255})

	// Render EbitenUI canvas (Parent banner & nested cards)
	g.ui.Draw(screen)

	// Render HUD information overlay
	hudY := 20.0
	g.drawText(screen, "=== SENJAUI HIERARCHICAL NINE-SLICE BANNER SYSTEM ===", 20, hudY, color.RGBA{255, 215, 0, 255})
	hudY += 20
	g.drawText(screen, fmt.Sprintf("Controls: [Space] Toggle | [1-4] Anim: %s | [R] Resize Parent", g.animStyle), 20, hudY, color.RGBA{180, 190, 210, 255})
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
	ebiten.SetWindowTitle("SenjaUI - Hierarchical Nine-Slice Banner & Card System")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatalf("Game loop error: %v", err)
	}
}

