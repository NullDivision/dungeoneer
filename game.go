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
	player       *entity
	playerCastle entity
	mapMatrix    [][]int
}

func render(game game, screen tcell.Screen) {
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
			}

			screen.SetContent(j, i, entityChar, nil, tcell.StyleDefault)
		}
	}
	screen.Sync()
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
	game := game{enemyCastle, &player, playerCastle, mapMatrix}
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
			render(game, screen)
		case <-ticker.C:
			render(game, screen)
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
