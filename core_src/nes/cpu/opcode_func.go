package cpu



// those function 0 or 1 to indicate whether 
// there need to be another clock cycle
type FUNC_Operate func(*Cpu) 

func GetOpOperate( opcode uint8 ) FUNC_Operate {
    return _cmp_opcode_lookup[opcode].operate
}

func GetOpName(opcode uint8) string {
    return _cmp_opcode_lookup[opcode].name
}

/*
Bitwise:   AND EOR ORA
           ASL LSR ROL ROR

Branch:    BCC BCS BEQ BMI
           BNE BPL BVC BVS

Compare:   BIT CMP CPX CPY

Flags:     CLC CLD CLI CLV
           SEC SED SEI

Jump:      JMP JSR RTI RTS

Math:      ADC SBC

Memory:    LDA LDX LDY
           STA STX STY DEC INC

Registers: TAX TAY TXA TYA
           DEX DEY INX INY

Stack:     PHA PHP PLA PLP TSX TXS

Other:     BRK NOP

Instructions that will not affect flags:
    Branch Instruction, and
    JMP, RTS, JSR,
    STA, STX, STY,
    NOP


/*
---- those 18 instruction need fetch data
BIT CPX CPY
--- RMW ----
DEC INC ASL LSR ROL ROR 
--- Extra Cycle ---
LDA LDX LDY CMP ADC SBC AND EOR ORA 
*/


// ADC and SBC are problematic
// A  M  R | V | A^R | A^M |~(A^M) |
// 0  0  0 | 0 |  0  |  0  |   1   |
// 0  0  1 | 1 |  1  |  0  |   1   |
// 0  1  0 | 0 |  0  |  1  |   0   |
// 0  1  1 | 0 |  1  |  1  |   0   |  so V = ~(A^M) & (A^R)
// 1  0  0 | 0 |  1  |  1  |   0   |
// 1  0  1 | 0 |  0  |  1  |   0   |
// 1  1  0 | 1 |  1  |  0  |   1   |
// 1  1  1 | 0 |  0  |  0  |   1   |
//
// Instruction: Add with Carry In
// Function:    A = A + M + C
// Flags Out:   C, V, N, Z
func ADC(self *Cpu) {
    // A += M + C 
    // with the carry bit, we can chain together additions of 
    // 8-bit words into large bit words. 
    // when 1 overflows, it sets the carry bit as an input to the next addition

    self.fetch()
    // Add is performed in 16-bit domain for emulation to capture any
    // carry bit, which will exist in bit 8 of the 16-bit word
    temp := uint16(self.A) + uint16(self.fetched) + self.getFlag( FLAG_CPU_C )
    self.SetFlagsNZC( temp , FLAG_CPU_C | FLAG_CPU_Z | FLAG_CPU_N )

    self.setFlag( FLAG_CPU_V ,  0 != ( (uint16(self.A)^temp) &^ uint16( self.A^self.fetched ) )&0x80  )

    // Load the result into the accumulator 
    self.A = uint8(temp & 0xFF)

}

// Instruction: Bitwise Logic AND
// Function:    A = A & M
// Flags Out:   N, Z
func AND(self *Cpu) {
    self.fetch()
    self.A = self.A & self.fetched

    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_Z | FLAG_CPU_N  )

}

// Instruction: Arithmetic Shift Left
// Function:    A = C <- (A << 1) <- 0
// Flags Out:   N, Z, C
func ASL(self *Cpu)  {
    self.performBitShift( true, false  )
}

// Instruction: Branch if Carry Clear
// Function:    if(C == 0) pc = address
func BCC(self *Cpu) {
    if self.getFlag( FLAG_CPU_C ) == 0 {
        self.performBranch()
    }
}



// Instruction: Branch if Carry Set
// Function:    if(C == 1) pc = address
func BCS(self *Cpu) {
    if self.getFlag( FLAG_CPU_C ) == 1 {
        self.performBranch()
    }
}

