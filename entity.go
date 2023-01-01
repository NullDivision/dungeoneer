package main

type location struct {
	x, y int
}

type entity struct {
	health int
	location
	target    *entity
	maxHealth int
}

func (e entity) isNearby(target *entity) bool {
	// Check if target is in a 3x3 centered around the entity
	isHorizontallyAdjacent := target.x >= e.x-1 && target.x <= e.x+1
	isVerticallyAdjacent := target.y >= e.y-1 && target.y <= e.y+1

	return isHorizontallyAdjacent && isVerticallyAdjacent
}

func (e entity) findTarget(entities []*entity) *entity {
	for _, target := range entities {
		if e.isNearby(target) {
			return target
		}
	}

	return nil
}
