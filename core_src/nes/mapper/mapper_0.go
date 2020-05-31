package mapper

type Mapper_0 struct {
    nPRGBanks uint8
    nCHRBanks uint8

}

func NewMapper_0( prg_rom_chunks , chr_rom_chunks uint8 ) *Mapper_0 {
    m := &Mapper_0{}
    m.nPRGBanks = prg_rom_chunks
    m.nCHRBanks = chr_rom_chunks
    m.Reset()
    return m
}

// when cpu read/write to its system bus
// cartridge is only interested if the address is within the 
// range 0x8000 - 0xFFFF
func (self *Mapper_0) CpuMapRead( addr uint16, data *uint8 ) (bool,uint) {
    // if PRGROM is 16KB
    //     CPU Address Bus          PRG ROM
    //     0x8000 -> 0xBFFF: Map    0x0000 -> 0x3FFF
    //     0xC000 -> 0xFFFF: Mirror 0x0000 -> 0x3FFF
    // if PRGROM is 32KB
    //     CPU Address Bus          PRG ROM
    //     0x8000 -> 0xFFFF: Map    0x0000 -> 0x7FFF    
    var mapped_addr uint = 0
    if addr >= 0x8000 && addr <= 0xFFFF {
        // mask to ensure we are offsetting from 0
        if self.nPRGBanks > 1 {
            mapped_addr = uint(addr & 0x7FFF)
        } else {
            // mirror
            mapped_addr = uint(addr & 0x3FFF)
        }
        return true,mapped_addr
    }
    return false,mapped_addr
}
func (self *Mapper_0) CpuMapWrite( addr uint16, data uint8) (bool,uint) {
    var mapped_addr uint = 0
    if addr >= 0x8000 && addr <= 0xFFFF {
        // mask to ensure we are offsetting from 0
        if self.nPRGBanks > 1 {
            mapped_addr = uint(addr & 0x7FFF)
        } else {
            // mirror
            mapped_addr = uint(addr & 0x3FFF)
        }
        return true, mapped_addr
    }
    return false, mapped_addr
}

func (self *Mapper_0) PpuMapRead( addr uint16) (bool,uint) {
    var mapped_addr uint = 0
    // pattern table
    if addr >= 0x0000 && addr <= 0x1FFF {
        mapped_addr = uint(addr)
        return true, mapped_addr
    }
    return false, mapped_addr
}
func (self *Mapper_0) PpuMapWrite( addr uint16) (bool,uint) {
    var mapped_addr uint = 0
    // pattern table
    // generally ppu write is no sense
    // in case of that the cartridge has no ROM at that position,
    // it may have pattern RAM, 
    // in which case we would want to the pattern memory
    if addr >= 0x0000 && addr <= 0x1FFF {
        if self.nCHRBanks == 0 {
            mapped_addr = uint(addr)
            return true, mapped_addr
        }
    }
    return false, mapped_addr
}

func (self *Mapper_0) Reset() {
}

// Get Mirror mode if mapper is in control
func (self *Mapper_0) Mirror() int {
    return MIRROR_HARDWARE
}

func (self *Mapper_0) IrqState() bool {
    return false
}
func (self *Mapper_0) IrqClear() {
}

func (self *Mapper_0) Scanline() {
}

