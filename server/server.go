package server

// Commands a character to target coordinates. If the target is
// minable, and the character can carry resources, it will mine. If
// the target is buildable and the character has resources, it will
// deposit those resources.
func CommandGo(s *Server, you *game.Character, x game.Coord, y game.Coord) error {
	if !standsInBounds(x, y, you) {
		return OutOfBoundsError
	}
	you.XTarget = x
	you.YTarget = y
}

// Command a character to build a new foundation for a building at the
// given coordinates. A culture has one or zero planned buildings at a time.
func CommandStartHouse(s *Server, you *Character, us *Culture, kind HouseKind,
	x game.Coord, y game.Coord) {
	if !fitsInBounds(x, y, kind) {
		return OutOfBoundsError
	}
	you.XTarget = x
	you.YTarget = y
	us.PlannedBuilding = game.Building{
		X:       x,
		Y:       y,
		Kind:    kind,
		Culture: us,
	}
}
