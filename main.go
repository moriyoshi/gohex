package main

import (
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
)

type Game struct {
    width int
    height int
    title string
}

func (game Game) Run(f func()) {
	for glfw.WindowParam(glfw.Opened) == 1 {
        f()
		glfw.SwapBuffers()
	}
}

func (game Game) KeyPressed(key int) bool {
    return glfw.Key(key) == glfw.KeyPress
}

func launch(game Game) {
	glfw.OpenWindowHint(glfw.WindowNoResize, 1)

	if err := glfw.OpenWindow(game.width, game.height, 0, 0, 0, 0, 16, 0, glfw.Windowed); err != nil {
		showError(err)
		return
	}
	defer glfw.CloseWindow()

	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle(game.title)

	if err := gl.Init(); err != nil {
		showError(err)
        return
	}

    if err := gameMain(game); err != nil {
        showError(err)
        return
    }
}

func main() {
	if err := glfw.Init(); err != nil {
		showError(err)
		return
	}
	defer glfw.Terminate()

    launch(Game{ 640, 480, "HexaGOn" })
}
