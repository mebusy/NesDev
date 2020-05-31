package cart

import (
    "io/ioutil"
    "log"
    "nes/mapper"
    "path"
    // "fmt"
    "encoding/binary"
    "os"
    "nes/tools"
)

type Cartridge struct {
    _RPGMemory []uint8
    _CHRMemory []uint8

    nMapperID int
    nPRGBanks int  // how many banks
    nCHRBanks int

    mapperRW  mapper.MAPPER_RW

    hw_mirror int
    battery bool

    gameHash string
    sram [8*1024]uint8
}

type Header struct {
    // name [4]byte
    prg_rom_chunks uint8
    chr_rom_chunks uint8
    Control1 uint8
    Control2 uint8
    prg_ram_size uint8
    tv_system1  uint8
    tv_system2  uint8
    // unused [5]byte
}


func NewCartridge( filename string ) *Cartridge {

    cart := &Cartridge {}

    var err error
    cart.gameHash, err  = tools.HashFile( filename )
    if err != nil {
        log.Fatal( err )
    }

    rom, err := ioutil.ReadFile( filename )
    if err != nil {
        log.Fatal( err )
    }

    header := Header{}
    idx := 4  // NES  name
    header.prg_rom_chunks = rom[idx] ; idx++
    header.chr_rom_chunks = rom[idx] ; idx++
    header.Control1 = rom[idx] ; idx++
    header.Control2 = rom[idx] ; idx++
    header.prg_ram_size = rom[idx] ; idx++
    header.tv_system1 = rom[idx] ; idx++
    header.tv_system2 = rom[idx] ; idx++
    idx += 5

    if header.Control1 & 0x04 != 0 {
        // 512 bytes for training information
        idx += 512
    }
    mapper1 := header.Control1 >> 4
    mapper2 := header.Control2 >> 4
    mapperID := mapper1 | mapper2<<4

    // mirroring type
    mirror1 := header.Control1 & 1
    mirror2 := (header.Control1 >> 3) & 1
    mirror := mirror1 | mirror2<<1

    // Determine Mapper ID
    cart.nMapperID = int( mapperID)  // uint8, >> is ok

    cart.hw_mirror = int(mirror)

    cart.battery = (header.Control1 >> 1) & 1 != 0
    // "Discover" File format 
    // There are actually 3 file format, but now I'm interested in is type 1
    nFileType := 1
    if (header.Control2 & 0x0C) == 0x08 {
        nFileType = 2
    }
    if nFileType == 1 {
        cart.nPRGBanks = int(header.prg_rom_chunks)
        cart._RPGMemory = make( []uint8, cart.nPRGBanks *16*1024 )
        idx += copy( cart._RPGMemory, rom[idx:] )

        cart.nCHRBanks = int(header.chr_rom_chunks)
        if cart.nCHRBanks == 0 {
            // Create CHR RAM
            cart._CHRMemory = make( []uint8, 1  *8*1024 )
        } else {
            // Allocate for ROM
            cart._CHRMemory = make( []uint8, cart.nCHRBanks *8*1024 )
            idx += copy( cart._CHRMemory, rom[idx:] )
        }

    } else if nFileType == 2 {
        cart.nPRGBanks = ((int(header.prg_ram_size) & 0x07) << 8) | int(header.prg_rom_chunks)
        cart._RPGMemory = make( []uint8, cart.nPRGBanks *16*1024 )
        idx += copy( cart._RPGMemory, rom[idx:] )

        cart.nCHRBanks = ((int(header.prg_ram_size) & 0x38) << 8) | int(header.chr_rom_chunks)
        if cart.nCHRBanks == 0 {
            // Create CHR RAM
            cart._CHRMemory = make( []uint8, 1  *8*1024 )
        } else {
            // Allocate for ROM
            cart._CHRMemory = make( []uint8, cart.nCHRBanks *8*1024 )
            idx += copy( cart._CHRMemory, rom[idx:] )
        }
        panic( "for debug..." )
    }

    if idx != len(rom) {
        log.Fatalf( "rom size is %d, but only %d bytes were read", len(rom), idx )
    }

    // load appropriate mapper
    switch cart.nMapperID {
    case 0:
        cart.mapperRW = mapper.NewMapper_0( header.prg_rom_chunks , header.chr_rom_chunks  )
    case 1:
        cart.mapperRW = mapper.NewMapper_1( header.prg_rom_chunks , header.chr_rom_chunks, cart.sram[:] )
    case 2:
        cart.mapperRW = mapper.NewMapper_2( header.prg_rom_chunks , header.chr_rom_chunks )
    case 3:
        cart.mapperRW = mapper.NewMapper_3( header.prg_rom_chunks , header.chr_rom_chunks  )
    case 4:
        cart.mapperRW = mapper.NewMapper_4( header.prg_rom_chunks , header.chr_rom_chunks , cart.sram[:] )
    case 66:
        cart.mapperRW = mapper.NewMapper_66( header.prg_rom_chunks , header.chr_rom_chunks  )
    default:
        log.Fatal( "unsupport mapper:" , cart.nMapperID )
    }

    log.Printf( "cartridge instanciated, mapper:%d, mirror:%d, RPG bank:%d, CHR bank:%d ", cart.nMapperID, cart.hw_mirror, cart.nPRGBanks, cart.nCHRBanks  )
    return cart
}

