package game

import (
	"container/list"
	"fmt"
	"log"
	"math"
)

const defaultShadowSize = 1

type Terrain struct {
	Board         [][]interface{}
	Width, Height int
}

type Location struct {
	X, Y   int
	Offset float64
}

type CharacterType struct {
	// TODO rounding errors at short ticks!!
	MovePerTick, WorkPerTick, MaxCarry float64
	Width, Height                      int
}

type Character struct {
	Carrying float64
	Culture  *Culture
	Location Location
	Target   interface{}
	Type     *CharacterType
}

type HouseType struct {
	MaxResources  float64
	Width, Height int
}

type House struct {
	Type          *HouseType
	Culture       *Culture
	Location      Location
	ResourcesLeft float64
}

type Culture struct {
	Characters    list.List
	PlannedHouses map[*House]bool
	BuiltHouses   map[*House]bool
}

type Game struct {
	Cultures []*Culture
	terrain  Terrain
}

func DumpTerrain(terrain Terrain) {
	chars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ?"
	nextChar := 0
	thingToName := make(map[interface{}]string)
	thingToName[nil] = "-"

	for row := 0; row < terrain.Height; row++ {
		for col := 0; col < terrain.Width; col++ {
			obj := terrain.Board[col][row]
			name, ok := thingToName[obj]
			if !ok {
				thingToName[obj] = string(chars[nextChar])
				if nextChar < len(chars)-1 {
					nextChar++
				}
				name = thingToName[obj]
			}
			fmt.Printf("%s", name)
		}
		fmt.Printf("\n")
	}

	for thing, name := range thingToName {
		fmt.Printf("%s : %v\n", name, thing)
	}
}

func isTerrainClear(who interface{}, terrain Terrain, x, y, width, height int) bool {
	if x < 0 || x+width > terrain.Width {
		return false
	}
	if y < 0 || y+height > terrain.Height {
		return false
	}

	for checkX := 0; checkX < width; checkX++ {
		for checkY := 0; checkY < height; checkY++ {
			tryX, tryY := x+checkX, y+checkY
			occupant := terrain.Board[tryX][tryY]
			if nil != occupant && who != occupant {
				return false
			}
		}
	}

	return true
}

const maxShortMoveSide = 8 // maxShortMoveSide must be less than sqrt MAX_INT
const maxFringeLength = maxShortMoveSide * maxShortMoveSide

type tile struct {
	x, y int
}

type steps struct {
	count              [maxShortMoveSide][maxShortMoveSide]int
	oX, oY, dirX, dirY int
}

func distSquared(a, b tile) int {
	dx, dy := a.x-b.x, a.y-b.y
	return dx*dx + dy*dy
}

func writeStep(count int, s *steps, t tile) {
	stepsX := (t.x - s.oX) * s.dirX
	stepsY := (t.y - s.oY) * s.dirY
	fmt.Printf("Writing step %d to (org:%d,%d) (tile:%d,%d) (step:%d,%d)\n",
		count, s.oX, s.oY, t.x, t.y, stepsX, stepsY)
	s.count[stepsX][stepsY] = count
}

func readStep(s *steps, t tile) int {
	stepsX := (t.x - s.oX) * s.dirX
	stepsY := (t.y - s.oY) * s.dirY
	return s.count[stepsX][stepsY]
}

func checkMove(who *Character, terrain Terrain, s *steps, x, y int) bool {
	stepX := (x - s.oX) * s.dirX
	stepY := (y - s.oY) * s.dirY

	if x < 0 || x > terrain.Width {
		return false
	}
	if y < 0 || y > terrain.Height {
		return false
	}
	if stepX < 0 || stepX >= maxShortMoveSide {
		return false
	}
	if stepY < 0 || stepY >= maxShortMoveSide {
		return false
	}
	if s.count[stepX][stepY] != -1 {
		return false
	}

	return isTerrainClear(
		who,
		terrain,
		x, y,
		who.Type.Width,
		who.Type.Height,
	)
}

