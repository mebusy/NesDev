package cpu

import (
    // "fmt"
    // "log"
)

func IsValidOpcode( opcode uint8 ) bool {
    lo := opcode & 0xF
    hi := (opcode >>4) & 0xF

    if lo == 1 || lo == 5 || lo == 6 || lo == 8 || lo == 0xD {
        return true
    }

    if ( lo == 0 || lo == 9 )  && hi != 8 {
        return true
    }

    if lo == 0xE && hi != 9 {
        return true
    }

    if lo == 2 && hi == 0xA {
        return true
    }

    if lo == 4 && ( hi==2 || ( hi>=8 && hi!=0xD && hi!=0xF )  ) {
        return true
    }

    if lo == 0xA && ( (hi&1==0) || hi==9 || hi==0xB ) {
        return true
    }

    if lo == 0xC && hi>0 && ( (hi&1==0) || hi==0xB ) {
        return true
    }

    return false
}


func IsAAddrModeOpcode( opcode uint8 ) bool {
    return  (opcode & 0x1F) == 0x0A && (opcode & 0x80) == 0
}

func IsReadModifyWriteOpcode( opcode uint8 ) bool {
    if !IsValidOpcode( opcode ) {
        return false
    }

    if opcode == 0xCA || opcode == 0xEA {
        return false  // exception
    }

    aaa := (opcode>>5)&7
    // bbb := (opcode>>2)&7
    cc := opcode&3

    // ASL, LSR, ROL, ROR  *5
    // DEC, INC,  *4
    if cc == 2 && (aaa < 4 || aaa > 5) {
        return true
    }

    return false
}

func IsPullStackOpcode ( opcode uint8 ) bool {
    // PLA, PLP, RTI, RTS
    return opcode == 0x40 || opcode == 0x60 || opcode == 0x28 || opcode == 0x68
}

func IsNormalStackOpcode ( opcode uint8 ) bool {
    // 4  PHA PLA PHP PLP  , do not include SP transfer instruction TSX, TXS
    lo := opcode & 0xF
    hi := (opcode >>4) & 0xF
    return hi&9 == 0 && lo == 0x8
}

func IsMemoryWriteOpcode( opcode uint8 ) bool {
    if !IsValidOpcode( opcode ) {
        return false
    }

    lo := opcode & 0xF
    hi := (opcode >>4) & 0xF

    aaa := (opcode>>5) & 0x7
    cc := opcode&0x3


    if ( lo == 6 || lo == 0xE ) && hi >= 0xC {
        return true   // DEC, INC,  *4
    }
    if cc <= 2 && aaa == 4 {
        // exception  DEY, TYA, TXA, TXS
        if (hi==8||hi==9) && (lo==8||lo==0xA) {
            return false
        }
        // BCC
        if opcode == 0x90 {
            return false
        }
        return true
    }


    return false
}

func  IsExtraCycleInvalidOpcode( opcode uint8 ) bool {
    bValid := IsValidOpcode( opcode )
    if bValid {
        return false
    }
    // normally , invalid opcode won't take extra cycle
    switch (opcode) {
    case 0x1C:
        fallthrough
    case 0x3C:
        fallthrough
    case 0x5C:
        fallthrough
    case 0x7C:
        fallthrough
    case 0xDC:
        fallthrough
    case 0xFC:
        return true
    }

    return false
}

func  IsExtraCycleValidOpcode( opcode uint8 ) bool {

    bValid := IsValidOpcode( opcode )
    if !bValid {
        return false
    }

    //NOP  0x1C 0x3C 0x5C 0x7C _ _ 0xDC 0xFC 
    aaa := (opcode>>5)&7
    bbb := (opcode>>2)&7
    cc := opcode&3

    // ADC SBC AND ORA EOR CMP,  LDA
    if (cc==1 && aaa!=4)  {
        return true
    }

    if (cc&1==0) && (aaa==5) && (bbb==0 || bbb&1==1) {
        return true
    }
    // LDX LDY


    return false
}


