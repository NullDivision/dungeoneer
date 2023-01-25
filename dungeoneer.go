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
	header := fmt.Sprintf(
		"P:%d E:%d P$: %d E$: %d",
		len(game.player.units),
		len(game.enemy.units),
		game.player.money,
		game.enemy.money,
	)

	for i := range header {
		screen.SetContent(i, 0, rune(header[i]), nil, tcell.StyleDefault)
	}

	// Render map
	for i := range game.mapMatrix {
		for j := range game.mapMatrix[i] {
			entityChar := mapChar

			if i == game.enemy.castle.y && j == game.enemy.castle.x {
				entityChar = castleChar
			} else if i == game.player.castle.y && j == game.player.castle.x {
				entityChar = castleChar
			} else if i == game.playerAvatar.y && j == game.playerAvatar.x {
				entityChar = playerChar
			} else {
				for k := range game.enemy.units {
					if i == game.enemy.units[k].y && j == game.enemy.units[k].x {
						entityChar = unitChar
					}
				}
				for k := range game.player.units {
					if i == game.player.units[k].y && j == game.player.units[k].x {
						entityChar = unitChar
					}
				}
			}

			screen.SetContent(j, i+1, entityChar, nil, tcell.StyleDefault)
		}
	}
	screen.Sync()
}

func updateEntityTargets(game *game) {
	// player castle needs to check if there are any units to target
	if game.player.castle.target == nil {
		game.player.castle.target = game.player.castle.findTarget(game.enemy.units)
	}
	// enemy castle needs to check if there are any units to target
	if game.enemy.castle.target == nil {
		// Include the player in targeting
		game.enemy.castle.target = game.enemy.castle.findTarget(append(game.player.units, game.playerAvatar))
	}
	// Check if any of the player units are next to enemy units
	for i := range game.player.units {
		if game.player.units[i].target == nil {
			game.player.units[i].target = game.player.units[i].findTarget(game.enemy.units)
		}
	}
	// Check if any of the player units are next to enemy units
	for i := range game.enemy.units {
		if game.enemy.units[i].target == nil {
			// Include the player in targeting
			game.enemy.units[i].target = game.enemy.units[i].findTarget(append(game.player.units, game.playerAvatar))
		}
	}
	// If player has a target, check if it's still next to them and if so, hit it
	if game.playerAvatar.target == nil || !game.playerAvatar.isNearby(game.playerAvatar.target) {
		game.playerAvatar.target = game.playerAvatar.findTarget(append(game.enemy.units, &game.enemy.castle))
	}
}

func updateDamage(game *game) {
	// Hit target
	for i := range game.enemy.units {
		if game.enemy.units[i].target != nil {
			game.enemy.units[i].target.health -= 1
			if game.enemy.units[i].target.health <= 0 {
				game.enemy.units[i].target = nil
			}
		}
	}

	// Hit target
	for i := range game.player.units {
		if game.player.units[i].target != nil {
			game.player.units[i].target.health -= 1
			if game.player.units[i].target.health <= 0 {
				game.player.units[i].target = nil
			}
		}
	}

	if game.playerAvatar.target != nil {
		game.playerAvatar.target.health -= 1
	}
}

func processEntities(game *game) {
	// Targeting
	updateEntityTargets(game)

	// Move units
	for i := range game.enemy.units {
		log.Println("Enemy unit:", i, "target:", game.enemy.units[i].target)
		if game.enemy.units[i].target != nil {
			continue
		}

		if game.enemy.units[i].x >= game.enemy.units[i].y {
			game.enemy.units[i].x--
		} else {
			game.enemy.units[i].y--
		}
	}
	for i := range game.player.units {
		if game.player.units[i].target != nil {
			continue
		}

		if game.player.units[i].x <= game.player.units[i].y || game.player.units[i].y == game.enemy.castle.y {
			game.player.units[i].x++
		} else {
			game.player.units[i].y++
		}
	}

	log.Println("Player units:", len(game.player.units))
	log.Println("Enemy units:", len(game.enemy.units))

	updateDamage(game)

	// Check if any of the player units are dead
	for i := len(game.player.units) - 1; i >= 0; i-- {
		if game.player.units[i].health <= 0 {
			game.enemy.money++
			game.player.units = append(game.player.units[:i], game.player.units[i+1:]...)
		}
	}

	// Check if any of the enemy units are dead
	for i := len(game.enemy.units) - 1; i >= 0; i-- {
		if game.enemy.units[i].health <= 0 {
			game.player.money++
			game.enemy.units = append(game.enemy.units[:i], game.enemy.units[i+1:]...)
		}
	}

	// If player dies, move them back to the castle and reset their health
	if game.playerAvatar.health <= 0 {
		game.enemy.money += 5
		game.playerAvatar.location = game.player.castle.location
		game.playerAvatar.health = game.playerAvatar.maxHealth
	}
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
				game.playerAvatar.y--
			case KeyDown:
				game.playerAvatar.y++
			case KeyLeft:
				game.playerAvatar.x--
			case KeyRight:
				game.playerAvatar.x++
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
				game.spawnUnits()
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
