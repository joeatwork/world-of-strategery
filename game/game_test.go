package game

import (
	"math"
	"testing"
	"time"
)

var loc0x0 = Location{
	X:       0,
	Y:       0,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc0x1 = Location{
	X:       0,
	Y:       1,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc1x0 = Location{
	X:       1,
	Y:       0,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc1x1 = Location{
	X:       1,
	Y:       1,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc2x2 = Location{
	X:       2,
	Y:       2,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc0x2 = Location{
	X:       2,
	Y:       2,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc2x0 = Location{
	X:       2,
	Y:       2,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc0x4 = Location{
	X:       0,
	Y:       4,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc3x0 = Location{
	X:       3,
	Y:       0,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc3x3 = Location{
	X:       3,
	Y:       3,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc3x4 = Location{
	X:       3,
	Y:       4,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var loc6x8 = Location{
	X:       6,
	Y:       8,
	OffsetX: 0.0,
	OffsetY: 0.0,
}

var workerType = &CharacterType{
	MovePerSec: 1,
	WorkPerSec: 4,
	MaxCarry:   10,
	Width:      2,
	Height:     2,
}

var houseType = &HouseType{
	MaxResources: 100,
	Width:        1,
	Height:       1,
}

func TestChooseMoveZeroTime(t *testing.T) {
	dt0, _ := time.ParseDuration("0ns")
	choice := chooseMove(loc3x4, loc6x8, 2, dt0)
	if choice != loc3x4 {
		t.Errorf("expected no motion, got %v", choice)
	}
}

func TestChooseMovePositive(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc6x8, 5, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveNegative(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc6x8, loc0x0, 5, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveOvershot(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc3x4, 10, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveXOnly(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc3x0, 3, dt)
	if choice != loc3x0 {
		t.Errorf("expected %v, got %v", loc3x0, choice)
	}
}

func TestChooseMoveYOnly(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc0x4, 4, dt)
	if choice != loc0x4 {
		t.Errorf("expected %v, got %v", loc0x4, choice)
	}
}

func TestAddCharacterSimple(t *testing.T) {
	game := NewGame(1, 4, 4)

	character, err := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc1x1,
	)

	if err != nil {
		t.Errorf("unexpected error adding character: %v", err)
	}

	found := false
	for e := game.Cultures[0].Characters.Front(); e != nil; e = e.Next() {
		chr := e.Value.(*Character)
		found = character == chr
		if found {
			break
		}
	}

	if !found {
		t.Errorf("adding character didn't add to list")
	}

	for x := 0; x < game.terrain.Width; x++ {
		for y := 0; y < game.terrain.Height; y++ {
			if (x == 1 || x == 2) &&
				(y == 1 || y == 2) {
				if game.terrain.Board[x][y] != character {
					t.Errorf("character didn't appear where placed at %d,%d", x, y)
				}
			} else if game.terrain.Board[x][y] != nil {
				t.Errorf("character appeared in the wrong place at %d, %d", x, y)
			}
		}
	}
}

func TestAddCharacterOrigin(t *testing.T) {
	game := NewGame(1, 4, 4)
	character, err := AddCharacter(game.terrain, game.Cultures[0], workerType, loc0x0)

	if err != nil {
		t.Errorf("unexpected error adding character: %v", err)
	}

	for x := 0; x < game.terrain.Width; x++ {
		for y := 0; y < game.terrain.Height; y++ {
			if (x == 0 || x == 1) &&
				(y == 0 || y == 1) {
				if game.terrain.Board[x][y] != character {
					t.Errorf("character didn't appear where placed at %d,%d", x, y)
				}
			} else if game.terrain.Board[x][y] != nil {
				t.Errorf("character appeared in the wrong place at %d, %d", x, y)
			}
		}
	}
}

func TestAddCharacterEdge(t *testing.T) {
	game := NewGame(1, 4, 4)
	character, err := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc2x2,
	)

	if err != nil {
		t.Errorf("unexpected error adding character: %v", err)
	}

	for x := 0; x < game.terrain.Width; x++ {
		for y := 0; y < game.terrain.Height; y++ {
			if (x == 2 || x == 3) &&
				(y == 2 || y == 3) {
				if game.terrain.Board[x][y] != character {
					t.Errorf("character didn't appear where placed at %d,%d", x, y)
				}
			} else if game.terrain.Board[x][y] != nil {
				t.Errorf("character appeared in the wrong place at %d, %d", x, y)
			}
		}
	}
}

func TestAddCharacterCollision(t *testing.T) {
	game := NewGame(1, 4, 4)
	_, err := AddCharacter(game.terrain, game.Cultures[0], workerType, loc1x1)
	if err != nil {
		t.Errorf("unexpected error adding character: %v", err)
	}

	overlaps := [...]Location{
		Location{0, 0, 0.0, 0.0},
		Location{0, 1, 0.0, 0.0},
		Location{1, 0, 0.0, 0.0},
		Location{1, 1, 0.0, 0.0},
		Location{0, 2, 0.0, 0.0},
		Location{2, 0, 0.0, 0.0},
		Location{2, 2, 0.0, 0.0},
	}

	for _, loc := range overlaps {
		_, err2 := AddCharacter(game.terrain, game.Cultures[0], workerType, loc)
		if err2 == nil {
			t.Errorf("placing characer at %v allowed overlap", loc)
		}
	}
}

func TestAddCharacterOutOfBounds(t *testing.T) {
	game := NewGame(1, 4, 4)
	outofbounds := [...]Location{
		Location{-1, -1, 0.0, 0.0},
		Location{-1, 0, 0.0, 0.0},
		Location{0, -1, 0.0, 0.0},
		Location{3, 3, 0.0, 0.0},
		Location{3, 0, 0.0, 0.0},
		Location{0, 3, 0.0, 0.0},
		Location{4, 4, 0.0, 0.0},
		Location{4, 0, 0.0, 0.0},
		Location{0, 4, 0.0, 0.0},
		Location{-1, 3, 0.0, 0.0},
	}

	for _, loc := range outofbounds {
		_, err := AddCharacter(game.terrain, game.Cultures[0], workerType, loc)
		if err == nil {
			t.Errorf("placing character at %v allowed out of bounds", loc)
		}
	}
}

func TestAttemptShortMoveHappyPath(t *testing.T) {
	game := NewGame(1, 4, 4)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)

	attemptShortMove(character, game.terrain, loc2x2)
	if character.Location != loc2x2 {
		t.Errorf("expected move to %v got %v", loc2x2, character.Location)
	}
}

func TestAttemptShortMoveOutOfBounds(t *testing.T) {
	outofbounds := []Location{
		Location{-1, -1, 0.0, 0.0},
		Location{-1, 0, 0.0, 0.0},
		Location{0, -1, 0.0, 0.0},
		Location{3, 3, 0.0, 0.0},
		Location{3, 0, 0.0, 0.0},
		Location{0, 3, 0.0, 0.0},
		Location{4, 4, 0.0, 0.0},
		Location{4, 0, 0.0, 0.0},
		Location{0, 4, 0.0, 0.0},
		Location{-1, 3, 0.0, 0.0},
	}

	for _, loc := range outofbounds {
		game := NewGame(1, 4, 4)
		character, _ := AddCharacter(
			game.terrain,
			game.Cultures[0],
			workerType,
			loc0x0,
		)
		attemptShortMove(character, game.terrain, loc)
		goalX, goalY := FloatsFromLocation(loc)
		atX, atY := FloatsFromLocation(character.Location)
		oldDist := math.Hypot(goalX, goalY)
		newDist := math.Hypot(goalX-atX, goalY-atY)
		if oldDist < newDist {
			t.Errorf("Out of bounds shortMove increased distance: from %v to %v",
				loc, character.Location)
		}

	}
}

func TestAttemptShortMoveObstructed(t *testing.T) {
	game := NewGame(1, 8, 8)
	walker, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)

	for i := 0; i < game.terrain.Height; i = i + workerType.Height {
		_, err := AddCharacter(
			game.terrain,
			game.Cultures[0],
			workerType,
			Location{4, i, 0.0, 0.0},
		)

		if err != nil {
			t.Fatalf("Can't add blocker character at %d,%d", 4, i)
		}
	}

	attemptShortMove(walker, game.terrain, Location{6, 6, 0.0, 0.0})
	if walker.Location != loc0x0 {
		DumpTerrain(game.terrain)
		t.Errorf("Unexpected end of obstructed move. expected %v got %v",
			loc0x0, walker.Location)
	}
}

func TestAttemptShortMoveWalkAround(t *testing.T) {
	game := NewGame(1, 8, 8)
	walker, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{0, 4, 0.0, 0.0},
	)

	// Blockers
	AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{4, 4, 0.0, 0.0},
	)

	AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{4, 2, 0.0, 0.0},
	)

	AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{4, 6, 0.0, 0.0},
	)

	endpoint := Location{6, 4, 0.0, 0.0}
	attemptShortMove(walker, game.terrain, endpoint)
	if walker.Location != endpoint {
		DumpTerrain(game.terrain)
		t.Errorf("Unexpected end of walkaround move. expected %v got %v",
			endpoint, walker.Location)
	}
}

func TestAttemptMoveHappyPath(t *testing.T) {
	twidth := 4
	theight := 4
	for x := 0; x < twidth-workerType.Width; x++ {
		for y := 0; y < theight-workerType.Height; y++ {
			game := NewGame(1, twidth, theight)
			character, _ := AddCharacter(
				game.terrain,
				game.Cultures[0],
				workerType,
				loc0x0,
			)
			attempt := Location{x, y, 0.0, 0.0}
			attemptMove(character, game.terrain, attempt)

			if character.Location != attempt {
				t.Errorf("Unobstructed move to %v failed", attempt)
			}

			for spotx, col := range game.terrain.Board {
				for spoty, cell := range col {
					if spotx >= x &&
						spotx < x+character.Type.Width &&
						spoty >= y &&
						spoty < y+character.Type.Height {
						if cell != character {
							t.Errorf(
								"character moved to %d,%d didn't appear at %d,%d",
								x, y, spotx, spoty,
							)
						}
					} else {
						if cell != nil {
							t.Errorf(
								"character moved to %d,%d appeared at %d,%d",
								x, y, spotx, spoty,
							)
						}
					}
				}
			}
		}
	}
}

func TestAttemptMoveOutOfBounds(t *testing.T) {
	outofbounds := []Location{
		Location{-1, -1, 0.0, 0.0},
		Location{-1, 0, 0.0, 0.0},
		Location{0, -1, 0.0, 0.0},
		Location{3, 3, 0.0, 0.0},
		Location{3, 0, 0.0, 0.0},
		Location{0, 3, 0.0, 0.0},
		Location{4, 4, 0.0, 0.0},
		Location{4, 0, 0.0, 0.0},
		Location{0, 4, 0.0, 0.0},
		Location{-1, 3, 0.0, 0.0},
	}

	for _, loc := range outofbounds {
		game := NewGame(1, 4, 4)
		character, _ := AddCharacter(
			game.terrain,
			game.Cultures[0],
			workerType,
			loc0x0,
		)
		attemptMove(character, game.terrain, loc)
		goalX, goalY := FloatsFromLocation(loc)
		atX, atY := FloatsFromLocation(character.Location)
		oldDist := math.Hypot(goalX, goalY)
		newDist := math.Hypot(goalX-atX, goalY-atY)
		if oldDist < newDist {
			t.Errorf("out of bounds move increased distance: from %v to %v",
				loc, character.Location)
		}
	}
}

func TestAttemptMoveNonzeroOffset(t *testing.T) {
	attempts := []struct {
		attempt Location
		expect  Location
	}{
		{attempt: Location{0, 0, 0.0, 0.1}, expect: Location{0, 0, 0.0, 0.1}},
		{attempt: Location{0, 0, 0.1, 0.0}, expect: Location{0, 0, 0.1, 0.0}},
		{attempt: Location{1, 1, 0.2, 0.9}, expect: Location{1, 1, 0.2, 0.9}},
	}

	for _, pair := range attempts {
		game := NewGame(1, 4, 4)
		character, _ := AddCharacter(
			game.terrain,
			game.Cultures[0],
			workerType,
			Location{0, 0, 0.1, 0.9},
		)
		attemptMove(character, game.terrain, pair.attempt)
		if character.Location != pair.expect {
			t.Errorf("nonzero offset move failed, expected %v got %v",
				pair.expect, character.Location)
		}
	}
}

func TestAttemptMoveLong(t *testing.T) {
	game := NewGame(1, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		Location{0, 0, 0, 0},
	)
	target := Location{14, 14, 0.7, 0.9}
	attemptMove(character, game.terrain, target)
	if character.Location != target {
		t.Errorf("long move failed: expected %v got %v",
			target, character.Location)
	}
}

func TestAttemptMoveObstructed(t *testing.T) {
	game := NewGame(1, 32, 32)
	walker, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)

	for i := 0; i < game.terrain.Height; i = i + workerType.Height {
		_, err := AddCharacter(
			game.terrain,
			game.Cultures[0],
			workerType,
			Location{17, i, 0.0, 0.0},
		)

		if err != nil {
			t.Fatalf("Can't add blocker character at %d,%d", 4, i)
		}
	}

	expected := Location{14, 14, 0.0, 0.0}
	attemptMove(walker, game.terrain, Location{30, 30, 0.3, 0.4})
	if walker.Location != expected {
		DumpTerrain(game.terrain)
		t.Errorf("long obstructed move unexpected result. expected %v got %v",
			expected, walker.Location)
	}
}

func TestTooManyPlannedHouses(t *testing.T) {
	game := NewGame(1, 1, 1)
	for i := 0; i < maxPlansAllowedPerCulture; i++ {
		PlanHouse(
			game.Cultures[0],
			houseType,
			loc0x0,
		)
	}

	if len(game.Cultures[0].PlannedHouses) != maxPlansAllowedPerCulture {
		t.Fatalf("Tried to plan maxAllowed Houses (%d), planned %d instead",
			maxPlansAllowedPerCulture, len(game.Cultures[0].PlannedHouses))
	}

	lastHouse := PlanHouse(
		game.Cultures[0],
		houseType,
		loc0x0,
	)

	if len(game.Cultures[0].PlannedHouses) != maxPlansAllowedPerCulture {
		t.Errorf("Adding one to maxAllowed house plans yielded unexpected %d",
			len(game.Cultures[0].PlannedHouses))
	}

	if _, ok := game.Cultures[0].PlannedHouses[lastHouse]; !ok {
		t.Errorf("Couldn't add one last house to a full set of plans")
	}
}

func TestInsideOfShadow(t *testing.T) {
	width, height := 16, 16
	for x := 0; x+workerType.Width < width; x++ {
		for y := 0; y+workerType.Height < height; y++ {
			game := NewGame(1, width, height)
			house := PlanHouse(
				game.Cultures[0],
				houseType,
				Location{4, 4, 0.0, 0.0},
			)

			character, _ := AddCharacter(
				game.terrain,
				game.Cultures[0],
				workerType,
				Location{x, y, 0.0, 0.0},
			)

			xShadow := x >= 2 && x < 6
			yShadow := y >= 2 && y < 6
			shadow := xShadow && yShadow

			if insideOfShadow(1, character, house) != shadow {
				t.Errorf("Wanted insideOfShadow(%v) to be %v but it wasn't",
					character.Location, shadow)
			}
		}
	}

}

func TestRerankPlannedToBuilt(t *testing.T) {
	game := NewGame(1, 16, 16)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	house.ResourcesLeft = house.ResourcesLeft + 1
	rerankHouse(game.terrain, house)

	if _, planned := game.Cultures[0].PlannedHouses[house]; planned {
		t.Errorf("Planned house with positive resources still planned")
	}

	if _, built := game.Cultures[0].BuiltHouses[house]; !built {
		t.Errorf("Planned house with positive resources not built")
	}

	for x, column := range game.terrain.Board {
		expectX := x >= house.Location.X && x < house.Location.X+house.Type.Width
		for y, obj := range column {
			expectY := y >= house.Location.Y && y < house.Location.Y+house.Type.Height
			if (obj == house) != (expectX && expectY) {
				t.Errorf("House built == %v at strange location %d,%d",
					obj == house, x, y)
			}
		}
	}
}

func TestRerankBuiltToDemolished(t *testing.T) {
	game := NewGame(1, 16, 16)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	house.ResourcesLeft = house.ResourcesLeft + 1
	rerankHouse(game.terrain, house)
	house.ResourcesLeft = 0
	rerankHouse(game.terrain, house)

	if _, planned := game.Cultures[0].PlannedHouses[house]; planned {
		t.Errorf("Demolished house still planned")
	}

	if _, built := game.Cultures[0].BuiltHouses[house]; built {
		t.Errorf("Demolished house still built")
	}

	for x, column := range game.terrain.Board {
		for y, obj := range column {
			if obj == house {
				t.Errorf("house zombie at %d,%d", x, y)
			}
		}
	}
}

func TestReevaluateBuildOk(t *testing.T) {
	game := NewGame(1, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	character.Carrying = character.Type.MaxCarry
	character.Target = house

	reevaluateTargetHouse(character)
	if character.Target != house {
		t.Errorf("Character lost interest in building for no reason")
	}
}

func TestReevaluateMineOk(t *testing.T) {
	game := NewGame(2, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[1],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	house.ResourcesLeft = house.Type.MaxResources
	character.Target = house

	reevaluateTargetHouse(character)
	if character.Target != house {
		t.Errorf("Character lost interest in building for no reason")
	}
}

func TestReevaluateBuildComplete(t *testing.T) {
	game := NewGame(1, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	character.Carrying = character.Type.MaxCarry
	house.ResourcesLeft = house.Type.MaxResources
	character.Target = house

	reevaluateTargetHouse(character)
	if character.Target != nil {
		t.Errorf("Character still trying to build completed structure")
	}
}

func TestReevaluateOutOfBuildResources(t *testing.T) {
	game := NewGame(1, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	character.Carrying = 0
	character.Target = house

	reevaluateTargetHouse(character)
	if character.Target != nil {
		t.Errorf("Character still trying to build with no resources")
	}
}

func TestReevaluateOutOfCarryCapacity(t *testing.T) {
	game := NewGame(2, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[1],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	house.ResourcesLeft = house.Type.MaxResources
	character.Target = house
	character.Carrying = character.Type.MaxCarry

	reevaluateTargetHouse(character)
	if character.Target != nil {
		t.Errorf("Character still trying to mine without carrying capacity")
	}
}

func TestReevaluateHouseDemolished(t *testing.T) {
	game := NewGame(1, 16, 16)
	character, _ := AddCharacter(
		game.terrain,
		game.Cultures[0],
		workerType,
		loc0x0,
	)
	house := PlanHouse(
		game.Cultures[0],
		houseType,
		Location{4, 4, 0.0, 0.0},
	)
	character.Carrying = character.Type.MaxCarry
	character.Target = house
	UnplanHouse(house)
	reevaluateTargetHouse(character)
	if character.Target != nil {
		t.Errorf("Character still trying to build demolished house")
	}
}
