package main

import (
    "math"
    "math/rand"
    "sort"
    gl "github.com/chsc/gogl/gl21"
)

type ObjectType int;

const (
    Wall = ObjectType(1)
    Reduction = ObjectType(2)
    Rotation = ObjectType(3)
)

type Object struct {
    t ObjectType
    position float64
    width float64
    segments [6]bool
    newRotationSpeed float64
    newRotationSpeedRatio float64
}

type ObjectList []Object

type Game struct {
    pg *Playground
    bgradius    float64
    hexradius   float64
    trembler    float64
    myDistance  float64
    mySize      float64
    objects     ObjectList
}

type ReductionAnimationState struct {
    position    float64
    prevRotationSpeed float64
    d           float64
    orig        [6]float64
    dest        [6]float64
}

type BlastAnimationState struct {
    t float64
}

type GameState struct {
    game *Game
    rotz        float64
    rotationSpeed    float64
    bgcorners   [6]float64
    bgcolor     [3]gl.Float
    pog         float64 // point of game
    speed       float64
    myPosition  float64
    tremble     float64
    scanOffset  int
    gameOver    bool
    hit         bool
    perturbation float64
    reduction   *ReductionAnimationState
    blast       *BlastAnimationState
}

func (objects ObjectList) Len() int { return len(objects) }

func (objects ObjectList) Less(i, j int) bool {
    return objects[i].position < objects[j].position
}

func (objects ObjectList) Swap(i, j int) {
    objects[i], objects[j] = objects[j], objects[i]
}

func (game *Game) initScene() error {
    gl.Enable(gl.TEXTURE_2D)
    gl.Enable(gl.DEPTH_TEST)

    gl.ClearColor(0., 0., 0., 0.)
    gl.ClearDepth(1)
    gl.DepthFunc(gl.LEQUAL)

    gl.Viewport(0, 0, gl.Sizei(game.pg.width), gl.Sizei(game.pg.height))
    gl.MatrixMode(gl.PROJECTION)
    gl.LoadIdentity()
    gl.Frustum(-1, 1, -1, 1, 1.0, 10.0)
    gl.MatrixMode(gl.MODELVIEW)
    gl.LoadIdentity()
    return nil
}

func (game *Game) destroyScene() {
}

