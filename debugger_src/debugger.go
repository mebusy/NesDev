package main

import (
    // "log"
    "os"
    "time"
    // "github.com/mattn/go-runewidth"
    "github.com/nsf/termbox-go"
    "nes"
    "runtime"
    "nes/cart"

    // "sync/atomic"
)


var nes_bus *nes.Bus
const PC_programStart = 0x8000


var bEmulationRun bool  // if true, run emulation in real time
var fResidualTime float64

var chan_keycode = make(chan int32)
const FAKE_SPACE_KEY int32 = 901

func main() {
    // start nes
    bus := nes.NewBus()

    /*
    code := `
A2 0A 8E 00 00 A2 03 8E
01 00 AC 00 00 A9 00 18
6D 01 00 88 D0 FA 8D 02
00 EA EA EA
`
    bus.DebugLoadCode( code, uint16(PC_programStart) )
    /*/
    cartridge := cart.NewCartridge( "../nestest.nes" )
    bus.InsertCartridge( cartridge )
    //*/


    bus.Reset()
    nes_bus = bus

    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()

    termbox.SetOutputMode( termbox.Output256 )

    go func() {
        for {
            ev := termbox.PollEvent()
            if ev.Type == termbox.EventKey {
                if ev.Key == termbox.KeyCtrlC {
                    termbox.Close()
                    os.Exit(1)
                } else if  ev.Key == termbox.KeySpace {
                    chan_keycode <- FAKE_SPACE_KEY
                } else {
                    chan_keycode <- ev.Ch
                }
            }
        }
    }()


    // main loop
    draw()

    time_now := float64(time.Now().UnixNano()) / float64( time.Second )
    for {
        t := float64(time.Now().UnixNano()) / float64( time.Second )
        OnUserUpdate( t - time_now )

        runtime.Gosched()
        time_now = t
    }
}
