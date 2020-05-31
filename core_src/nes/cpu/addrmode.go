package cpu

import (
    "log"
)



// Addressing Modes , 12 in total

// those function 0 or 1 to indicate whether 
// there need to be another clock cycle
type FUNC_AddressingMode func(*Cpu) int

// implied
func IMP(self *Cpu) int {
    // there is actually no data as part of the instruction
    // it doesn't need do anything 

    // here , implied also means that it could be operating upon the accumulator
    // so I'm going to set my fetched variable to content of A 
    // TODO add constraints on Accumulator Addressing only ?
    self.fetched = self.A
    return 0
}
// immediate
func IMM(self *Cpu) int {
    // the data is supplied as part of the instruction 
    // it's going to be the next byte
    // all of my addressmode are going to set `addr_abs` variable
    // so the instruction knows where to read the data from when it need to
    self.addr_abs = self.PC; self.PC++

    return 0
}
// zero page
func ZP0(self *Cpu) int {
    // addr : 0xHHLL 
    // HH is the page , while LL is the offset in page

    // zero page addressing means the byte of data we are interested in reading 
    // can be found with page 0 , 0x0000~0x00FF
    self.addr_abs = uint16(self.readNextInstructionBytePC())

    return 0
}
// zero page, X-indexed
func ZPX(self *Cpu) int {
    // zero page , with X register offset
    // works like an array, where the starting address of that array is within page 0
    self.addr_abs =uint16( self.readNextInstructionBytePC() + self.X )
    self.addr_abs &= 0x00FF
    return 0
}
// zero page, Y-indexed
func ZPY(self *Cpu) int {
    // zero page , with Y register offset
    // works like an array, where the starting address of that array is within page 0
    self.addr_abs =uint16( self.readNextInstructionBytePC() + self.Y )
    self.addr_abs &= 0x00FF
    return 0
}
// absolute
func ABS(self *Cpu) int {
    lo := uint16( self.readNextInstructionBytePC() )
    hi := uint16( self.readNextInstructionBytePC() )
    self.addr_abs = (hi<<8)|lo

    return 0
}
// abs, X-indexed
func ABX(self *Cpu) int {
    lo := uint16( self.readNextInstructionBytePC() )
    hi := uint16( self.readNextInstructionBytePC() )
    self.addr_abs = (hi<<8)|lo

    self.addr_abs += uint16(self.X)

    // caveat in addressing mode 
    // cross boundary
    if (self.addr_abs & 0xFF00) != (hi<<8) {
        return 1
    }

    return 0
}
// abs, Y-indexed
func ABY(self *Cpu) int {
    lo := uint16( self.readNextInstructionBytePC() )
    hi := uint16( self.readNextInstructionBytePC() )
    self.addr_abs = (hi<<8)|lo

    self.addr_abs += uint16(self.Y)

    // caveat in addressing mode 
    // cross boundary
    if (self.addr_abs & 0xFF00) != (hi<<8) {
        return 1
    }

    return 0
}

// Note: The next 3 address modes use indirection (aka Pointers!)

// indirect
func IND(self *Cpu) int {
    // the complicated one that is effectively 6502's way of implementing pointers
    ptr_lo := uint16( self.readNextInstructionBytePC() )
    ptr_hi := uint16( self.readNextInstructionBytePC() )

    ptr := (ptr_hi<<8)|ptr_lo

    // if the lower byte of the ptr is equal to 0xFF, 
    // then the high byte of the final address we need to add 1 to it ,
    // we're effectively changing the page.  
    // For this particular instruction, that doen't actually happen.
    // nesdev.com/6502bugs.txt
    // An indirect JMP (xxFF) will fail because the MSB will be fetched from address xx00 instead of page xx+1.
    if ptr_lo == 0xFF {  // page boundary hardware bug
        self.addr_abs = ( uint16(self.read(ptr & 0xFF00 ))<<8) | uint16(self.read(ptr+0))
    } else {  // behavior normally
        self.addr_abs = ( uint16(self.read(ptr + 1      ))<<8) | uint16(self.read(ptr+0))
    }

    return 0
}


