package main

import (
	"log"
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/joeatwork/world-of-strategery/game"
)

func notSupportedMarshal(v interface{}) (msg []byte, payloadType byte, err error) {
	return nil, websocket.UnknownFrame, websocket.ErrNotSupported
}

func notSupportedUnmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	return websocket.ErrNotSupported
}

func statusMarshal(v interface{}) (msg []byte, payloadType byte, err error) {
	return []byte("STATUS"), websocket.TextFrame, nil
}

func ordersUnmarshal(msg []byte, payloadType byte, v interface{}) (err error) {
	orders := v.(*game.Orders)
	*orders = game.Orders{}
	return nil
}

func main() {
	ordersCodec := websocket.Codec{notSupportedMarshal, ordersUnmarshal}
	statusCodec := websocket.Codec{statusMarshal, notSupportedUnmarshal}

	gameLoop := game.RunGameLoop()

	handler := websocket.Handler(func(ws *websocket.Conn) {
		go func() {
			for !gameLoop.IsStopped() {
				status := gameLoop.ReadLatestStatus()
				if err := statusCodec.Send(ws, status); err != nil {
					log.Printf("can't write, %v", err)
					gameLoop.Stop()
				}
			}

			log.Printf("writer terminated")
		}()

		go func() {
			for !gameLoop.IsStopped() {
				var orders game.Orders
				if err := ordersCodec.Receive(ws, &orders); err != nil {
					log.Printf("can't read, %v", err)
					gameLoop.Stop()
				} else {
					gameLoop.WriteOrders(orders)
				}
			}

			log.Printf("reader terminated")
		}()
	})

	http.Handle("/game", handler)
}
