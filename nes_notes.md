# NesDev

[wiki NesDev](http://wiki.nesdev.com/)

[olc source code](https://github.com/OneLoneCoder/olcNES)

## CPU

- Specification
    - 2A03 = 6502 + audio
    - 8-bit 
    - no builtin memory
    - 16-bit bus,  can address 64k
- Memory 
    - connect to CPU via bus
    - mapped to $0000-$07FF
- APU
    - mapped to $4000-$4017
- Cartrige
    - mapped to $4020-$FFFF , to the end.


## PPU 

- Specification
    - 2C02
    - mapped to $2000-$2007
    - has its own bus,  can address 16k
- Graphics Memory (mapping from cartridge)
    - 8k, PPU $0000-$1FFF
- VRam 
    - 2k
    - PPU $2000-$27FF
- Palettes 
    - PPU $3F00-$3FFF
- Object Attribute Memory (OAM)
    - is NOT available via any bus


##  Clocks

- feed CPU & PPU
- every clock tick
    - PPU output a pixel to screen

----- 

1. [overview](nes_emulator_1.md)
2. [cup](nes_emulator_2.md)
3. [bus,cartridge,mapper...](nes_emulator_3.md)
4. [ppu BG](nes_emulator_4.md)
5. [ppu Sprite](nes_emulator_5.md)
6. [apu](nes_emulator_6.md)
7. [mapper](nes_emulator_7.md)

----

## Instructions

- https://www.masswerk.at/6502/6502_instruction_set.htm
- http://c64os.com/post/6502instructions