func attemptShortMove(who *Character, terrain Terrain, goal Location, walkDistance float64) float64 {
	if walkDistance < 0 {
		panic("walkDistance must not be negative")
	}

	totalDistance := walkDistance + who.Location.Offset

	// offset preserves the leftover walkDistance
	// between calls to attemptShortMove, so that
	//
	//    attemptShortMove(..., far_away, 0.4)
	//    attemptShortMove(..., far_away, 0.4)
	//    attemptShortMove(..., far_away, 0.4)
	//
	// Ends with a move of one tile and an offset of 0.2
	//
	// As a special case, if the character arrives at goal, then it's
	// offset will be goal.Offset - the character will decline to move
	// all of walkDistance.

	stepDistancef, offset := math.Modf(totalDistance)
	stepDistance := int(stepDistancef)

	visionOffsetX, visionOffsetY := 0, 0
	dirX, dirY := 1, 1
	dx, dy := goal.X-who.Location.X, goal.Y-who.Location.Y
	if dx < maxShortMoveSide && dx >= 0 {
		visionOffsetX = (dx - maxShortMoveSide) / 2
	}
	if dx > -maxShortMoveSide && dx < 0 {
		visionOffsetX = (maxShortMoveSide + dx) / 2
	}
	if dy < maxShortMoveSide && dy >= 0 {
		visionOffsetY = (dy - maxShortMoveSide) / 2
	}
	if dy < -maxShortMoveSide && dy < 0 {
		visionOffsetY = (maxShortMoveSide + dy) / 2
	}
	if visionOffsetX == 0 {
		visionOffsetX = 1
	}
	if visionOffsetY == 0 {
		visionOffsetY = 1
	}
	if dx < 0 {
		dirX = -1
	}
	if dy < 0 {
		dirY = -1
	}

	fmt.Printf("TODO VISON OFFSETS %d => %d, %d => %d\n",
		dx, visionOffsetX, dy, visionOffsetY)

	steps := &steps{
		// Because of dirX, the sign of visionOffset isn't what you think it should be
		oX:   who.Location.X + visionOffsetX,
		oY:   who.Location.Y + visionOffsetY,
		dirX: dirX,
		dirY: dirY,
	}

	// if steps.oX < 0 {
	// 	steps.oX = 0
	// }
	// if steps.oY < 0 {
	// 	steps.oY = 0
	// }
	// if steps.ox > TOO BIG // EDGE CASE super skinny
	// if steps.oy > TOO BIG // EDGE CASE super short fields?

	for x, _ := range steps.count {
		for y := range steps.count[x] {
			steps.count[x][y] = -1
		}
	}

	fmt.Printf("TODO steps %v\n", steps)

	var fringe [maxFringeLength]tile
	goalTile := tile{goal.X, goal.Y}
	fringe[0] = tile{who.Location.X, who.Location.Y}
	fringeStart := 0
	fringeEnd := 1
	writeStep(0, steps, fringe[0])

	for fringeStart < fringeEnd {
		fmt.Printf("TODO FRINGE %d:%d < %d: %v\n",
			fringeStart, fringeEnd, maxFringeLength, fringe[fringeStart:fringeEnd])
		fmt.Printf("TODO STEPS:\n")
		for x, _ := range steps.count {
			fmt.Printf("TODO > [")
			for y, _ := range steps.count[x] {
				fmt.Printf("%2d ", steps.count[x][y])
			}
			fmt.Printf("]\n")
		}

		current := fringe[fringeStart]
		currentDist := readStep(steps, current)
		fringeStart++
		if currentDist < stepDistance {
			if checkMove(who, terrain, steps, current.x+1, current.y) {
				fringe[fringeEnd] = tile{current.x + 1, current.y}
				writeStep(currentDist+1, steps, fringe[fringeEnd])
				fringeEnd++
			}
			if checkMove(who, terrain, steps, current.x, current.y+1) {
				fringe[fringeEnd] = tile{current.x, current.y + 1}
				writeStep(currentDist+1, steps, fringe[fringeEnd])
				fringeEnd++
			}
			if checkMove(who, terrain, steps, current.x-1, current.y) {
				fringe[fringeEnd] = tile{current.x - 1, current.y}
				writeStep(currentDist+1, steps, fringe[fringeEnd])
				fringeEnd++
			}
			if checkMove(who, terrain, steps, current.x, current.y-1) {
				fringe[fringeEnd] = tile{current.x, current.y - 1}
				writeStep(currentDist+1, steps, fringe[fringeEnd])
				fringeEnd++
			}
		}
	}

	fmt.Printf("TODO Fringe check complete: %d*%d,%d*%d toward %v\n",
		steps.oX, steps.dirX, steps.oY, steps.dirY, goalTile)
	for x, _ := range steps.count {
		fmt.Printf("TODO > [")
		for y, _ := range steps.count[x] {
			fmt.Printf("%2d ", steps.count[x][y])
		}
		fmt.Printf("]\n")
	}

	// We've found the shortest distance to every reachable point in
	// our "vision" range.

	closestPoint := tile{who.Location.X, who.Location.Y}
	closestSteps := readStep(steps, closestPoint)
	closestDistSquared := distSquared(closestPoint, goalTile)

	for x, _ := range steps.count {
		for y, _ := range steps.count[x] {
			pt := tile{
				x: steps.oX + (x * steps.dirX),
				y: steps.oY + (y * steps.dirY),
			}
			// fmt.Printf("TODO Distance check %d,%d => %v\n", x, y, pt)
			ptSteps := readStep(steps, pt)
			// fmt.Printf("TODO     ... distance %v steps\n", ptSteps)

			if ptSteps != -1 {
				newDSquared := distSquared(pt, goalTile)
				if newDSquared < closestDistSquared {
					closestPoint = pt
					closestSteps = ptSteps
					closestDistSquared = newDSquared
					fmt.Printf("TODO new closest point %v\n", closestPoint)
				}
			}
		}
	}

	if closestPoint == goalTile && offset > goal.Offset {
		// If the character arrives at goal, it won't continue
		// walking, so we discard leftover offset
		offset = goal.Offset
	}

	// closestPoint now contains the closest reachable point to goal
	// within our search area.

	// closestSteps now contains the distance in steps to closestPoint

	for x := 0; x < who.Type.Width; x++ {
		for y := 0; y < who.Type.Height; y++ {
			oldX := who.Location.X + x
			oldY := who.Location.Y + y
			terrain.Board[oldX][oldY] = nil
		}
	}

	originalOffset := who.Location.Offset
	who.Location.X = closestPoint.x
	who.Location.Y = closestPoint.y
	who.Location.Offset = offset

	for x := 0; x < who.Type.Width; x++ {
		for y := 0; y < who.Type.Height; y++ {
			newX := who.Location.X + x
			newY := who.Location.Y + y
			terrain.Board[newX][newY] = who
		}
	}

	// TODO this could be negative?
	return float64(closestSteps) + (who.Location.Offset - originalOffset)
}

