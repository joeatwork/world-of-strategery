package game

import (
	"sync"
	"time"
)

// Interact with the GameLoop via
// 1) possibly blocking, buffered writes of Orders
// 2) Quick, "last snapshot" reads of Gamestatus
type GameLoop struct {
	status     GameStatus
	orders     chan<- Orders
	stopped    bool
	statusLock sync.RWMutex
	stopLock   sync.RWMutex
}

func (l *GameLoop) ReadLatestStatus() GameStatus {
	l.statusLock.RLock()
	defer l.statusLock.RUnlock()
	return l.status
}

func (l *GameLoop) WriteOrders(orders Orders) {
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
	orders := make(chan Orders)
	shared := &GameLoop{}
	shared.status = ReadStatus(g)
	shared.orders = orders

	go func() {
		lastTime := time.Now().UnixNano() / 1000000
		for shared.IsStopped() {
			select {
			case orders := <-orders:
				ApplyOrders(g, orders)
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
