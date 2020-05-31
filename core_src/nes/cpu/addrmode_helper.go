package cpu

import (
)


func _name2AddressingModeFunc( name string ) FUNC_AddressingMode {
    switch name {
    case "IMP": return IMP
    case "IMM": return IMM
    case "ZP0": return ZP0
    case "ZPX": return ZPX
    case "ZPY": return ZPY
    case "REL": return REL
    case "ABS": return ABS
    case "ABX": return ABX
    case "ABY": return ABY
    case "IND": return IND
    case "IZX": return IZX
    case "IZY": return IZY
    default:
        panic( "unknown addressing mode" )
    }
}



var _map_addrmode_cycle map[string] int = map[string] int{
    "IMP": 1, // rule 1
    "IMM": 2,
    "ZP0": 2+1, // 
    "ZPX": 2+1, // rule 2
    "ZPY": 2+1, // rule 2
    "REL": 2,   // 2**
    "ABS": 3+1, // mem read
    "ABX": 3+1,
    "ABY": 3+1,
    "IND": 3+2,  // JMP(6C) only, 
    "IZX": 2+1+1+1,  // rule 2
    "IZY": 2+1+1+1,
}


var _tbl_bbbcc_addrmode_name [4][8]string
func init() {
    cc := 1
    _tbl_bbbcc_addrmode_name[cc] = [8]string{
        "IZX", "ZP0", "IMM", "ABS",  "IZY", "ZPX", "ABY", "ABX",
    }
    cc = 2
    // here IMP is actually the ACC 
    _tbl_bbbcc_addrmode_name[cc] = [8]string {
        "IMM", "ZP0", "IMP", "ABS",  "nil", "ZPX", "nil", "ABX",
    }
    cc = 0
    _tbl_bbbcc_addrmode_name[cc] = [8]string {
        "IMM", "ZP0", "nil", "ABS",  "nil", "ZPX", "nil", "ABX",
    }

    // No instructions have the form aaabbb11.
}

