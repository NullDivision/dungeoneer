package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

const mapChar = '.'
const playerChar = '@'
const castleChar = '#'
const unitChar = 'U'

type entity struct {
	x, y int
}

// Get the terminal with defaults coming from the environment.
// Env values are "WINDOW_WIDTH", "WINDOW_HEIGHT"
func getWindowSize(screen tcell.Screen) (int, int) {
	width, height := screen.Size()
	envWidth, err := strconv.Atoi(os.Getenv("WINDOW_WIDTH"))

	if err != nil {
		envWidth = width
	}

	envHeight, err := strconv.Atoi(os.Getenv("WINDOW_HEIGHT"))

	if err != nil {
		envHeight = height
	}

	return envWidth, envHeight
}

func exit(screen tcell.Screen) {
	screen.Fini()
	os.Exit(0)
}

type game struct {
	enemyCastle  entity
	enemyUnits   []*entity
	player       *entity
	playerCastle entity
	playerUnits  []*entity
	mapMatrix    [][]int
}

func update(game game, screen tcell.Screen) {
	// Check collisions
	if game.player.x == game.enemyCastle.x && game.player.y == game.enemyCastle.y {
		exit(screen)
	}

	// Move units
	for i := range game.enemyUnits {
		if game.enemyUnits[i].x > game.enemyUnits[i].y {
			game.enemyUnits[i].x--
		} else {
			game.enemyUnits[i].y--
		}
	}

	for i := range game.playerUnits {
		if game.playerUnits[i].x < game.playerUnits[i].y {
			game.playerUnits[i].x++
		} else {
			game.playerUnits[i].y++
		}
	}

	// Render map
	for i := range game.mapMatrix {
		for j := range game.mapMatrix[i] {
			entityChar := mapChar

			if i == game.enemyCastle.y && j == game.enemyCastle.x {
				entityChar = castleChar
			} else if i == game.playerCastle.y && j == game.playerCastle.x {
				entityChar = castleChar
			} else if i == game.player.y && j == game.player.x {
				entityChar = playerChar
			} else {
				for k := range game.enemyUnits {
					if i == game.enemyUnits[k].y && j == game.enemyUnits[k].x {
						entityChar = unitChar
					}
				}
				for k := range game.playerUnits {
					if i == game.playerUnits[k].y && j == game.playerUnits[k].x {
						entityChar = unitChar
					}
				}
			}

			screen.SetContent(j, i, entityChar, nil, tcell.StyleDefault)
		}
	}
	screen.Sync()
}

func spawnUnits(game *game) {
	game.enemyUnits = append(game.enemyUnits, &entity{game.enemyCastle.x, game.enemyCastle.y})
	game.playerUnits = append(game.playerUnits, &entity{game.playerCastle.x, game.playerCastle.y})
}

func run(keyChannel chan tcell.Key, screen tcell.Screen) {
	width, height := getWindowSize(screen)
	player := entity{}
	playerCastle := entity{}
	enemyCastle := entity{width - 1, height - 1}
	mapMatrix := make([][]int, height)
	for i := range mapMatrix {
		mapMatrix[i] = make([]int, width)
	}
	game := game{
		enemyCastle:  enemyCastle,
		enemyUnits:   make([]*entity, 0),
		player:       &player,
		playerCastle: playerCastle,
		playerUnits:  make([]*entity, 0),
		mapMatrix:    mapMatrix,
	}
	// Create a ticker to update the game
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case ev := <-keyChannel:
			switch ev {
			case tcell.KeyEscape:
				exit(screen)
			case tcell.KeyUp:
				player.y--
			case tcell.KeyDown:
				player.y++
			case tcell.KeyLeft:
				player.x--
			case tcell.KeyRight:
				player.x++
			}
			update(game, screen)
		case tick := <-ticker.C:
			if tick.Second()%5 == 0 {
				spawnUnits(&game)
			}

			update(game, screen)
		}
	}
}

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}

	if err := screen.Init(); err != nil {
		panic(err)
	}

	width, height := getWindowSize(screen)

	screen.SetSize(width, height)

	mapMatrix := make([][]int, height)
	for i := range mapMatrix {
		mapMatrix[i] = make([]int, width)
	}

	// Create a channel to listen for events
	keyChannel := make(chan tcell.Key)

	go run(keyChannel, screen)

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			keyChannel <- ev.Key()
		}
	}
}
