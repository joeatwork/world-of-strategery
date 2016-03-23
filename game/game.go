package game

type Coord int8

const MaxCoord Coord = 255

type CharacterKind int

const (
	Worker CharacterKind = iota
)

type HouseKind int

const (
	Tower HouseKind = iota
)

// Houses, Swarm Characters, and Walls live here
type Terrain struct {
}

// You can't build or mine this, and you can't step on it either.
type Wall struct {
}

// Mobile, subject to orders, can mine and build
type Character struct {
	XTarget  Coord
	YTarget  Coord
	Carrying int
	Type     CharacterType
}

// Immobile, can be owned by a culture, can be mined and built
type House struct {
	X       Coord
	Y       Coord
	Kind    HouseKind
	Culture *Culture
}

type Swarm struct {
}

// This is a "side" - one culture per player
type Culture struct {
}

type characterAction int

const (
	building characterAction = iota
	mining
	moving
	idle
)

type characterAction struct {
	characterIndex  int
	characterAction characterAction
}

func (cs []characterAction) Len() int {
	return len(cs)
}

func (cs []characterAction) Less(i, j int) bool {
	if cs[i].characterAction == cs[j].characterAction {
		return cs[j].characterIndex < cs[i].characterIndex
	}

	return cs[i].characterAction < cs[j].characterAction
}

func (cs []characterAction) Swap(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}

func Tick(game Game) {
	// Sort Characters into builders, miners, and movers
	for i, character := range game.Characters {
		game.CharacterActionBuffer[i] = CharacterState{
			characterIndex: i,
		}

		switch {
		case canBuild(character):
			game.CharacterActionBuffer[i].characterAction = building
		case canMine(character):
			game.CharacterActionBuffer[i].characterAction = mining
		case canMove(character):
			game.CharacterActionBuffer[i].characterAction = moving
		default:
			game.CharacterActionBuffer[i].characterAction = idle
		}
	}

	sort.Sort(game.CharacterActionBuffer) // TODO character action is part of vision

	for _, characterAction := range characterActionBuffer {
		character := game.Characters[characterAction.characterIndex]
		switch characterAction.characterAction {
		case building:
			// TODO build
		case mining:
			// TODO mine
		case moving:
			// TODO path finding
		case idle:
			// TODO idle?
		default:
			panic("unknown character action %d", characterAction.characterAction)
		}

		// TODO Calculate vision
	}
}
