package main

import (
    "fmt"
    "image"
    "image/color"
    "strings"

    "github.com/golang/freetype/truetype"
    "golang.org/x/image/font"
    "golang.org/x/image/font/gofont/gomono"

    "log"

    "github.com/hajimehoshi/ebiten"
    "github.com/hajimehoshi/ebiten/text"
)

var (
    uiImage       *ebiten.Image 
    uiFont        font.Face
    // uiFontMHeight int
    uiFontMWidth int
    uiFontSize int  = 12
)

const (
    ANCHOR_LEFT int = 1<< iota
    ANCHOR_RIGHT
    ANCHOR_TOP
    ANCHOR_BOTTOM
)

func init() {
    tt, err := truetype.Parse(gomono.TTF)
    if err != nil {
        log.Fatal(err)
    }
    uiFont = truetype.NewFace(tt, &truetype.Options{
        Size:    float64(uiFontSize),
        DPI:     72,
        Hinting: font.HintingFull,
    })
    // b, _, _ := uiFont.GlyphBounds('M')
    // uiFontMHeight = (b.Max.Y - b.Min.Y).Ceil()
    uiFontMWidth = textWidth("M")

    // log.Println( "uiFontMHeight:", uiFontMHeight, "uiFontSize:" , uiFontSize  )
    uiImage, err = ebiten.NewImage(16*uiFontMWidth,16*uiFontSize, ebiten.FilterDefault)
    if err != nil {
        log.Fatal( err )
    }
    // uiImage.Fill(color.Black)
    for i:=0;i<256; i++ {
        y := (i/16) * uiFontSize - 2 + uiFontSize
        x := (i%16) * uiFontMWidth
        // log.Println(i,uiFontMWidth, x,y)
        text.Draw( uiImage, fmt.Sprintf( "%c",i ) , uiFont, x,y, color.White )
    }
}
func textWidth(str string) int {
    maxW := 0
    for _, line := range strings.Split(str, "\n") {
        b, _ := font.BoundString(uiFont, line)
        w := (b.Max.X - b.Min.X).Ceil()
        if maxW < w {
            maxW = w
        }
    }
    return maxW
}


func drawText( img *ebiten.Image , str string, x,y int, color color.Color, anchor int ) int {
    sw,sh := img.Size()
    if anchor & ANCHOR_RIGHT != 0 {
        x = sw - x - len( str )*uiFontMWidth
    }

    if anchor & ANCHOR_BOTTOM != 0 {
        y = sh - y - uiFontSize
    }

    _drawUIText( img, str, x,y, color )
    return len(str)*uiFontMWidth
}

func _drawUIText( img *ebiten.Image , str string, x,y int, color color.Color ) {
    for i:=0 ;i<len(str) ;i++ {
        ch := int(str[i])
        suby := (ch/16) * uiFontSize
        subx := (ch%16) * uiFontMWidth
        subImage := uiImage.SubImage(image.Rect( subx ,suby , subx+ uiFontMWidth,suby+ uiFontSize)).(*ebiten.Image)
        drawImage( img, subImage , x + i*uiFontMWidth  ,y, 0, color )
    }
}

func drawImage( img_src *ebiten.Image, img_dst  *ebiten.Image , x,y int ,
        anchor int, color color.Color ) {
    sw,sh := img_src.Size()
    dw,dh := img_dst.Size()
    op := &ebiten.DrawImageOptions{}
    dst_x := x
    dst_y := y
    if anchor & ANCHOR_RIGHT != 0 {
        dst_x = sw - x - dw
    }
    if anchor & ANCHOR_BOTTOM != 0 {
        dst_y = sh - y - dh
    }
    // log.Println( dst_x, dst_y  )
    op.GeoM.Translate( float64(dst_x),float64(dst_y) )
    r,g,b,_ := color.RGBA()
    op.ColorM.Scale( float64(r)/65535, float64(g)/65535, float64(b)/65535, 1 )
    img_src.DrawImage( img_dst , op )
}

