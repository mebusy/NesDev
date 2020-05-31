package mapper


type Mapper_1 struct {
    Mapper_0
    // nPRGBanks uint8
    // nCHRBanks uint8
    nLoadRegister uint8
    nLoadRegisterCount uint8
    nControlRegister uint8

    nCHRBankSelect4Lo uint8
    nCHRBankSelect4Hi uint8
    nCHRBankSelect8 uint8       // whole 8k

    nPRGBankSelect16Lo uint8
    nPRGBankSelect16Hi uint8
    nPRGBankSelect32 uint8   // whole 32k 

    mirrormode  int
    // if you wanna save game, dump RAMStatic to file
    RAMStatic  []uint8

}

func NewMapper_1( prg_rom_chunks , chr_rom_chunks uint8, sram []uint8 ) *Mapper_1 {
    m := &Mapper_1{}
    m.nPRGBanks = prg_rom_chunks
    m.nCHRBanks = chr_rom_chunks
    m.mirrormode = MIRROR_HORIZONTAL
    m.RAMStatic = sram

    m.Reset()
    return m
}


func (self *Mapper_1) CpuMapRead( addr uint16, data *uint8 ) (bool,uint) {

    var mapped_addr uint = 0
    if addr >= 0x6000 && addr <= 0x7FFF {
        // Read is from static ram on cartridge
        mapped_addr = READFROMSRAM
        // Read data from RAM
        *data = self.RAMStatic[addr & 0x1FFF]  // 8k mapping
        // Signal mapper has handled request
        return true, mapped_addr
    }

    if addr >= 0x8000 {
        // PRG bank mode
        if self.nControlRegister & 0b01000 != 0 {
            // 16K Mode
            if addr >= 0x8000 && addr <= 0xBFFF {
                mapped_addr = uint(self.nPRGBankSelect16Lo) * 0x4000 + uint(addr & 0x3FFF)
                return true, mapped_addr
            }

            if addr >= 0xC000 && addr <= 0xFFFF {
                mapped_addr = uint(self.nPRGBankSelect16Hi) * 0x4000 + uint(addr & 0x3FFF)
                return true, mapped_addr
            }
        } else {
            // 32K Mode
            mapped_addr = uint(self.nPRGBankSelect32) * 0x8000 + uint(addr & 0x7FFF)
            return true, mapped_addr
        }
    }

    return false, 0
}

