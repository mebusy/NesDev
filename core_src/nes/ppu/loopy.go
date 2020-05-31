package ppu

/*
uint16_t coarse_x : 5;
uint16_t coarse_y : 5;
uint16_t nametable_x : 1;
uint16_t nametable_y : 1;
uint16_t fine_y : 3;
uint16_t unused : 1;
*/
type LoopyReg struct {
    reg uint16
}

const (
    _ uint16  = 1 << iota
    _
    _
    _
    _

    _
    _
    _
    _
    _

    LOOPY_REG_NAMETABLE_X  // 10
    LOOPY_REG_NAMETABLE_Y  // 11
)

func (self *LoopyReg) CoarseX() uint16 {
    var mask uint16 = 0x1F
    var bitoff uint16 = 0
    return (self.reg>>bitoff) & mask
}
func (self *LoopyReg) CoarseY() uint16 {
    var mask uint16 = 0x1F
    var bitoff uint16 = 5
    return (self.reg>>bitoff) & mask
}

func (self *LoopyReg) NametableX() uint16 {
    var mask uint16 = 0x1
    var bitoff uint16 = 10
    return (self.reg>>bitoff) & mask
}
func (self *LoopyReg) NametableY() uint16 {
    var mask uint16 = 0x1
    var bitoff uint16 = 11
    return (self.reg>>bitoff) & mask
}

func (self *LoopyReg) FineY() uint16 {
    var mask uint16 = 0x7
    var bitoff uint16 = 12
    return (self.reg>>bitoff) & mask
}

func (self *LoopyReg) SetCoarseX( val uint16) {
    var mask uint16 = 0x1F
    var bitoff uint16 = 0
    self.reg &^= mask<<bitoff
    self.reg |= (val&mask)<<bitoff
}
func (self *LoopyReg) SetCoarseY( val uint16) {
    var mask uint16 = 0x1F
    var bitoff uint16 = 5
    self.reg &^= mask<<bitoff
    self.reg |= (val&mask)<<bitoff
}

func (self *LoopyReg) SetNametableX( val uint16) {
    var mask uint16 = 0x1
    var bitoff uint16 = 10
    self.reg &^= mask<<bitoff
    self.reg |= (val&mask)<<bitoff
}
func (self *LoopyReg) SetNametableY(val uint16 ) {
    var mask uint16 = 0x1
    var bitoff uint16 = 11
    self.reg &^= mask<<bitoff
    self.reg |= (val&mask)<<bitoff
}

func (self *LoopyReg) SetFineY(val uint16)  {
    var mask uint16 = 0x7
    var bitoff uint16 = 12
    self.reg &^= mask<<bitoff
    self.reg |= (val&mask)<<bitoff
}


/*
All of the cycles marked in red imply that we're doing some additional manipulation of the loopy registers. 
And there are 4 essential functions 
    1. incrementing in x-direction
    2. incrementing in y-direction
    3. resetting x-axis
    4. resetting y-axis
*/
func (self *Ppu) IncrementScrollX() {
    // Note: pixel perfect scrolling horizontally is handled by the 
    // data shifters. Here we are operating in the spatial domain of 
    // tiles, 8x8 pixel blocks.

    // Ony if rendering is enabled
    if self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0 {
        // A single name table is 32x30 tiles. As we increment horizontally
        // we may cross into a neighbouring nametable, or wrap around to
        // a neighbouring nametable
        if self.vram_addr.CoarseX() == 31 {
            // Leaving nametable so wrap address round
            self.vram_addr.SetCoarseX( 0 )
            // Flip target nametable bit, so now we're indexing into the other nametable
            // vram_addr.nametable_x = ~vram_addr.nametable_x;
            self.vram_addr.reg ^= LOOPY_REG_NAMETABLE_X
        } else {
            // Staying in current nametable, so just increment
            self.vram_addr.SetCoarseX( self.vram_addr.CoarseX() +1 )
        }
    }
}


func (self *Ppu) IncrementScrollY() {
    // Incrementing vertically is more complicated. The visible nametable
    // is 32x30 tiles, but in memory there is enough room for 32x32 tiles.

    // In addition, the NES doesnt scroll vertically in chunks of 8 pixels

    // Ony if rendering is enabled
    if self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0 {
        // If possible, just increment the fine y offset
        if self.vram_addr.FineY() < 7 {
            self.vram_addr.SetFineY( self.vram_addr.FineY()+1 )
        } else {
            // we need to increment the row, potentially wrapping into neighbouring vertical nametables.

            // Reset fine y offset
            self.vram_addr.SetFineY( 0 )

            // Check if we need to swap vertical nametable targets
            if self.vram_addr.CoarseY() == 29 {
                // we do , so reset coarse y offset
                self.vram_addr.SetCoarseY( 0 )
                // And flip the target nametable bit
                // vram_addr.nametable_y = ~vram_addr.nametable_y
                self.vram_addr.reg ^= LOOPY_REG_NAMETABLE_Y

            } else if self.vram_addr.CoarseY() == 31 {
                // In case the pointer is in the attribute memory, we just wrap around the current nametable
                self.vram_addr.SetCoarseY( 0 )
            } else {
                // None of the above boundary/wrapping conditions apply so just increment the coarse y offset
                // vram_addr.coarse_y++;
                self.vram_addr.SetCoarseY(  self.vram_addr.CoarseY()+1 )
            }
        } // end >= 7
    } // end in rendering
}

// Transfer the temporarily stored horizontal nametable access information into the "pointer". 
// ote that fine x scrolling is not part of the "pointer" ddressing mechanism
func (self *Ppu) TransferAddressX() {
    // Ony if rendering is enabled
    if self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0 {
        // vram_addr.nametable_x = tram_addr.nametable_x;
        self.vram_addr.SetNametableX( self.tram_addr.NametableX() )
        // vram_addr.coarse_x    = tram_addr.coarse_x;
        self.vram_addr.SetCoarseX(  self.tram_addr.CoarseX() )
    }
}

func (self *Ppu) TransferAddressY() {
    // Ony if rendering is enabled
    if self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0 {
        // vram_addr.fine_y      = tram_addr.fine_y;
        self.vram_addr.SetFineY( self.tram_addr.FineY() )
        // vram_addr.nametable_y = tram_addr.nametable_y;
        self.vram_addr.SetNametableY( self.tram_addr.NametableY() )
        // vram_addr.coarse_y    = tram_addr.coarse_y;
        self.vram_addr.SetCoarseY(  self.tram_addr.CoarseY() )
    }
}

