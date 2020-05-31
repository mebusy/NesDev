package nes

import (
    "log"
    "nes/cpu"
    "nes/tools"
    "strconv"
    "strings"

    // "io/ioutil"
    "nes/cart"
    "nes/ppu"
    // "nes/sprite"
    // "nes/color"
    "nes/apu"
)

/*
Bus Read & Write

One of the reason we don't need to emulate the R/W signal
is that it is implied by which function is called.
We're either writing to the bus or reading from the bus.
*/
type Bus struct {

    // devices on bus
    cpu *cpu.Cpu
    // cpu ram
    pram [2*1024]uint8

    ppu *ppu.Ppu

    apu *apu.Apu

    cartridge *cart.Cartridge

    Controller [2]uint8
    // store the snap-shot of the input
    // when the corresponding memory address is written to
    controller_state [2]uint8

    SystemClockCounter int64

    // handle DMA
    dma_page uint8
    dma_count uint8  // together with dma_page to form a 16bit address
    dma_data uint8

    dma_transfer bool  // whethe dma is in process
    dma_dummy bool

    // At any point the emulation cat interrogate the instance of BUS(i.e. NES) and
    // say what is your current audio output
    dAudioTimePerSystemSample float64
    dAudioTimePerNESClock float64
    dAudioTime float64
}


func NewBus() *Bus {
    log.Println( "bus instanciated" )

    // create cpu
    _cpu := cpu.NewCPU()

    // create ppu
    _ppu := ppu.NewPpu()

    // create apu
    _apu := apu.NewApu()

    // create bus
    _bus := &Bus{ dma_dummy:true }
    _bus.resetPram()

    _bus.cpu = _cpu
    _bus.cpu.ConnectToBus( _bus )
    _bus.ppu = _ppu
    _bus.apu = _apu

    return _bus
}

func (self *Bus) resetPram() {
    for i := range self.pram {
        self.pram[i] = 0
    }
    log.Println( "ram reseted" )
}

// only CPU can read/write bus
func (self *Bus) CpuWrite(addr uint16, data uint8) {
    // give the cartridge of 1st depth on all R/W
    if self.cartridge.CpuWrite( addr, data ) {

    } else if addr >= 0x0000 && addr <= 0x1FFF {
        // write to PRAM
        self.pram[addr & 0x07FF] = data
    } else if addr >= 0x2000 && addr <= 0x3FFF {
        // targeting PPU, which only have 8 addressable elements 
        // mirroring through the range 0x2000 ~ 0x3FFF
        // 8 byte , mirror 
        self.ppu.CpuWrite( addr & 0x07 , data )
    } else if (addr >= 0x4000 && addr<= 0x4013) || addr == 0x4015 || addr == 0x4017 {
        // NES APU
        self.apu.CpuWrite( addr, data )
    } else if addr == 0x4014 {   // DMA
        // A write to this address initiates a DMA transfer
        self.dma_page = data
        self.dma_count = 0
        self.dma_transfer = true
    } else if addr >= 0x4016 && addr <= 0x4017 {
        self.controller_state[addr & 1] = self.Controller[addr & 1]
    }
}

// bReadonly is used for the disassembler.
func (self *Bus) CpuRead(addr uint16, bReadonly bool) uint8 {
    var data uint8 = 0

    if self.cartridge.CpuRead( addr , &data ) {

    } else  if addr >= 0x0000 && addr <= 0x1FFF {
        // 2K PRAM, mirror
        data = self.pram[addr & 0x07FF]
    } else if addr >= 0x2000 && addr <= 0x3FFF {
        data = self.ppu.CpuRead( addr & 0x07, bReadonly )
    } else if addr == 0x4015 {
        // APU Read Status
        data = self.apu.CpuRead(addr)
    } else if addr >= 0x4016 && addr <= 0x4017 {
        // Read out the MSB of the controller status word
        data = tools.B2i(self.controller_state[addr & 1] & 0x80 != 0)
        self.controller_state[addr & 1] <<= 1
    }
    return data
}

// System Interface

// reset NES
func (self *Bus) Reset() {
    // verified
    self.cartridge.Reset()
    self.ppu.Reset()
    self.cpu.Reset()
    self.resetPram()

    self.SystemClockCounter = 0

    self.dma_page = 0
    self.dma_count = 0
    self.dma_data = 0
    self.dma_dummy = true
    self.dma_transfer = false
}

