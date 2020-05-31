module emulator

go 1.13

require (
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/hajimehoshi/ebiten v1.11.1
	github.com/hajimehoshi/oto v0.5.4
	golang.org/x/image v0.0.0-20200430140353-33d19683fad8
	nes v0.0.0
)

replace nes v0.0.0 => ../core_src/nes
