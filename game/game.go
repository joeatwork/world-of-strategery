package game

import (
	"container/list"
	"log"
	"time"
)

type Terrain struct {
	Board      [][]interface{}
	XMax, YMax int
}

type Location struct {
	X, Y float32
}

type Area struct {
	Location      Location
	Height, Width uint8
}

type Character struct {
	Culture    *Culture
	MovePerSec float32
	WorkPerSec float32
	Area       Area
	Carrying   float32
	MaxCarry   float32
	OffsetX    float32
	OffsetY    float32
	Target     interface{}
}

type House struct {
	Culture       *Culture
	Area          Area
	ResourcesLeft float32 // TODO - float resources means rounding errors!
	MaxResources  float32
}

type Culture struct {
	Characters    list.List // <House>
	PlannedHouses map[*House]bool
	BuiltHouses   map[*House]bool
}

type Game struct {
	cultures   []Culture
	lastUpdate time.Time
}

// Propose a new location given an origin, target, speed and duration.
func chooseMove(from Location, to Location,
	speedPerSec float32, dt time.Duration) (float32, float32) {
	var dirX, dirY float32
	deltaX := from.X - to.X
	deltaY := from.Y - to.Y
	absX := deltaX
	absY := deltaY
	if deltaX < 0 {
		absX = -deltaX
	}
	if deltaY < 0 {
		absY = -deltaY
	}

	if absX < absY {
		if deltaY < 0 {
			dirX, dirY = 0, -1
		}
		if deltaY > 0 {
			dirX, dirY = 0, 1
		}
	} else {
		if deltaX < 0 {
			dirX, dirY = -1, 0
		}
		if deltaX > 0 {
			dirX, dirY = 1, 0
		}
	}

	moveX := dirX * speedPerSec / float32(dt.Seconds())
	moveY := dirY * speedPerSec / float32(dt.Seconds())

	return moveX, moveY
}

// Attempts to move the character from their current position to
// moveX, moveY. May move the character a shorter distance, or no
// distance at all.
func attemptMove(who *Character, terrain *Terrain, moveX, moveY float32) {
	moveX += who.OffsetX
	moveY += who.OffsetY
	for moveX > 1 && moveY > 1 && moveX < -1 && moveY < -1 {
		origin := who.Area.Location
		width := who.Area.Width
		height := who.Area.Height

		destX := int(origin.X)
		destY := int(origin.Y)
		if moveX > 0 {
			destX, moveX = destX+1, moveX-1
		}
		if moveX < 0 {
			destX, moveX = destX-1, moveX+1
		}
		if moveY > 0 {
			destY, moveY = destY+1, moveY-1
		}
		if moveY < 0 {
			destY, moveY = destY-1, moveY+1
		}

		if destX < 0 {
			destX = 0
		}
		if destY < 0 {
			destY = 0
		}
		if destX+width > terrain.XMax {
			destX = MaxCoord - width
		}
		if destY+height > terrain.YMax {
			destY = MaxCoord - height
		}

		if destX != origin.X || destY != origin.Y {
			clear := true
		WHILECLEAR:
			for x := 0; x < width; x++ {
				for y := 0; y < height; y++ {
					tryX := destX + x
					tryY := destY + y
					occupant := terrain.Board[tryX][tryY]
					if nil != occupant || who != occupant {
						clear = false
						break WHILECLEAR
					}
				}
			}

			if clear {
				for x := 0; x < width; x++ {
					for y := 0; y < height; y++ {
						oldX := origin.X + x
						oldY := origin.Y + y
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

				who.Area.Location.X = destX
				who.Area.Location.Y = destY
			} // if clear
		} // if dest != origin
	} // for move

	if moveX >= 1 || moveX <= -1 || moveY >= 1 || moveY <= -1 {
		panic("proposed move not consumed in the attempt")
	}

	who.OffsetX = moveX
	who.OffsetY = moveY
}

const shadowSize float32 = 1

func insideOfShadow(who *Character, target *House) {
	shadowXMin := target.Area.Location.X - shadowSize
	shadowYMin := target.Area.Location.Y - shadowSize
	shadowXMax := target.Area.Location.X + target.Area.Width + shadowSize
	shadowYMax := target.Area.Location.Y + target.Area.Height + shadowSize
	whoXMin := who.Area.Location.X
	whoYMin := who.Area.Location.Y
	whoXMax := who.Area.Location.X + who.Area.Width
	whoYMax := who.Area.Location.Y + who.Area.Height
	xOverlap :=
		(whoXMin > shadowXMin && whoXMin < shadowXMax) ||
			(whoXMax > shadowXMin && whoXMax < shadowXMax)

	if xOverlap {
		return (whoYMin > shadowYMin && whoYMin < shadowYMax) ||
			(whoYMax > shadowYMin && whoYMax < shadowYMax)
	}
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

	if transfer > target.MaxResource-target.ResourcesLeft {
		transfer = target.MaxResource - target.ResourcesLeft
	}
	if transfer > who.Carrying {
		transfer = who.Carrying
	}

	target.ResourcesLeft = target.ResourcesLeft + transfer
	who.Carrying = who.Carrying - transfer
}

func rerankHouse(house *House) {
	if house.ResourcesLeft == 0 {
		delete(Culture.BuiltHouses, house)
	}
}

func reevaluateTargetHouse(who *Character) {
	house := who.Target.(*House)
	if _, ok := house.Culture.PlannedHouses[house]; !ok {
		if _, ok := house.Culture.BuildHouses[house]; !ok {
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
	dt := now - game.lastUpdate

	// TODO shouldn't iterate by culture, or one team gets to move
	// before the other
	for _, culture := range Game.cultures {
		for e := culture.Characters.Front(); e != nil; e = e.Next() {
			who := e.Value.(*Character)
			switch target := who.Target.(type) {
			case *Location:
				x, y := chooseMove(
					who.Area.Location,
					target,
					who.MovePerSec,
					dt,
				)
				attemptMove(who, terrain, x, y)
			case *House:
				if insideOfShadow(who, target) {
					if who.Culture == target.Culture {
						build(who, target)
					} else {
						mine(who, target)
					}
					rerankHouse(target)
				} else {
					workTarget := Location{
						X: target.Area.Location.X + target.Area.Width/2,
						Y: target.Area.Locatoin.Y + target.Area.Height/2,
					}
					x, y := chooseMove(
						who.Area.Location,
						workTarget,
						who.MovePerSec,
						dt,
					)
					attemptMove(who, terrain, x, y)
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
