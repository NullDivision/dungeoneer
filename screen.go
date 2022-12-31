package main

import (
	"os"
	"strconv"

	"github.com/gdamore/tcell/v2"
)

func showMessage(screen tcell.Screen, message string) {
	for i := range message {
		screen.SetContent(i, 0, rune(message[i]), nil, tcell.StyleDefault)
	}
	screen.Sync()
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

func makeScreen() (tcell.Screen, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return screen, err
	}
	if err = screen.Init(); err != nil {
		panic(err)
	}

	return screen, err
}
