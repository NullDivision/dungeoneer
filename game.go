package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
)

const mapChar = '.'
const playerChar = '@'

type player struct {
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

func render(mapMatrix [][]int, player player, screen tcell.Screen) {
	// Render map
	for i := range mapMatrix {
		for j := range mapMatrix[i] {
			if i == player.y && j == player.x {
				screen.SetContent(j, i, playerChar, nil, tcell.StyleDefault)
				continue
			}

			screen.SetContent(j, i, mapChar, nil, tcell.StyleDefault)
		}
	}
	screen.Sync()
}

func run(keyChannel chan tcell.Key, screen tcell.Screen, player player) {
	width, height := getWindowSize(screen)
	mapMatrix := make([][]int, height)
	for i := range mapMatrix {
		mapMatrix[i] = make([]int, width)
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
			render(mapMatrix, player, screen)
		case <-ticker.C:
			render(mapMatrix, player, screen)
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

	// Init player
	player := player{0, 0}

	// Create a channel to listen for events
	keyChannel := make(chan tcell.Key)

	go run(keyChannel, screen, player)

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			keyChannel <- ev.Key()
		}
	}
}
