package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestSkillTreeGameLoopExecution(t *testing.T) {
	game := NewSkillTreeGame()
	screen := ebiten.NewImage(screenWidth, screenHeight)

	// Test Update & Draw initial frame
	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	game.Draw(screen)

	// Simulate selecting and unlocking power_strike node
	nodeStrike, ok := game.graph.GetNode("power_strike")
	if !ok {
		t.Fatalf("expected power_strike node to exist in graph")
	}

	game.unlockNode(nodeStrike)
	if !nodeStrike.Data.Unlocked {
		t.Fatalf("expected power_strike to be unlocked")
	}

	// Verify next frame renders with newly unlocked node
	if err := game.Update(); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	game.Draw(screen)

	// Test camera panning and drawing
	game.graph.Pan(50, -30)
	game.graph.SetZoom(1.2)
	game.Draw(screen)
}

