package apu

import (
    // "log"
)

type Sequencer struct {
    sequence uint32  // some data stored in the sequencer
    timer uint16  // counter value
    reload uint16
    output uint8  // instant output for the sequencer

    new_sequence uint32
}

type SEQ_FUNC func(*uint32)

func (self *Sequencer) clock( bEnable bool, fn SEQ_FUNC ) uint8 {
    if bEnable {
        self.timer--
        if self.timer == 0xFFFF {
            self.timer = self.reload + 1
            fn( &self.sequence )
            // lsb
            self.output = uint8(self.sequence) & 0x1
            // log.Println( "req rotated", self.sequence, self.output )
        }
    }

    return self.output
}
