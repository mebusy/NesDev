package cpu

import (
    "log"
)

const (
    SP_BASE uint16 = 0x100
    SP_INIT uint8 = 0xFF
    CPU_FREQUENCY int = 1789773
)

type Cpu struct {
    A uint8
    X uint8
    Y uint8
    SP uint8
    PC uint16
    Status FLAG_CPU

    // fetch
    fetched uint8
    /*
    Depending on the addressing mode, we might want to read from 
    different location of the memory,  so I'm going to store that 
    location in the variable:  addr_abs
    */
    addr_abs uint16
    /*
    Branch instructions can only jump a certain distance from the 
    location where the instruction was called. 
    so they jump to a relative address.
    */
    addr_rel uint16
    // opcode currently working with
    opcode  uint8
    // circles left for the duration of this instruction
    cycles int

}

func NewCPU() *Cpu {
    log.Println( "cpu instanciated" )
    return &Cpu { A:0,X:0,Y:0,SP:0,PC:0,Status:0,
        fetched:0,addr_abs:0,addr_rel:0,opcode:0,cycles:0 }
}

type BUS_RW interface {
    CpuWrite( addr uint16, data uint8 )
    CpuRead ( addr uint16, bReadonly bool ) uint8
}

var bus BUS_RW   // bus in cpu module
func (self *Cpu) ConnectToBus ( _bus BUS_RW ) {
    log.Println( "cpu connected to bus" )
    bus = _bus
}

// the R/W on bus, is actually done by CPU
func (self *Cpu) write( addr uint16, data uint8  ) {
    bus.CpuWrite(addr, data)
}

func (self *Cpu) read ( addr uint16 ) uint8 {
    // In normal operation "read only" is set to false. 
    // This may seem odd. Some devices on the bus may change state when they are read from,
    //  and this is INTENTIONAL under normal circumstances.
    // However the disassembler will want to read the data at an address without changing the state of the devices on the bus
    return bus.CpuRead(addr,false)
}

func (self *Cpu) readNextInstructionBytePC() uint8 {
    d := self.read( self.PC )
    // mostly, when you perform a read, you're going to increment PC
    self.PC ++
    return d
}

func (self *Cpu) readWord( addr uint16 ) uint16 {
    lo := uint16(self.read(addr + 0))
    hi := uint16(self.read(addr + 1))
    // log.Printf( "PC: $%02X%02X,  %04X", hi, lo,  (hi << 8) | lo )
    return (hi << 8) | lo
}


type FLAG_CPU uint8

/*
C: is set either by the user to inform an operation that
    we want to use a carry bit , or it is set by the 
    operation itself
Z: is set mostly whenever the result of an operation equals 0
B: indicates that the break operation has been called 
*/
const (
    FLAG_CPU_C FLAG_CPU = 1<<iota    // Carry Bit
    FLAG_CPU_Z      // Zero
    FLAG_CPU_I      // Disable Interrupts
    FLAG_CPU_D      // Decimal Mode (unused in this impl)
    FLAG_CPU_B      // Break
    FLAG_CPU_U      // Unused
    FLAG_CPU_V      // Overflow
    FLAG_CPU_N      // Negative
)

func (self *Cpu) getFlag( f FLAG_CPU ) uint16 {
    if ( self.Status & f ) != 0 {
        return 1
    } else {
        return 0
    }
}

func (self *Cpu) setFlag( f FLAG_CPU , v bool ) {
    if v {
        self.Status |= f
    } else {
        self.Status &^= f
    }
}

func (self *Cpu) SetFlagsNZC( result uint16, flags FLAG_CPU ) {

    if flags & FLAG_CPU_Z != 0 {
        self.setFlag( FLAG_CPU_Z , (result&0xFF) == 0 )

        flags &^= FLAG_CPU_Z
    }
    if flags & FLAG_CPU_N != 0 {
        self.setFlag( FLAG_CPU_N , (result & 0x80) != 0 )

        flags &^= FLAG_CPU_N
    }
    if flags & FLAG_CPU_C != 0 {
        self.setFlag( FLAG_CPU_C , (result&0xFF00) != 0 )

        flags &^= FLAG_CPU_C
    }

    if flags != 0 {
        log.Fatalf( "[fatal] unknow flag in SetFlags : %02X  " , flags )
    }
}


// for debugging / logging stuff
var clock_count = 0

