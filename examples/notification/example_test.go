package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/senjalab/senjaui"
)

func TestGameLoopExecution(t *testing.T) {
	game := NewGame()
	screen := ebiten.NewImage(screenWidth, screenHeight)

	// Test initial frame
	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	game.Draw(screen)

	// Test animation frames
	for i := 0; i < 30; i++ {
		if err := game.Update(); err != nil {
			t.Fatalf("Update frame %d error: %v", i, err)
		}
		game.Draw(screen)
	}

	// Test resizing parent
	game.banner.SetSize(620, 350)
	game.Update()
	game.Draw(screen)

	// Test clicking child card and bubbling
	game.cardAccept.TriggerContentClick("card-accept:btn-accept")
	if game.lastClickLog != `Parent Banner caught click from: "card-accept:btn-accept"` {
		t.Fatalf("unexpected lastClickLog: %q", game.lastClickLog)
	}

	// Test closing and reopening
	game.banner.Close()
	for i := 0; i < 20; i++ {
		game.Update()
		game.Draw(screen)
	}
	if game.banner.IsOpen() {
		t.Fatalf("expected banner to close")
	}

	game.banner.SetAnimation(senjaui.AnimFadeIn, 0.2, senjaui.EaseLinear)
	game.banner.Open()
	for i := 0; i < 20; i++ {
		game.Update()
		game.Draw(screen)
	}
	if !game.banner.IsOpen() {
		t.Fatalf("expected banner to be open")
	}

	// Test dynamic image/skin swap
	prevSkin := game.currentSkin
	game.cycleSkin()
	if game.currentSkin == prevSkin {
		t.Fatalf("expected currentSkin to advance after cycleSkin")
	}
	game.Update()
	game.Draw(screen)
}


