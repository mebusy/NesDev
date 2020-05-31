package apu

type Lengthcounter struct {
    counter uint8
}

func (self *Lengthcounter) clock( bEnable, bHalt bool  ) uint8 {
    if !bEnable {
        self.counter = 0
    } else {
        if self.counter > 0 && !bHalt {
            self.counter--
        }
    }
    return self.counter
}
