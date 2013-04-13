package main

import (
    "math"
    "math/rand"
	gl "github.com/chsc/gogl/gl21"
)

type Wall struct {
    segment int
    position float64
    width float64
}

var (
	rotz        float64
    rotSpeed    float64     = math.Pi * 0.02
    bgdivision  int         = 6
    bgcolor     []gl.Float  = []gl.Float{1, 1, 0}
    bgradius    float64     = 20
    hexradius   float64     = .5
    perturbation float64   = .04
    walls       []Wall      = nil
    pog         float64     = 0.
    myPosition  float64     = 0.
    myDistance  float64     = .1
)

func initScene(game Game) (err error) {
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.DEPTH_TEST)

	gl.ClearColor(0., 0., 0., 0.)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	gl.Viewport(0, 0, gl.Sizei(game.width), gl.Sizei(game.height))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Frustum(-1, 1, -1, 1, 1.0, 10.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	return
}

func destroyScene() {
}

func drawBackground() {
	gl.Begin(gl.TRIANGLES)
    defer gl.End()

    for i := 0; i < bgdivision; i += 1 {
        x1, y1 := gl.Float(math.Cos(float64(i) * math.Pi * 2 / float64(bgdivision)) * bgradius),
                  gl.Float(math.Sin(float64(i) * math.Pi * 2 / float64(bgdivision)) * bgradius)
        x2, y2 := gl.Float(math.Cos(float64(i + 1) * math.Pi * 2/ float64(bgdivision)) * bgradius),
                  gl.Float(math.Sin(float64(i + 1) * math.Pi * 2 / float64(bgdivision)) * bgradius)

        {
            var brightness gl.Float
            if i % 2 == 0 {
                brightness = .5
            } else {
                brightness = .8
            }
            gl.Color4f(bgcolor[0] * brightness, bgcolor[1] * brightness, bgcolor[2] * brightness, 1)
        }

        gl.Normal3f(0, 0, 1)
        gl.TexCoord2f(0, 0)
        gl.Vertex3f(0, 0, 1)
        gl.TexCoord2f(1, 0)
        gl.Vertex3f(x1, y1, 1)
        gl.TexCoord2f(1, 1)
        gl.Vertex3f(x2, y2, 1)
    }
}

func drawCentralHexagon() {
	gl.Begin(gl.LINE_LOOP)
	defer gl.End()
    gl.Color4f(bgcolor[0], bgcolor[1], bgcolor[2], 1.)
    gl.LineWidth(16.)

    for i := 0; i < bgdivision; i += 1 {
        x, y := gl.Float(math.Cos(float64(i) * math.Pi * 2 / float64(bgdivision)) * hexradius),
                gl.Float(math.Sin(float64(i) * math.Pi * 2 / float64(bgdivision)) * hexradius)
        gl.Vertex3f(x, y, 1)
    }
}

func drawWalls() {
	gl.Begin(gl.QUADS)
	defer gl.End()

    gl.Color4f(bgcolor[0], bgcolor[1], bgcolor[2], 1.)

    for i := 0; i < len(walls); i += 1 {
        wall := walls[i]
        segment := wall.segment
        offset := wall.position - pog
        width := wall.width
        if offset < bgradius && offset + width >= 0. {
            x1, y1 := gl.Float(math.Cos(float64(segment) * math.Pi * 2 / float64(bgdivision)) * math.Max(offset, hexradius)),
                      gl.Float(math.Sin(float64(segment) * math.Pi * 2 / float64(bgdivision)) * math.Max(offset, hexradius))
            x2, y2 := gl.Float(math.Cos(float64(segment + 1) * math.Pi * 2/ float64(bgdivision)) * math.Max(offset, hexradius)),
                      gl.Float(math.Sin(float64(segment + 1) * math.Pi * 2 / float64(bgdivision)) * math.Max(offset, hexradius))
            x3, y3 := gl.Float(math.Cos(float64(segment) * math.Pi * 2 / float64(bgdivision)) * math.Max((offset + width), hexradius)),
                      gl.Float(math.Sin(float64(segment) * math.Pi * 2 / float64(bgdivision)) * math.Max((offset + width), hexradius))
            x4, y4 := gl.Float(math.Cos(float64(segment + 1) * math.Pi * 2/ float64(bgdivision)) * math.Max((offset + width), hexradius)),
                      gl.Float(math.Sin(float64(segment + 1) * math.Pi * 2 / float64(bgdivision)) * math.Max((offset + width), hexradius))
            gl.Vertex3f(x1, y1, 1)
            gl.Vertex3f(x3, y3, 1)
            gl.Vertex3f(x4, y4, 1)
            gl.Vertex3f(x2, y2, 1)
        }
    }
}

func drawMyTriangle() {
	gl.Begin(gl.TRIANGLES)
	defer gl.End()

    p := myPosition
    o := hexradius + myDistance
    s := .05
    x1, y1 := math.Cos((p - s) * math.Pi * 2 / float64(bgdivision)) * o,
              math.Sin((p - s) * math.Pi * 2 / float64(bgdivision)) * o
    x2, y2 := math.Cos((p + s) * math.Pi * 2 / float64(bgdivision)) * o,
              math.Sin((p + s) * math.Pi * 2 / float64(bgdivision)) * o
    z := math.Sqrt(math.Pow(x1 - x2, 2) + math.Pow(y1 - y2, 2))
    x3, y3 := math.Cos(p * math.Pi * 2 / float64(bgdivision)) * (o + z),
              math.Sin(p * math.Pi * 2 / float64(bgdivision)) * (o + z)

    gl.Color4f(bgcolor[0], bgcolor[1], bgcolor[2], 1.)
    gl.Vertex3f(gl.Float(x1), gl.Float(y1), 1)
    gl.Vertex3f(gl.Float(x2), gl.Float(y2), 1)
    gl.Vertex3f(gl.Float(x3), gl.Float(y3), 1)
}

func drawScene() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Rotatef(gl.Float(rotz), 0, 0, 1)
	gl.Translatef(0, 0, gl.Float(-3. + rand.Float64() * perturbation))

	rotz += rotSpeed * 180 / math.Pi
    pog += .02

    drawBackground()
    drawCentralHexagon()
    drawWalls()
    drawMyTriangle()
}

func gameMain(game Game) error {
	if err := initScene(game); err != nil {
		return err
	}
	defer destroyScene()

    walls = make([]Wall, 100)
    var j int = 0
    for i := 0; i < len(walls) / 2; i += 1 {
        walls[j] = Wall { i % 6, float64(i) * .7, .15 }; j += 1
        walls[j] = Wall { (i + 3) % 6, float64(i) * .7, .15 }; j += 1
    }
    game.Run(func () {
        drawScene()
        // rotate against the rotation so the triangle virtually pauses
        // at the same position
        myPosition -= rotSpeed * 0.95
        if (game.KeyPressed('Z')) {
           myPosition += .06
        }
        if (game.KeyPressed('X')) {
           myPosition -= .06
        }
    })
    return nil
}
