package ppu

import "nes/tools"

func (self *Ppu) clearSpriteScanline() {
    for _, entry := range self.spriteScanline {
        entry.Y = 0xFF  // y-coord is 255 means it is not visible
        entry.X = 0xFF
        entry.ID = 0xFF
        entry.Attribute = 0xFF
    }

    self.sprite_count = 0
    // self.bSpriteZeroHitPossible = false
    // self.bSpriteZeroBeingRendered = false
}

func (self *Ppu) evaluateSprite () {

    oam := self.OAM

    self.bSpriteZeroHitPossible = false

    for i:=0; i<len(oam); i+=4 {
        if self.sprite_count > 8 {  // stop  if we got 9 sprites...
            break
        }
        objattr := ObjAttribEntry{ oam[i],oam[i+1],oam[i+2],oam[i+3] }
        diff := self.scanline - int(objattr.Y)
        spriteSize := 8
        if self.ctrl_reg & CTRL_PPU_SPR_SIZE != 0 {
            spriteSize = 16
        }
        if diff >=0 && diff < spriteSize {
            if self.sprite_count < 8 {
                // Is this sprite  sprite zero?
                if i == 0 {
                    // It is, so its may possible trigger a sprite zero hit
                    self.bSpriteZeroHitPossible = true
                }

                self.spriteScanline[self.sprite_count] = objattr
            }
            self.sprite_count++
        }
    }
    if self.sprite_count > 8 {
        self.status_reg |= ST_PPU_SPR_OVERFLOW
        self.sprite_count = 8
    }
}

func (self *Ppu) loadSpriteShifters() {
    for i:=0; i< int(self.sprite_count); i++ {
        var sprite_pattern_bits_lo uint8
        var sprite_pattern_bits_hi uint8
        // the address are influenced by server things
        //  the current sprite mode: 8x8 / 8x16
        //  the current pattern table that's been chosen
        //  and the tile ID of sprite it self
        var sprite_pattern_addr_lo uint16
        var sprite_pattern_addr_hi uint16

        var row_offset uint16 = uint16(self.scanline) - uint16(self.spriteScanline[i].Y)

        if self.ctrl_reg & CTRL_PPU_SPR_SIZE == 0 {
            // 8x8

            if self.spriteScanline[i].Attribute & 0x80 == 0 {
                // normal, not flip vertically
                // reading directly from pattern table
                sprite_pattern_addr_lo =
                    (uint16(self.CtrlRegPatternSPR()) << 12) | // which pattern table
                    (uint16(self.spriteScanline[i].ID) << 4)  |  // each tile 16bytes
                    row_offset  // row offset

            } else {
                // flip vertically
                sprite_pattern_addr_lo =
                    (uint16(self.CtrlRegPatternSPR()) << 12) | // which pattern table
                    (uint16(self.spriteScanline[i].ID) << 4)  |  // each tile 16bytes
                    (7-row_offset)  // row offset
            }
        } else {
            // 8x16

            if self.spriteScanline[i].Attribute & 0x80 == 0 {
                // normal, not flip vertically
                // there are 2 tiles , which tile should be read ?
                if row_offset < 8 {
                    // top half
                    sprite_pattern_addr_lo =
                        ((uint16(self.spriteScanline[i].ID) & 0x1)<<12) |  // which pattern tbl
                        ((uint16(self.spriteScanline[i].ID) & 0xFE)<<4) |
                        (row_offset & 0x7)
                } else {
                    // bottom half
                    sprite_pattern_addr_lo =
                        ((uint16(self.spriteScanline[i].ID) & 0x1)<<12) |
                        (((uint16(self.spriteScanline[i].ID) & 0xFE)+1 )<<4) |  // diff: +1
                        (row_offset & 0x7)
                }
            } else {
                // flip vertically
                // there are 2 tiles , which tile should be read ?
                if row_offset < 8 {
                    // top half
                    sprite_pattern_addr_lo =
                        ((uint16(self.spriteScanline[i].ID) & 0x1)<<12) |
                        (((uint16(self.spriteScanline[i].ID) & 0xFE)+1 )<<4) |  // diff: +1
                        ((7-row_offset) & 0x7)
                } else {
                    // bottom half
                    sprite_pattern_addr_lo =
                        ((uint16(self.spriteScanline[i].ID) & 0x1)<<12) |  // which pattern tbl
                        ((uint16(self.spriteScanline[i].ID) & 0xFE)<<4) |
                        ((7-row_offset) & 0x7)
                }
            }
        }
        sprite_pattern_addr_hi = sprite_pattern_addr_lo + 8
        sprite_pattern_bits_lo = self.PpuRead( sprite_pattern_addr_lo, false )
        sprite_pattern_bits_hi = self.PpuRead( sprite_pattern_addr_hi, false )

        // before we place those data into shifters there's one more thing to do : flipX
        if self.spriteScanline[i].Attribute & 0x40 != 0 {
            // flipX
            sprite_pattern_bits_lo = tools.Flipbyte( sprite_pattern_bits_lo )
            sprite_pattern_bits_hi = tools.Flipbyte( sprite_pattern_bits_hi )
        }

        // load to the shifters
        self.sprite_shifter_pattern_lo[i] = sprite_pattern_bits_lo
        self.sprite_shifter_pattern_hi[i] = sprite_pattern_bits_hi
    } // end for
}


func (self *Ppu) renderFG() (uint8, uint8, uint8) {
    var fg_pixel uint8 = 0
    var fg_palette uint8 = 0
    var fg_priority uint8 = 0

    if self.mask_reg & MASK_PPU_RENDER_SPR != 0 {
        if (self.mask_reg & MASK_PPU_RENDER_SPR_LEFT != 0) || (self.cycle >= 9) {

            self.bSpriteZeroBeingRendered = false

            for i:=0 ; i< int(self.sprite_count) ; i++ {
                if self.spriteScanline[i].X == 0 {
                    fg_pixel_lo := tools.B2i(self.sprite_shifter_pattern_lo[i]&0x80 !=0)
                    fg_pixel_hi := tools.B2i(self.sprite_shifter_pattern_hi[i]&0x80 !=0)
                    fg_pixel = (fg_pixel_hi << 1) | fg_pixel_lo

                    // + 0x04 because the first 4 palette is used for background
                    fg_palette = (self.spriteScanline[i].Attribute & 0x3) + 0x04
                    // the priority with the background, it allows sprite to go behind bg tiles
                    fg_priority = tools.B2i((self.spriteScanline[i].Attribute & 0x20) == 0) // yes, ==0 here
                    // if it is not transparent , we done
                    if fg_pixel != 0 {
                        if i==0 {
                            self.bSpriteZeroBeingRendered = true
                        }
                        break
                    }
                }
            }
        }
    }

    return fg_pixel, fg_palette, fg_priority
}





