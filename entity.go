package main

type location struct {
	x, y int
}

func (e1 location) isOverlapping(e2 location) bool {
	return e1.x == e2.x && e1.y == e2.y
}

type entity struct {
	health int
	location
	maxHealth int
}
