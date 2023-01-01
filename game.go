package main

type game struct {
	enemyCastle  entity
	enemyMoney   int
	enemyUnits   []*entity
	paused       bool
	player       *entity
	playerCastle entity
	playerMoney  int
	playerUnits  []*entity
	mapMatrix    [][]int
}

func isEndState(game game) bool {
	if game.playerCastle.health <= 0 {
		return true
	}

	if game.enemyCastle.health <= 0 {
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
		enemyCastle:  enemyCastle,
		player:       &entity{health: 2, maxHealth: 2},
		playerCastle: playerCastle,
		mapMatrix:    mapMatrix,
	}
}
