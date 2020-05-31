package mapper


type Mapper_2 struct {
    Mapper_0
    // nPRGBanks uint8
    // nCHRBanks uint8
    nPRGBankSelectLo uint8  // bank: $8000-$BFFF
    nPRGBankSelectHi uint8
}

func NewMapper_2( prg_rom_chunks , chr_rom_chunks uint8 ) *Mapper_2 {
    m := &Mapper_2{}
    m.nPRGBanks = prg_rom_chunks
    m.nCHRBanks = chr_rom_chunks
    m.Reset()
    return m
}


func (self *Mapper_2) CpuMapRead( addr uint16, data *uint8 ) (bool,uint) {
    var mapped_addr uint = 0
    // lo
    if addr >= 0x8000 && addr <= 0xBFFF {
        mapped_addr = uint(uint(self.nPRGBankSelectLo) * 0x4000 + uint(addr & 0x3FFF))
        return true, mapped_addr
    }
    // hi
    if addr >= 0xC000 && addr <= 0xFFFF {
        mapped_addr = uint(uint(self.nPRGBankSelectHi) * 0x4000 + uint(addr & 0x3FFF))
        return true, mapped_addr
    }

    return false, mapped_addr
}

func (self *Mapper_2) CpuMapWrite( addr uint16, data uint8) (bool,uint) {
    if addr >= 0x8000 && addr <= 0xFFFF {
        self.nPRGBankSelectLo = data & 0x0F;
    }

    // Mapper has handled write, but do not update ROMs
    return false, 0

}

func (self *Mapper_2) Reset() {
    self.nPRGBankSelectLo = 0
    self.nPRGBankSelectHi = self.nPRGBanks -1
}


