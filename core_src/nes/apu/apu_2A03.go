package apu

import (
    "log"
    "nes/cpu"
)

type Apu struct {

    Ch_sample chan float32

    clock_counter uint32
    frame_clock_counter uint32  // used in maintaining the musical timing

    // channels [5]*Channel
    sequencers [5]*Sequencer
    oscpulses [5]*OscPulse
    envelopes [5]*Envelope
    lengthcounters [5]*Lengthcounter
    sweepers [5]*Sweeper

    dGlobalTime float64

    bUseRawMode bool

    enable  [5]bool
    halt    [5]bool
    samples  [5]float32
    outputs  [5]float32
}

var length_table = []uint8 {
    10, 254, 20,  2, 40,  4, 80,  6,
        160,   8, 60, 10, 14, 12, 26, 14,
        12,  16, 24, 18, 48, 20, 96, 22,
        192,  24, 72, 26, 16, 28, 32, 30 }

func NewApu() *Apu {
    log.Println( "apu instanciated" )
    apu := &Apu { Ch_sample: nil}


    for i:= PULSE1_CHANNEL ; i<= WAVE_CHANNEL ; i++ {
        // apu.channels[i] = &Channel{}
        apu.sequencers[i] = &Sequencer{}
        apu.oscpulses[i] = NewOscPulse()
        apu.envelopes[i] = &Envelope{}
        apu.lengthcounters[i] = &Lengthcounter{}
        apu.sweepers[i] = &Sweeper{}
    }

    apu.sequencers[NOISE_CHANNEL].sequence = 0xDBDB


    return apu
}


func (self *Apu) CpuRead(addr uint16) uint8 {
    panic("did not implement APU READ")
}