/*
indicate to the CPU that we want one clock cycle to occur
*/
func (self *Cpu) Clock( ) {
    // each instruction requires serveral clock cycles to execute 
    //  only going to do the execution when internal `cycles` is equal to 0
    if self.cycles == 0 {
        // first, read next instruction byte
        self.opcode = self.readNextInstructionBytePC()

        // Always set the Unused status flag bit to 1
        self.setFlag( FLAG_CPU_U , true )


        // get cycles
        self.cycles = GetOpCycles( self.opcode )
        // perform fetch of intermediate DATA using the required addressing mode
        additional_cycle1 := GetOpAddrMode( self.opcode )( self )
        // Perform operation
        GetOpOperate( self.opcode )( self )

        additional_cycle2 := 0
        if IsExtraCycleValidOpcode( self.opcode ) || IsExtraCycleInvalidOpcode( self.opcode ) {
            additional_cycle2 = 1
        }

        // add an extra cycle if both addrmode and instruction 
        // tend to potentially need mmore 1 cycle
        self.cycles += additional_cycle1 & additional_cycle2

        // Always set the Unused status flag bit to 1
        self.setFlag( FLAG_CPU_U , true )

    }

    // every time call clock function 
    // 1 cycle has elapsed
    self.cycles --

    // for debugging / logging stuff
    clock_count++;
}

/*
Interrupts:
reset/irq/nmi can occur any time they need to behave asynchronously.
and they interrupt the processor from doing its current job/
however it will finish the current instruction its executing.
Depending upon the type interrupts , various things on process change.
*/


// Forces the 6502 into a known state. 
// This is hard-wired inside the CPU. 

func (self *Cpu) Reset( ) {
    // verified

    // Generally the program data resides somewhere further along in the address memory.

    // when reset is called on the 6502, it looks directly to location 0xFFFC to 
    //  try and read the 16-bit address.  
    // Typically the programmer would set the value at location 0xFFFC at compile time.
    // So the chip knows that in the event of `reset` , it should always look at this address to 
    //  get the address to set its PC to.

    self.addr_abs = 0xFFFC
    lo := uint16(self.read(self.addr_abs + 0))
    hi := uint16(self.read(self.addr_abs + 1))
    // Set PC
    self.PC = (hi << 8) | lo

    // The registers are set to 0x00, 
    self.A = 0
    self.X = 0
    self.Y = 0

    // Your reset routine will normally include LDX #$FF, TXS to initialize the stack, since it it not self-initializing.
    // so why need we do it ? 
    self.SP = SP_INIT

    // The status register is cleared except for unused bit which remains at 1.  
    self.Status = 0x00 | FLAG_CPU_U

    // Clear internal helper variables
    self.addr_abs = 0
    self.addr_rel = 0
    self.fetched = 0

    // important
    // reset takes some time
    self.cycles = 8

    log.Println( "cpu reseted" )
}


// for irq , and nmi interrupt
// When there is an interrupt request, the current instruction is allowed to finish.
//   (which I facilitate by doing the whole thing when cycles == 0) 
// When the routine that services the interrupt has finished, the status register
//  and program counter can be restored to how they where before it occurred. 
// This is impemented by the "RTI" instruction.
func (self *Cpu) performInterrupt( interrutp_entry uint16 , take_cycles int ) {
        // store current PC
        self.push2stackWord( self.PC )

        self.setFlag( FLAG_CPU_B , false )
        self.setFlag( FLAG_CPU_U , true )

        // store current status
        self.push2stackByte( uint8(self.Status) )

        // like reset, a hard-coded address is integrated to get the new value of PC.
        self.addr_abs = interrutp_entry
        lo := uint16(self.read(self.addr_abs + 0))
        hi := uint16(self.read(self.addr_abs + 1))
        // Set PC
        self.PC = (hi << 8) | lo

        // important
        self.setFlag( FLAG_CPU_I , true )
        // important
        // interrupt takes some time
        self.cycles = take_cycles

}

/* 
interrupt request
it is the stardard interrupts, and can be ignored depending on 
    whether the interrupt flag is set or not 
*/
func (self *Cpu) Irq( ) {
    // can be ignored if the disable interrupt bit has been SET
    if self.getFlag( FLAG_CPU_I ) == 0 {
        // interrupt want to run a certain piece of code to service the interrupt  
        self.performInterrupt( 0xFFFE, 7 )
        // log.Println("irq")
    }
}

/* 
non-maskable interrupt request can NOT be disabled. 
used for communication between CPU and PPU
when PPU entering VBLANK, it send NMI request to CPU.
*/
func (self *Cpu) Nmi( ) {
    // non-maskable interrupt is exactly the same except nothign can stop this
    self.performInterrupt( 0xFFFA, 8 )
}

// Internal Helper function

// fetch data from appropriate source, stored in fetched variable
// Some instructions dont have to fetch data as the source is implied by the instruction.
//  For example "INX" increments the X register. There is no additional data required.
// For all other addressing modes, the data resides at the location held within 
//  addr_abs variable, so it is read from there. 
func (self *Cpu) fetch( ) uint8 {
    addrmode_name := GetOpAddrModeName( self.opcode )

    if addrmode_name != "IMP" {
        self.fetched = self.read( self.addr_abs )
    }
    // implied mode will fetch the data in IMP() method

    // return fetched for convenience
    return self.fetched
}


// helper function
func (self *Cpu) IsInstructionComplete() bool {
    return self.cycles == 0
}

