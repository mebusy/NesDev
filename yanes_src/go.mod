module gones

go 1.13

require (
	github.com/go-gl/gl v0.0.0-20231021071112-07e5d0ea2e71 // indirect
	github.com/go-gl/glfw v0.0.0-20240118000515-a250818d05e3
	github.com/gordonklaus/portaudio v0.0.0-20230709114228-aafa478834f5 // indirect
	github.com/mebusy/simpleui v0.0.0-20240220094802-ecce9f4c57ea // indirect
	// github.com/mebusy/simpleui v0.0.0
	nes v0.0.0
)

replace nes v0.0.0 => ../core_src/nes

// replace github.com/mebusy/simpleui v0.0.0 => ../../mebusy_git_simpleui