func (self *Apu) CpuWrite(addr uint16, data uint8) {
    switch addr {
    // Pulse 1
    case 0x4000:
        // set duty cycle
        ch := PULSE1_CHANNEL
        switch (data & 0xC0) >> 6 {
        case 0x00: self.sequencers[ch].new_sequence = 0b01000000; self.oscpulses[ch].dutycycle = 0.125
        case 0x01: self.sequencers[ch].new_sequence = 0b01100000; self.oscpulses[ch].dutycycle = 0.250
        case 0x02: self.sequencers[ch].new_sequence = 0b01111000; self.oscpulses[ch].dutycycle = 0.500
        case 0x03: self.sequencers[ch].new_sequence = 0b10011111; self.oscpulses[ch].dutycycle = 0.750
        }
        self.sequencers[ch].sequence = self.sequencers[ch].new_sequence
        self.halt[ch] = (data & 0x20) != 0
        self.envelopes[ch].volume = uint16(data) & 0x0F
        self.envelopes[ch].disable = (data & 0x10) != 0
    case 0x4001:
        ch := PULSE1_CHANNEL
        self.sweepers[ch].enabled = data & 0x80 != 0
        self.sweepers[ch].period = (data & 0x70) >> 4
        self.sweepers[ch].down = data & 0x08 != 0
        self.sweepers[ch].shift = data & 0x07
        self.sweepers[ch].reload = true
    case 0x4002:
        ch := PULSE1_CHANNEL
        self.sequencers[ch].reload = (self.sequencers[ch].reload & 0xFF00) | uint16(data)
    case 0x4003:
        ch := PULSE1_CHANNEL
        self.sequencers[ch].reload = ((uint16(data) & 0x07) << 8) | (self.sequencers[ch].reload & 0x00FF)
        self.sequencers[ch].timer = self.sequencers[ch].reload
        self.sequencers[ch].sequence = self.sequencers[ch].new_sequence
        self.lengthcounters[ch].counter = length_table[(data & 0xF8) >> 3]
        self.envelopes[ch].start = true

    // Pulse 2
    case 0x4004:
        ch := PULSE2_CHANNEL
        switch (data & 0xC0) >> 6 {
        case 0x00: self.sequencers[ch].new_sequence = 0b01000000; self.oscpulses[ch].dutycycle = 0.125
        case 0x01: self.sequencers[ch].new_sequence = 0b01100000; self.oscpulses[ch].dutycycle = 0.250
        case 0x02: self.sequencers[ch].new_sequence = 0b01111000; self.oscpulses[ch].dutycycle = 0.500
        case 0x03: self.sequencers[ch].new_sequence = 0b10011111; self.oscpulses[ch].dutycycle = 0.750
        }
        self.sequencers[ch].sequence = self.sequencers[ch].new_sequence
        self.halt[ch] = (data & 0x20) != 0
        self.envelopes[ch].volume = uint16(data) & 0x0F
        self.envelopes[ch].disable = (data & 0x10) != 0
    case 0x4005:
        ch := PULSE2_CHANNEL
        self.sweepers[ch].enabled = data & 0x80 != 0
        self.sweepers[ch].period = (data & 0x70) >> 4
        self.sweepers[ch].down = data & 0x08 != 0
        self.sweepers[ch].shift = data & 0x07
        self.sweepers[ch].reload = true
    case 0x4006:
        ch := PULSE2_CHANNEL
        self.sequencers[ch].reload = (self.sequencers[ch].reload & 0xFF00) | uint16(data)
    case 0x4007:
        ch := PULSE2_CHANNEL
        self.sequencers[ch].reload = ((uint16(data) & 0x07) << 8) | (self.sequencers[ch].reload & 0x00FF)
        self.sequencers[ch].timer = self.sequencers[ch].reload
        self.sequencers[ch].sequence = self.sequencers[ch].new_sequence
        self.lengthcounters[ch].counter = length_table[(data & 0xF8) >> 3]
        self.envelopes[ch].start = true

    // Triangle
    case 0x4008:
    case 0x4009:
    case 0x400A:
    case 0x400B:

    // Noise
    case 0x400C:
        ch := NOISE_CHANNEL
        self.envelopes[ch].volume = (uint16(data) & 0x0F)
        self.envelopes[ch].disable = (data & 0x10) != 0
        self.halt[ch] = (data & 0x20) != 0
    case 0x400D:
    case 0x400E:
        ch := NOISE_CHANNEL
        switch (data & 0x0F) {
        case 0x00: self.sequencers[ch].reload = 0
        case 0x01: self.sequencers[ch].reload = 4
        case 0x02: self.sequencers[ch].reload = 8
        case 0x03: self.sequencers[ch].reload = 16
        case 0x04: self.sequencers[ch].reload = 32
        case 0x05: self.sequencers[ch].reload = 64
        case 0x06: self.sequencers[ch].reload = 96
        case 0x07: self.sequencers[ch].reload = 128
        case 0x08: self.sequencers[ch].reload = 160
        case 0x09: self.sequencers[ch].reload = 202
        case 0x0A: self.sequencers[ch].reload = 254
        case 0x0B: self.sequencers[ch].reload = 380
        case 0x0C: self.sequencers[ch].reload = 508
        case 0x0D: self.sequencers[ch].reload = 1016
        case 0x0E: self.sequencers[ch].reload = 2034
        case 0x0F: self.sequencers[ch].reload = 4068
        }
    case 0x400F:
        // self.envelopes[PULSE1_CHANNEL].start = true
        // self.envelopes[PULSE2_CHANNEL].start = true
        self.envelopes[NOISE_CHANNEL].start = true
        self.lengthcounters[NOISE_CHANNEL].counter = length_table[(data & 0xF8) >> 3]

    // DMC
    case 0x4010:
    case 0x4011:
    case 0x4012:
    case 0x4013:

    case 0x4015: // status 
        // the whole channle can be enabled or disabled
        // self.enable[PULSE1_CHANNEL] = (data & 0x01) != 0
        // self.enable[TRIANGLE_CHANNEL] = (data & 0x04) != 0
        // ...
        for i:=PULSE1_CHANNEL; i<=NOISE_CHANNEL; i++ {
            self.enable[i] = (data & (1<<i) ) != 0

            if !self.enable[i] {
                // apu.pulse1.lengthValue = 0
                self.lengthcounters[i].counter = 0
            }
        }
        // DMC
        /*
        if !apu.dmc.enabled {
            apu.dmc.currentLength = 0
        } else {
            if apu.dmc.currentLength == 0 {
                apu.dmc.restart()
            }
        }
        //*/

    case 0x4017: // frame count
        // log.Println( "write frame counter", data  )
    }
}