// indirect, zero page, X
func IZX(self *Cpu) int {
    // ZPX first
    //  somewhere in the zero page
    t := uint16( self.readNextInstructionBytePC() )
    //  and from that location, we offset that one byte address  by X
    // then IND
    lo := uint16(self.read( (t + uint16(self.X)  )& 0xFF ))
    hi := uint16(self.read( (t + uint16(self.X)+1)& 0xFF ))

    self.addr_abs = (hi<<8)|lo

    return 0
}
// indirect, zero page, Y
func IZY(self *Cpu) int {
    // IND in zero page first 
    t := uint16( self.readNextInstructionBytePC() )
    lo := uint16(self.read( (t  )& 0xFF ))
    hi := uint16(self.read( (t+1)& 0xFF ))

    self.addr_abs = (hi<<8)|lo

    // then Y-index
    self.addr_abs += uint16(self.Y)

    // page boundary check
    if (self.addr_abs & 0xFF00) != (hi<<8) {
        return 1
    }
    return 0
}

// relative
// This address mode is exclusive to branch instructions.
func REL(self *Cpu) int {
    self.addr_rel = uint16( self.readNextInstructionBytePC() )

    // only applies to branching instructions
    // and branching instructions can only jump a location that's 
    //  in the vicinity of the branch instruction
    // in fact, the address must reside within -128 to +127 of the branch instruction

    // the single byte that I read back is effectively unsigned
    //  but in order to jump backwards, it needs to be a signed data. 
    if (self.addr_rel & 0x80) != 0  {
        self.addr_rel |= 0xFF00   // expand the negative number to 16-bitwise
    }
    return 0
}


// helper
func GetOpAddrMode( opcode uint8 ) FUNC_AddressingMode {
    return _name2AddressingModeFunc(  GetOpAddrModeName( opcode )  )
}

func GetOpAddrModeName( opcode uint8 ) string {
    if !IsValidOpcode( opcode ) {
        return "IMP"
    }

    aaa := (opcode>>5)&7
    bbb := (opcode>>2)&7
    cc := opcode&3

    // general rules
    basic_addrmode := _tbl_bbbcc_addrmode_name[cc][bbb]

    // exceptions
    if cc == 2 {
        // with STX and LDX, "zero page,X" addressing becomes "zero page,Y"
        if bbb == 5 &&
            (aaa == 0x04 || aaa == 0x05)  {   // 100 STX, 101 LDX
            return "ZPY"
        }
        // and with LDX, "absolute,X" becomes "absolute,Y".
        if bbb == 7 && aaa == 5 {
            return "ABY"
        }

        // Note that bbb = 100 and 110 are missing.
        if bbb == 6 &&
            (aaa == 0x04 || aaa == 0x05)  {  // TXA, TSX 
            return "IMP"
        }
    } else if cc == 0 {
        // the only 1 indirect instruction
        if aaa == 3 && bbb == 3 {  // JMP ind
            return "IND"
        }

        // The conditional branch instructions all have the form xxy10000.
        // The flag indicated by xx is compared with y, 
        // and the branch is taken if they are equal.
        // BNE, BEQ, BCC, BCS, BPL, BMI, BVC, BVS
        if bbb == 4 {
            return "REL"
        }
        // 4 interrupt and subroutine instructions:
        // BR,JSR abs,RTI,RTS
        // (JSR is the only absolute-addressing instruction that doesn't fit the aaabbbcc pattern.)
        if opcode == 0x00 {  // BRK
            return "IMP"
        } else if opcode == 0x20 { // JSR abs
            return "ABS"
        } else if opcode == 0x40 || opcode == 0x60 {
            // RTI, RTS
            return "IMP"
        }


        // Other single-byte instructions: part 1
        // PHP	PLP	PHA	PLA	DEY	TAY	INY	INX
        //deasawCLC	SEC	CLI	SEI	TYA	CLV	CLD	SED 
        if opcode & 0xF == 0x08 {
            return "IMP"
        }
    }
    return basic_addrmode
}



