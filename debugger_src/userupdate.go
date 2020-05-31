package main

import (
    "time"
    "log"
)

func OnUserUpdate( fElapsedTime float64 ) {
    // log.Println( fElapsedTime )
    var keycode int32 = 0

    select {
    case keycode = <-chan_keycode:
        log.Println( keycode )
    case <- time.After( time.Nanosecond ):
    }
    if keycode == FAKE_SPACE_KEY {
        bEmulationRun = !bEmulationRun
        log.Println( "Emulation Run:", bEmulationRun )
    }

    if bEmulationRun {
        // NTSC:  60Hz frame rate
        if fResidualTime > 0 {
            // most of the time, emulator do nothing
            fResidualTime -= fElapsedTime
        } else {
            fResidualTime += (1.0 / 60.0) - fElapsedTime;

            nes_bus.DebugSingleFrame(true)
            draw()
        }
    } else {
        switch keycode {
        case 115 : // S
            nes_bus.DebugStepInstruction()
        case 114: //R
            nes_bus.Reset()
        case 101: //E
            generatePatternImage( nes_bus )
            generatePaletteImage( nes_bus )
            /*
            nes_bus.DebugStepInstruction()
            PC = int( nes_bus.DebugDumpCpu().PC )

            nes_bus.DebugStepInstruction()
            PC = int( nes_bus.DebugDumpCpu().PC )
            cnt := 0
            for {
                if PC == PC_programStart  {
                    log.Println( "program end" )
                    break
                }
                nes_bus.DebugPressStart()

                nes_bus.DebugStepInstruction()
                PC = int( nes_bus.DebugDumpCpu().PC )

                cnt = (cnt+1)%1000
                if cnt == 0 {
                }
                runtime.Gosched()
                draw()
            }
            //*/
        case 102: // F
            nes_bus.DebugSingleFrame(false)
        }

        draw()
    } // end if !bEmulatorRun
}