func (self *Bus) Clock() bool {
    self.ppu.Clock()

    self.apu.Clock()

    // cpu's clock is 3 times slower
    if self.SystemClockCounter %3 == 0 {
        // stop the CPU from being clocked if a 
        // DMA transfer is happening
        if self.dma_transfer {
            // We need to wait until the next even CPU clock cycle
            // before DMA transfer starts...
            // The peripheral responsible for handling this transfer
            //  is in some way synchronized with the CPU clock. And as 
            //  with most digital design it's constantly respongding to 
            //  to changes in input, but we only take the output when 
            //  we know that it's going to be valid. 
            // The DMA device is synchronized with every other CPU clock,
            //  which means we may have to wait 1 or 2 clock before the DAM
            //  can start.

            if self.dma_dummy {
                // waiting for synchronize
                // clock 0*n is for cpu, 1*n is for dma waitting  , 
                if self.SystemClockCounter % 2 == 1 {
                    self.dma_dummy = false
                }
            } else {
                // clock 2*n is for starting DAM
                // on even cycle, I read data from CPU
                if self.SystemClockCounter % 2 == 0 {
                    self.dma_data = self.CpuRead( (uint16(self.dma_page)<<8)|uint16( self.dma_count ) ,false )
                } else {
                    // on odd cycle, I write data to PPU
                    self.ppu.DMATransfer( self.dma_data  )

                    self.dma_count++
                    // finish DMA
                    if self.dma_count == 0 {
                        self.dma_transfer = false
                        self.dma_dummy = true
                    }
                }
            }

        } else {
            // normal cpu clock
            self.cpu.Clock()
        }
    }

    // Synchronizing with Audio
    bAudioSampleReady := false
    self.dAudioTime += self.dAudioTimePerNESClock
    if self.dAudioTime >= self.dAudioTimePerSystemSample {
        self.dAudioTime = 0
        bAudioSampleReady = true
        self.apu.SendOutputSample()
    }

    // nmi interrupt was emitted from the PPU to the CPU
    if self.ppu.Nmi {
        self.ppu.Nmi = false
        self.cpu.Nmi()
    }

    // check if cartridge is requesting IRQ
    if self.cartridge.GetMapper().IrqState() {
        self.cartridge.GetMapper().IrqClear()
        self.cpu.Irq()
    }

    self.SystemClockCounter++

    return bAudioSampleReady
}
func (self *Bus) InsertCartridge( cart *cart.Cartridge ) {
    self.cartridge = cart
    self.ppu.ConnectCartridge( cart )
}

func (self *Bus) SetAudioSampleRate( sample_rate int ) {
    log.Println( "audio sample_rate setted:", sample_rate )
    self.dAudioTimePerSystemSample = 1/float64(sample_rate)
    self.dAudioTimePerNESClock = 1.0 / 5369318.0 // PPU Clock Frequency
    // log.Println( self.dAudioTimePerNESClock , self.dAudioTimePerSystemSample ,  self.dAudioTimePerSystemSample / self.dAudioTimePerNESClock )
}

func (self *Bus) SetAudioChannel(channel chan float32) {
    self.apu.Ch_sample = channel
}

func (self *Bus) StepSeconds( seconds float64 ) {
    /*
    cycles := int64(float64(cpu.CPU_FREQUENCY) * 3 * seconds )
    target_clock_counter := self.SystemClockCounter  + cycles
    for self.SystemClockCounter < target_clock_counter {
        self.DebugStepInstruction()
    }
    /*/
    self.DebugSingleFrame( true )
    //*/
}


// =================== DEBUG ======================

// set dummy palette for exporting Pattern Image
func (self *Bus) Debug_SetDebugPalette()  {
    for i:=0 ; i<32; i++ {
        self.ppu.DebugSetPalette( i, i )
    }
}


// addr : specify the start address of programm data
func (self *Bus) DebugLoadCode2PRam( codes string , addr uint16 ) {
    // TODO
    // set Reset Vector
    // [0xFFFC] = uint8(addr&0xFF)
    // [0xFFFD] = uint8( (addr>>8)&0xFF )

    // Dont forget to set IRQ and NMI vectors if you want to play with those
    // just for test
    // [0xFFFE] = uint8(addr&0xFF)
    // [0xFFFF] = uint8( (addr>>8)&0xFF )

    // ok, rom call override vector setting
    for i, c := range strings.Fields( codes ) {
        v , err := strconv.ParseInt( c , 16, 0 )
        if err != nil {
            log.Fatal( err )
        }

        self.pram[ int(addr) + i ] = uint8(v)
    }
}


func (self *Bus) DebugDumpCpu() cpu.Cpu {
    return *self.cpu
}
func (self *Bus) GetApu() *apu.Apu {
    return self.apu
}
func (self *Bus) GetPpu() *ppu.Ppu {
    return self.ppu
}
func (self *Bus) GetCartrideg() *cart.Cartridge {
    return self.cartridge
}

// func (self *Bus) DebugSetPC( addr uint16 )  {
//     self.cpu.PC = addr
// }

func (self *Bus) DebugStepInstruction() {
    // Clock enough times to execute a whole CPU instruction
    for {
        self.Clock()
        if self.cpu.IsInstructionComplete() {
            break
        }
    }
    // CPU clock runs slower than system clock, so it may be
    // complete for additional system clock cycles. Drain
    // those out
    for {
        self.Clock()
        if !self.cpu.IsInstructionComplete() {
            break
        }
    }
}


func (self *Bus) DebugSingleFrame( bIgnoreInstuction bool) {
    // Clock enough times to draw a single frame
    for {
        self.Clock()
        if self.ppu.Frame_complete {
            break
        }
    }
    if !bIgnoreInstuction {
        // Use residual clock cycles to complete current instruction
        for {
            self.Clock()
            if self.cpu.IsInstructionComplete() {
                break
            }
        }
    }
    // Reset frame completion flag
    self.ppu.Frame_complete = false;
}

