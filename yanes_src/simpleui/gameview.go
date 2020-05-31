package simpleui

import (
    "github.com/go-gl/gl/v2.1/gl"
    "github.com/go-gl/glfw/v3.1/glfw"
    "image"
    // "image/color"
    // "fmt"
    // "log"
    //"nes"
    // "nes/sql"
)

type CustomViewIF interface {
    View
    SetGLWindow(window *glfw.Window)
    SetAudioDevice(*Audio)
    OnKey(glfw.Key)
    TextureBuff() []uint8
    Title() string
}

const padding = 0

type GameView struct {
    director   *Director
    customView CustomViewIF
    texture    uint32

    ppuScreen  *image.RGBA
}

func NewGameView(director *Director, customView CustomViewIF) View {
    texture := createGLTexture()
    screen := image.NewRGBA(image.Rect(0, 0, width, height))
    return &GameView{director:director, 
                    customView:customView, 
                    texture:texture,
                    ppuScreen: screen}
}

func (view *GameView) Enter() {
    gl.ClearColor(0, 0, 0, 1)

    view.director.window.SetKeyCallback(view.onKey)

    if view.customView != nil {
        view.director.SetTitle(view.customView.Title() )
        view.customView.SetGLWindow(view.director.window)
        view.customView.SetAudioDevice(view.director.audio)
        view.customView.Enter()
    }

}

func (view *GameView) Exit() {
    view.director.window.SetKeyCallback(nil)

    if view.customView != nil {
        view.customView.Exit()
    }
}

// qibinyi, main update

func (view *GameView) Update(t, dt float64) {
    if dt > 1 {
        dt = 0
    }


    /*
        window := view.director.window


            if joystickReset(glfw.Joystick1) {
                view.director.ShowMenu()
            }
            if joystickReset(glfw.Joystick2) {
                view.director.ShowMenu()
            }
            if readKey(window, glfw.KeyEscape) {
                view.director.ShowMenu()
            }


        //*/

    if view.customView != nil {
        view.customView.Update(t, dt)

        copy(view.ppuScreen.Pix, view.customView.TextureBuff())

    }

    // console.DebugSingleFrame(true)
    gl.BindTexture(gl.TEXTURE_2D, view.texture)

    setTexture(view.ppuScreen)

    drawBuffer(view.director.window)
    gl.BindTexture(gl.TEXTURE_2D, 0)

}

func (view *GameView) onKey(window *glfw.Window,
    key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
    if action == glfw.Press {
        if view.customView != nil {
            view.customView.OnKey(key)
        }
    }
}

func drawBuffer(window *glfw.Window) {
    w, h := window.GetFramebufferSize()
    s1 := float32(w) / float32(width)
    s2 := float32(h) / float32(height)
    f := float32(1 - padding)
    var x, y float32
    if s1 >= s2 {
        x = f * s2 / s1
        y = f
    } else {
        x = f
        y = f * s1 / s2
    }
    gl.Begin(gl.QUADS)
    gl.TexCoord2f(0, 1)
    gl.Vertex2f(-x, -y)
    gl.TexCoord2f(1, 1)
    gl.Vertex2f(x, -y)
    gl.TexCoord2f(1, 0)
    gl.Vertex2f(x, y)
    gl.TexCoord2f(0, 0)
    gl.Vertex2f(-x, y)
    gl.End()
}
