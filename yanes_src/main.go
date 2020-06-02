package main

import (
    "log"
    "os"
    "runtime/pprof"
    "github.com/mebusy/simpleui"
    "nes"
    "nes/cart"
    "bytes"
    "io/ioutil"
    "fmt"
    "path"
    "flag"
)



var workingDir string

func init() {
    workingDir = path.Join( simpleui.HomeDir() , ".yanes" )
    os.MkdirAll(workingDir, os.ModePerm)
}


var flag_scale = flag.Int("s", 2, "scale")
var flag_rompath = flag.String("p", "", "nes room path")
var flag_help = flag.Bool("h", false, "print help")

func main() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal("could not create CPU profile: ", err)
    }
    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal("could not start CPU profile: ", err)
    }
    defer pprof.StopCPUProfile()

    flag.Parse()

    fmt.Println()

    if *flag_help {
        flag.PrintDefaults()
        return
    }

    if *flag_rompath == "" {
        flag.PrintDefaults()
        fmt.Println()
        log.Fatal( "you must specify an NES rom" )
        return
    }

    w,h := 256+128,300
    view := NewNesView(w,h)

    // new cartridge
    cartridge := cart.NewCartridge(*flag_rompath)

    // new NES 
    view.console = nes.NewBus()
    view.console.InsertCartridge(cartridge)

    // must call it
    simpleui.SetWindow( w,h, *flag_scale  )
    simpleui.Run( view )
}


func DumpNameTable2File( directory string , tblName [2][1024]uint8 ) {
    os.MkdirAll(directory, os.ModePerm)
    for i, nametable := range tblName {
        var buffer bytes.Buffer
        for j, v := range nametable {
            buffer.WriteString( fmt.Sprintf( "%02X", v ) )
            if ((j+1)&0x1F) == 0 {
                buffer.WriteString( "\n" )
            }
        }

        err := ioutil.WriteFile( fmt.Sprintf( "%s/nt%d.txt", directory, i ) , buffer.Bytes()  , 0644)
        if err != nil {
            log.Fatal(err)
        }
    }
    log.Println( "name table dumped to" , directory )
}

