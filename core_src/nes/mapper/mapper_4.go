package mapper

import (
    // "log"
)

type Mapper_4 struct {
    nPRGBanks uint8
    nCHRBanks uint8

    // mapper instructure
    nTargetRegister uint8
    bPRGBankMode bool
    bCHRInversion bool

    mirrormode int

    pRegister [8]int
    // internal array to store the bank offsets
    // for the CHR and PRG ROM
    pCHRBank  [8]uint32
    pPRGBank  [4]uint32

    // status flags
    bIRQActive bool
    bIRQEnable bool
    bIRQUpdate bool

    // counter
    nIRQCounter uint16
    nIRQReload uint16

    RAMStatic []uint8
}

func NewMapper_4( prg_rom_chunks , chr_rom_chunks uint8, sram []uint8 ) *Mapper_4 {
    m := &Mapper_4{}
    m.nPRGBanks = prg_rom_chunks
    m.nCHRBanks = chr_rom_chunks
    m.mirrormode = MIRROR_HORIZONTAL
    m.RAMStatic = sram

    m.Reset()
    return m
}


func (self *Mapper_4) CpuMapRead( addr uint16, data *uint8 ) (bool,uint) {
    var mapped_addr uint = 0
    if addr >= 0x6000 && addr <= 0x7FFF {
        // Write is to static ram on cartridge
        mapped_addr = READFROMSRAM
        // Write data to RAM
        *data = self.RAMStatic[addr & 0x1FFF]
        // Signal mapper has handled request
        return true, mapped_addr
    }

    if addr >= 0x8000 && addr <= 0xFFFF {
        idx_chunk := (addr - 0x8000) / 0x2000
        mapped_addr = uint(self.pPRGBank[idx_chunk] + uint32(addr & 0x1FFF))
        return true, mapped_addr
    }
    return false, 0
}


func (self *Mapper_4) CpuMapWrite( addr uint16, data uint8) (bool,uint) {
    var mapped_addr uint = 0
    if addr >= 0x6000 && addr <= 0x7FFF {
        // Write is to static ram on cartridge
        mapped_addr = READFROMSRAM
        // Write data to RAM
        self.RAMStatic[addr & 0x1FFF] = data
        // Signal mapper has handled request
        return true, mapped_addr
    }

    if addr >= 0x8000 && addr <= 0x9FFF {
        // Bank Select
        if addr & 0x0001 == 0  {  // even
            self.nTargetRegister = data & 0x07
            self.bPRGBankMode = (data & 0x40) != 0
            self.bCHRInversion = (data & 0x80) != 0
        } else {  // odd
            // Update target register
            self.pRegister[self.nTargetRegister] = int(data)
            self.updateBankOfset()
        }
        return false, 0
    }

    // handle mirroring
    if addr >= 0xA000 && addr <= 0xBFFF {
        if addr & 0x0001 == 0 {  // even
            // Mirroring
            if (data & 0x01) != 0 {
                self.mirrormode = MIRROR_HORIZONTAL
            } else {
                self.mirrormode = MIRROR_VERTICAL
            }
            // log.Printf( "mapper4 set mirror:%d, with %d->$%X" , self.mirrormode, data, addr  )
        } else {  // odd
            // PRG Ram Protect
            // TODO:
        }
        return false, 0
    }

    // handle IRQ
    if addr >= 0xC000 && addr <= 0xDFFF {
        if addr & 0x0001 == 0{ // even
            self.nIRQReload = uint16(data)
        } else { // odd
            self.nIRQCounter = 0x0000
        }
        return false, 0
    }

    // Enable/Disable IRQ
    if addr >= 0xE000 && addr <= 0xFFFF {
        if addr & 0x0001 == 0 {  // even
            self.bIRQEnable = false
            self.bIRQActive = false
        } else { // odd
            self.bIRQEnable = true
        }
        return false, 0
    }

    return false, 0
}





func (self *Mapper_4) PpuMapRead( addr uint16) (bool,uint) {
    var mapped_addr uint = 0

    if addr >= 0x0000 && addr <= 0x1FFF {
        idx_chunk := addr / 0x400
        mapped_addr = uint(self.pCHRBank[idx_chunk] + uint32(addr & 0x03FF))
        return true, mapped_addr
    }
    return false, 0
}

func (self *Mapper_4) PpuMapWrite( addr uint16) (bool,uint) {
    var mapped_addr uint = 0
    if addr >= 0x0000 && addr <= 0x1FFF {
        if self.nCHRBanks == 0 {
            /*
            mapped_addr = uint(addr)
            /*/
            idx_chunk := addr / 0x400
            mapped_addr = uint(self.pCHRBank[idx_chunk] + uint32(addr & 0x03FF))
            //*/
            return true, mapped_addr
        }
    }
    return false, 0
}