// Communication with Main Bus
// return false/true to tell the calling system whether 
//  the cartridge is handling that read or write
func (self *Cartridge) CpuWrite(addr uint16, data uint8) bool {
    if ok, mapped_addr := self.mapperRW.CpuMapWrite( addr, data); ok {
        if mapped_addr == mapper.READFROMSRAM {
            // Mapper has actually set the data value, for example cartridge based RAM
        } else {
            // Mapper has produced an offset into cartridge bank memory
            self._RPGMemory[mapped_addr] = data
        }
        return true
    }

    return false
}

func (self *Cartridge) CpuRead(addr uint16, data *uint8 ) bool {
    // var mapped_addr int = 0

    // if cpu read cartridge
    //  1. translate address
    //  2. read data from PROG ROM
    //  3. return data and true
    if ok, mapped_addr := self.mapperRW.CpuMapRead( addr, data); ok {
        if mapped_addr == mapper.READFROMSRAM {
            // Mapper has actually set the data value, for example cartridge based RAM
        } else {
            // Mapper has produced an offset into cartridge bank memory
            *data = self._RPGMemory[mapped_addr]
        }
        return true
    }
    return false
}

// Communication with PPU Bus
func (self *Cartridge) PpuWrite(addr uint16, data uint8) bool {
    if ok, mapped_addr := self.mapperRW.PpuMapWrite( addr); ok {
        self._CHRMemory[mapped_addr] = data
        return true
    }
    return false
}

func (self *Cartridge) PpuRead(addr uint16, data *uint8) bool {
    if ok, mapped_addr := self.mapperRW.PpuMapRead( addr); ok {
        *data = self._CHRMemory[mapped_addr]
        return true
    }
    return false
}

func (self *Cartridge) Reset() {
    if self.mapperRW != nil {
        self.mapperRW.Reset()
    }
}

func (self *Cartridge) Mirror() int {
    m := self.mapperRW.Mirror();
    if (m == mapper.MIRROR_HARDWARE) {
        // Mirror configuration was defined
        // in hardware via soldering
        return self.hw_mirror;
    } else {
        // Mirror configuration can be
        // dynamically set via mapper
        return m;
    }
}

func (self *Cartridge) GetMapper() mapper.MAPPER_RW {
    return self.mapperRW
}

func (self *Cartridge) Save( directory string ) {
    if !self.battery {
        return
    }
    filepath :=  path.Join( directory , self.gameHash + ".sram" )
    err := ioutil.WriteFile( filepath , self.sram[:]  , 0644)
    if err != nil {
        log.Println( "save sram failed: ", err  )
        return
    }
    log.Println( "sram saved" )
}

func (self *Cartridge) Load( directory string ) {
    if !self.battery {
        return
    }
    filepath :=  path.Join( directory , self.gameHash + ".sram" )
    file, err := os.Open(filepath)
    if err != nil {
        log.Println( "load sram: no .sram file found" )
        return
    }
    defer file.Close()

    if err := binary.Read(file, binary.LittleEndian, self.sram[:] ); err != nil {
        log.Println( "load sram failed:", err )
        return
    }
    log.Println( "sram loaded" )
}




