package main

import (
    "image/color"
    "fmt"

    "github.com/hajimehoshi/ebiten"
    // "github.com/hajimehoshi/ebiten/ebitenutil"
    "github.com/hajimehoshi/ebiten/inpututil"

    "nes"
    "nes/cpu"
    // "time"
    // "log"
)

// Game implements ebiten.Game interface.
type Game struct{
    Design_Width, Design_Height int
    Width, Height int
    nes  *nes.Bus

    imgPattern [2]*ebiten.Image
    imgPalette *ebiten.Image
    imgPalIndicator *ebiten.Image
    imgScreen *ebiten.Image

    imgPageZero *ebiten.Image
    imgPagePRG *ebiten.Image

    imgRegisterStatus *ebiten.Image
    imgDisassembler *ebiten.Image

    bEmulationRun bool

    debugPalette int

    PC int
    cpu cpu.Cpu
}


func NewGame() *Game {
    game := &Game{
        Width: 720*WINDOW_SCALE,
        Height: 480*WINDOW_SCALE,
        nes : nes.NewBus(),
    }

    game.imgPageZero, _ =  ebiten.NewImage( 440,196, ebiten.FilterDefault )
    game.imgPagePRG, _ =  ebiten.NewImage( 440,196, ebiten.FilterDefault )

    game.imgRegisterStatus, _ = ebiten.NewImage( 256,76, ebiten.FilterDefault )
    game.imgDisassembler, _ = ebiten.NewImage( 256,240, ebiten.FilterDefault )

    // game.InsertCartridge( "../f1.nes" )
    // game.InsertCartridge( "../smb.nes" )
    // game.InsertCartridge( "../donkeykong.nes" )
    // game.InsertCartridge( "../nestest.nes" )
    game.InsertCartridge( "../doae2.nes" )
    game.ResetNES()
    game.UpdatePatternTable( -1 )

    startSoundSampleLoop( game.nes )

    return game
}

// Update proceeds the game state.
// Update is called every tick (1/60 [s] by default).
func (g *Game) Update(screen *ebiten.Image) error {

    // Write your game's logical update.
    if inpututil.IsKeyJustPressed(ebiten.KeySpace ) {
        g.bEmulationRun = !g.bEmulationRun
    }
    debug := true
    if debug {
        if inpututil.IsKeyJustPressed( ebiten.KeyE ) {
            g.debugPalette = (g.debugPalette+1) % 8
            g.UpdatePatternTable( g.debugPalette )
        }
    }
    g.UpdatePalette()


    g.cpu = g.nes.DebugDumpCpu()
    g.PC = int(g.cpu.PC)
    if g.bEmulationRun {

        g.UpdateController()
        g.nes.DebugSingleFrame(true)
    } else {
        if inpututil.IsKeyJustPressed( ebiten.KeyS ) {
            g.nes.DebugStepInstruction()
        } else if inpututil.IsKeyJustPressed( ebiten.KeyF ) {
            g.nes.DebugSingleFrame(true)
        } else if inpututil.IsKeyJustPressed( ebiten.KeyR ) {
            g.nes.Reset()
        }

    }

    return nil
}

// Draw draws the game screen.
// Draw is called every frame (typically 1/60[s] for 60Hz display).
func (g *Game) Draw(screen *ebiten.Image) {
    // Write your game's rendering.

    if g.bEmulationRun {
        g.UpdateScreen()
        if g.imgScreen != nil {
            drawImage( screen, g.imgScreen , 0, uiFontSize, 0, color.White )
        }
    } else {
        // draw zero page
        g.draw1PageMemory( g.imgPageZero, 0)
        drawImage( screen, g.imgPageZero, 0, uiFontSize , 0 , color.White )

        // ============= draw program data ===================
        start_addr := (g.PC&^0x3F) - 6* 16
        if start_addr < 0 {
            start_addr = 0
        }
        _,h := g.imgPageZero.Size()
        g.draw1PageMemory( g.imgPagePRG, start_addr )
        drawImage( screen, g.imgPagePRG, 0, uiFontSize + h + 8 , 0 , color.White )

    }

    // draw pattern image
    for i, img := range g.imgPattern {
        if img != nil {
            w,_ := img.Size()
            drawImage( screen, img, (1-i)*w ,0, ANCHOR_BOTTOM | ANCHOR_RIGHT , color.White)
        }
    }
    if g.imgPalette != nil {
        drawImage( screen, g.imgPalette, 0 , 132, ANCHOR_BOTTOM | ANCHOR_RIGHT , color.White)
    }
    if g.imgPalIndicator != nil {
        w,_ := g.imgPalIndicator.Size()
        drawImage( screen, g.imgPalIndicator, (8-1-g.debugPalette)*w , 132, ANCHOR_BOTTOM | ANCHOR_RIGHT , color.White)
    }

    // draw register status 
    g.UpdateRegisterStatus()
    drawImage( screen, g.imgRegisterStatus, 0,uiFontSize, ANCHOR_TOP | ANCHOR_RIGHT , color.White)

    g.UpdateDisassembler()
    drawImage( screen, g.imgDisassembler, 0,uiFontSize+78, ANCHOR_TOP | ANCHOR_RIGHT , color.White)

    // ebitenutil.DebugPrint(screen , fmt.Sprintf( "%d", g.nes.SystemClockCounter )   )
    drawText( screen , fmt.Sprintf( "fps:%2d clock: %d", int(ebiten.CurrentFPS()), g.nes.SystemClockCounter ), 0,0,color.White, 0 )

    // draw help info
    drawText( screen, "Space:switch run/stop  S:single Step F:single Frame", 0, 0 , color.White , ANCHOR_BOTTOM)

    // drawImage( screen, uiImage, 0,0,0, color.White)
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
// If you don't have to adjust the screen size with the outside size, just return a fixed size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    return g.Width/WINDOW_SCALE, g.Height/WINDOW_SCALE
}



func startSoundSampleLoop( nes *nes.Bus ) {
    sampleRate := 44100
    channelNum := 1
    bitDepthInBytes := 1
    buffSizeInBytes := 512


    go func() {
        // ticker := time.NewTicker( time.Second / time.Duration(sampleRate)   )
        soundPlayer = NewSoundPlayer( sampleRate, channelNum, bitDepthInBytes, buffSizeInBytes )
        defer soundPlayer.Close()

        p := soundPlayer.context.NewPlayer()
        apu := nes.GetApu()
        // apu.SetHasAudioComsumer(true)
        sample := make( []byte, 1,1 )
        for {
            sample[0] = byte((<-apu.Ch_sample)*127)
            // log.Println( sample )
            _ = sample
            p.Write( sample )
        }
    }()
}

func SoundOut( nChannel int, fGlobalTime float64, fTimeStep float64 ) {

}
