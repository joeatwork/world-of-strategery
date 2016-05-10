package game

import "testing"

func TestBuildAndMine(t *testing.T) {
	game := NewGame(4, 4)
	AddCulture(game)
	AddCulture(game)
	redHouse := PlanHouse(
		game.Cultures[0],
		houseType,
		loc0x0,
	)
	red, err := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{2, 0, 0.0},
	)
	if err != nil {
		DumpTerrain(game.terrain)
		t.Fatalf("can't place red for BuildAndMine scenario: %v", err)
	}

	green, err := AddCharacter(
		game.terrain,
		game.Cultures[1],
		workerType,
		Location{0, 2, 0.0},
	)
	if err != nil {
		DumpTerrain(game.terrain)
		t.Fatalf("can't place green for BuildAndMine scenario: %v", err)
	}

	for redHouse.ResourcesLeft < redHouse.Type.MaxResources {
		if red.Carrying == 0 {
			red.Carrying = red.Type.MaxCarry
			red.Target = redHouse
		}

		carryingBefore := red.Carrying
		locBefore := red.Location

		Tick(game, 1)

		if carryingBefore == red.Carrying && locBefore == red.Location {
			DumpTerrain(game.terrain)
			t.Fatalf("Gridlock trying to build")
		}
	}

	if red.Target != nil {
		t.Errorf("Red failed to reevaluate after building red house")
	}

	red.Target = &Location{2, 2, 0.0}
	for *red.Target.(*Location) != red.Location {
		locBefore := red.Location
		Tick(game, 1)

		if locBefore == red.Location {
			DumpTerrain(game.terrain)
			t.Fatalf("Gridlock trying to move red away from house: stuck at %v",
				red.Location)
		}
	}

	green.Target = redHouse
	for green.Carrying < green.Type.MaxCarry {
		carryingBefore := green.Carrying
		locBefore := green.Location

		Tick(game, 1)

		if carryingBefore == green.Carrying && locBefore == green.Location {
			DumpTerrain(game.terrain)
			t.Fatalf("Gridlock trying to mine")
		}
	}

	if green.Target != nil {
		t.Errorf("Green failed to reevaluate after mining red house")
	}
}