func (state *GameState) drawBackground() {
    gl.Begin(gl.TRIANGLES)
    defer gl.End()

    for i, c1 := range(state.bgcorners) {
        c2 := state.bgcorners[(i + 1) % len(state.bgcorners)]
        x1, y1 := gl.Float(math.Cos(c1) * state.game.bgradius),
                  gl.Float(math.Sin(c1) * state.game.bgradius)
        x2, y2 := gl.Float(math.Cos(c2) * state.game.bgradius),
                  gl.Float(math.Sin(c2) * state.game.bgradius)

        {
            var brightness gl.Float
            if i % 2 == 0 {
                brightness = .5
            } else {
                brightness = .8
            }
            gl.Color4f(state.bgcolor[0] * brightness, state.bgcolor[1] * brightness, state.bgcolor[2] * brightness, 1)
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

func (state *GameState) drawCentralHexagon() {
    gl.LineWidth(4.)
    defer gl.LineWidth(1.)

    gl.Begin(gl.LINE_LOOP)
    defer gl.End()
    gl.Color4f(state.bgcolor[0], state.bgcolor[1], state.bgcolor[2], 1.)

    r := state.game.hexradius + state.tremble

    for _, c := range(state.bgcorners) {
        x, y := gl.Float(math.Cos(c) * r),
                gl.Float(math.Sin(c) * r)
        gl.Vertex3f(x, y, 1)
    }
}

func (state *GameState) drawObjects() {
    gl.Begin(gl.QUADS)
    defer gl.End()

    hexradius := state.game.hexradius
    objects := state.game.objects

    gl.Color4f(state.bgcolor[0], state.bgcolor[1], state.bgcolor[2], 1.)

    hit := false

    for i := state.scanOffset; i < len(objects); i += 1 {
        object := objects[i]
        offset := object.position - state.pog
        width := object.width
        segments := object.segments
        if offset >= state.game.bgradius {
            break
        }
        if offset + width >= 0. {
            switch object.t {
            case Wall:
                for i, c1 := range state.bgcorners {
                    if segments[i] {
                        c2 := state.bgcorners[(i + 1) % len(state.bgcorners)]
                        corners := [4][2]float64 {
                            { math.Cos(c1) * math.Max(offset, hexradius),
                              math.Sin(c1) * math.Max(offset, hexradius), },
                            { math.Cos(c1) * math.Max((offset + width), hexradius),
                              math.Sin(c1) * math.Max((offset + width), hexradius), },
                            { math.Cos(c2) * math.Max((offset + width), hexradius),
                              math.Sin(c2) * math.Max((offset + width), hexradius), },
                            { math.Cos(c2) * math.Max(offset, hexradius),
                              math.Sin(c2) * math.Max(offset, hexradius), },
                        }
                        for _, corner := range(corners) {
                            gl.Vertex3f(gl.Float(corner[0]), gl.Float(corner[1]), 1)
                        }
                        hit = hit || state.hittest(corners)
                    }
                }
            case Reduction:
                if state.reduction == nil && offset < 0 {
                    ncorners := 6
                    for i := 0; i < 6; i += 1 {
                        if segments[i] { ncorners -= 1}
                    }
                    dest := [6]float64 {}
                    c := 0
                    for i := 0; i < 6; i += 1 {
                        dest[i] = math.Pi * 2. * float64(c) / float64(ncorners)
                        if !segments[i] { c += 1 }
                    }
                    state.reduction = &ReductionAnimationState {
                        position: state.pog,
                        prevRotationSpeed: state.rotationSpeed,
                        d: width,
                        orig: state.bgcorners,
                        dest: dest,
                    }
                    if object.newRotationSpeedRatio != 0 {
                        state.rotationSpeed *= object.newRotationSpeedRatio
                    } else {
                        state.rotationSpeed = object.newRotationSpeed
                    }
                }
            case Rotation:
                if offset < state.speed {
                    if object.newRotationSpeedRatio != 0 {
                        state.rotationSpeed *= object.newRotationSpeedRatio
                    } else {
                        state.rotationSpeed = object.newRotationSpeed
                    }
                }
            }
        } else {
            state.scanOffset = i
        }
    }
    state.hit = hit
}


func (state *GameState) myTriangleCoords(r float64) [3][2]float64 {
    t := state.tremble * .5
    p := state.myPosition
    o := state.game.hexradius + t + state.game.myDistance
    x, y := math.Cos(p) * o, math.Sin(p) * o
    return [3][2]float64 {
        { x + math.Cos(p) * r,
          y + math.Sin(p) * r, },
        { x + math.Cos(p + math.Pi * 2 / 3) * r,
          y + math.Sin(p + math.Pi * 2 / 3) * r, },
        { x + math.Cos(p + math.Pi * 4 / 3) * r,
          y + math.Sin(p + math.Pi * 4 / 3) * r }, }
}

func (state *GameState) drawMyTriangle() {
    gl.Begin(gl.TRIANGLES)
    defer gl.End()

    gl.Color4f(state.bgcolor[0], state.bgcolor[1], state.bgcolor[2], 1.)
    for _, p := range(state.myTriangleCoords(state.game.mySize)) {
        gl.Vertex3f(gl.Float(p[0]), gl.Float(p[1]), 1)
    }
}

func (state *GameState) drawBlast() {
    gl.LineWidth(2.)
    defer gl.LineWidth(1.)
    gl.Begin(gl.LINE_LOOP)
    defer gl.End()

    gl.Color4f(state.bgcolor[0], state.bgcolor[1], state.bgcolor[2], 1.)
    for _, p := range(state.myTriangleCoords(state.game.mySize + state.blast.t)) {
        gl.Vertex3f(gl.Float(p[0]), gl.Float(p[1]), 1)
    }
    state.blast.t += .005
    if state.blast.t >= state.game.mySize * 1.5 {
        state.blast.t = 0.
    }
}


func (state *GameState) drawScene() {
    gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

    gl.MatrixMode(gl.MODELVIEW)
    gl.LoadIdentity()
    gl.Rotatef(gl.Float(state.rotz), 0, 0, 1)
    gl.Translatef(0, 0, gl.Float(-3. + rand.Float64() * state.perturbation))

    state.drawBackground()
    state.drawCentralHexagon()
    state.drawObjects()
    state.drawMyTriangle()
    if state.gameOver {
        state.drawBlast()
    }
}

func (state *GameState) hittest(corners [4][2]float64) bool {
    points := state.myTriangleCoords(state.game.mySize)
    for _, point := range(points) {
        prevCorner := corners[len(corners) - 1]
        hit := 0
        for _, corner := range(corners) {
            if ((point[0] - prevCorner[0]) * (corner[1] - prevCorner[1]) -
                 (point[1] - prevCorner[1]) * (corner[0] - prevCorner[0]) >= 0) {
                hit += 1
            }
            prevCorner = corner
        }
        if (hit == 0 || hit == 4) { return true }
    }
    return false
}

func gameMain(pg Playground) error {
    game := Game {
        pg           : &pg,
        bgradius     : 3.5,
        hexradius    : .5,
        trembler     : .06,
        myDistance   : .1,
        mySize       : .025,
    }
    {
        objects := make(ObjectList, 100)
        for i, _ := range(objects) {
            c := i % 6
            objects[i] = Object {
                t        : Wall,
                position : 3 + float64(i) * .7,
                width    : .15,
                segments : [6]bool {
                    c == 0 || c == 1 || c == 3 || c == 4,
                    c == 1 || c == 2 || c == 4 || c == 5,
                    c == 2 || c == 2 || c == 5 || c == 0,
                    c == 3 || c == 4 || c == 0 || c == 1,
                    c == 4 || c == 5 || c == 1 || c == 2,
                    c == 5 || c == 0 || c == 2 || c == 3,
                },
            }
        }
        objects[49] = Object {
            t: Reduction,
            position: 13.15,
            width: .5,
            segments: [6]bool { false, false, false, true, false, false },
            newRotationSpeedRatio: 0.2,
        }
        objects[30] = Object {
            t: Reduction,
            position: 28.15,
            width: 1.25,
            segments: [6]bool { true, false, false, false, false, false },
            newRotationSpeedRatio: 0.2,
        }
        objects[13] = Object {
            t: Rotation,
            position: 5.5,
            newRotationSpeedRatio: -1,
        }
        objects[19] = Object {
            t: Rotation,
            position: 9.5,
            newRotationSpeedRatio: -1,
        }
        sort.Sort(objects)
        game.objects = objects
    }

    if err := game.initScene(); err != nil {
        return err
    }
    defer game.destroyScene()

    state := GameState {
        game        : &game,
        pog         : 0.,
        rotationSpeed    : math.Pi * .005,
        bgcorners   : [6]float64{},
        bgcolor     : [3]gl.Float{0.5, 0.8, 1.},
        myPosition  : 0.,
        speed       : .01,
        perturbation: .04,
    }
    {
        for i := 0; i < 6; i += 1 {
            state.bgcorners[i] = math.Pi * 2 * float64(i) / 6.
        }
    }

    pg.Run(func () {
        state.drawScene()

        if (!state.gameOver && state.hit) {
            state.gameOver = true
            state.blast = &BlastAnimationState { t: 0. }
        }
        if (state.gameOver) {
        } else {
            state.tremble = rand.Float64() * state.game.trembler

            // move against the rotation so the triangle virtually pauses
            // at the same position
            state.myPosition -= state.rotationSpeed
            if (pg.KeyPressed('Z')) {
               state.myPosition += .08
            }
            if (pg.KeyPressed('X')) {
               state.myPosition -= .08
            }

            state.pog += state.speed

            if state.reduction != nil {
                r := math.Min((state.pog - state.reduction.position) / state.reduction.d, 1.)
                for i := 0; i < 6; i += 1 {
                    state.bgcorners[i] = state.reduction.orig[i] * (1. - r) + state.reduction.dest[i] * r
                }
                if r >= 1. {
                    state.rotationSpeed = state.reduction.prevRotationSpeed
                    state.reduction = nil
                }
            }
        }
        state.rotz += state.rotationSpeed * 180 / math.Pi
    })
    return nil
}