func GetOpCycles( opcode uint8 ) int {
    // hard-code
    if opcode == 0x0 {
        return 7
    }
    if opcode == 0x20 || opcode == 0x40 || opcode == 0x60 {
        return 6
    }

    addrmode_name := GetOpAddrModeName( opcode )

    // code base
    cycle := _map_addrmode_cycle[addrmode_name]

    // 1. All single-byte instructions waste a cycle reading and ignoring the byte that comes immediately after the instruction 
    // (this means no instruction can take less than two cycles).
    if addrmode_name == "IMP" {
        cycle += 1
    }

    if IsNormalStackOpcode( opcode ) {
        cycle += 1
    }

    // 2. Zero page,X, zero page,Y, and (zero page,X) addressing modes 
    // spend an extra cycle reading the unindexed zero page address.
    if addrmode_name == "ZPX" ||
        addrmode_name == "ZPY" ||
        addrmode_name == "IZX" {
        cycle += 1
    }

    // 3. Absolute,X, absolute,Y, and (zero page),Y addressing modes need an extra cycle 
    // if the indexing crosses a page boundary, or if the instruction writes to memory.
    if addrmode_name == "ABX" ||
        addrmode_name == "ABY" ||
        addrmode_name == "IZY" {

        // 3.0 
        if IsMemoryWriteOpcode( opcode ) || IsReadModifyWriteOpcode(opcode) {
            cycle += 1
        }
        // 3.5 page boundary 
        // Done , implemented in clock() function

    }

    // 4. The conditional branch instructions require an extra cycle if the branch actually happens, 
    // and a second extra cycle if the branch happens and crosses a page boundary.
    // TODO will be implmented in  the instruction operations

    // 5. Read-modify-write instruction (ASL, DEC, INC, LSR, ROL, ROR) 
    // need a cycle for the modify stage (except in accumulator mode, which doesn't access memory).
    if IsReadModifyWriteOpcode( opcode ) && !IsAAddrModeOpcode( opcode ) {
        cycle += 2   // modify ?
    }

    // 6. Instructions that pull data off the stack (PLA, PLP, RTI, RTS) need an extra cycle to increment the stack pointer 
    // (because the stack pointer points to the first empty address on the stack, not the last used address).
    if IsPullStackOpcode( opcode ) {
        cycle += 1
    }

    // 7. RTS needs an extra cycle (in addition to the single-byte penalty and the pull-from-stack penalty) to increment the return address.
    if opcode == 0x60 {
        cycle += 1
    }

    // 8. JSR spends an extra cycle juggling the return address internally.
    if opcode == 0x20 {
        cycle += 1
    }

    // 9. Hardware interrupts take the same number of cycles as a BRK instruction

    // JMP abs
    if opcode == 0x4C {
        cycle -= 1
    }

    // for invalide 
    if !IsValidOpcode( opcode ) {
        bbb := (opcode>>2)&7
        cc := opcode&3
        lo := opcode & 0xF
        hi := (opcode >>4) & 0xF

        if cc < 3 {
            if bbb == 0  {
                return cycle
            }

            // bbb == 1,3,5
            dummy_cycle := -1
            for j:=0 ; j<3; j++ {
                _code := uint8((int(opcode)&^3) + j)
                if IsValidOpcode(  _code )  {
                    tmp := GetOpCycles(_code )
                    if dummy_cycle == -1 || tmp < dummy_cycle  {
                        dummy_cycle = tmp
                    }
                }
            }

            if dummy_cycle > 0 {
                cycle = dummy_cycle
            }
        } else if  cc == 3 {
            cycle_10 := GetOpCycles( opcode -1 )
            cycle_01 :=  GetOpCycles( opcode -2 )
            cycle = cycle_10
            // fmt.Printf( "\t\t %02X , %d,%d\n", opcode, cycle_01, cycle_10  )

            // bbb == 1,3,5,7
            if bbb & 1 == 1 {
                if cycle_01 > cycle {
                    cycle = cycle_01
                }
            } else {
                // 0,4,6
                if lo == 0xB && (hi&1==0) {
                    // 0B,2B,4B,6B, 8B,AB,CB,EB
                    cycle = 2
                } else if hi>=8 && hi <= 0xB && (lo == 3 || lo == 0xB)  {
                    // 83,93,a3,b3, 9b,bb , max
                    if cycle_01 > cycle {
                        cycle = cycle_01
                    }
                } else {
                    // other  1+2
                    cycle = cycle_01 + cycle_10 + 1
                    if cycle > 8 {
                        cycle = 8
                    }
                }
            } // end 0,4,6
        }
    }

    // defensive check
    if cycle < 2 {
        log.Fatal( "instruction cycles should >= 2" )
        cycle = 2
    }

    return cycle  // no instruction can take less than two cycles
}


