package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	mapChar    = '.'
	playerChar = '@'
	castleChar = '#'
	unitChar   = 'U'
)
const headerHeight = 1

type playerKey uint8

const (
	KeyUp playerKey = iota
	KeyDown
	KeyLeft
	KeyRight
	KeyEscape
	KeyPause
)

func exit(screen tcell.Screen) {
	screen.Fini()
	os.Exit(0)
}

type game struct {
	enemyCastle  entity
	enemyUnits   []*entity
	paused       bool
	player       *entity
	playerCastle entity
	playerUnits  []*entity
	mapMatrix    [][]int
}

func renderMap(game game, screen tcell.Screen) {
	header := fmt.Sprintf("P:%d E:%d", len(game.playerUnits), len(game.enemyUnits))

	for i := range header {
		screen.SetContent(i, 0, rune(header[i]), nil, tcell.StyleDefault)
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

			screen.SetContent(j, i+1, entityChar, nil, tcell.StyleDefault)
		}
	}
	screen.Sync()
}

func moveNpcs(game *game) {
	// Move units
	for i := range game.enemyUnits {
		if game.enemyUnits[i].x >= game.enemyUnits[i].y {
			game.enemyUnits[i].x--
		} else {
			game.enemyUnits[i].y--
		}
	}

	for i := range game.playerUnits {
		if game.playerUnits[i].x <= game.playerUnits[i].y || game.playerUnits[i].y == game.enemyCastle.y {
			game.playerUnits[i].x++
		} else {
			game.playerUnits[i].y++
		}
	}
}

func isEndState(game game) bool {
	if game.player.isOverlapping(game.enemyCastle) {
		return true
	}

	// Check if any of the player units are on the enemy castle
	for i := range game.playerUnits {
		if game.playerUnits[i].isOverlapping(game.enemyCastle) {
			return true
		}
	}

	// Check if any of the enemy units are on the player castle
	for i := range game.enemyUnits {
		if game.enemyUnits[i].isOverlapping(game.playerCastle) {
			return true
		}
	}

	return false
}

func update(game *game, screen tcell.Screen, isTick bool) {
	if isTick {
		moveNpcs(game)
	}

	// Check collisions
	if isEndState(*game) {
		exit(screen)
	}

	log.Println("Player units:", len(game.playerUnits))
	log.Println("Enemy units:", len(game.enemyUnits))

	// Eliminate unit if it's on the same tile as the player
	for i := range game.enemyUnits {
		if game.enemyUnits[i].isOverlapping(*game.player) {
			game.enemyUnits = append(game.enemyUnits[:i], game.enemyUnits[i+1:]...)
		}
	}

	// Eliminate both units if they're on the same tile
	for i := len(game.playerUnits) - 1; i >= 0; i-- {
		for j := len(game.enemyUnits) - 1; j >= 0; j-- {
			if game.playerUnits[i].isOverlapping(*game.enemyUnits[j]) {
				game.playerUnits = append(game.playerUnits[:i], game.playerUnits[i+1:]...)
				game.enemyUnits = append(game.enemyUnits[:j], game.enemyUnits[j+1:]...)
			}
		}
	}

	// Render map
	renderMap(*game, screen)
}

func spawnUnits(game *game) {
	game.enemyUnits = append(game.enemyUnits, &entity{game.enemyCastle.x, game.enemyCastle.y})
	game.playerUnits = append(game.playerUnits, &entity{game.playerCastle.x, game.playerCastle.y})
}

func run(keyChannel chan playerKey, screen tcell.Screen) {
	width, height := getWindowSize(screen)
	player := entity{}
	playerCastle := entity{}
	enemyCastle := entity{width - 1, height - headerHeight - 1}
	mapMatrix := make([][]int, height-headerHeight)
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
		// TODO: abstract this away somehow
		select {
		case ev := <-keyChannel:
			switch ev {
			case KeyEscape:
				exit(screen)
			case KeyUp:
				player.y--
			case KeyDown:
				player.y++
			case KeyLeft:
				player.x--
			case KeyRight:
				player.x++
			case KeyPause:
				showMessage(screen, "Game paused")
				game.paused = true
			default:
				showMessage(screen, "Unknown key"+string(ev))
				game.paused = true
			}

			if !game.paused {
				update(&game, screen, false)
			}
		case tick := <-ticker.C:
			if game.paused {
				continue
			}

			log.Println("Tick", tick)

			if tick.Second()%5 == 0 {
				spawnUnits(&game)
			}

			update(&game, screen, true)
		}
	}
}

func handleKeyboardEvents(screen tcell.Screen, keyChannel chan playerKey) {
	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape:
				keyChannel <- KeyEscape
			case tcell.KeyUp:
				keyChannel <- KeyUp
			case tcell.KeyDown:
				keyChannel <- KeyDown
			case tcell.KeyLeft:
				keyChannel <- KeyLeft
			case tcell.KeyRight:
				keyChannel <- KeyRight
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					keyChannel <- KeyEscape
				case 'p':
					keyChannel <- KeyPause
				}
			}
		}
	}
}

func main() {
	screen, err := makeScreen()
	if err != nil {
		panic(err)
	}

	width, height := getWindowSize(screen)

	screen.SetSize(width, height)

	mapMatrix := make([][]int, height)
	for i := range mapMatrix {
		mapMatrix[i] = make([]int, width)
	}

	// Create a channel to listen for events
	keyChannel := make(chan playerKey)
	file, err := os.Create("debug.log")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	log.SetOutput(file)

	log.Println("Starting game")

	go run(keyChannel, screen)
	handleKeyboardEvents(screen, keyChannel)
}
