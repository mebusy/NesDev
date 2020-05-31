package main

import (
    "log"
    "fmt"
    "nes/ppu"
    // "image"
    "image/color"
    "nes/cart"
    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/ebitenutil"
    "github.com/hajimehoshi/ebiten/inpututil"
    "os/user"
    "path"
)


var homeDir string

func init() {
	u, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	homeDir = u.HomeDir
}

func (g *Game) InsertCartridge( filename string ) {
    cartridge := cart.NewCartridge( filename )
    cartridge.Load( path.Join( homeDir, ".yanes" ) )
    g.nes.InsertCartridge( cartridge )
}

func (g *Game) ResetNES() {
    g.nes.Reset()
}

func (g *Game) UpdatePalette( ) {
    // ==============================
    spr := g.nes.GetPpu().GetPaletteSprite()
    if g.imgPalette == nil {
        var err error
        g.imgPalette, err = ebiten.NewImage( spr.Width, spr.Height , ebiten.FilterDefault  )
        if err != nil {
            log.Fatal( err )
        }
    }
    g.imgPalette.ReplacePixels( spr.Pix )
    // ==============================
    if g.imgPalIndicator == nil {
        var err error
        g.imgPalIndicator, err = ebiten.NewImage( spr.Width/8, spr.Height , ebiten.FilterDefault  )
        if err != nil {
            log.Fatal( err )
        }

        // update only once
        w,h := g.imgPalIndicator.Size()
        /*
        for i:=0 ; i<w*4/5 ;i++ {
            g.imgPalIndicator.Set( i,0 , color.White )
            g.imgPalIndicator.Set( i,h-1 , color.White )
        }
        for i:=0 ; i<h; i++ {
            g.imgPalIndicator.Set( 0,i , color.White )
            g.imgPalIndicator.Set( w*4/5-1+1 , i , color.White )
        }
        */
        _ = h
        ebitenutil.DrawLine( g.imgPalIndicator , 0,0, float64(w*4/5),0, color.White )
        ebitenutil.DrawLine( g.imgPalIndicator , 0,float64(h-1), float64(w*4/5),float64(h-1), color.White )
        ebitenutil.DrawLine( g.imgPalIndicator, 1,0, 1, float64(h-1), color.White )
        ebitenutil.DrawLine( g.imgPalIndicator, float64(w*4/5),0, float64(w*4/5), float64(h-1), color.White )
    }
}


func (g *Game) UpdatePatternTable( pal int) {
    if pal < 0 {
        pal = g.debugPalette
    }
    for i, img := range g.imgPattern {
        spr := g.nes.GetPpu().GetPatternTable( i,pal )
        if img == nil {
            var err error
            g.imgPattern[i], err = ebiten.NewImage( spr.Width, spr.Height , ebiten.FilterDefault  )
            if err != nil {
                log.Fatal( err )
            }
            img = g.imgPattern[i]
        }
        img.ReplacePixels( spr.Pix )
    }


}


func (g *Game) draw1PageMemory( img *ebiten.Image,  start_addr int ) {
    w, h := img.Size()
    img.Clear()
    ebitenutil.DrawRect(img,0,0, float64(w),float64(h), color.White )
    ebitenutil.DrawRect(img,1,1, float64(w-2),float64(h-2), color.Black )

    x := 0
    cur_y := 0
    margin := 2
    for j :=0 ; j< 16 ; j++ {
        for i :=0 ; i< 16 ; i++ {

            addr := uint16(j*16 + i + start_addr )
            if  i == 0 {
                x = 0
                // drawString( fmt.Sprintf( "$%04X: " , addr ) )
                str := fmt.Sprintf( "$%04X:" , addr )
                drawText(img, str , margin+ x , margin+cur_y*uiFontSize, color.White, 0 )
                x += textWidth(str) + uiFontSize
                x&^=7
            }

            val := g.nes.CpuRead( addr, false  )
            var col color.Color = color.White
            if addr == uint16(g.PC) {
                col = color.RGBA{ R:0, G:255, B:255, A:255}
            }
            str := fmt.Sprintf( "%02X" , val )
            drawText(img,str  , margin+x  , margin+cur_y*uiFontSize, col, 0 )
            x += textWidth(str) + uiFontSize
            x&^=7
        }
        cur_y++
    }
}

