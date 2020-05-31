package main

import (
    "log"
    "runtime/pprof"
    "os"
    "github.com/hajimehoshi/ebiten"
    // "time"
    // "runtime"
)



const WINDOW_SCALE = 1

func main() {
    // runtime.GOMAXPROCS(1)

    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal("could not create CPU profile: ", err)
    }
    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal("could not start CPU profile: ", err)
    }
    defer pprof.StopCPUProfile()



    game := NewGame()


    // Sepcify the window size as you like. Here, a doulbed size is specified.
    ebiten.SetWindowSize(game.Width, game.Height)
    ebiten.SetWindowTitle("yet another NES-go")
    // Call ebiten.RunGame to start your game loop.
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}