func attemptMove(who *Character, terrain Terrain, goal Location, walkDistance float64) float64 {
	fmt.Printf("TODO AttemptMove %v => %v\n", who.Location, goal)
	movedTotal := float64(0)
	moveRemaining := walkDistance
	for moveRemaining > 0 {
		movedNext := attemptShortMove(who, terrain, goal, moveRemaining)
		movedTotal = movedTotal + movedNext
		moveRemaining = moveRemaining - movedNext
		fmt.Printf("TODO Next ShortMove => %v movedNext (%.2f) total %.2f remaining %.2f\n",
			who.Location, movedNext, movedTotal, moveRemaining)
		if movedNext == 0 {

			break
		}
	}

	fmt.Printf("TODO Final Position: %v\n", who.Location)
	return movedTotal
}

func insideOfShadow(shadowSize int, who *Character, target *House) bool {
	targetX, targetY := target.Location.X, target.Location.Y
	whoXMin, whoYMin := who.Location.X, who.Location.Y
	whoXMax := whoXMin + who.Type.Width
	whoYMax := whoYMin + who.Type.Height
	shadowXMin := targetX - shadowSize
	shadowYMin := targetY - shadowSize
	shadowXMax := targetX + target.Type.Width + shadowSize
	shadowYMax := targetY + target.Type.Height + shadowSize
	xOverlap :=
		(whoXMin > shadowXMin && whoXMin < shadowXMax) ||
			(whoXMax > shadowXMin && whoXMax < shadowXMax)

	if xOverlap {
		return (whoYMin > shadowYMin && whoYMin < shadowYMax) ||
			(whoYMax > shadowYMin && whoYMax < shadowYMax)
	}

	return false
}

func mine(who *Character, target *House, dt float64) {
	transfer := who.Type.WorkPerTick * dt

	// Mining
	if transfer > target.ResourcesLeft {
		transfer = target.ResourcesLeft
	}
	if transfer > who.Type.MaxCarry-who.Carrying {
		transfer = who.Type.MaxCarry - who.Carrying
	}

	target.ResourcesLeft = target.ResourcesLeft - transfer
	who.Carrying = who.Carrying + transfer
}

func build(terrain Terrain, who *Character, target *House, dt float64) {
	if _, built := target.Culture.BuiltHouses[target]; !built {
		// You can't build a house if it's position is obstructed.
		clear := isTerrainClear(
			target,
			terrain,
			target.Location.X,
			target.Location.Y,
			target.Type.Width,
			target.Type.Height,
		)
		if !clear {
			return
		}
	}

	transfer := who.Type.WorkPerTick * dt

	if transfer > target.Type.MaxResources-target.ResourcesLeft {
		transfer = target.Type.MaxResources - target.ResourcesLeft
	}
	if transfer > who.Carrying {
		transfer = who.Carrying
	}

	target.ResourcesLeft = target.ResourcesLeft + transfer
	who.Carrying = who.Carrying - transfer
}

func rerankHouse(terrain Terrain, house *House) {
	if house.ResourcesLeft == 0 {
		delete(house.Culture.BuiltHouses, house)
		for x := 0; x < house.Type.Width; x++ {
			for y := 0; y < house.Type.Height; y++ {
				oldX := house.Location.X + x
				oldY := house.Location.Y + y
				terrain.Board[oldX][oldY] = nil
			}
		}
	}

	_, planned := house.Culture.PlannedHouses[house]
	if house.ResourcesLeft > 0 && planned {
		delete(house.Culture.PlannedHouses, house)
		house.Culture.BuiltHouses[house] = true
		for x := 0; x < house.Type.Width; x++ {
			for y := 0; y < house.Type.Height; y++ {
				newX := house.Location.X + x
				newY := house.Location.Y + y
				terrain.Board[newX][newY] = house
			}
		}
	}
}

