package apu

import (
    "nes/tools"
)

type Sweeper struct {
    enabled bool
    down bool
    reload bool
    shift uint8
    timer uint8
    period uint8
    change uint16
    mute bool
}

func (self *Sweeper) track( target uint16 ) {
    if self.enabled {
        self.change = target >> uint16(self.shift)
        self.mute = (target < 8) || (target > 0x7FF)
    }
}


func (self *Sweeper) clock( target *uint16, channel bool ) bool {
    changed := false
    if self.timer == 0 && self.enabled && self.shift > 0 && !self.mute {
        if *target >= 8 && self.change < 0x07FF {
            if self.down {
                *target -= self.change - uint16(tools.B2i(channel))
            } else {
                *target += self.change
            }
            changed = true
        }
    }

    // if self.enabled 
    {
        if self.timer == 0 || self.reload {
            self.timer = self.period
            self.reload = false
        } else {
            self.timer--
        }
        self.mute = (*target < 8) || (*target > 0x7FF)
    }

    return changed
}



