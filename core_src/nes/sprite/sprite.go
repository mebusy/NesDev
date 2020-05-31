package sprite

import (
    "nes/color"
    // "log"
)

// represent a 2-D RGBA image
type Sprite struct {
    Width int
    Height int
    Pix []uint8
}

func NewSprite( width, height int  ) *Sprite {
    sprite := &Sprite{ Width:width, Height:height }
    sprite.Pix = make( []uint8, sprite.Width * sprite.Height * 4 )
    return sprite
}

func (self *Sprite) SetPixel( x,y int , color color.COLOR  ) {
    if x < 0 || y < 0 {
        return
    }
    if x >= self.Width || y >= self.Height {
        return
    }

    offset := (self.Width*y + x) * 4
    // log.Println( x,y,offset )
    //*
    self.Pix[offset+0] = color.R
    self.Pix[offset+1] = color.G
    self.Pix[offset+2] = color.B
    self.Pix[offset+3] = 0xFF

    // if bTransparent {
    //     self.Pix[offset+3] = 0x00
    // } else {
    //     self.Pix[offset+3] = 0xFF
    // }
    /*/
    self.Pix[offset+0] = 255
    self.Pix[offset+1] = 0
    self.Pix[offset+2] = 0
    self.Pix[offset+3] = 0xFF
    //*/
}