func reevaluateTargetHouse(who *Character) {
	house := who.Target.(*House)
	if _, ok := house.Culture.PlannedHouses[house]; !ok {
		if _, ok := house.Culture.BuiltHouses[house]; !ok {
			goto abandon // House has gone away
		}
	}

	if house.Culture == who.Culture { // Building
		if who.Carrying == 0 {
			goto abandon // nothing left to build with
		}
		if house.ResourcesLeft >= house.Type.MaxResources {
			goto abandon // our work is done
		}
	} else { // Mining
		if who.Carrying >= who.Type.MaxCarry {
			goto abandon // can't carry any more
		}
	}

	return

abandon:
	who.Target = nil
	return
}

func NewGame(numCultures, width, height int) *Game {
	ret := Game{
		Cultures: make([]*Culture, numCultures),
		terrain: Terrain{
			Board:  make([][]interface{}, width),
			Width:  width,
			Height: height,
		},
	}

	for i, _ := range ret.Cultures {
		ret.Cultures[i] = &Culture{
			PlannedHouses: make(map[*House]bool),
			BuiltHouses:   make(map[*House]bool),
		}
	}

	for i, _ := range ret.terrain.Board {
		ret.terrain.Board[i] = make([]interface{}, ret.terrain.Height)
	}

	return &ret
}

type CantPlaceCharacterError struct {
	msg string
}

func (e *CantPlaceCharacterError) Error() string {
	return e.msg
}

// TODO can't call this with private terrain!?
func AddCharacter(terrain Terrain, culture *Culture,
	ctype *CharacterType, loc Location) (*Character, *CantPlaceCharacterError) {
	positionClear := isTerrainClear(
		nil,
		terrain,
		int(loc.X),
		int(loc.Y),
		ctype.Width,
		ctype.Height,
	)

	if !positionClear {
		return nil, &CantPlaceCharacterError{
			"can't place character, position is occupied or out of bounds",
		}
	}

	character := &Character{
		Culture:  culture,
		Location: loc,
		Type:     ctype,
	}

	for x := 0; x < character.Type.Width; x++ {
		for y := 0; y < character.Type.Height; y++ {
			placeX, placeY := int(loc.X)+x, int(loc.Y)+y
			terrain.Board[placeX][placeY] = character
		}
	}

	culture.Characters.PushBack(character)
	return character, nil
}

const maxPlansAllowedPerCulture = 255

func PlanHouse(culture *Culture, houseType *HouseType, loc Location) *House {
	if len(culture.PlannedHouses) >= maxPlansAllowedPerCulture {
		for k, _ := range culture.PlannedHouses {
			delete(culture.PlannedHouses, k)
			break
		}
	}

	ret := &House{
		Type:          houseType,
		Culture:       culture,
		Location:      loc,
		ResourcesLeft: 0,
	}

	culture.PlannedHouses[ret] = true
	return ret
}

func UnplanHouse(house *House) {
	delete(house.Culture.PlannedHouses, house)
}

func Tick(game *Game, dt float64) {
	// TODO shouldn't iterate by culture, or one team gets to move
	// before the other
	for _, culture := range game.Cultures {
		for e := culture.Characters.Front(); e != nil; e = e.Next() {
			who := e.Value.(*Character)
			switch target := who.Target.(type) {
			case *Location:
				TODO_Old_location := who.Location
				distance := who.Type.MovePerTick * dt
				attemptMove(who, game.terrain, *target, distance)
				fmt.Printf("TODO attempting move %v => %v (ended %v)\n",
					TODO_Old_location, target, who.Location)
			case *House:
				if insideOfShadow(defaultShadowSize, who, target) {
					if who.Culture == target.Culture {
						build(game.terrain, who, target, dt)
					} else {
						mine(who, target, dt)
						fmt.Printf("TODO attempting mine\n")
					}
					rerankHouse(game.terrain, target)
				} else {
					workTarget := Location{
						X:      target.Location.X + (target.Type.Width / 2),
						Y:      target.Location.Y + (target.Type.Height / 2),
						Offset: 0.0,
					}
					distance := who.Type.MovePerTick * dt
					attemptMove(who, game.terrain, workTarget, distance)
					fmt.Printf("TODO attempting move to house %v (ended %v)\n",
						target, who.Location)
				}
				reevaluateTargetHouse(who)
			case nil:
				// Nothing to do
			default:
				log.Panicf("unexpected character target type %T\n", target)
			}
		}
	}
}