func (self *Mapper_4) Reset() {
    self.nTargetRegister = 0x00
    self.bPRGBankMode = false
    self.bCHRInversion = false
    self.mirrormode = MIRROR_HORIZONTAL

    self.bIRQActive = false
    self.bIRQEnable = false
    self.bIRQUpdate = false
    self.nIRQCounter = 0x0000
    self.nIRQReload = 0x0000

    for i:=0; i< len(self.pPRGBank); i++ {
        self.pPRGBank[i] = 0
    }
    for i:=0; i< len(self.pCHRBank); i++ {
        self.pCHRBank[i] = 0
        self.pRegister[i] = 0
    }

    self.pPRGBank[0] = 0 * 0x2000
    self.pPRGBank[1] = 1 * 0x2000

    // by default it's in this fixed upper 16k mode
    self.pPRGBank[2] = (uint32(self.nPRGBanks) * 2 - 2) * 0x2000
    self.pPRGBank[3] = (uint32(self.nPRGBanks) * 2 - 1) * 0x2000
}

func (self *Mapper_4) IrqState() bool {
    return self.bIRQActive
}

// called by bus if an IRQ is triggered
func (self *Mapper_4) IrqClear() {
    self.bIRQActive = false
}

func (self *Mapper_4) Scanline() {
    if self.nIRQCounter == 0 {
        self.nIRQCounter = self.nIRQReload
    } else {
        self.nIRQCounter--
    }

    if self.nIRQCounter == 0 && self.bIRQEnable {
        self.bIRQActive = true
    }
}

func (self *Mapper_4) Mirror() int {
    return self.mirrormode
}

func (self *Mapper_4) updateBankOfset() {
    // Update Pointer Table
    if self.bCHRInversion {
        self.pCHRBank[0] = self.chrBankOffset( self.pRegister[2] )
        self.pCHRBank[1] = self.chrBankOffset( self.pRegister[3] )
        self.pCHRBank[2] = self.chrBankOffset( self.pRegister[4] )
        self.pCHRBank[3] = self.chrBankOffset( self.pRegister[5] )
        self.pCHRBank[4] = self.chrBankOffset( (self.pRegister[0] & 0xFE) )
        self.pCHRBank[5] = self.chrBankOffset( (self.pRegister[0] | 0x01) )
        self.pCHRBank[6] = self.chrBankOffset( (self.pRegister[1] & 0xFE) )
        self.pCHRBank[7] = self.chrBankOffset( (self.pRegister[1] | 0x01) )
    } else {
        self.pCHRBank[0] = self.chrBankOffset( (self.pRegister[0] & 0xFE) )
        self.pCHRBank[1] = self.chrBankOffset( (self.pRegister[0] | 0x01) )
        self.pCHRBank[2] = self.chrBankOffset( (self.pRegister[1] & 0xFE) )
        self.pCHRBank[3] = self.chrBankOffset( (self.pRegister[1] | 0x01) )
        self.pCHRBank[4] = self.chrBankOffset( self.pRegister[2] )
        self.pCHRBank[5] = self.chrBankOffset( self.pRegister[3] )
        self.pCHRBank[6] = self.chrBankOffset( self.pRegister[4] )
        self.pCHRBank[7] = self.chrBankOffset( self.pRegister[5] )
    }

    if self.bPRGBankMode {
        self.pPRGBank[0] = self.prgBankOffset( -2 )
        self.pPRGBank[2] = self.prgBankOffset( self.pRegister[6] )
    } else {
        self.pPRGBank[0] = self.prgBankOffset( self.pRegister[6] )
        self.pPRGBank[2] = self.prgBankOffset( -2 )
    }

    self.pPRGBank[1] = self.prgBankOffset( self.pRegister[7] )
    self.pPRGBank[3] = self.prgBankOffset( -1 ) // last 8k bank

    // log.Printf( "%+v, reg6:%x, reg7:%x" , self.pPRGBank , self.pRegister[6], self.pRegister[7]  )
}

func (self *Mapper_4) prgBankOffset(index int) uint32 {
    if index >= 0x80 {
        index -= 0x100
    }
    index %= int(self.nPRGBanks) * 2
    offset := index * 0x2000
    if offset < 0 {
        offset += int(self.nPRGBanks)*16*1024
    }
    return uint32(offset)
}

// CHR ROM 按 1k 拆分bank
func (self *Mapper_4) chrBankOffset(index int) uint32 {
    if index >= 0x80 {
        index -= 0x100
    }

    nActualCHRBank := self.nCHRBanks
    if nActualCHRBank == 0 {
        nActualCHRBank = 1
    }
    index %= int(nActualCHRBank) * 8
    offset := index * 0x0400
    if offset < 0 {
        offset += int(nActualCHRBank) * 8 * 1024
    }
    return uint32(offset)
}