// Instruction: Branch if Equal
// Function:    if(Z == 1) pc = address
func BEQ(self *Cpu) {
    if self.getFlag( FLAG_CPU_Z ) == 1 {
        self.performBranch()
    }
}

// Test Bits in Memory with Accumulator
func BIT(self *Cpu) {
    self.fetch()
    temp := self.A & self.fetched
    // bits 7 and 6 of operand are transfered to bit 7 and 6 of status register (N,V);
    self.SetFlagsNZC( uint16(self.fetched) , FLAG_CPU_N )
    self.setFlag( FLAG_CPU_V, (self.fetched & (1 << 6)) != 0 )

    // the zeroflag is set to the result of operand AND accumulator.
    self.SetFlagsNZC( uint16(temp), FLAG_CPU_Z  )
}

// Instruction: Branch if Negative
// Function:    if(N == 1) pc = address
func BMI(self *Cpu) {
    if self.getFlag( FLAG_CPU_N ) == 1 {
        self.performBranch()
    }
}

// Instruction: Branch if Not Equal
// Function:    if(Z == 0) pc = address
func BNE(self *Cpu) {
    if self.getFlag( FLAG_CPU_Z ) == 0 {
        self.performBranch()
    }
}

// Instruction: Branch if Positive
// Function:    if(N == 0) pc = address
func BPL(self *Cpu) {
    if self.getFlag( FLAG_CPU_N ) == 0 {
        self.performBranch()
    }
}

// Instruction: Break
// Function:    Program Sourced Interrupt
func BRK(self *Cpu) {
    // BRK causes a non-maskable interrupt and increments the program counter by one. ?? (it seems that it actually effectively do irq)
    // Therefore an RTI will go to the address of the BRK +2 so that BRK may be used to replace a two-byte instruction for debugging and the subsequent RTI will be correct.
    self.PC++

    self.setFlag( FLAG_CPU_I, true )
    self.push2stackWord( self.PC )

    self.setFlag( FLAG_CPU_B , true )
    self.push2stackByte( uint8(self.Status) )
    self.setFlag( FLAG_CPU_B , false )

    self.PC = self.readWord( 0xFFFE )

}

// Instruction: Branch if Overflow Clear
// Function:    if(V == 0) pc = address
func BVC(self *Cpu) {
    if self.getFlag( FLAG_CPU_V ) == 0 {
        self.performBranch()
    }
}

// Instruction: Branch if Overflow Set
// Function:    if(V == 1) pc = address
func BVS(self *Cpu) {
    if self.getFlag( FLAG_CPU_V ) == 1 {
        self.performBranch()
    }
}

// Instruction: Clear Carry Flag
// Function:    C = 0
func CLC(self *Cpu) {
    self.setFlag( FLAG_CPU_C , false )
}

// Instruction: Clear Decimal Flag
// Function:    D = 0
func CLD(self *Cpu) {
    self.setFlag( FLAG_CPU_D , false )
}

// Instruction: Disable Interrupts / Clear Interrupt Flag
// Function:    I = 0
func CLI(self *Cpu) {
    self.setFlag( FLAG_CPU_I , false )
}

// Instruction: Clear Overflow Flag
// Function:    V = 0
func CLV(self *Cpu) {
    self.setFlag( FLAG_CPU_V , false )
}

// Instruction: Compare Accumulator
// Function:    C <- A >= M      Z <- (A - M) == 0
// Flags Out:   N, C, Z
func CMP(self *Cpu) {
    self.fetch()
    self.compare( self.A, self.fetched )
}

// Instruction: Compare X Register
// Function:    C <- X >= M      Z <- (X - M) == 0
// Flags Out:   N, C, Z
func CPX(self *Cpu) {
    self.fetch()
    self.compare( self.X, self.fetched )
}

// Instruction: Compare Y Register
// Function:    C <- Y >= M      Z <- (Y - M) == 0
// Flags Out:   N, C, Z
func CPY(self *Cpu) {
    self.fetch()
    self.compare( self.Y, self.fetched )
}

