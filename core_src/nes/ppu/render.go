package ppu

func (self *Ppu) render() {
    // renderingEnabled := self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0
    // visibleLine := self.scanline >= 0 && self.scanline < 240
    // visibleCycle := self.cycle >= 1 && self.cycle <= 256
    // if renderingEnabled  && visibleLine && visibleCycle 
    {
        bg_pixel, bg_palette := self.renderBG()

        fg_pixel, fg_palette, fg_priority := self.renderFG()

        if self.cycle < 8+1 && ((self.mask_reg &  MASK_PPU_RENDER_BG_LEFT) == 0)  {
            bg_pixel = 0
        }
        if self.cycle < 8+1 && ((self.mask_reg &  MASK_PPU_RENDER_SPR_LEFT) == 0)  {
            fg_pixel = 0
        }

        var pixel, palette uint8 = 0,0
        if bg_pixel == 0 && fg_pixel == 0 {
            // both bg / fg is transparent, draw backgroud color
            pixel = 0
            palette = 0
        } else if bg_pixel == 0 && fg_pixel > 0 {
            pixel = fg_pixel
            palette = fg_palette
        } else if bg_pixel > 0 && fg_pixel == 0 {
            pixel = bg_pixel
            palette = bg_palette
        } else if bg_pixel > 0 && fg_pixel > 0 {
            if fg_priority != 0 {
                pixel = fg_pixel
                palette = fg_palette
            } else {
                pixel = bg_pixel
                palette = bg_palette
            }
            /*
            if self.bSpriteZeroHitPossible && self.bSpriteZeroBeingRendered {
                // make sure it's rendering both bg and fg
                if self.mask_reg & MASK_PPU_RENDER_BG != 0  && self.mask_reg & MASK_PPU_RENDER_SPR != 0 {
                    // checks what happens at the left-hand edge of the screen
                    // thost bit basically decides whether the 8 pixels to the left of the screen should be drawn or not
                    // The left edge of the screen has specific switches to control its appearance.
                    // This is used to smooth inconsistencies when scrolling (since sprites x coord must be >= 0)
                    if self.mask_reg & ( MASK_PPU_RENDER_BG_LEFT | MASK_PPU_RENDER_SPR_LEFT ) == 0 {
                        // left disabled
                        if self.cycle >= 9 && self.cycle < 258 {
                            self.status_reg |= ST_PPU_SPR_ZERO_HIT
                        }
                    } else {
                        if self.cycle >= 1 && self.cycle < 255 {
                            self.status_reg |= ST_PPU_SPR_ZERO_HIT
                        }

                    }
                }
            }
            /*/
            if self.bSpriteZeroHitPossible && self.bSpriteZeroBeingRendered {
                // always miss 0
                if self.cycle >= 1 && self.cycle < 255+1 {
                    self.status_reg |= ST_PPU_SPR_ZERO_HIT
                }
            }
            //*/
        }
        // Now we have a final pixel colour, and a palette for this cycle
        self.sprScreen.SetPixel(self.cycle - 1, self.scanline, self.GetColourFromPaletteRam( int(palette), pixel) )
    }
}
