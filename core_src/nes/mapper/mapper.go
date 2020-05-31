package mapper

const (
    MIRROR_HARDWARE = iota -1
    MIRROR_HORIZONTAL
    MIRROR_VERTICAL
    MIRROR_ONESCREEN_LO
    MIRROR_ONESCREEN_HI
)

const (
    READFROMSRAM = 0xFFFFFFFF
)

// mapper not actually provide any data 
// it is just translate the address
type MAPPER_RW interface {
    CpuMapRead( addr uint16, data *uint8 ) (bool,uint)
    CpuMapWrite( addr uint16, data uint8 ) (bool,uint)
    PpuMapRead( addr uint16 ) (bool,uint)
    PpuMapWrite( addr uint16) (bool,uint)
    Reset()
    Mirror() int
    IrqState() bool
    IrqClear()
    Scanline()
}

func init() {
    if MIRROR_HORIZONTAL != 0 {
        panic ( "mapper const error!" )
    }
}

