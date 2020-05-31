package ppu

import (
    "log"
    "fmt"
    // "math/rand"
    "nes/cart"
    "nes/mapper"
    // "nes/palette"
    "nes/sprite"
    // "nes/color"
    // "nes/tools"
)

type Ppu struct {
    // split 2K VRAM into two 1k table name
    // In fact NES does have the ability to address 4 name tables,
    //  there's some trickery involved in that
    tblName [2][1024]uint8  // VRAM 

    tblPalette [32]uint8    // Palette , don't read data from this variable directly

    // the following 3 is for emulator purpose ?
    sprScreen *sprite.Sprite // represent the full screen
    sprScreenSnapshot *sprite.Sprite // represent the screen snapshot
    // sprNameTable [2]  // represent graphical depiction of name tables
    sprPatternTable [2]*sprite.Sprite // represent the depiction of pattern tables
    sprPalette *sprite.Sprite

    cartridge *cart.Cartridge  // PPU need to read cartridge

    Frame_complete bool // represent when a frame is complete

    // private
    scanline    int   // row on screen    (y)
    cycle       int   // column of screen (x)

    status_reg STATUS_PPU
    ctrl_reg CTRL_PPU
    mask_reg MASK_PPU
    Nmi bool

    // I need to know wheter I'm writing to the low byte or high byte. 
    address_latch uint8
    // when we read data from the PPU, it is in fact delayed by one cycle.
    //  so we need to buffer that byte
    ppu_data_buffer uint8
    // I need a 16-bit variable to store the compiled address
    // ppu_address uint16

    // loopy register instead
    vram_addr  LoopyReg
    tram_addr  LoopyReg
    fine_x     uint8

    // varibles for storing the PPU pre-loading data
    bg_next_tile_id uint8
    bg_next_tile_attrib uint8
    bg_next_tile_lsb uint8
    bg_next_tile_msb uint8

    // shift register
    bg_shifter_pattern_lo uint16
    bg_shifter_pattern_hi uint16
    bg_shifter_attrib_lo uint16
    bg_shifter_attrib_hi uint16

    OAM [256]uint8
    oam_addr uint8

    spriteScanline [8]ObjAttribEntry // max 8 sprites per scanline
    sprite_count uint8   // how many sprites found this scanline
    sprite_shifter_pattern_lo [8]uint8
    sprite_shifter_pattern_hi [8]uint8

    // determine when we're doing the sprite evaluation and scrolling through the OAM memory
    // if we don't select sprite 0, then we can't possibly have a sprite 0 hit
    bSpriteZeroHitPossible bool
    // to remember if I'm currently drawing sprite zero between cycle updates
    bSpriteZeroBeingRendered bool

    odd_frame bool
}


func NewPpu() *Ppu {
    log.Println( "ppu instanciated" )
    ppu := &Ppu {}
    // palScreen = palette.COLORS[:]  // not use
    ppu.sprPatternTable[0] = sprite.NewSprite( 128,128 )
    ppu.sprPatternTable[1] = sprite.NewSprite( 128,128 )
    ppu.sprPalette = sprite.NewSprite( 8*5 * 6  , 6 )
    ppu.sprScreen = sprite.NewSprite( 256,240 )
    ppu.sprScreenSnapshot = sprite.NewSprite( 256,240 )

    return ppu
}


func (self *Ppu) Reset() {
    // not finish
    log.Println( "ppu reseted" )

    self.fine_x = 0
    self.address_latch = 0
    self.ppu_data_buffer = 0
    self.scanline = 0
    self.cycle = 0

    self.bg_next_tile_id = 0
    self.bg_next_tile_attrib = 0
    self.bg_next_tile_lsb = 0
    self.bg_next_tile_msb = 0

    self.bg_shifter_attrib_hi = 0
    self.bg_shifter_attrib_lo = 0
    self.bg_shifter_pattern_hi = 0
    self.bg_shifter_pattern_lo = 0

    self.status_reg = 0
    self.mask_reg = 0
    self.ctrl_reg = 0

    self.vram_addr.reg = 0
    self.tram_addr.reg = 0

    // self.scanline_trigger = false
    self.odd_frame = false
}

