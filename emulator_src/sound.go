package main

import (
    "github.com/hajimehoshi/oto"
    "log"
)

type SoundPlayer struct {
    context *oto.Context
}

func NewSoundPlayer(sampleRate, channelNum, bitDepthInBytes, buffSizeInBytes int) *SoundPlayer {

    sp := &SoundPlayer{}
    // NewContext creates and holds ready-to-use Player objects.
    // go newDriver <- Context.mux
    c, err := oto.NewContext(sampleRate, channelNum, bitDepthInBytes, buffSizeInBytes )
    if err != nil {
        log.Fatal(err)
    }

    sp.context = c
    return sp
}

func (self *SoundPlayer) Close() {
    self.context.Close()
}

var soundPlayer *SoundPlayer


