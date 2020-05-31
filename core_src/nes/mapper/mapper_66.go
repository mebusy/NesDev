package mapper


type Mapper_66 struct {
    Mapper_0
    // nPRGBanks uint8
    // nCHRBanks uint8
    nPRGBankSelect uint8
    nCHRBankSelect uint8
}

func NewMapper_66( prg_rom_chunks , chr_rom_chunks uint8 ) *Mapper_66 {
    m := &Mapper_66{}
    m.nPRGBanks = prg_rom_chunks
    m.nCHRBanks = chr_rom_chunks
    m.Reset()
    return m
}


func (self *Mapper_66) CpuMapRead( addr uint16, data *uint8 ) (bool,uint) {
    var mapped_addr uint = 0
    // pattern table
    if addr >= 0x8000 && addr <= 0xFFFF {
        mapped_addr = uint(self.nPRGBankSelect) * 0x8000 + uint(addr&0x7FFF)
        return true, mapped_addr
    }
    return false, mapped_addr
}

func (self *Mapper_66) CpuMapWrite( addr uint16, data uint8) (bool,uint) {
    if addr >= 0x8000 && addr <= 0xFFFF {
        self.nCHRBankSelect = data & 0x03
        self.nPRGBankSelect = (data & 0x30) >> 4
    }

    // Mapper has handled write, but do not update ROMs
    return false, 0
}

func (self *Mapper_66) PpuMapRead( addr uint16) (bool,uint) {
    var mapped_addr uint = 0
    // pattern table
    if addr >= 0x0000 && addr <= 0x1FFF {
        mapped_addr = uint(self.nCHRBankSelect) * 0x2000 + uint(addr)
        return true, mapped_addr
    }
    return false, mapped_addr
}

func (self *Mapper_66) PpuMapWrite( addr uint16) (bool,uint) {
    return false, 0
}

func (self *Mapper_66) Reset() {
    self.nCHRBankSelect = 0
    self.nPRGBankSelect = 0
}