// Communication with Main Bus
func (self *Ppu) CpuWrite(addr uint16, data uint8) {
    switch addr {
    case 0x0000: // control
        self.ctrl_reg = CTRL_PPU(data)

        self.tram_addr.SetNametableX(  uint16(self.CtrlRegNameTableX())  )
        self.tram_addr.SetNametableY(  uint16(self.CtrlRegNameTableY())  )
    case 0x0001: // mask
        self.mask_reg = MASK_PPU(data)
    case 0x0002: // status 
        // you can't write to status register
    case 0x0003: // OAM address
        self.oam_addr = data
        // log.Println( "set oam_addr to:", self.oam_addr )
    case 0x0004: // OAM data
        self.OAM[self.oam_addr] = data
        self.oam_addr ++  // Address should increment on $2004 write
    case 0x0005: // Scroll
        if self.address_latch == 0 {
            self.fine_x = data & 0x7  // bottom 3 bits
            self.tram_addr.SetCoarseX(  uint16(data >> 3) ) // top 5 bits
        } else {
            self.tram_addr.SetFineY(  uint16(data & 0x7) )
            self.tram_addr.SetCoarseY(  uint16(data >> 3) )
        }
        self.address_latch = (self.address_latch+1)&1
    case 0x0006: // PPU Address
        // if I write to the address register , 
        //  I need to set the high byte first, and then the low byte
        if self.address_latch == 0 {
            // why not simply `ppu_address = data ?`  because 
            //  sometimes the address_latch will be reset from outside
            // self.ppu_address = (self.ppu_address & 0x00FF) | (uint16(data)<<8)
            self.tram_addr.reg = ( self.tram_addr.reg & 0x00FF ) | (uint16(data) << 8)
        } else {
            // self.ppu_address = (self.ppu_address & 0xFF00) | uint16(data)
            self.tram_addr.reg = ( self.tram_addr.reg & 0xFF00 ) | uint16(data)
            // once a full 16bit address info has been written the vram_addr is updated
            self.vram_addr.reg = self.tram_addr.reg
        }
        self.address_latch = (self.address_latch+1)&1

    case 0x0007: // PPU data
        // self.PpuWrite( self.ppu_address, data)
        self.PpuWrite( self.vram_addr.reg, data)
        // it would be very tedious for the programmer to have to write
        //  2 bytes address and 1 byte data.
        // most of the time programmers will be writing data to successive locations. 
        //  the PPU provides a facility for this: it has an auto increment on the PPU address.

        // so now you can write a stream of data to PPU, 
        //  but how to write the name table memory but in vertically orientation ?
        //  the control register has a bit called increment mode
        if self.ctrl_reg & CTRL_PPU_INCREMENT_MODE != 0 {
            // self.ppu_address += 32
            self.vram_addr.reg += 32
        } else {
            // self.ppu_address ++
            self.vram_addr.reg ++
        }
    }
}

func (self *Ppu) cpuPeek(addr uint16) uint8 {
    var data uint8 = 0
    switch addr {
    case 0x0000: // Control
        data = uint8(self.ctrl_reg)
        break;
    case 0x0001: // Mask
        data = uint8(self.mask_reg)
        break;
    case 0x0002: // Status
        data = uint8(self.status_reg)
        break;
    case 0x0003: // OAM Address
        break;
    case 0x0004: // OAM Data
        break;
    case 0x0005: // Scroll
        break;
    case 0x0006: // PPU Address
        break;
    case 0x0007: // PPU Data
        break;
    }
    return data
}

