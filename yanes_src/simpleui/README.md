

```
type View interface {
	Enter()
	Exit()
	Update(t, dt float64)
}

type CustomViewIF interface {
	View
	SetGLWindow(window *glfw.Window)
	SetAudioDevice(*Audio)
	OnKey(glfw.Key)
	TextureBuff() []uint8
    Title() string
}

simpleui.Run( your_custom_view )
```
