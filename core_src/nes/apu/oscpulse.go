package apu

import (
    "math"
)

type OscPulse struct {
    frequency float64
    dutycycle float64
    amplitude float64
    harmonics float64
}


func NewOscPulse() *OscPulse {
    osc := &OscPulse{ amplitude:1, harmonics:2 }
    return osc
}


const (
    DOUBLE_PI = 2 * math.Pi
    HALF_PI = math.Pi / 2
    TWO_OVER_PI = 2 / math.Pi
)


func (self *OscPulse) sample( t float64 ) float64 {
    var a float64  // a,b represent the sample values of the underlying sine wave forms
    var b float64
    var p float64 = self.dutycycle * DOUBLE_PI

    /*
    for n := 1.0; n < self.harmonics ; n++ {
        c := n * self.frequency * 2.0 * math.Pi * t
        a += -approxsin( c ) / n
        b += -approxsin( c - p*n ) / n
    }
    return (( 2*self.amplitude / math.Pi) * (a-b)  * 0.48)   // y1-y2 may generate wave form with ampl ±2.x
    /*/
    _, fract := math.Modf( self.frequency* t + 0)
    a = TWO_OVER_PI * self.amplitude * ( math.Pi*( fract ) - HALF_PI )
    _, fract = math.Modf( fract + p)
    b = TWO_OVER_PI * self.amplitude * ( math.Pi*( fract ) - HALF_PI )

    return (a-b) // * 0.5
    //*/

}


