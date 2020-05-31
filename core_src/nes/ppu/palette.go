package ppu

import (
    "nes/color"
    "nes/sprite"
)

func (self *Ppu) GetPaletteSprite( ) *sprite.Sprite {
    width := self.sprPalette.Height
    height := width
    entryPerPalette := 4+1

    for palette:=0 ; palette < 8 ; palette++ {
        for entry:=0 ; entry < entryPerPalette ; entry ++ {
            // now to draw each color block
            clr := self.GetColourFromPaletteRam(palette, uint8(entry)%4)
            if entry == 4 {
                clr = color.COLOC_BLACK
            }

            x_off := (palette*entryPerPalette + entry) * width
            for y:=0; y<height; y++ {
                for x:=0; x< width; x++ {
                    if y > 0 && y < height -1 {
                        continue   // ignore middle lines
                    }
                    if entry == 0 && y== height-1 && x==0 {
                        self.sprPalette.SetPixel (x_off + x,y, color.COLOC_RED )
                    } else {
                        self.sprPalette.SetPixel (x_off + x,y, clr )
                    }
                }
            }

            stride := self.sprPalette.Width * 4
            for y:= 1; y<height-1; y++ {
                copy( self.sprPalette.Pix[ stride*y: stride*(y+1) ] , self.sprPalette.Pix  )
            }

        }
    }
    return self.sprPalette
}
