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

func processEntities(game *game) {
	// player castle needs to check if there are any units to target
	if game.playerCastle.target == nil {
		game.playerCastle.target = game.playerCastle.findTarget(game.enemyUnits)
	}
	// enemy castle needs to check if there are any units to target
	if game.enemyCastle.target == nil {
		// Include the player in targeting
		game.enemyCastle.target = game.enemyCastle.findTarget(append(game.playerUnits, game.player))
	}

	// Check if any of the player units are next to enemy units
	for i := range game.playerUnits {
		if game.playerUnits[i].target == nil {
			game.playerUnits[i].target = game.playerUnits[i].findTarget(game.enemyUnits)
		}
	}
	// Check if any of the player units are next to enemy units
	for i := range game.enemyUnits {
		if game.enemyUnits[i].target == nil {
			// Include the player in targeting
			game.enemyUnits[i].target = game.enemyUnits[i].findTarget(append(game.playerUnits, game.player))
		}
	}

	// Move units
	for i := range game.enemyUnits {
		log.Println("Enemy unit:", i, "target:", game.enemyUnits[i].target)
		if game.enemyUnits[i].target != nil {
			continue
		}

		if game.enemyUnits[i].x >= game.enemyUnits[i].y {
			game.enemyUnits[i].x--
		} else {
			game.enemyUnits[i].y--
		}
	}

	for i := range game.playerUnits {
		if game.playerUnits[i].target != nil {
			continue
		}

		if game.playerUnits[i].x <= game.playerUnits[i].y || game.playerUnits[i].y == game.enemyCastle.y {
			game.playerUnits[i].x++
		} else {
			game.playerUnits[i].y++
		}
	}

	log.Println("Player units:", len(game.playerUnits))
	log.Println("Enemy units:", len(game.enemyUnits))

	// Hit target
	for i := range game.enemyUnits {
		if game.enemyUnits[i].target != nil {
			game.enemyUnits[i].target.health -= 1
			if game.enemyUnits[i].target.health <= 0 {
				game.enemyUnits[i].target = nil
			}
		}
	}

	// Hit target
	for i := range game.playerUnits {
		if game.playerUnits[i].target != nil {
			game.playerUnits[i].target.health -= 1
			if game.playerUnits[i].target.health <= 0 {
				game.playerUnits[i].target = nil
			}
		}
	}

	// If player has a target, check if it's still next to them and if so, hit it
	if game.player.target == nil || !game.player.isNearby(game.player.target) {
		game.player.target = game.player.findTarget(append(game.enemyUnits, &game.enemyCastle))
	}
	if game.player.target != nil {
		game.player.target.health -= 1
	}

	// Check if any of the player units are dead
	for i := len(game.playerUnits) - 1; i >= 0; i-- {
		if game.playerUnits[i].health <= 0 {
			game.playerUnits = append(game.playerUnits[:i], game.playerUnits[i+1:]...)
		}
	}

	// Check if any of the enemy units are dead
	for i := len(game.enemyUnits) - 1; i >= 0; i-- {
		if game.enemyUnits[i].health <= 0 {
			game.enemyUnits = append(game.enemyUnits[:i], game.enemyUnits[i+1:]...)
		}
	}

	// If player dies, move them back to the castle and reset their health
	if game.player.health <= 0 {
		game.player.location = game.playerCastle.location
		game.player.health = game.player.maxHealth
	}
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

func update(game *game, screen tcell.Screen, isTick bool) {
	if isTick {
		processEntities(game)
	}

	// Check collisions
	if isEndState(*game) {
		exit(screen)
	}

	// Render map
	renderMap(*game, screen)
}

func spawnUnits(game *game) {
	game.enemyUnits = append(
		game.enemyUnits,
		&entity{health: 1, location: game.enemyCastle.location, maxHealth: 1},
	)
	game.playerUnits = append(
		game.playerUnits,
		&entity{health: 1, location: game.playerCastle.location, maxHealth: 1},
	)
}

func run(keyChannel chan playerKey, screen tcell.Screen) {
	width, height := getWindowSize(screen)
	game := makeNewGame(width, height)
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
				game.player.y--
			case KeyDown:
				game.player.y++
			case KeyLeft:
				game.player.x--
			case KeyRight:
				game.player.x++
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
