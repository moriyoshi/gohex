package main

import (
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
)

type Playground struct {
    width int
    height int
    title string
}

func (pg Playground) Run(f func()) {
	for glfw.WindowParam(glfw.Opened) == 1 {
        f()
		glfw.SwapBuffers()
	}
}

func (pg Playground) KeyPressed(key int) bool {
    return glfw.Key(key) == glfw.KeyPress
}

func launch(pg Playground) {
	glfw.OpenWindowHint(glfw.WindowNoResize, 1)

	if err := glfw.OpenWindow(pg.width, pg.height, 0, 0, 0, 0, 16, 0, glfw.Windowed); err != nil {
		showError(err)
		return
	}
	defer glfw.CloseWindow()

	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle(pg.title)

	if err := gl.Init(); err != nil {
		showError(err)
        return
	}

    if err := gameMain(pg); err != nil {
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

    launch(Playground{ 640, 480, "HexaGOn" })
}