func (self *Apu) Clock() {
    bQuarterFrameClock := false
    bHalfFrameClock := false

    self.dGlobalTime += (0.3333333333 / float64(cpu.CPU_FREQUENCY))

    // 6 times slow that ppu clock
    if self.clock_counter % 6 == 0 {
        self.frame_clock_counter ++

        // 4-Step Sequence Mode
        if self.frame_clock_counter == 3729 {
            bQuarterFrameClock = true
        }

        if self.frame_clock_counter == 7457 {
            bQuarterFrameClock = true
            bHalfFrameClock = true
        }

        if self.frame_clock_counter == 11186 {
            bQuarterFrameClock = true
        }

        if self.frame_clock_counter == 14916 {
            bQuarterFrameClock = true
            bHalfFrameClock = true
            self.frame_clock_counter = 0
        }

        // Update functional units

        // Quater frame "beats" adjust the volume envelope
        if (bQuarterFrameClock) {
            // pulse1_env.clock(pulse1_halt);
            // pulse2_env.clock(pulse2_halt);
            // noise_env.clock(noise_halt);
            self.envelopes[PULSE1_CHANNEL].clock( self.halt[PULSE1_CHANNEL] )
            self.envelopes[PULSE2_CHANNEL].clock( self.halt[PULSE2_CHANNEL] )
            self.envelopes[NOISE_CHANNEL].clock( self.halt[NOISE_CHANNEL] )
        }
        // Half frame "beats" adjust the note length and
        // frequency sweepers
        if (bHalfFrameClock) {
            // pulse1_lc.clock(pulse1_enable, pulse1_halt);
            // pulse2_lc.clock(pulse2_enable, pulse2_halt);
            // noise_lc.clock(noise_enable, noise_halt);
            self.lengthcounters[PULSE1_CHANNEL].clock( self.enable[PULSE1_CHANNEL],self.halt[PULSE1_CHANNEL] )
            self.lengthcounters[PULSE2_CHANNEL].clock( self.enable[PULSE2_CHANNEL],self.halt[PULSE2_CHANNEL] )
            self.lengthcounters[NOISE_CHANNEL].clock( self.enable[NOISE_CHANNEL],self.halt[NOISE_CHANNEL] )

            // pulse1_sweep.clock(pulse1_seq.reload, 0);
            // pulse2_sweep.clock(pulse2_seq.reload, 1);
            self.sweepers[PULSE1_CHANNEL].clock( &self.sequencers[PULSE1_CHANNEL].reload, false )
            self.sweepers[PULSE2_CHANNEL].clock( &self.sequencers[PULSE2_CHANNEL].reload, true )
        }

        // Update Pulse1 Channel
        for i:= PULSE1_CHANNEL; i<= PULSE2_CHANNEL; i++ {
            ch := i
            seq := self.sequencers[ch]
            osc := self.oscpulses[ch]
            env := self.envelopes[ch]
            lc := self.lengthcounters[ch]
            sweep := self.sweepers[ch]
            seq.clock(
                self.enable[ch],
                func(s *uint32) {
                    // Shift 8-bit right by 1 bit, wrapping around
                    *s = ((*s & 0x0001) << 7) | ((*s & 0x00FE) >> 1)
                } )
            // self.samples[ch] = float32(seq.output)

            osc.frequency = float64(cpu.CPU_FREQUENCY) / (16.0 * (float64)(seq.reload + 1))
            osc.amplitude = (float64)(env.output -1) / 16.0
            self.samples[ch] = float32(osc.sample(self.dGlobalTime))
            if (lc.counter > 0 && seq.timer >= 8 && !sweep.mute && env.output > 2) {
                self.outputs[ch] += (self.samples[ch] - self.outputs[ch]) * 0.5
            } else {
                self.outputs[ch] = 0
            }
        } // end 2 pulse wave

        {
            ch := NOISE_CHANNEL
            seq := self.sequencers[ch]
            // osc := self.oscpulses[ch]
            env := self.envelopes[ch]
            lc := self.lengthcounters[ch]
            // sweep := self.sweepers[ch]
            seq.clock(
                self.enable[ch],
                func(s *uint32) {
                    *s = (((*s & 0x0001) ^ ((*s & 0x0002) >> 1)) << 14) | ((*s & 0x7FFF) >> 1)
                } )


            // if lc.counter > 0 && seq.timer >= 8 {
            if (lc.counter > 0 && seq.timer >= 8 && env.output > 2) {
                self.outputs[ch] = float32(seq.output) * ((float32(env.output)-1) / 16.0)
                // 0 ~ 1 
            } else {
                self.outputs[ch] = 0
            }
        }

        for i:= PULSE1_CHANNEL; i<=NOISE_CHANNEL; i++ {
            ch := i
            if !self.enable[ch] {
                self.outputs[ch] = 0
            }
        }

    } // end %6 == 0

    for i:= PULSE1_CHANNEL; i<=PULSE2_CHANNEL; i++ {
        ch := i
        self.sweepers[ch].track(self.sequencers[ch].reload);
    }

    self.clock_counter++
}

func (self *Apu) Reset() {
    log.Println( "apu reseted" )
}

func (self *Apu) SendOutputSample()  {
    if self.Ch_sample != nil {
        /*
        output := ((1.0 * self.outputs[PULSE1_CHANNEL]) - 0.8) * 0.1 +
                    ((1.0 * self.outputs[PULSE2_CHANNEL]) - 0.8) * 0.1 +
                    ((2.0 * (self.outputs[NOISE_CHANNEL] - 0.5))) * 0.1
        /*/
        output := (self.outputs[PULSE1_CHANNEL]  +  self.outputs[PULSE2_CHANNEL]  +
                     (2.0 * (self.outputs[NOISE_CHANNEL] - 0.5)) +
                     0 ) * 0.1
        //*/

        self.Ch_sample <- output
    }
}