var cpu_flags = []string{ "N","V","-","B","D","I","Z","C" }
func (g *Game) UpdateRegisterStatus(  ) {
    img := g.imgRegisterStatus
    w, h := img.Size()
    img.Clear()
    ebitenutil.DrawRect(img,0,0, float64(w),float64(h), color.White )
    ebitenutil.DrawRect(img,1,1, float64(w-2),float64(h-2), color.Black )

    margin := 2
    x := margin
    y := margin
    x += drawText(  img, "STATUS: ", x, y, color.White, 0 )
    x += uiFontSize
    for i,v := range cpu_flags {
        col := color.RGBA{ R:255, G:0, B:0, A:255 }
        if g.cpu.Status & ( 1<<(7-i) ) != 0 {
            col = color.RGBA{ R:0, G:255, B:0, A:255 }
        }
        // drawStringFg( fmt.Sprintf( "%s ",v ), termbox.ColorRed  )
        x += drawText( img, fmt.Sprintf( "%s ",v ), x , y , col, 0 )
        x += uiFontSize
    }

    y += uiFontSize
    x = margin
    drawText( img, fmt.Sprintf("PC: $%04X", g.PC), x,y ,color.White ,0  )  ; y += uiFontSize
    drawText( img, fmt.Sprintf("A:  $%02X [%3d]", g.cpu.A, g.cpu.A), x,y ,color.White ,0  )  ; y += uiFontSize
    drawText( img, fmt.Sprintf("X:  $%02X [%3d]", g.cpu.X, g.cpu.X), x,y ,color.White ,0  )  ; y += uiFontSize
    drawText( img, fmt.Sprintf("Y:  $%02X [%3d]", g.cpu.Y, g.cpu.Y), x,y ,color.White ,0  )  ; y += uiFontSize
    drawText( img, fmt.Sprintf("SP: $%02X [%3d]", g.cpu.SP, g.cpu.SP), x,y ,color.White ,0  )  ; y += uiFontSize

}

func (g *Game) UpdateScreen( ) {
    spr := g.nes.GetPpu().GetScreenSnapshot()
    if g.imgScreen == nil {
        var err error
        g.imgScreen, err = ebiten.NewImage( spr.Width, spr.Height , ebiten.FilterDefault  )
        if err != nil {
            log.Fatal(err)
        }
    }
    //*
    g.imgScreen.ReplacePixels(spr.Pix)
    /*/
    // debug 
    imgP0 := g.imgPattern[0]
    if imgP0 == nil {
        return
    }
    tblName := g.nes.DumpNameTalbe0()
    g.imgScreen.Clear()
    for y:=0; y<30; y++ {
        for x:=0; x<32; x++ {
            id := tblName[ y*32 + x ]
            subx := int(id&0xF) << 3
            suby := int(id>>4) << 3
            subImage := imgP0.SubImage(image.Rect( subx ,suby , subx+ 8,suby+ 8)).(*ebiten.Image)
            // g.imgScreen.DrawImage
            drawImage( g.imgScreen ,  subImage , x*8, y*8, 0 )
        }
    }
    //*/
}

func (g *Game) UpdateDisassembler() {
    img := g.imgDisassembler
    w, h := img.Size()
    img.Clear()
    ebitenutil.DrawRect(img,0,0, float64(w),float64(h), color.White )
    ebitenutil.DrawRect(img,1,1, float64(w-2),float64(h-2), color.Black )

    if g.bEmulationRun {
        oam := g.nes.GetPpu().OAM
        for i:=0; i<len(oam); i+=4 {
            objattr := ppu.ObjAttribEntry{ oam[i],oam[i+1],oam[i+2],oam[i+3] }
            drawText(img, fmt.Sprintf( "%2d: (%02X,%02X) ID:%02X, AT:%02X", i/4,
                objattr.X,objattr.Y, objattr.ID,objattr.Attribute ), 4, 4 + i*3, color.White , 0  )
        }
    } else {
        // show disassemble
    }
}

func (g *Game) UpdateController() {
    cid := 0
    g.nes.Controller[cid] = 0

    sensitive := 0
    if inpututil.KeyPressDuration( ebiten.KeyD) > sensitive {
        g.nes.Controller[cid] |= 1  // right
    }
    if inpututil.KeyPressDuration( ebiten.KeyA) > sensitive {
        g.nes.Controller[cid] |= 2  // left
    }
    if inpututil.KeyPressDuration( ebiten.KeyS) > sensitive {
        g.nes.Controller[cid] |= 4  // down
    }
    if inpututil.KeyPressDuration( ebiten.KeyW) > sensitive {
        g.nes.Controller[cid] |= 8  // up
    }
    if inpututil.IsKeyJustPressed( ebiten.KeyEnter) {
        g.nes.Controller[cid] |= 0x10  // start
    }
    if inpututil.IsKeyJustPressed( ebiten.KeyShift) {
        g.nes.Controller[cid] |= 0x20  // select
    }
    if inpututil.KeyPressDuration( ebiten.KeyJ) > sensitive {
        g.nes.Controller[cid] |= 0x40  // 
    }
    if inpututil.KeyPressDuration( ebiten.KeyK ) > sensitive {
        g.nes.Controller[cid] |= 0x80  // 
    }
}

