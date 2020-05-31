package ppu

import (
    "nes/tools"
)

func (self *Ppu) loadNextBgTileID () {
    self.bg_next_tile_id = self.PpuRead( 0x2000 | ( self.vram_addr.reg & 0x0FFF ), false )
}


// Prime the "in-effect" background tile shifters ready for outputting next 8 pixels in scanline.
func (self *Ppu) LoadBackgroundShifters () {
    self.bg_shifter_pattern_lo = (self.bg_shifter_pattern_lo & 0xFF00) | uint16(self.bg_next_tile_lsb)
    self.bg_shifter_pattern_hi = (self.bg_shifter_pattern_hi & 0xFF00) | uint16(self.bg_next_tile_msb)

    var color uint8 = 0
    if self.bg_next_tile_attrib & 0x1 != 0 {
        color = 0xFF
    }
    self.bg_shifter_attrib_lo  = (self.bg_shifter_attrib_lo & 0xFF00) | uint16(color)

    color = 0
    if self.bg_next_tile_attrib & 0x2 != 0 {
        color = 0xFF
    }
    self.bg_shifter_attrib_hi  = (self.bg_shifter_attrib_hi & 0xFF00) | uint16(color)
}


// Every cycle the shifters storing pattern and attribute information shift their contents by 1 bit.
func (self *Ppu) UpdateShifters () {
    if self.mask_reg & MASK_PPU_RENDER_BG != 0 {
        // Shifting background tile pattern row
        self.bg_shifter_pattern_lo <<= 1;
        self.bg_shifter_pattern_hi <<= 1;
        // Shifting palette attributes by 1
        self.bg_shifter_attrib_lo <<= 1;
        self.bg_shifter_attrib_hi <<= 1;
    }

    if self.mask_reg & MASK_PPU_RENDER_SPR != 0 && self.cycle>=1 && self.cycle <258 {
        for i:=0; i< int(self.sprite_count) ; i++ {
            // before I start shifting 
            // I want to decrement the x-coordinate of the sprites
            //  I only want to start shifting once that x-coord become 0
            if self.spriteScanline[i].X > 0 {
                self.spriteScanline[i].X --
            } else {
                // the scanline has hit the sprite, start shifting
                self.sprite_shifter_pattern_lo[i] <<= 1
                self.sprite_shifter_pattern_hi[i] <<= 1
            }
        }
    }
}


func (self *Ppu) renderBG() (uint8, uint8) {
    // here, put a pixel
    // now we've got a cycle and scanline tracking architecture.
    var bg_pixel uint8 = 0  // The 2-bit pixel to be rendered
    var bg_palette uint8 = 0  // The 2-bit index of the palette the pixel indexes

    // We only render backgrounds if the PPU is enabled to do so.
    // Note if background rendering is disabled, the pixel and palette combine to form 0x00.
    // This will fall through the colour tables to yield the current background colour in effect.
    if self.mask_reg & MASK_PPU_RENDER_BG != 0 {
        if (self.mask_reg & MASK_PPU_RENDER_BG_LEFT != 0) || (self.cycle >= 9) {
            // Handle Pixel Selection by selecting the relevant bit depending upon fine x scolling.
            var bit_mux uint16 = 0x8000 >> self.fine_x

            // Select Plane pixels by extracting from the shifter at the required location.
            p0_pixel := tools.B2i((self.bg_shifter_pattern_lo & bit_mux) != 0)
            p1_pixel := tools.B2i((self.bg_shifter_pattern_hi & bit_mux) != 0)
            // Combine to form pixel index
            bg_pixel = (p1_pixel << 1) | p0_pixel

            // Get palette
            bg_pal0 := tools.B2i((self.bg_shifter_attrib_lo & bit_mux) != 0)
            bg_pal1 := tools.B2i((self.bg_shifter_attrib_hi & bit_mux) != 0)
            bg_palette = (bg_pal1 << 1) | bg_pal0
        }
    }

    return bg_pixel, bg_palette
}

