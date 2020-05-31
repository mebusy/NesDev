package main

import (
    "log"
    "nes"
    "nes/cpu"
    "nes/ppu"
    "nes/cart"
)

var  code = `
A2 0A 8E 00 00 A2 03 8E
01 00 AC 00 00 A9 00 18
6D 01 00 88 D0 FA 8D 02
00 EA EA EA
`

func main() {

    cartridge := cart.NewCartridge( "../nestest.nes" )
    bus := nes.NewBus()

    // bus.DebugLoadCode( code, 0x8000 )

    bus.InsertCartridge( cartridge )

    cpu.TestAddrMode()
    cpu.TestOpcodeCategory()

    ppu.Test()

    bus.SetAudioSampleRate( 44100 )
    bus.Reset()

    bus.Debug_SetDebugPalette()
    generatePatternImage( bus )
    generatePaletteImage( bus )

    log.Println( "emulator start" )
}