// Instruction: Decrement Value at Memory Location
// Function:    M = M - 1
// Flags Out:   N, Z
func DEC(self *Cpu) {
    self.fetch()
    temp := self.fetched -1
    self.write( self.addr_abs, temp )
    self.SetFlagsNZC( uint16(temp), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Decrement X Register
// Function:    X = X - 1
// Flags Out:   N, Z
func DEX(self *Cpu) {
    self.X--
    self.SetFlagsNZC( uint16(self.X), FLAG_CPU_N | FLAG_CPU_Z )
}

func DEY(self *Cpu) {
    self.Y--
    self.SetFlagsNZC( uint16(self.Y), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Bitwise Logic XOR
// Function:    A = A xor M
// Flags Out:   N, Z
func EOR(self *Cpu) {
    self.fetch()
    self.A = self.A ^ self.fetched

    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_N | FLAG_CPU_Z  )

}

// Instruction: Increment Value at Memory Location
// Function:    M = M + 1
// Flags Out:   N, Z
func INC(self *Cpu) {
    self.fetch()
    // INC does not care about carry ?
    temp := self.fetched + 1
    self.write( self.addr_abs, temp )
    self.SetFlagsNZC( uint16(temp), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Increment X Register
// Function:    X = X + 1
// Flags Out:   N, Z
func INX(self *Cpu) {
    self.X++
    self.SetFlagsNZC( uint16(self.X), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Increment Y Register
// Function:    Y = Y + 1
// Flags Out:   N, Z
func INY(self *Cpu) {
    self.Y++
    self.SetFlagsNZC( uint16(self.Y), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Jump To Location
// Function:    pc = address
func JMP(self *Cpu) {
    self.PC = self.addr_abs
}

// Instruction: Jump To Sub-Routine
// Function:    Push current pc to stack, pc = address
func JSR(self *Cpu) {
    // JSR pushes the address-1 of the next operation on to the stack
    //  before transferring program control to the following address.
    //  Subroutines are normally terminated by an RTS op code.
    self.PC--
    self.push2stackWord(self.PC)
    self.PC = self.addr_abs
}

// Instruction: Load The Accumulator
// Function:    A = M
// Flags Out:   N, Z
func LDA(self *Cpu) {
    self.fetch()
    self.A = self.fetched;
    self.SetFlagsNZC( uint16(self.A), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Load The X Register
// Function:    X = M
// Flags Out:   N, Z
func LDX(self *Cpu) {
    self.fetch()
    self.X = self.fetched;
    self.SetFlagsNZC( uint16(self.X), FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Load The Y Register
// Function:    Y = M
// Flags Out:   N, Z
func LDY(self *Cpu) {
    self.fetch()
    self.Y = self.fetched;
    self.SetFlagsNZC( uint16(self.Y), FLAG_CPU_N | FLAG_CPU_Z )
}

/*
LSR  Logic Shift One Bit Right (Memory or Accumulator)

     0 -> [76543210] -> C             N Z C I D V
                                      0 + + - - -
*/
func LSR(self *Cpu) {
    self.performBitShift( false, false )
}

func NOP(self *Cpu) {
    // Sadly not all NOPs are equal, Ive added a few here
    // based on https://wiki.nesdev.com/w/index.php/CPU_unofficial_opcodes
    // and will add more based on game compatibility, and ultimately
    // I'd like to cover all illegal opcodes too
}

// Instruction: Bitwise Logic OR
// Function:    A = A | M
// Flags Out:   N, Z
func ORA(self *Cpu) {
    self.fetch()
    self.A = self.A | self.fetched

    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_N | FLAG_CPU_Z  )
}


// Instruction: Push Accumulator to Stack
// Function:    A -> stack
func PHA(self *Cpu) {
    // the stack is somewhere in the memory
    // 6502 has hard-coded into it a base lococation for the Stack Pointer 0x100
    self.push2stackByte( self.A )

}

// Instruction: Push Status Register to Stack
// Function:    status -> stack
// Note:        Break flag is set to 1 before push
func PHP(self *Cpu) {
    self.push2stackByte( uint8( self.Status | FLAG_CPU_B | FLAG_CPU_U )  )

    self.setFlag( FLAG_CPU_B , false  )
    self.setFlag( FLAG_CPU_U , false  )
}

// Instruction: Pop Accumulator off Stack
// Function:    A <- stack
// Flags Out:   N, Z
func PLA(self *Cpu) {
    self.A = self.pullFromStackByte()
    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_N | FLAG_CPU_Z )
}

func PLP(self *Cpu) {
    self.Status = FLAG_CPU(self.pullFromStackByte())
    self.setFlag( FLAG_CPU_U , true  )
}

// ROL  Rotate One Bit Left (Memory or Accumulator)
/*
     C <- [76543210] <- C             N Z C I D V
                                      + + + - - -
*/
func ROL(self *Cpu) {
    self.performBitShift( true, true  )
}

/*
ROR  Rotate One Bit Right (Memory or Accumulator)

     C -> [76543210] -> C             N Z C I D V
                                      + + + - - -
*/
func ROR(self *Cpu) {
    self.performBitShift( false, true )
}

// return from interrupt : irq and nmi
func RTI(self *Cpu) {
    self.Status = FLAG_CPU(self.pullFromStackByte())

    // self.status &^= FLAG_CPU_B
    // self.status &^= FLAG_CPU_U
    self.setFlag( FLAG_CPU_B , false  )
    self.setFlag( FLAG_CPU_U , false  )
    // Q? how about FLAG_CPU_I ?

    self.PC = self.pullFromStackWord()

    // Note that unlike RTS, 
    // the return address on the stack is the actual address
}

// Return from Subroutine
func RTS(self *Cpu) {
    // RTS pulls the top two bytes off the stack (low byte first) 
    //  and transfers program control to that address+1. 
    // It is used, as expected, to exit a subroutine invoked via 
    //  JSR which pushed the address-1.
    self.PC = self.pullFromStackWord()
    self.PC++
}

// Instruction: Subtraction with Borrow In
// Function:    A = A - M - (1-C)
//          =>  A = A + -M -1 + C
//          =>  A = A + ~M+1 -1 + C  = A + ~M + C
// Flags Out:   C, V, N, Z
func SBC(self *Cpu) {

    self.fetch()
    // ~M
    inv_value := ^self.fetched

    temp := uint16(self.A) + uint16(inv_value) + self.getFlag( FLAG_CPU_C )
    self.SetFlagsNZC( temp , FLAG_CPU_C | FLAG_CPU_Z | FLAG_CPU_N )

    // 
    self.setFlag( FLAG_CPU_V ,  0 != ( (uint16(self.A)^temp) &^ uint16( self.A^inv_value ) )&0x80  )

    // overflow explained: 
    //  http://www.righto.com/2012/12/the-6502-overflow-flag-explained.html
    //  http://www.6502.org/tutorials/vflag.html


    // Load the result into the accumulator 
    self.A = uint8(temp & 0xFF)

}

// Instruction: Set Carry Flag
// Function:    C = 1
func SEC(self *Cpu) {
    self.setFlag( FLAG_CPU_C , true )
}

// Instruction: Set Decimal Flag
// Function:    D = 1
func SED(self *Cpu) {
    self.setFlag( FLAG_CPU_D , true )
}

// Instruction: Set Interrupt Flag / Enable Interrupts
// Function:    I = 1
func SEI(self *Cpu) {
    self.setFlag( FLAG_CPU_I , true )
}

// Instruction: Store Accumulator at Address
// Function:    M = A
func STA(self *Cpu) {
    self.write( self.addr_abs , self.A )
}

// Instruction: Store X Register at Address
// Function:    M = X
func STX(self *Cpu) {
    self.write( self.addr_abs , self.X )
}

// Instruction: Store Y Register at Address
// Function:    M = Y
func STY(self *Cpu) {
    self.write( self.addr_abs , self.Y )
}

// Instruction: Transfer Accumulator to X Register
// Function:    X = A
// Flags Out:   N, Z
func TAX(self *Cpu) {
    self.X = self.A
    self.SetFlagsNZC( uint16(self.X) , FLAG_CPU_N | FLAG_CPU_Z )
}

func TAY(self *Cpu) {
    self.Y = self.A
    self.SetFlagsNZC( uint16(self.Y) , FLAG_CPU_N | FLAG_CPU_Z )
}


// Instruction: Transfer X Register to Accumulator
// Function:    A = X
// Flags Out:   N, Z
func TXA(self *Cpu) {
    self.A = self.X
    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Transfer Stack Pointer to X Register
// Function:    X = stack pointer
// Flags Out:   N, Z
func TSX(self *Cpu) {
    self.X = self.SP
    self.SetFlagsNZC( uint16(self.X) , FLAG_CPU_N | FLAG_CPU_Z )
}

// Instruction: Transfer X Register to Stack Pointer
// Function:    stack pointer = X
func TXS(self *Cpu) {
    self.SP = self.X
}

// Instruction: Transfer Y Register to Accumulator
// Function:    A = Y
// Flags Out:   N, Z
func TYA(self *Cpu) {
    self.A = self.Y
    self.SetFlagsNZC( uint16(self.A) , FLAG_CPU_N | FLAG_CPU_Z )
}

func XXX(self *Cpu) {
}

// helper 
func (self *Cpu) performBranch() {
    // the branch actually happen
    self.cycles ++
    self.addr_abs = self.PC + self.addr_rel

    // check cross boundary
    if (self.addr_abs & 0xFF00) != (self.PC & 0xFF00) {
        self.cycles ++
    }

    self.PC = self.addr_abs
}

func (self *Cpu) push2stackByte( data uint8  ) {
    self.write( SP_BASE + uint16(self.SP) , data )
    self.SP--
}

func (self *Cpu) pullFromStackByte( ) uint8 {
    self.SP++  // point to top of stack
    return self.read( SP_BASE + uint16(self.SP) )
}

func (self *Cpu) push2stackWord( data uint16  ) {
    // push hi
    self.push2stackByte( uint8((data>>8)&0xFF) )
    // push lo
    self.push2stackByte( uint8((data  ) &0xFF) )
}

func (self *Cpu) pullFromStackWord( ) uint16 {
    // pull lo
    lo := uint16(self.pullFromStackByte())
    // pull hi
    hi := uint16(self.pullFromStackByte())

    return (hi<<8) | lo
}

// A-B, set N,Z,C
func (self *Cpu) compare(a, b uint8 ) {
    temp := uint16(a) - uint16(b)
    self.SetFlagsNZC( temp, FLAG_CPU_N | FLAG_CPU_Z )
    self.setFlag( FLAG_CPU_C , a >= b )
}

/*
ASL        区别
    ROL    左移 |C进位 , 左移，C进位都按照最终计算结果 设置
    LSR    右移 ， 右移，最低位都移 C进位
    ROR    右移 |C进位<<7  ,
*/
func (self *Cpu) performBitShift( bLeft, bCarry bool  ) {

    self.fetch()  // * 4 Instruction 
    var temp uint16
    if bLeft {
        temp = uint16(self.fetched) << 1    // shift first
        if bCarry {
            temp |= self.getFlag(FLAG_CPU_C)   // + Carry bit
        }

        self.SetFlagsNZC( temp, FLAG_CPU_C )  // then set Carry
    } else {
        temp = uint16(self.fetched) >> 1  //  shift
        if bCarry {
            temp |= self.getFlag(FLAG_CPU_C)<<7   // + Carry bit
        }

        self.setFlag( FLAG_CPU_C , self.fetched & 0x1 != 0  )  // use last bit to set CarryBit
    }

    self.SetFlagsNZC( temp, FLAG_CPU_Z | FLAG_CPU_N )

    if GetOpAddrModeName( self.opcode ) == "IMP" {
        // write to A
        self.A = uint8(temp & 0xFF)
    } else {
        // write to destination
        self.write(self.addr_abs , uint8(temp & 0xFF) )
    }

}
