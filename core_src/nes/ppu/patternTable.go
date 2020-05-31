package ppu

import (
    "nes/color"
    "nes/sprite"
)


// This function draw the CHR ROM for a given pattern table into
// an olc::Sprite, using a specified palette. Pattern tables consist
func (self *Ppu) GetPatternTable( i int, palette int ) *sprite.Sprite {
    for nTileY :=0 ; nTileY < 16; nTileY++ {
        for nTileX :=0 ; nTileX < 16; nTileX++ {
            // 16 tile per row * 16 bytes per tile = 256 bytes
            nOffset := nTileY*256 + nTileX*16

            // each tile
            for row :=0 ; row < 8 ; row ++ {
                // each row of tile is 1 lsb byte + 1 msb byte 
                tile_lsb := self.PpuRead( uint16(i*0x1000+nOffset+row+0), false )
                tile_msb := self.PpuRead( uint16(i*0x1000+nOffset+row+8), false )
                for col :=0 ; col < 8 ; col ++ {
                    pixel := (tile_lsb & 0x01) + ((tile_msb & 0x01)<<1)
                    tile_lsb >>= 1
                    tile_msb >>= 1

                    self.sprPatternTable[i].SetPixel(
                        nTileX * 8 + (7 - col),
                        nTileY * 8 + row,
                        self.GetColourFromPaletteRam(palette, pixel) )
                }
            }
        }
    }
    // Finally return the updated sprite representing the pattern table
    return self.sprPatternTable[i]
}


func (self *Ppu) GetColourFromPaletteRam( palette int ,  pixel uint8) color.COLOR  {
    // This is a convenience function that takes a specified palette and pixel
    // index and returns the appropriate screen colour.
    // "0x3F00"       - Offset into PPU addressable range where palettes are stored
    // "palette << 2" - Each palette is 4 bytes in size
    // "pixel"        - Each pixel index is either 0, 1, 2 or 3
    // "& 0x3F"       - Stops us reading beyond the bounds of the palScreen array
    // 0,1,2,3 equvilent to 1,2,3,4 because of the mirroring ?  
    return color.COLORS[self.PpuRead(0x3F00 + (uint16(palette) << 2) + uint16(pixel), false ) & 0x3F];

    // Note: We dont access tblPalette directly here, instead we know that ppuRead()
    // will map the address onto the seperate small RAM attached to the PPU bus.
}


func (self *Ppu) GetScreenSnapshot( ) *sprite.Sprite {
    return self.sprScreenSnapshot
}