func (self *Mapper_1) CpuMapWrite( addr uint16, data uint8) (bool,uint) {
    var mapped_addr uint = 0
    if addr >= 0x6000 && addr <= 0x7FFF {
        // Write is to static ram on cartridge
        mapped_addr = READFROMSRAM
        // Write data to RAM
        self.RAMStatic[addr & 0x1FFF] = data
        // Signal mapper has handled request
        return true, mapped_addr
    }

    if addr >= 0x8000 {
        if data & 0x80 != 0 {
            // MSB is set, so reset serial loading
            self.nLoadRegister = 0x00
            self.nLoadRegisterCount = 0
            self.nControlRegister = self.nControlRegister | 0x0C
        } else {
            // Load data in serially into load register
            // It arrives LSB first, so implant this at
            // bit 5. After 5 writes, the register is ready
            self.nLoadRegister >>= 1
            self.nLoadRegister &^= 1<<4 // mebusy, add
            self.nLoadRegister |= (data & 0x01) << 4
            self.nLoadRegisterCount++

            if self.nLoadRegisterCount == 5 {
                // Get Mapper Target Register, by examining
                // bits 13 & 14 of the address
                nTargetRegister := uint8((addr >> 13) & 0x03)
                if nTargetRegister == 0 { // 0x8000 - 0x9FFF
                    // Set Control Register
                    self.nControlRegister = self.nLoadRegister & 0x1F

                    switch (self.nControlRegister & 0x03) {
                    case 0: self.mirrormode = MIRROR_ONESCREEN_LO
                    case 1: self.mirrormode = MIRROR_ONESCREEN_HI
                    case 2: self.mirrormode = MIRROR_VERTICAL
                    case 3: self.mirrormode = MIRROR_HORIZONTAL
                    }
                } else if nTargetRegister == 1 { // 0xA000 - 0xBFFF 
                    // Set CHR Bank Lo
                    if self.nControlRegister & 0b10000 != 0 {
                        // 4K CHR Bank at PPU 0x0000
                        self.nCHRBankSelect4Lo = self.nLoadRegister & 0x1F
                    } else {
                        // 8K CHR Bank at PPU 0x0000
                        self.nCHRBankSelect8 = self.nLoadRegister & 0x1E
                    }
                } else if nTargetRegister == 2 { // 0xC000 - 0xDFFF 
                    // Set CHR Bank Hi
                    if self.nControlRegister & 0b10000 != 0 {
                        // 4K CHR Bank at PPU 0x1000
                        self.nCHRBankSelect4Hi = self.nLoadRegister & 0x1F
                    }
                } else if (nTargetRegister == 3) { // 0xE000 - 0xFFFF
                    // Configure PRG Banks
                    nPRGMode := (self.nControlRegister >> 2) & 0x03

                    if (nPRGMode == 0 || nPRGMode == 1) {
                        // Set 32K PRG Bank at CPU 0x8000
                        self.nPRGBankSelect32 = (self.nLoadRegister & 0x0E) >> 1
                    } else if (nPRGMode == 2) {
                        // Fix 16KB PRG Bank at CPU 0x8000 to First Bank
                        self.nPRGBankSelect16Lo = 0
                        // Set 16KB PRG Bank at CPU 0xC000
                        self.nPRGBankSelect16Hi = self.nLoadRegister & 0x0F
                    } else if (nPRGMode == 3) {
                        // Set 16KB PRG Bank at CPU 0x8000
                        self.nPRGBankSelect16Lo = self.nLoadRegister & 0x0F
                        // Fix 16KB PRG Bank at CPU 0xC000 to Last Bank
                        self.nPRGBankSelect16Hi = self.nPRGBanks - 1
                    }
                }

                // 5 bits were written, and decoded, so
                // reset load register
                self.nLoadRegister = 0x00
                self.nLoadRegisterCount = 0
            } // end if nLoadRegisterCount == 5
        } // end  data & 0x80 == 0
    } // end if addr >= 0x8000

    // Mapper has handled write, but do not update ROMs
    return false, 0
}

func (self *Mapper_1) PpuMapRead( addr uint16) (bool,uint) {
    var mapped_addr uint = 0
    if addr < 0x2000 {
        if self.nCHRBanks == 0 {
            mapped_addr = uint(addr)
            return true, mapped_addr
        } else {
            if self.nControlRegister & 0b10000 != 0{
                // 4K CHR Bank Mode
                if addr >= 0x0000 && addr <= 0x0FFF {
                    mapped_addr = uint(self.nCHRBankSelect4Lo) * 0x1000 + uint(addr & 0x0FFF)
                    return true, mapped_addr
                }

                if addr >= 0x1000 && addr <= 0x1FFF {
                    mapped_addr = uint(self.nCHRBankSelect4Hi) * 0x1000 + uint(addr & 0x0FFF)
                    return true, mapped_addr
                }
            } else {
                // 8K CHR Bank Mode
                mapped_addr = uint(self.nCHRBankSelect8) * 0x1000 + uint(addr & 0x1FFF)
                return true, mapped_addr
            }
        }
    } // end if addr < 0x2000 {

    return false, mapped_addr
}

func (self *Mapper_1) Reset() {
    self.nControlRegister = 0x1C
    self.nLoadRegister = 0x00
    self.nLoadRegisterCount = 0x00

    self.nCHRBankSelect4Lo = 0
    self.nCHRBankSelect4Hi = 0
    self.nCHRBankSelect8 = 0

    self.nPRGBankSelect32 = 0
    self.nPRGBankSelect16Lo = 0
    self.nPRGBankSelect16Hi = self.nPRGBanks - 1
}

func (self *Mapper_1) Mirror() int {
    return self.mirrormode
}

