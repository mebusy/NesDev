package color


type COLOR struct {
    R uint8
    G uint8
    B uint8
}

var COLORS [64]COLOR

// 64 palette Nes could use
func init() {
    COLORS[0x00] = COLOR{84, 84, 84}
    COLORS[0x01] = COLOR{0, 30, 116}
    COLORS[0x02] = COLOR{8, 16, 144}
    COLORS[0x03] = COLOR{48, 0, 136}
    COLORS[0x04] = COLOR{68, 0, 100}
    COLORS[0x05] = COLOR{92, 0, 48}
    COLORS[0x06] = COLOR{84, 4, 0}
    COLORS[0x07] = COLOR{60, 24, 0}
    COLORS[0x08] = COLOR{32, 42, 0}
    COLORS[0x09] = COLOR{8, 58, 0}
    COLORS[0x0A] = COLOR{0, 64, 0}
    COLORS[0x0B] = COLOR{0, 60, 0}
    COLORS[0x0C] = COLOR{0, 50, 60}
    COLORS[0x0D] = COLOR{0, 0, 0}
    COLORS[0x0E] = COLOR{0, 0, 0}
    COLORS[0x0F] = COLOR{0, 0, 0}

    COLORS[0x10] = COLOR{152, 150, 152}
    COLORS[0x11] = COLOR{8, 76, 196}
    COLORS[0x12] = COLOR{48, 50, 236}
    COLORS[0x13] = COLOR{92, 30, 228}
    COLORS[0x14] = COLOR{136, 20, 176}
    COLORS[0x15] = COLOR{160, 20, 100}
    COLORS[0x16] = COLOR{152, 34, 32}
    COLORS[0x17] = COLOR{120, 60, 0}
    COLORS[0x18] = COLOR{84, 90, 0}
    COLORS[0x19] = COLOR{40, 114, 0}
    COLORS[0x1A] = COLOR{8, 124, 0}
    COLORS[0x1B] = COLOR{0, 118, 40}
    COLORS[0x1C] = COLOR{0, 102, 120}
    COLORS[0x1D] = COLOR{0, 0, 0}
    COLORS[0x1E] = COLOR{0, 0, 0}
    COLORS[0x1F] = COLOR{0, 0, 0}

    COLORS[0x20] = COLOR{236, 238, 236}
    COLORS[0x21] = COLOR{76, 154, 236}
    COLORS[0x22] = COLOR{120, 124, 236}
    COLORS[0x23] = COLOR{176, 98, 236}
    COLORS[0x24] = COLOR{228, 84, 236}
    COLORS[0x25] = COLOR{236, 88, 180}
    COLORS[0x26] = COLOR{236, 106, 100}
    COLORS[0x27] = COLOR{212, 136, 32}
    COLORS[0x28] = COLOR{160, 170, 0}
    COLORS[0x29] = COLOR{116, 196, 0}
    COLORS[0x2A] = COLOR{76, 208, 32}
    COLORS[0x2B] = COLOR{56, 204, 108}
    COLORS[0x2C] = COLOR{56, 180, 204}
    COLORS[0x2D] = COLOR{60, 60, 60}
    COLORS[0x2E] = COLOR{0, 0, 0}
    COLORS[0x2F] = COLOR{0, 0, 0}

    COLORS[0x30] = COLOR{236, 238, 236}
    COLORS[0x31] = COLOR{168, 204, 236}
    COLORS[0x32] = COLOR{188, 188, 236}
    COLORS[0x33] = COLOR{212, 178, 236}
    COLORS[0x34] = COLOR{236, 174, 236}
    COLORS[0x35] = COLOR{236, 174, 212}
    COLORS[0x36] = COLOR{236, 180, 176}
    COLORS[0x37] = COLOR{228, 196, 144}
    COLORS[0x38] = COLOR{204, 210, 120}
    COLORS[0x39] = COLOR{180, 222, 120}
    COLORS[0x3A] = COLOR{168, 226, 144}
    COLORS[0x3B] = COLOR{152, 226, 180}
    COLORS[0x3C] = COLOR{160, 214, 228}
    COLORS[0x3D] = COLOR{160, 162, 160}
    COLORS[0x3E] = COLOR{0, 0, 0}
    COLORS[0x3F] = COLOR{0, 0, 0}
}

var (
    COLOC_BLACK = COLOR{0, 0, 0}
    COLOC_RED = COLOR{255, 0, 0}
)
