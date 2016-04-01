package server

// Commands a character to target coordinates. Character will try to
// get there, and then go idle.
func CommandGo(s *Server, you *game.Character, where *game.Location) error {
	you.Target = where
}

// Commands a character to got to work mining or building
// what. Character will try to get there, do as much work as it can,
// and then go idle
func CommandWork(s *Server, you *game.Character, what *game.House) error {
	you.Target = what
}

// Propose adding a house. The proposed house (if unblocked) will be
// eligible for building.
func CommandStartHouse(s *Server, us *game.Culture,
	kind HouseKind, where *game.Location) {
	house := game.House{
		Location: *where,
		Kind:     kind,
		Culture:  us,
	}
	us.PlannedHouses[house] = true
}

// Abandon the plan to build a house.
func AbandonPlannedHouse(s *Server, us *game.Culture, what *game.House) error {
	delete(us.PlannedHouses, what)
}
