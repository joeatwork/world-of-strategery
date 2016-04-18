package game

import (
	"container/list"
	"fmt"
	"log"
	"math"
	"time"
)

const defaultShadowSize float64 = 1

type Terrain struct {
	Board         [][]interface{}
	Width, Height int
}

type Location struct {
	X, Y             int
	OffsetX, OffsetY float64
}

type CharacterType struct {
	MovePerSec, WorkPerSec, MaxCarry float64
	Width, Height                    int
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
	Cultures   []*Culture
	lastUpdate time.Time
	terrain    Terrain
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

func FloatsFromLocation(l Location) (float64, float64) {
	return float64(l.X) + l.OffsetX, float64(l.Y) + l.OffsetY
}

func LocationFromFloats(x, y float64) Location {
	xf, offsetX := math.Modf(x)
	yf, offsetY := math.Modf(y)
	return Location{
		X:       int(xf),
		Y:       int(yf),
		OffsetX: offsetX,
		OffsetY: offsetY,
	}
}

// Propose a new location given an origin, target, speed and duration.
// return value is an absolute location
func chooseMove(from Location, to Location,
	speedPerSec float64, dt time.Duration) Location {
	if dt == 0 {
		return from
	}
	fromX, fromY := FloatsFromLocation(from)
	toX, toY := FloatsFromLocation(to)

	deltaX := toX - fromX
	deltaY := toY - fromY
	wholeJourneyDist := math.Hypot(deltaX, deltaY)
	canMoveDist := speedPerSec / dt.Seconds()
	canMovePortion := canMoveDist / wholeJourneyDist
	if canMovePortion > 1.0 {
		canMovePortion = 1.0
	}

	canMoveX := deltaX * canMovePortion
	canMoveY := deltaY * canMovePortion

	return LocationFromFloats(
		fromX+canMoveX,
		fromY+canMoveY,
	)
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

const maxMoveStack = 64
const maxShortMoveSide = 8

type tilePosition struct {
	x, y int
}

// abs(who.Location.X - goal.X) must be less than maxShortMoveSide
// abs(who.Location.Y - goal.Y) must be less than maxShortMoveSide
//
// who.Location will not be further away from goal after a call to
// attemptShortMove
func attemptShortMove(who *Character, terrain Terrain, goal Location) {
	var marks [maxShortMoveSide][maxShortMoveSide]bool
	var stack [maxMoveStack]tilePosition

	dx := goal.X - who.Location.X
	dy := goal.Y - who.Location.Y
	if dx == 0 && dy == 0 {
		who.Location.OffsetX = goal.OffsetX
		who.Location.OffsetY = goal.OffsetY
		return
	}
	if dx >= maxShortMoveSide || dx <= -maxShortMoveSide {
		panic("attemptShortMove called with X move beyond max size")
	}
	if dy >= maxShortMoveSide || dy <= -maxShortMoveSide {
		panic("attemptShortMove called with Y move beyond max size")
	}

	var regionOriginX int
	var regionOriginY int
	if dx < 0 {
		marginX := (maxShortMoveSide + dx) / 2
		regionOriginX = goal.X - marginX
	} else {
		marginX := (maxShortMoveSide - dx) / 2
		regionOriginX = who.Location.X - marginX
	}
	if dy < 0 {
		marginY := (maxShortMoveSide + dy) / 2
		regionOriginY = goal.Y - marginY
	} else {
		marginY := (maxShortMoveSide - dy) / 2
		regionOriginY = who.Location.Y - marginY
	}

	marks[who.Location.X-regionOriginX][who.Location.Y-regionOriginY] = true
	checkAndMark := func(tryX, tryY int) bool {
		markX := tryX - regionOriginX
		if markX < 0 || markX >= maxShortMoveSide {
			return false
		}
		markY := tryY - regionOriginY
		if markY < 0 || markY >= maxShortMoveSide {
			return false
		}

		if marks[markX][markY] {
			return false
		}

		clear := isTerrainClear(
			who,
			terrain,
			tryX,
			tryY,
			who.Type.Width,
			who.Type.Height,
		)

		if !clear {
			return false
		}

		marks[markX][markY] = true
		return true
	}

	stack[0].x = who.Location.X
	stack[0].y = who.Location.Y
	stackDepth := 1
	goalTile := tilePosition{
		x: goal.X,
		y: goal.Y,
	}

	var current tilePosition
	for stackDepth > 0 {
		current = stack[stackDepth-1]
		if current == goalTile {
			break
		}

		switch {
		case checkAndMark(current.x+1, current.y):
			stack[stackDepth].x = current.x + 1
			stack[stackDepth].y = current.y
			stackDepth++
		case checkAndMark(current.x, current.y+1):
			stack[stackDepth].x = current.x
			stack[stackDepth].y = current.y + 1
			stackDepth++
		case checkAndMark(current.x-1, current.y):
			stack[stackDepth].x = current.x - 1
			stack[stackDepth].y = current.y
			stackDepth++
		case checkAndMark(current.x, current.y-1):
			stack[stackDepth].x = current.x
			stack[stackDepth].y = current.y - 1
			stackDepth++
		default:
			stackDepth--
		}
	}

	newDx, newDy := goal.X-current.x, goal.Y-current.y
	if dx*dx+dy*dy <= newDx*newDx+newDy*newDy {
		return
	}

	for x := 0; x < who.Type.Width; x++ {
		for y := 0; y < who.Type.Height; y++ {
			oldX := who.Location.X + x
			oldY := who.Location.Y + y
			terrain.Board[oldX][oldY] = nil
		}
	}

	for x := 0; x < who.Type.Width; x++ {
		for y := 0; y < who.Type.Height; y++ {
			newX := current.x + x
			newY := current.y + y
			terrain.Board[newX][newY] = who
		}
	}

	if current == goalTile {
		who.Location = goal
	} else {
		who.Location = Location{
			X:       current.x,
			Y:       current.y,
			OffsetX: 0,
			OffsetY: 0,
		}
	}
}

// Attempts to move the character from their current position to the
// absolute position moveX, moveY. Move will stop if obstructed by
// terrain bounds or other objects. Updates the terrain and location
// of the given character.
func attemptMove(who *Character, terrain Terrain, goal Location) {
	// The only guarantees the current implementation of attemptMove
	// gives is
	//
	// 1) If the smallest rectangle containing who.Location and goal
	// is empty, who will be moved to goal
	//
	// 2) who.Location will never be further away from goal after a
	// call to attemptMove than it was before the call
	for {
		dx, dy := goal.X-who.Location.X, goal.Y-who.Location.Y
		nextGoal := goal
		switch {
		case dx >= maxShortMoveSide:
			nextGoal.X = who.Location.X + (maxShortMoveSide - 1)
			nextGoal.OffsetX = 0.0
		case dx <= -maxShortMoveSide:
			nextGoal.X = who.Location.X - (maxShortMoveSide + 1)
			nextGoal.OffsetX = 0.0
		}

		switch {
		case dy > maxShortMoveSide:
			nextGoal.Y = who.Location.Y + (maxShortMoveSide - 1)
			nextGoal.OffsetY = 0.0
		case dx < -maxShortMoveSide:
			nextGoal.Y = who.Location.Y - (maxShortMoveSide + 1)
			nextGoal.OffsetY = 0.0
		}

		start := who.Location
		attemptShortMove(who, terrain, nextGoal)
		if who.Location == start {
			break
		}
	}
}

func insideOfShadow(shadowSize float64, who *Character, target *House) bool {
	targetX, targetY := FloatsFromLocation(target.Location)
	whoXMin, whoYMin := FloatsFromLocation(who.Location)
	whoXMax := whoXMin + float64(who.Type.Width)
	whoYMax := whoYMin + float64(who.Type.Height)
	shadowXMin := targetX - shadowSize
	shadowYMin := targetY - shadowSize
	shadowXMax := targetX + float64(target.Type.Width) + shadowSize
	shadowYMax := targetY + float64(target.Type.Height) + shadowSize
	xOverlap :=
		(whoXMin > shadowXMin && whoXMin < shadowXMax) ||
			(whoXMax > shadowXMin && whoXMax < shadowXMax)

	if xOverlap {
		return (whoYMin > shadowYMin && whoYMin < shadowYMax) ||
			(whoYMax > shadowYMin && whoYMax < shadowYMax)
	}

	return false
}

func mine(who *Character, target *House, dt time.Duration) {
	transfer := who.Type.WorkPerSec * dt.Seconds()

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

func build(terrain Terrain, who *Character, target *House, dt time.Duration) {
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

	transfer := who.Type.WorkPerSec * dt.Seconds()

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

func NewGame(numCultures, width, height int) Game {
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

	return ret
}

type CantPlaceCharacterError struct {
	msg string
}

func (e *CantPlaceCharacterError) Error() string {
	return e.msg
}

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

	return character, nil
}

const maxPlansAllowedPerCulture = 255

func PlanHouse(culture *Culture, htype *HouseType, loc Location) *House {
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

func Tick(game Game, now time.Time) {
	if game.lastUpdate.IsZero() {
		game.lastUpdate = now
		return
	}

	dt := now.Sub(game.lastUpdate)

	// TODO shouldn't iterate by culture, or one team gets to move
	// before the other
	for _, culture := range game.Cultures {
		for e := culture.Characters.Front(); e != nil; e = e.Next() {
			who := e.Value.(*Character)
			switch target := who.Target.(type) {
			case *Location:
				choice := chooseMove(
					who.Location,
					*target,
					who.Type.MovePerSec,
					dt,
				)
				attemptMove(who, game.terrain, choice)
			case *House:
				if insideOfShadow(defaultShadowSize, who, target) {
					if who.Culture == target.Culture {
						build(game.terrain, who, target, dt)
					} else {
						mine(who, target, dt)
					}
					rerankHouse(game.terrain, target)
				} else {
					targetX, targetY := FloatsFromLocation(target.Location)
					workTarget := LocationFromFloats(
						targetX+float64(target.Type.Width)/2,
						targetY+float64(target.Type.Height)/2,
					)
					choice := chooseMove(
						who.Location,
						workTarget,
						who.Type.MovePerSec,
						dt,
					)
					attemptMove(who, game.terrain, choice)
				}
				reevaluateTargetHouse(who)
			case nil:
				// Nothing to do
			default:
				log.Panicf("unexpected character target type %T\n", target)
			}
		}
	}

	game.lastUpdate = now
}
