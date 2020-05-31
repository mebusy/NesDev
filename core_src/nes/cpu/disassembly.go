package cpu

/*
import (
    "fmt"
)

func (self *Cpu) disassemble( nStart , nStop uint16 ) {
    var addr uint32 = uint32(nStart)
    var value, lo, hi uint8 = 0
    var line_addr uint16 = 0

    mapLines := map[uint16]string {}

    // Starting at the specified address we read an instruction byte
    // which in turn yields information from the lookup table as to 
    // how many additional bytes we need to read and what the addressing mode is.

    for addr <= uint32(nStop) {
        line_addr = uint16(addr)

        // Prefix line with instruction address
        sInst = fmt.Sprintf("$%04X: ", addr)
        // Read instruction, and get its readable name
        opcode := bus.CpuRead( uint16(addr), true )
    }
}
//*/