func (self *Ppu) CpuRead(addr uint16, bReadonly bool) uint8 {
    var data uint8 = 0

    if bReadonly {
        // Reading from PPU registers can affect their contents
        // so this read only option is used for examining the
        // state of the PPU without changing its state. This is
        // really only used in debug mode.
        return self.cpuPeek( addr )
    }

    switch addr {
    case 0x0000: // control
    case 0x0001: // mask
    case 0x0002: // status 
        // we know that our program was getting stuck reading the status reg.
        // so I'm going to hack in something just to make some progress
        // self.status_reg |= ST_PPU_VBLANK  // removed, be correctly handled in clock()

        // act of reading is changing the state of the device
        // we are only interested in the top 3 bits
        //  the unused bits tend to be filled with noise or more likely
        //  what was last on the internal data buffer of PPU
        data = (uint8(self.status_reg) & 0xE0) | (self.ppu_data_buffer & 0x1F)
        // readint the status register also clears the vertical blank
        //  whether or not you are in vertical blank is irrelevant
        //  As soon as you read the statue to determine if you're in vblank,
        //  it gets reset to 0
        self.status_reg &^= ST_PPU_VBLANK
        // reading from the status register will also set 
        //  our address_latch back to 0
        self.address_latch = 0

    case 0x0003: // OAM address
        // reading address register make no sense
    case 0x0004: // OAM data
        data = self.OAM[self.oam_addr]
    case 0x0005: // Scroll
    case 0x0006: // PPU Address
        // you can not read from the address register
        // it doesn't make any sense
    case 0x0007: // PPU data
        // we can read from data register
        // but it delayed by one read
        data = self.ppu_data_buffer
        // self.ppu_data_buffer = self.PpuRead( self.ppu_address , false )
        self.ppu_data_buffer = self.PpuRead( self.vram_addr.reg , false )

        // this delayed read is true for almost all of the PPU address range
        //  except for where our palettes reside.
        // there are various hardware reasons why this could be the case

        // &0x3FFF
        // The two MSB within the PPU memory address should be completely ignored in all
        // circumstances, effectively mirroring the 0000-3FFF address range within the whole 0000-FFFF region,
        // for a total of 4 times.
        if self.vram_addr.reg & 0x3FFF >= 0x3f00 {
            data = self.ppu_data_buffer
            // Palette read should also read VRAM into read buffer
            // Palette RAM consists of twenty-eight 6-bit words of DRAM embedded within the PPU and accessible when the VRAM address is between $3F00 and $3FFF (inclusive). 
            // When you read PPU $3F00-$3FFF, you get immediate data from Palette RAM (without the 1-read delay usually present when reading from VRAM) 
            // and the PPU will also fetch nametable data from the corresponding address (which is mirrored from PPU $2F00-$2FFF). 
            // This phenomenon does not occur during writes (as it would result in corrupting the contents of the nametables when writing to the palette) 
            // and only happens during reading (since it has no noticeable side effects).
            self.ppu_data_buffer = self.PpuRead( self.vram_addr.reg - 0x1000 , false )
        }

        // auto increment address
        if self.ctrl_reg & CTRL_PPU_INCREMENT_MODE != 0 {
            // self.ppu_address += 32
            self.vram_addr.reg += 32
        } else {
            // self.ppu_address ++
            self.vram_addr.reg ++
        }
    }
    return data
}

// Communication with PPU Bus
func (self *Ppu) PpuWrite(addr uint16, data uint8) {
    // mask the address just in case the PPU ever 
    // tries to its bus in a location beyond its addressable
    addr &= 0x3FFF

    if self.cartridge.PpuWrite(addr, data ) {

    } else if addr >= 0x2000 && addr <= 0x3EFF {
        addr &= 0x0FFF  // 4k , address offset start from name tables
        if self.cartridge.Mirror() == mapper.MIRROR_VERTICAL {
            // Vertical
            if addr >= 0x0000 && addr <= 0x03FF {
                self.tblName[0][addr & 0x03FF] = data
            }
            if addr >= 0x0400 && addr <= 0x07FF {
                self.tblName[1][addr & 0x03FF] = data
            }
            if addr >= 0x0800 && addr <= 0x0BFF {
                self.tblName[0][addr & 0x03FF] = data  // M
            }
            if addr >= 0x0C00 && addr <= 0x0FFF {
                self.tblName[1][addr & 0x03FF] = data  // M
            }
        } else if self.cartridge.Mirror() == mapper.MIRROR_HORIZONTAL {
            // Horizontal
            if addr >= 0x0000 && addr <= 0x03FF {
                self.tblName[0][addr & 0x03FF] = data
            }
            if addr >= 0x0400 && addr <= 0x07FF {
                self.tblName[0][addr & 0x03FF] = data  // M
            }
            if addr >= 0x0800 && addr <= 0x0BFF {
                self.tblName[1][addr & 0x03FF] = data
            }
            if addr >= 0x0C00 && addr <= 0x0FFF {
                self.tblName[1][addr & 0x03FF] = data  // M
            }
        } else {
            panic( fmt.Sprintf( "unsupoported mirror mode: %d" , self.cartridge.Mirror() ) )
        }
    } else if addr >= 0x3F00 && addr <= 0x3FFF {
        addr &= 0x001F
        // mirroring
        if (addr == 0x0010) {
            addr = 0x0000
        }
        if (addr == 0x0014) {
            addr = 0x0004
        }
        if (addr == 0x0018) {
            addr = 0x0008
        }
        if (addr == 0x001C) {
            addr = 0x000C
        }
        self.tblPalette[addr] = data
    }

}

