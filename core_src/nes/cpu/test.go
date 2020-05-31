package cpu

import (
    "fmt"
    "log"
    "bytes"
    "strings"
)


// expect_total will be re-calc if it is == -1
func checkCategory( func_category  func ( uint8 ) bool , func_name string, expect_total , expect_kind int, bPrint bool ) {
    cnt := 0
    m := map[string]string {}
    var buffer bytes.Buffer
    for i:=0; i< 256; i++ {
        if func_category(uint8(i)) {
            cnt++
            m[ GetOpName( uint8(i) ) ] = ""
            if bPrint {
                buffer.WriteString( fmt.Sprintf( "%s(%02X), " , GetOpName( uint8(i) ) ,i ) )
            }
        }
    }

    if expect_total == -1 {
        expect_total = 0
        for k,_ := range m {
            expect_total += getOpcodeNumberByName( k )
        }
    }

    prefix := fmt.Sprintf( "[%s]: ", func_name )

    if bPrint {
        if cnt < 10 {
            fmt.Printf( "%s%v, %s\n", prefix, m, buffer.String()  )
        } else {
            fmt.Printf( "%s%v\n", prefix, m )
        }
    }
    if bPrint {
        prefix = strings.Repeat( " ", len(prefix) )
    }
    fmt.Printf( "%s%d opcode in %d kinds\n", prefix, cnt, len(m) )

    if cnt != expect_total || len(m) != expect_kind {
        log.Fatalf( "[fatal] %s error, total:%d vs %d, kind:%d vs %d" , func_name, cnt, expect_total, len(m), expect_kind )
    }
}


func TestOpcodeCategory() {
    checkCategory(  IsValidOpcode , "IsValidOpcode", -1, 56, false )
    checkCategory(  IsAAddrModeOpcode , "IsAAddrModeOpcode", 4, 4, true  )
    checkCategory(  IsReadModifyWriteOpcode , "IsReadModifyWriteOpcode", -1, 6, true )
    checkCategory(  IsPullStackOpcode , "IsPullStackOpcode", -1, 4, true )
    checkCategory(  IsNormalStackOpcode , "IsNormalStackOpcode", -1, 4, true )
    checkCategory(  IsMemoryWriteOpcode , "IsMemoryWriteOpcode", -1, 5, true )
    checkCategory(  IsExtraCycleValidOpcode , "IsExtraCycleValidOpcode", -1, 9, true )
    checkCategory(  IsExtraCycleInvalidOpcode , "IsExtraCycleInvalidOpcode", 6 , 1, true )

}


func TestAddrMode() {
    for i :=0; i< 256; i++ {
        // aaa := (i>>5)&7
        bbb := (i>>2)&7
        cc := i&3

        addrmode_name := GetOpAddrModeName( uint8(i) )
        addrmode_name_cmp := _cmp_opcode_lookup[i].cmp_addrMode_name

        // ensure the addressing mode of instruction is correct
        if cc == cc  {
            if addrmode_name != addrmode_name_cmp {
                log.Fatalf( "[Fatal] %02X,%s, %s", i, addrmode_name  ,  addrmode_name_cmp ) 
            } else if !IsValidOpcode( uint8(i) ) &&  GetOpName( uint8(i) ) != "???" {
                log.Printf( "%02X has the wrong name: %s ", i,GetOpName( uint8(i) ) )
            }
        }

        // ensure the cycle of instruction is correct
        if cc<4 || IsValidOpcode( uint8(i) ) {
            /*
            if addrmode == map_addrmode_pointer[ "IMP" ] || addrmode == map_addrmode_pointer[ "IMM" ] ||
                addrmode == map_addrmode_pointer[ "ZP0" ] || addrmode == map_addrmode_pointer[ "ZPX" ] || addrmode == map_addrmode_pointer[ "ZPY" ] ||
                addrmode == map_addrmode_pointer[ "IND" ] || addrmode == map_addrmode_pointer[ "REL" ] ||
                addrmode == map_addrmode_pointer[ "ABS" ] || addrmode == map_addrmode_pointer[ "ABX" ] || addrmode == map_addrmode_pointer[ "ABY" ] ||
                addrmode == map_addrmode_pointer[ "IZX" ] || addrmode == map_addrmode_pointer[ "IZY" ]   {
                cycle := GetOpCycles( uint8(i) )
                cycle_expected := _cmp_opcode_lookup[ i ].cycles
                if cycle != cycle_expected {
                    log.Printf( "%s(%02x) last %d cycles, but it has only %d\n", GetOpName( uint8(i) ) , i, cycle_expected, cycle )
                }
            }
            /*/
            cycle := GetOpCycles( uint8(i) )
            cycle_expected := _cmp_opcode_lookup[ i ].cmp_cycles
            if cycle != cycle_expected {
                fmt.Printf( "%s(%02x) last %d cycles, but it has only %d, bbb:%d \n", GetOpName( uint8(i) ) , i, cycle_expected, cycle, bbb )
                // fmt.Printf("\n")
            }
            //*/
        }
    } // end for 

    fmt.Println()

    // adc, sbc test
    /*
    var i uint8 = 0
    var j uint8 = 0
    for j = 0 ; j< 255 ; j++ {
        for i = 0 ; i< 255 ; i++ {
            if ^(i^j) != (i^(^j)) {
                log.Fatalf( "[fatal] %d,%d ",i,j  )
            }
        }
    }

    A := []int { 0,0,0,0,1,1,1,1 }
    M := []int { 0,0,1,1,0,0,1,1 }
    R := []int { 0,1,0,1,0,1,0,1 }

    fmt.Printf( "A M R V A^R A^M ~(A^M) M^R  (A^R)&(M^R)  V \n" )
    for i:=0 ; i< len(A); i++ {
        a_r := A[i]^R[i]
        a_m := A[i]^M[i]
        v := a_r &^ a_m
        m_r := M[i]^R[i]

        fmt.Printf( "%d %d %d %d  %d   %d    %d     %d       %d        %d \n", A[i], M[i], R[i], v, a_r, a_m, (^a_m)&1, m_r, a_r&m_r, v )
    }
    //*/

    // var t uint8 = 0
    // log.Println(  1 | 1 << 8 )
}


func getOpcodeNumberByName( name string ) int {
    cnt := 0
    for i:=0; i< 256; i++ {
        if _cmp_opcode_lookup[i].name == name &&  IsValidOpcode( uint8(i) )  {
            cnt ++
        }
    }
    return cnt
}
