package main

import (
    "nes"
    "image"
    "image/png"
    // "image/color"
    "os"
    "log"
    "fmt"
)


func generatePatternImage( bus *nes.Bus ) {
    for i:=0; i<2 ;i++ {
        sprPattern := bus.DumpCurrentPatternTable( i,0 )

        width := sprPattern.Width
        height := sprPattern.Height
        img := image.NewRGBA(image.Rect(0, 0, width , height ))

        copy( img.Pix, sprPattern.Pix )

        f, err := os.Create( fmt.Sprintf( "pattern_%02d.png" , i ) )
        if err != nil {
            log.Fatal(err)
        }

        if err := png.Encode(f, img); err != nil {
            f.Close()
            log.Fatal(err)
        }

        if err := f.Close(); err != nil {
            log.Fatal(err)
        }
    }
}

func generatePaletteImage( bus *nes.Bus ) {
    spr := bus.DumpPaletteSprite()
    width := spr.Width
    height := spr.Height
    img := image.NewRGBA(image.Rect(0, 0, width , height ))
    copy( img.Pix, spr.Pix )

    f, err := os.Create( "palette.png"  )
    if err != nil {
        log.Fatal(err)
    }

    if err := png.Encode(f, img); err != nil {
        f.Close()
        log.Fatal(err)
    }

    if err := f.Close(); err != nil {
        log.Fatal(err)
    }

}