func (self *Ppu) PpuRead(addr uint16, bReadonly bool) uint8 {
    var data uint8 = 0
    addr &= 0x3FFF

    // 2020.05.31 Read Palette data is hot operation
    // so move it to the top if ... though it looks ugly...
    if addr >= 0x3F00 && addr <= 0x3FFF {
        addr &= 0x001F
        // mirroring
        if (addr == 0x0010) {
            addr = 0x0000
        }
        if (addr == 0x0014) {
            addr = 0x0004
        }
        if (addr == 0x0018) {
            addr = 0x0008
        }
        if (addr == 0x001C) {
            addr = 0x000C
        }

        if self.mask_reg & MASK_PPU_GRAYSCALE  != 0 {
            data = self.tblPalette[addr] & 0x30
        } else {
            data = self.tblPalette[addr] & 0x3F
        }

    } else if self.cartridge.PpuRead(addr, &data ) {

    } else if addr >= 0x2000 && addr <= 0x3EFF {     // < 8k int total
        addr &= 0x0FFF  // 4k , address offset start from name tables
        if self.cartridge.Mirror() == mapper.MIRROR_VERTICAL {
            // Vertical
            if addr >= 0x0000 && addr <= 0x03FF {
                data = self.tblName[0][addr & 0x03FF]
            }
            if addr >= 0x0400 && addr <= 0x07FF {
                data = self.tblName[1][addr & 0x03FF]
            }
            if addr >= 0x0800 && addr <= 0x0BFF {
                data = self.tblName[0][addr & 0x03FF]  // M 
            }
            if addr >= 0x0C00 && addr <= 0x0FFF {
                data = self.tblName[1][addr & 0x03FF]  // M
            }
        } else if self.cartridge.Mirror() == mapper.MIRROR_HORIZONTAL {
            // Horizontal
            if addr >= 0x0000 && addr <= 0x03FF {
                data = self.tblName[0][addr & 0x03FF]
            }
            if addr >= 0x0400 && addr <= 0x07FF {
                data = self.tblName[0][addr & 0x03FF]  // M
            }
            if addr >= 0x0800 && addr <= 0x0BFF {
                data = self.tblName[1][addr & 0x03FF]
            }
            if addr >= 0x0C00 && addr <= 0x0FFF {
                data = self.tblName[1][addr & 0x03FF]  // M
            }
        } else {
            panic( fmt.Sprintf( "unsupoported mirror mode: %d" , self.cartridge.Mirror() ) )
        }
    }

    return data
}

func (self *Ppu) ConnectCartridge( cart *cart.Cartridge ) {
    self.cartridge = cart
}

func (self *Ppu) DMATransfer( data uint8 ) {
    self.OAM[ self.oam_addr ] = data
    self.oam_addr ++
}
func (self *Ppu) OAM_Address( ) uint8  {
    return self.oam_addr
}


