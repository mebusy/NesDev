
# Yet Another NES Emulator in Go

## Structure

```bash
.
├── core_src
│   └── nes
│       ├── apu
│       ├── cart
│       ├── color
│       ├── cpu
│       ├── mapper
│       ├── ppu
│       ├── sprite
│       └── tools
├── debugger_src  (deprecated)
├── emulator_src  (deprecated)
├── yanes_src
```

- **core_src** 
    - the core NES codes
- **yanes_src** 
    - NES emulator
    - UI: github.com/go-gl/glfw
    - Audio: github.com/gordonklaus/portaudio 


## How to run 

```bash
go run *.go -p <path to your nes-rom>
```

```bash
  -h    print help
  -p string
        nes room path
  -s int
        scale (default 2)
```

## TODO List

1. audio
    - triangle wave channel does not be implemented yet
    - DMA channel does not be implemented yet
2. mapper
    - mapper 2,3,66 have not been fully tested
    - add more mappers
3. improve performance in order to add more debugging features without decreasing fps


## Misc

[NES development notes](nes_notes.md)

