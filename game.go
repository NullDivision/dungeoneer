package main

type player struct {
	castle entity
	money  int
	units  []*entity
}

type game struct {
	enemy        *player
	paused       bool
	player       *player
	playerAvatar *entity
	mapMatrix    [][]int
}

func (g *game) spawnUnits() {
	g.enemy.units = append(
		g.enemy.units,
		&entity{health: 1, location: g.enemy.castle.location, maxHealth: 1},
	)
	g.player.units = append(
		g.player.units,
		&entity{health: 1, location: g.player.castle.location, maxHealth: 1},
	)
}

func isEndState(game game) bool {
	if game.player.castle.health <= 0 {
		return true
	}

	if game.enemy.castle.health <= 0 {
		return true
	}

	return false
}

func makeNewGame(width, height int) game {
	enemyCastle := entity{
		health:    1,
		maxHealth: 1,
		location:  location{width - 1, height - headerHeight - 1},
	}
	playerCastle := entity{health: 1, maxHealth: 1}
	mapMatrix := make([][]int, height-headerHeight)
	for i := range mapMatrix {
		mapMatrix[i] = make([]int, width)
	}

	return game{
		enemy:        &player{castle: enemyCastle},
		player:       &player{castle: playerCastle},
		playerAvatar: &entity{health: 2, maxHealth: 2},
		mapMatrix:    mapMatrix,
	}
}