func (self *Ppu) Clock() {
    if self.scanline >= -1 && self.scanline < 240 {

        if self.scanline == 0 && self.cycle == 0 && self.odd_frame && 
            (  self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0   ) {
            // "Odd Frame" cycle skip
            self.cycle = 1
        }


        // leaving the vertical blank
        if self.scanline == -1 && self.cycle == 1 {
            self.status_reg &^= ST_PPU_VBLANK

            // reset out sprite status 
            self.status_reg &^= ST_PPU_SPR_OVERFLOW

            self.status_reg &^= ST_PPU_SPR_ZERO_HIT

            for i:=0; i< len(self.sprite_shifter_pattern_lo); i++ {
                self.sprite_shifter_pattern_lo[i] = 0
                self.sprite_shifter_pattern_hi[i] = 0
            }
        }


        // for a bunch of cycles our particular scanline we want to
        //  extract the tileID, the attribute and tile pattern
        if (self.cycle >=2 && self.cycle < 258) || (self.cycle >= 321 && self.cycle < 338) { // not 337 ?
            // every visible cycle, we want to update the shifters
            self.UpdateShifters()

            // the repeated 8 cycles per tile, are for pre-loading info to render the next 8 pixel
            switch (self.cycle -1) & 0x7 {
            case 0:
                // and when the internal cycle counter loops aroude 8 pixels ( red cell )
                // we're going to load our backgroud 
                self.LoadBackgroundShifters()

                self.loadNextBgTileID()
            case 2:
                self.bg_next_tile_attrib = self.PpuRead( 0x23C0 |
                    (self.vram_addr.NametableY() << 11) |
                    (self.vram_addr.NametableX() << 10) |
                    ((self.vram_addr.CoarseY() >>2) <<3) |  // unit16, >> is ok
                     (self.vram_addr.CoarseX() >>2)  ,
                false )
                // split it into 2-bits
                if self.vram_addr.CoarseY() & 0x2 != 0 {
                    self.bg_next_tile_attrib >>= 4
                }
                if self.vram_addr.CoarseX() & 0x2 != 0 {
                    self.bg_next_tile_attrib >>= 2
                }
                self.bg_next_tile_attrib &= 0x3  // 2-bits
            case 4:
                self.bg_next_tile_lsb = self.PpuRead(
                    (uint16(self.CtrlRegPatternBG()) << 12) +
                    (uint16(self.bg_next_tile_id) << 4) +
                    self.vram_addr.FineY() + 0 ,
                false )
            case 6:
                self.bg_next_tile_msb = self.PpuRead(
                    (uint16(self.CtrlRegPatternBG()) << 12) +
                    (uint16(self.bg_next_tile_id) << 4) +
                    self.vram_addr.FineY() + 8 ,
                false )
            case 7:
                // we've gone through 8 pixels, we must be going onto the next tile
                self.IncrementScrollX()
            }  // end switch
        }

        // when we entering the end of the scanline , we want to 
        //  increment the Y-direction of our loopy register
        if self.cycle == 256 {
            // we've done with a visible row, we'll increment scroll Y
            self.IncrementScrollY()
        }
        if self.cycle == 257 {
            // but because we've increment the scrollY , the scrollX is still incorrect
            // we need reset it
            self.LoadBackgroundShifters()
            self.TransferAddressX()
        }
        // we need to set Y on the non visible scanline
        if self.scanline == -1 && self.cycle >=280 && self.cycle < 305 {
            self.TransferAddressY()
        }
        // Superfluous reads of tile id at end of scanline
        if self.cycle == 338 || self.cycle == 340 {
            self.loadNextBgTileID()
        }

        // Foreground Rendering  ===============================
        // Foreground Rendering  ===============================
        // Foreground Rendering  ===============================
        // Foreground Rendering  ===============================

        renderingEnabled := self.mask_reg & (MASK_PPU_RENDER_BG | MASK_PPU_RENDER_SPR) != 0
        visibleLine := self.scanline >= 0 && self.scanline < 240
        if renderingEnabled {  // must add this, because you can not set spriteoverflow flag when $2000 is 0
            if self.cycle == 257 {
                // must cleanup even scanline == -1
                self.clearSpriteScanline()
                // Secondly, clear out any residual information in sprite pattern shifters
                for i:=0; i< len(self.sprite_shifter_pattern_lo); i++ {
                    self.sprite_shifter_pattern_lo[i] = 0
                    self.sprite_shifter_pattern_hi[i] = 0
                }

                if visibleLine {
                    // perform sprite evaluation
                    self.evaluateSprite()
                }
            }
        }

        if self.cycle == 340 {
            self.loadSpriteShifters()
        }

    } // end if self.scanline >= -1 && self.scanline < 240


    self.render()

    // curiously on scaline 240, nothing happens
    if self.scanline == 240 {
        // Post Render Scanline -- Do Nothing
    }

    // entering the vertical blank, and emit a nmi interrupt
    if self.scanline >= 241 && self.scanline < 261 {
        if self.scanline == 241 && self.cycle == 1 {  // qibinyi ,TODO it can not pass ppu test
            // we entering the vertical blank, we set the vblank bit
            self.status_reg |= ST_PPU_VBLANK
            // if the enable non-maskable interrupt bit has been set in the
            //  control register, then set nmi varialbe to true
            if self.ctrl_reg & CTRL_PPU_ENABLE_NMI != 0 {
                self.Nmi = true
            }
        }

    }

    /*
    // noise debug
    c := 0x3F
    if rand.Intn(2) == 1 {
        c = 0x30
    }
    self.sprScreen.SetPixel( self.cycle-1, self.scanline, color.COLORS[ c ]  )
    //*/

    // Advance renderer -- it never stops, it's relentless
    self.cycle++

    // mapper
    if self.mask_reg & MASK_PPU_RENDER_BG != 0  || self.mask_reg & MASK_PPU_RENDER_SPR != 0 {
        if self.cycle == 260 && self.scanline < 240 {
            self.cartridge.GetMapper().Scanline()
        }
    }

    if self.cycle >= 341 {
        // reset cycle
        self.cycle = 0
        self.scanline++

        if self.scanline >= 261 {
            self.scanline = -1
            self.Frame_complete = true

            // for emulator displaying
            copy(self.sprScreenSnapshot.Pix , self.sprScreen.Pix)
            self.odd_frame = !self.odd_frame
        }
    }
}



// =============== for debug ========================
func (self *Ppu) DebugSetPalette( pal_entry, color_entry int )  {
    self.tblPalette[ pal_entry & 0x1F ] = uint8(color_entry & 0x3F)
}

func (self *Ppu) DumpNameTables(  ) [2][1024]uint8 {
    return self.tblName
}





