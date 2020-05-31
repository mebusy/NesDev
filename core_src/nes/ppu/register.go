package ppu

type STATUS_PPU uint8  // $2002

const (
    _ STATUS_PPU = 1 << iota  // bit 0
    _              // bit 1
    _              // bit 2
    _              // bit 3
    _              // bit 4
    ST_PPU_SPR_OVERFLOW         // bit 5
    ST_PPU_SPR_ZERO_HIT         // bit 6
    ST_PPU_VBLANK               // bit 7
)


type MASK_PPU uint8  // $2001

const (
    MASK_PPU_GRAYSCALE MASK_PPU  = 1 << iota
    MASK_PPU_RENDER_BG_LEFT
    MASK_PPU_RENDER_SPR_LEFT
    MASK_PPU_RENDER_BG
    MASK_PPU_RENDER_SPR
    MASK_PPU_ENHANCE_RED
    MASK_PPU_ENHANCE_GREEN
    MASK_PPU_ENHANCE_BLUE
)


type CTRL_PPU uint8   // $2000

const (
    CTRL_PPU_NAMETABLE_X    CTRL_PPU = 1 << iota
    CTRL_PPU_NAMETABLE_Y     // D1D0:  name table address 
    CTRL_PPU_INCREMENT_MODE  // by 1 , or by 32
    CTRL_PPU_PATTERN_SP         // 0=$0000,1=$1000
    CTRL_PPU_PATTERN_B          // 0=$0000,1=$1000
    CTRL_PPU_SPR_SIZE           // 8x8, or 8x16
    CTRL_PPU_SLAVEMODE          // unused
    CTRL_PPU_ENABLE_NMI         // execute NMI on vblack  1:enable
)

func (self *Ppu) CtrlRegNameTableX () CTRL_PPU {
    return self.ctrl_reg & 0x1
}
func (self *Ppu) CtrlRegNameTableY () CTRL_PPU {
    return (self.ctrl_reg>>1) & 0x1
}
func (self *Ppu) CtrlRegPatternSPR() CTRL_PPU {
    return (self.ctrl_reg>>3) & 0x1
}
func (self *Ppu) CtrlRegPatternBG() CTRL_PPU {
    return (self.ctrl_reg>>4) & 0x1
}
