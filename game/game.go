package game

import (
	"container/list"
	"log"
	"math"
	"time"
)

type Terrain struct {
	Board      [][]interface{}
	XMax, YMax int
}

type Location struct {
	X, Y float64
}

type Area struct {
	Location      Location
	Height, Width int
}

type Character struct {
	Culture    *Culture
	MovePerSec float64
	WorkPerSec float64
	Area       Area
	Carrying   float64
	MaxCarry   float64
	Target     interface{}
}

type House struct {
	Culture       *Culture
	Area          Area
	ResourcesLeft float64 // TODO - float resources means rounding errors!
	MaxResources  float64
}

type Culture struct {
	Characters    list.List // <House>
	PlannedHouses map[*House]bool
	BuiltHouses   map[*House]bool
}

type Game struct {
	cultures   []Culture
	lastUpdate time.Time
	terrain    *Terrain
}

// Propose a new location given an origin, target, speed and duration.
// return value is an absolute location
func chooseMove(from Location, to Location,
	speedPerSec float64, dt time.Duration) Location {
	if dt == 0 {
		return from
	}

	deltaX := to.X - from.X
	deltaY := to.Y - from.Y
	wholeJourneyDist := math.Hypot(deltaX, deltaY)
	canMoveDist := speedPerSec / dt.Seconds()
	canMovePortion := canMoveDist / wholeJourneyDist
	if canMovePortion > 1.0 {
		canMovePortion = 1.0
	}

	canMoveX := deltaX * canMovePortion
	canMoveY := deltaY * canMovePortion

	return Location{
		X: from.X + canMoveX,
		Y: from.Y + canMoveY,
	}
}

// Attempts to move the character from their current position to the
// absolute position moveX, moveY. May move the character a shorter
// distance, or no distance at all.
func attemptMove(who *Character, terrain *Terrain, attempt Location) {
	moveTileXf, moveOffsetX := math.Modf(attempt.X)
	moveTileYf, moveOffsetY := math.Modf(attempt.Y)
	moveTileX, moveTileY := int(moveTileXf), int(moveTileYf)
	width := who.Area.Width
	height := who.Area.Height

	if moveTileX < 0 {
		moveTileX = 0
	}
	if moveTileY < 0 {
		moveTileY = 0
	}
	if moveTileX > terrain.XMax {
		moveTileX = terrain.XMax
	}
	if moveTileY > terrain.YMax {
		moveTileY = terrain.YMax
	}

	originTileXf, _ := math.Modf(who.Area.Location.X)
	originTileYf, _ := math.Modf(who.Area.Location.Y)
	originTileX, originTileY := int(originTileXf), int(originTileYf)

stepforward:
	for originTileX != moveTileX || originTileY != moveTileY {
		destX := originTileX
		destY := originTileY
		switch {
		case originTileX < moveTileX:
			destX = destX + 1
		case originTileX > moveTileX:
			destX = destX - 1
		case originTileY < moveTileY:
			destY = destY + 1
		case originTileY > moveTileY:
			destY = destY - 1
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				tryX := destX + x
				tryY := destY + y
				occupant := terrain.Board[tryX][tryY]
				if nil != occupant || who != occupant {
					// TODO, if who is blocked in X direction they
					// should still advance along the Y direction
					break stepforward
				}
			}
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				oldX := originTileX + x
				oldY := originTileY + y
				terrain.Board[oldX][oldY] = nil
			}
		}

		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				newX := destX + x
				newY := destY + y
				terrain.Board[newX][newY] = who
			}
		}

		originTileX = destX
		originTileY = destY
	}

	if originTileX == moveTileX {
		who.Area.Location.X = float64(originTileX) + moveOffsetX
	} else {
		who.Area.Location.X = float64(originTileX)
	}

	if originTileY == moveTileY {
		who.Area.Location.Y = float64(originTileY) + moveOffsetY
	} else {
		who.Area.Location.Y = float64(originTileY)
	}
}

const shadowSize float64 = 1

func insideOfShadow(who *Character, target *House) bool {
	shadowXMin := target.Area.Location.X - shadowSize
	shadowYMin := target.Area.Location.Y - shadowSize
	shadowXMax := target.Area.Location.X + float64(target.Area.Width) + shadowSize
	shadowYMax := target.Area.Location.Y + float64(target.Area.Height) + shadowSize
	whoXMin := who.Area.Location.X
	whoYMin := who.Area.Location.Y
	whoXMax := who.Area.Location.X + float64(who.Area.Width)
	whoYMax := who.Area.Location.Y + float64(who.Area.Height)
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
	transfer := who.WorkPerSec * dt.Seconds()

	// Mining
	if transfer > target.ResourcesLeft {
		transfer = target.ResourcesLeft
	}
	if transfer > who.MaxCarry-who.Carrying {
		transfer = who.MaxCarry - who.Carrying
	}

	target.ResourcesLeft = target.ResourcesLeft - transfer
	who.Carrying = who.Carrying + transfer
}

func build(who *Character, target *House, dt time.Duration) {
	transfer := who.WorkPerSec * dt.Seconds()

	if transfer > target.MaxResources-target.ResourcesLeft {
		transfer = target.MaxResources - target.ResourcesLeft
	}
	if transfer > who.Carrying {
		transfer = who.Carrying
	}

	target.ResourcesLeft = target.ResourcesLeft + transfer
	who.Carrying = who.Carrying - transfer
}

func rerankHouse(house *House) {
	if house.ResourcesLeft == 0 {
		delete(house.Culture.BuiltHouses, house)
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
		if house.ResourcesLeft >= house.MaxResources {
			goto abandon // our work is done
		}
	} else { // Mining
		if who.Carrying >= who.MaxCarry {
			goto abandon // can't carry any more
		}
	}

	return

abandon:
	who.Target = nil
	return
}

func Tick(game Game, now time.Time) {
	dt := now.Sub(game.lastUpdate)

	// TODO shouldn't iterate by culture, or one team gets to move
	// before the other
	for _, culture := range game.cultures {
		for e := culture.Characters.Front(); e != nil; e = e.Next() {
			who := e.Value.(*Character)
			switch target := who.Target.(type) {
			case *Location:
				choice := chooseMove(
					who.Area.Location,
					*target,
					who.MovePerSec,
					dt,
				)
				attemptMove(who, game.terrain, choice)
			case *House:
				if insideOfShadow(who, target) {
					if who.Culture == target.Culture {
						build(who, target, dt)
					} else {
						mine(who, target, dt)
					}
					rerankHouse(target)
				} else {
					workTarget := Location{
						X: target.Area.Location.X + float64(target.Area.Width)/2,
						Y: target.Area.Location.Y + float64(target.Area.Height)/2,
					}
					choice := chooseMove(
						who.Area.Location,
						workTarget,
						who.MovePerSec,
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
