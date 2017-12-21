package game

import (
	"sync"
	"time"
)

// GameLoop manages an ongoing game as a concurrent process. You can ask the
// GameLoop for a snapshot of the game state, or send the GameLoop orders that
// it will relay into the game it contains.
type GameLoop struct {
	status     GameStatus
	orders     chan<- []Order
	stopped    bool
	statusLock sync.RWMutex
	stopLock   sync.RWMutex
}

// ReadLatestStatus returns a (possibly out of date) snapshot of the game status.
func (l *GameLoop) ReadLatestStatus() GameStatus {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

func (l *GameLoop) WriteOrders(orders []Order) {
	l.orders <- orders
}

func (l *GameLoop) Stop() {
	l.stopLock.Lock()
	defer l.stopLock.Unlock()
	l.stopped = true
}

func (l *GameLoop) IsStopped() bool {
	l.stopLock.RLock()
	defer l.stopLock.RUnlock()
	return l.stopped
}

func RunGameLoop() *GameLoop {
	g := &Game{}
	orders := make(chan []Order)
	shared := &GameLoop{}
	shared.status = ReadStatus(g)
	shared.orders = orders

	go func() {
		lastTime := time.Now().UnixNano() / 1000000
		for shared.IsStopped() {
			select {
			case orders := <-orders:
				for _, o := range orders {
					o.Apply(g)
				}
			default:
			}

			thisTime := time.Now().UnixNano()
			dt := float64(thisTime - lastTime)
			lastTime = thisTime
			Tick(g, dt)

			// TODO readStatus needs to be cheap, or needs to be on-demand
			workingStatus := ReadStatus(g)

			shared.statusLock.Lock()
			shared.status = workingStatus
			shared.statusLock.Unlock()
		}
	}()

	return shared
}
