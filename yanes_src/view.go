package main

import (
    "image"
    "image/color"
    "nes/ppu"
    "nes"
    "github.com/mebusy/simpleui"
    "fmt"
    "github.com/go-gl/glfw/v3.1/glfw"
)

type NesView struct {
    console *nes.Bus
    window *glfw.Window

    screenImage *image.RGBA
}

func NewNesView( w,h int) *NesView {
    view := &NesView{  }
    view.screenImage = image.NewRGBA(image.Rect(0, 0, w, h))
    return view
}


func (self *NesView) Title( ) string {
    return "Yet Another NES Emulator in Go"
}
func (self *NesView) SetGLWindow( window *glfw.Window ) {
    self.window = window
}
func (self *NesView) SetAudioDevice( audio *simpleui.Audio ) {
    self.console.SetAudioChannel(audio.GetAudioChannel())
    self.console.SetAudioSampleRate(int(audio.GetSampleRate()))
}
func (self *NesView) Enter() {
    // if err := view.console.LoadState(savePath(view.hash)); err == nil {
    //  return
    // } else {
    self.console.Reset()
    // }

    // if load state fail , then load sram
    // load sram
    self.console.GetCartrideg().Load( workingDir )
}

// to handle some speical event
func (self *NesView) OnKey( key glfw.Key ) {
    switch key {
    case glfw.KeyP:
        simpleui.Screenshot( "./debug/" , self.screenImage  )
    case glfw.KeyR:
        self.console.Reset()
    case glfw.KeyT:
        DumpNameTable2File( "./debug/", self.console.GetPpu().DumpNameTables() )
    }
}

func (self *NesView) updateControllers() {
    /*
        turbo := console.PPU.Frame%6 < 3
        k1 := readKeys(window, turbo)
        j1 := readJoystick(glfw.Joystick1, turbo)
        j2 := readJoystick(glfw.Joystick2, turbo)
        console.SetButtons1(combineButtons(k1, j1))
        console.SetButtons2(j2)
        //*/
    cid := 0
    self.console.Controller[cid] = 0

    if simpleui.ReadKey(self.window, glfw.KeyD) {
        self.console.Controller[cid] |= 1 // right
    }
    if simpleui.ReadKey(self.window, glfw.KeyA) {
        self.console.Controller[cid] |= 2 // left
    }
    if simpleui.ReadKey(self.window, glfw.KeyS) {
        self.console.Controller[cid] |= 4 // down
    }
    if simpleui.ReadKey(self.window, glfw.KeyW) {
        self.console.Controller[cid] |= 8 // up
    }
    if simpleui.ReadKey(self.window, glfw.KeyEnter) {
        self.console.Controller[cid] |= 0x10 // start
    }
    if simpleui.ReadKey(self.window, glfw.KeyRightShift) {
        self.console.Controller[cid] |= 0x20 // select
    }
    if simpleui.ReadKey(self.window, glfw.KeyJ) {
        self.console.Controller[cid] |= 0x40 //
    }
    if simpleui.ReadKey(self.window, glfw.KeyK) {
        self.console.Controller[cid] |= 0x80 //
    }
}

func (self *NesView) Exit() {
    self.console.SetAudioChannel(nil)
    self.console.SetAudioSampleRate(0)
    // save sram
    self.console.GetCartrideg().Save( workingDir )
    // sql.Close()
    // save state
    // view.console.SaveState(savePath(view.hash))
}

var fps_cnt = 0
var time_start float64
var framePerSecond int

func (self *NesView) Update(t, dt float64 ) {
    self.updateControllers()
    self.console.StepSeconds(dt)

    fps_cnt ++
    if time_start == 0 {
        time_start = t
    } else {
        framePerSecond =  int( float64(fps_cnt) /  (t - time_start) )
    }

}


func (self *NesView) TextureBuff() []uint8  {
    var y_off int

    {
        spr := self.console.GetPpu().GetScreenSnapshot()
        /*
        copy( self.screenImage.Pix , spr.Pix )
        /*/
        screenWidth := self.screenImage.Bounds().Size().X
        dst_stride := screenWidth << 2
        src_stride := spr.Width << 2

        simpleui.CopyStride( self.screenImage.Pix[ y_off * dst_stride : ], dst_stride, spr.Pix , src_stride , src_stride, spr.Height )
        //*/

        y_off += spr.Height+1

    }

    {
        palSpr := self.console.GetPpu().GetPaletteSprite()
        screenWidth := self.screenImage.Bounds().Size().X
        dst_stride := screenWidth << 2
        src_stride := palSpr.Width << 2

        simpleui.CopyStride( self.screenImage.Pix[ y_off * dst_stride : ], dst_stride, palSpr.Pix , src_stride , src_stride, palSpr.Height )

        y_off += palSpr.Height+1
    }

    // simpleui.DrawText( self.screenImage , 0,0, "ABC" , color.White )
    // simpleui.DrawCenteredText( self.screenImage , 0,0, "uvw" , color.White )
    oam := self.console.GetPpu().OAM
    cnt := 0
    for i:=0; i<len(oam); i+=4 {
        idx_spr := i/4
        objattr := ppu.ObjAttribEntry{ oam[i],oam[i+1],oam[i+2],oam[i+3] }
        if cnt >= 40 {
            break
        }
        str := fmt.Sprintf( "%2d %02x,%02x %02x %02x", idx_spr, objattr.X,objattr.Y, objattr.ID,objattr.Attribute )
        simpleui.DrawText( self.screenImage , 256+1, 0 + simpleui.FONT_SIZE*(cnt) , str , color.White )
        cnt++
    }

    // fps
    simpleui.DrawText( self.screenImage , 0,0, fmt.Sprintf( "fps:%2d", framePerSecond ) , color.White )

    return self.screenImage.Pix
    //*/
}
