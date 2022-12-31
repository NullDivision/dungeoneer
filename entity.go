package main

type entity struct {
	x, y int
}

func (e1 entity) isOverlapping(e2 entity) bool {
	return e1.x == e2.x && e1.y == e2.y
}
