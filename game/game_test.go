package game

import (
	"testing"
	"time"
)

var loc0x0 = Location{
	X: 0.0,
	Y: 0.0,
}

var loc0x4 = Location{
	X: 0.0,
	Y: 4.0,
}

var loc3x0 = Location{
	X: 3.0,
	Y: 0.0,
}

var loc3x4 = Location{
	X: 3.0,
	Y: 4.0,
}

var loc6x8 = Location{
	X: 6.0,
	Y: 8.0,
}

func TestChooseMoveZeroTime(t *testing.T) {
	dt0, _ := time.ParseDuration("0ns")
	choice := chooseMove(loc3x4, loc6x8, 2, dt0)
	if choice != loc3x4 {
		t.Errorf("expected no motion, got %v", choice)
	}
}

func TestChooseMovePositive(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc6x8, 5, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveNegative(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc6x8, loc0x0, 5, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveOvershot(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc3x4, 10, dt)
	if choice != loc3x4 {
		t.Errorf("expected %v, got %v", loc3x4, choice)
	}
}

func TestChooseMoveXOnly(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc3x0, 3, dt)
	if choice != loc3x0 {
		t.Errorf("expected %v, got %v", loc3x0, choice)
	}
}

func TestChooseMoveYOnly(t *testing.T) {
	dt, _ := time.ParseDuration("1s")
	choice := chooseMove(loc0x0, loc0x4, 4, dt)
	if choice != loc0x4 {
		t.Errorf("expected %v, got %v", loc0x4, choice)
	}
}
