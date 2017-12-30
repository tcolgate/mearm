package main

import (
	"fmt"
	"math"
	"runtime"

	"github.com/adammck/ik"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/camera/control"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/window"
)

type target struct {
	x, y, z float64
}

// Context structure passed to all tests
type Context struct {
	target *target

	GS       *gls.GLS              // OpenGL state
	Win      window.IWindow        // Window
	Renderer *renderer.Renderer    // pointer to renderer object
	Camera   *camera.Perspective   // current camera
	Gui      *gui.Panel            // GUI panel container for GUI tests
	Root     *gui.Root             // GUI root container
	Orbit    *control.OrbitControl // pointer to orbit camera controller
	Scene    *core.Node            // Node container for 3D tests
	AmbLight *light.Ambient        // pointer to ambient light
}

func init() {
	runtime.LockOSThread()
}

func main() {
	// Creates window and OpenGL context
	win, err := window.New("glfw", 800, 600, "Hello G3N", false)
	if err != nil {
		panic(err)
	}

	ctx := &Context{
		target: &target{0, 0, 0},
	}
	ctx.Win = win

	// Create OpenGL state
	gs, err := gls.New()
	if err != nil {
		panic(err)
	}
	ctx.GS = gs

	// Sets the OpenGL viewport size the same as the window size
	// This normally should be updated if the window is resized.
	width, height := win.GetSize()
	gs.Viewport(0, 0, int32(width), int32(height))

	// Creates scene for 3D objects
	scene := core.NewNode()

	ctx.Scene = scene

	axis := graphic.NewAxisHelper(2)
	scene.Add(axis)

	grid := graphic.NewGridHelper(100, 1, &math32.Color{0.4, 0.4, 0.4})
	scene.Add(grid)

	// Adds white ambient light to the scene
	ambLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.5)
	scene.Add(ambLight)
	ctx.AmbLight = ambLight

	// Adds a perspective camera to the scene
	// The camera aspect ratio should be updated if the window is resized.
	aspect := float32(width) / float32(height)
	camera := camera.NewPerspective(65, aspect, 0.01, 1000)
	camera.SetPosition(0, 0, 5)
	ctx.Camera = camera

	// Creates a renderer and adds default shaders
	rend := renderer.NewRenderer(gs)
	err = rend.AddDefaultShaders()
	if err != nil {
		panic(err)
	}
	ctx.Renderer = rend

	// Creates root panel for GUI
	ctx.Root = gui.NewRoot(gs, win)
	ctx.Gui = ctx.Root.GetPanel()

	// Sets window background color
	gs.ClearColor(0, 0, 0, 1.0)

	// Subscribe to window key events
	ctx.Win.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		// ESC terminates program
		if kev.Keycode == window.KeyEscape {
			ctx.Win.SetShouldClose(true)
			return
		}
		// F10 toggles full screen
		if kev.Keycode == window.KeyF10 {
			ctx.Win.SetFullScreen(!ctx.Win.FullScreen())
			return
		}
	})

	ctx.Win.Subscribe(window.OnChar, func(evname string, ev interface{}) {
		char := ev.(*window.CharEvent)
		switch char.Char {
		case 'h':
			ctx.target.x -= 0.02
		case 'l':
			ctx.target.x += 0.02
		case 'j':
			ctx.target.y -= 0.02
		case 'k':
			ctx.target.y += 0.02
		case 'i':
			ctx.target.z -= 0.02
		case 'm':
			ctx.target.z += 0.02
		default:
			return
		}
	})

	// Subscribe to window resize events
	ctx.Win.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) {
		winResizeEvent(ctx)
	})

	ctx.Root.SubscribeWin()
	ctx.Orbit = control.NewOrbitControl(ctx.Camera, ctx.Win)

	// Creates a wireframe sphere positioned at the center of the scene
	geom := geometry.NewSphere(0.01, 8, 8, 0, math.Pi*2, 0, math.Pi)
	mat := material.NewStandard(&math32.Color{1, 1, 1})
	mat.SetSide(material.SideDouble)
	mat.SetWireframe(false)
	sphere := graphic.NewMesh(geom, mat)
	sphere.SetPosition(
		float32(ctx.target.x),
		float32(ctx.target.y),
		float32(ctx.target.z))

	scene.Add(sphere)

	armmat := material.NewStandard(&math32.Color{1, 1, 1})
	armmat.SetSide(material.SideDouble)
	armmat.SetWireframe(true)

	armRoot := core.NewNode()

	joint1 := core.NewNode()
	armRoot.Add(joint1)

	len1 := float64(0.3)
	// Creates a wireframe sphere positioned at the center of the scene
	arm1 := geometry.NewBox(0.1, len1, 0.1, 1, 1, 1)
	box1 := graphic.NewMesh(arm1, armmat)
	box1.SetPosition(0, float32(len1/2), 0)
	joint1.Add(box1)

	joint2 := core.NewNode()
	joint2.SetPosition(0, float32(len1/2), 0)
	box1.Add(joint2)

	len2 := float64(0.8)
	// Creates a wireframe sphere positioned at the center of the scene
	arm2 := geometry.NewBox(0.1, len2, 0.1, 1, 1, 1)
	box2 := graphic.NewMesh(arm2, armmat)
	box2.SetPosition(0, float32(len2/2), 0)
	joint2.Add(box2)

	joint3 := core.NewNode()
	joint3.SetPosition(0, float32(len2/2), 0)
	box2.Add(joint3)

	len3 := float64(0.8)
	arm3 := geometry.NewBox(0.1, len3, 0.1, 1, 1, 1)
	box3 := graphic.NewMesh(arm3, armmat)
	box3.SetPosition(0, float32(len3/2), 0)
	joint3.Add(box3)

	scene.Add(armRoot)

	x := ik.MakeRootSegment(ik.MakeVector3(0, 0, 0))
	a := ik.MakeSegment(x, ik.Euler(0, -70, 0), ik.Euler(0, 70, 0), ik.MakeVector3(0, len1, 0))
	b := ik.MakeSegment(a, ik.Euler(0, 0, -70), ik.Euler(0, 0, 0), ik.MakeVector3(0, len2, 0))
	c := ik.MakeSegment(b, ik.Euler(0, 0, -120), ik.Euler(0, 0, -90), ik.MakeVector3(0, len3, 0))

	target := ik.MakeVector3(ctx.target.x, ctx.target.y, ctx.target.z)
	_, bestAngles := ik.Solve(x, target)

	// Restore the best Rotation
	a.SetRotation(&bestAngles[0])
	b.SetRotation(&bestAngles[1])
	c.SetRotation(&bestAngles[2])

	joint1.SetRotationX(float32(bestAngles[0].Heading))
	joint1.SetRotationY(float32(bestAngles[0].Pitch))
	joint1.SetRotationZ(float32(bestAngles[0].Bank))
	joint2.SetRotationX(float32(bestAngles[1].Heading))
	joint2.SetRotationY(float32(bestAngles[1].Pitch))
	joint2.SetRotationZ(float32(bestAngles[1].Bank))
	joint3.SetRotationX(float32(bestAngles[2].Heading))
	joint3.SetRotationY(float32(bestAngles[2].Pitch))
	joint3.SetRotationZ(float32(bestAngles[2].Bank))

	// Render loop
	for !win.ShouldClose() {
		// Clear buffers
		gs.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

		sphere.SetPosition(
			float32(ctx.target.x),
			float32(ctx.target.y),
			float32(ctx.target.z))

		target := ik.MakeVector3(ctx.target.x, ctx.target.y, ctx.target.z)
		_, bestAngles := ik.Solve(x, target)

		// Restore the best Rotation
		a.SetRotation(&bestAngles[0])
		b.SetRotation(&bestAngles[1])
		c.SetRotation(&bestAngles[2])

		joint1.SetRotationX(float32(bestAngles[0].Heading))
		joint1.SetRotationY(float32(bestAngles[0].Pitch))
		joint1.SetRotationZ(float32(bestAngles[0].Bank))
		joint2.SetRotationX(float32(bestAngles[1].Heading))
		joint2.SetRotationY(float32(bestAngles[1].Pitch))
		joint2.SetRotationZ(float32(bestAngles[1].Bank))
		joint3.SetRotationX(float32(bestAngles[2].Heading))
		joint3.SetRotationY(float32(bestAngles[2].Pitch))
		joint3.SetRotationZ(float32(bestAngles[2].Bank))

		base := 90 + (float32(bestAngles[0].Pitch) * 180 / math.Pi)
		baseP := base / 180

		right := 180 - (90 + (float32(bestAngles[1].Bank) * 180 / math.Pi))
		rightP := right / 180

		left := -1 * (180 - (right + (-1 * float32(bestAngles[2].Bank) * 180 / math.Pi)))
		leftP := left / 180

		fmt.Printf("b; %v (%v) r: %v (%v) l: %v (%v)\n", base, baseP, right, rightP, left, leftP)

		// Render the scene using the specified camera
		//rend.Render(scene, camera)
		rend.SetScene(scene)
		rend.Render(camera)

		// Update window and checks for I/O events
		win.SwapBuffers()
		win.PollEvents()
	}
}

// winResizeEvent is called when the window resize event is received
func winResizeEvent(ctx *Context) {

	// Sets view port
	width, height := ctx.Win.GetSize()
	ctx.GS.Viewport(0, 0, int32(width), int32(height))
	aspect := float32(width) / float32(height)

	// Sets camera aspect ratio
	ctx.Camera.SetAspect(aspect)

	// Sets GUI root panel size
	ctx.Root.SetSize(float32(width), float32(height))
}
