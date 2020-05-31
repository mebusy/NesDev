package main

import (
    // "math/rand"
    "github.com/nsf/termbox-go"
    "fmt"
)


var cur_x int
var cur_y int

var PC int

const BgColor = 0x12

func drawString( str string ) {
    for _,s := range str {
        termbox.SetCell(cur_x, cur_y, s , termbox.ColorWhite  , BgColor )
        cur_x++
    }
}
func drawStringFg( str string, fgColor termbox.Attribute  ) {
    for _,s := range str {
        termbox.SetCell(cur_x, cur_y, s , fgColor, BgColor )
        cur_x++
    }
}

func draw1PageMemory( start_addr int ) {
    for j :=0 ; j< 16 ; j++ {
        for i :=0 ; i< 16 ; i++ {

            addr := uint16(j*16 + i + start_addr )
            if  i == 0 {
                cur_x = 0
                drawString( fmt.Sprintf( "$%04X: " , addr ) )
            }

            val := nes_bus.CpuRead( addr, false  )
            if addr == uint16(PC) {
                drawStringFg( fmt.Sprintf( "%02X " , val ), termbox.ColorCyan | termbox.AttrBold )
            } else {
                drawString( fmt.Sprintf( "%02X " , val ) )
            }
        }
        cur_y++
    }
}

var cpu_flags = []string{ "N","V","-","B","D","I","Z","C" }
func draw() {
    cpu := nes_bus.DebugDumpCpu()
    PC = int(cpu.PC)

    // w, h := termbox.Size()
    // _,_ = w,h
    // termbox.Init()
    termbox.SetCursor(0,0)
    termbox.Clear(termbox.ColorWhite, BgColor )

    win_top_y := 4

    cur_x = 0
    cur_y = win_top_y

    // ============= draw 0 page mem ======================
    draw1PageMemory( 0 )

    cur_x = 0
    cur_y++

    // ============= draw program data ===================
    start_addr := (PC&^0x3F) - 6* 16
    if start_addr < 0 {
        start_addr = 0
    }
    draw1PageMemory( start_addr )

    // draw PC area
    win_x := 56

    cur_x  = win_x
    cur_y = win_top_y

    drawString( "STATUS: " )
    for i,v := range cpu_flags {
        if cpu.Status & ( 1<<(7-i) ) != 0 {
            drawStringFg( fmt.Sprintf( "%s ",v ), termbox.ColorGreen  )
        } else {
            drawStringFg( fmt.Sprintf( "%s ",v ), termbox.ColorRed  )
        }
    }
    cur_x  = win_x; cur_y++
    drawString( fmt.Sprintf("PC: $%04X", PC) )
    cur_x  = win_x; cur_y++
    drawString( fmt.Sprintf("A: $%02X [%2d]", cpu.A, cpu.A) )
    cur_x  = win_x; cur_y++
    drawString( fmt.Sprintf("X: $%02X [%2d]", cpu.X, cpu.X) )
    cur_x  = win_x; cur_y++
    drawString( fmt.Sprintf("Y: $%02X [%2d]", cpu.Y, cpu.Y) )
    cur_x  = win_x; cur_y++
    drawString( fmt.Sprintf("SP: $%02X [%2d]", cpu.SP, cpu.SP) )

    // draw hint
    cur_x  = 4; cur_y = 34 + win_top_y
    drawString( "Space:run/pause R:reset E:dumpPattern " )

    termbox.Flush()
}
