package ppu

import (
    "log"
)

func Test() {
    if LOOPY_REG_NAMETABLE_X != 1<<10 {
        log.Fatal( "LOOPY_REG_NAMETABLE_X  value error!" )
    }
    log.Println( "ppu tested" )
}
