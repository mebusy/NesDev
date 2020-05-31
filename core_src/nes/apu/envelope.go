package  apu

type Envelope struct {
    start  bool
    disable  bool
    divider_count uint16
    volume uint16
    output uint16
    decay_count uint16
}


func (self *Envelope) clock( bLoop bool ) {
    if !self.start {
        if self.divider_count == 0 {
            self.divider_count = self.volume

            if self.decay_count == 0 {
                if bLoop {
                    self.decay_count = 15
                }
            } else {
                self.decay_count--
            }
        } else {
            self.divider_count--
        }
    } else {
        self.start = false
        self.decay_count = 15
        self.divider_count = self.volume
    }

    if self.disable {
        self.output = self.volume
    } else {
        self.output = self.decay_count
    }
}

