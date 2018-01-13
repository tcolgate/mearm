package main

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/adammck/ik"
	"github.com/jroimartin/gocui"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
	"gobot.io/x/gobot/platforms/raspi"
)

type target struct {
	x, y, z float64
	claw    float64
}

// Context structure passed to all tests
type Context struct {
	sync.Mutex
	target  target
	targets chan target
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	ctx := &Context{
		target:  target{0.8, 1.1, 0, 45},
		//target:  target{90, 90, 45, 45},
		targets: make(chan target, 1),
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	g.SetKeybinding("", 'h', gocui.ModNone, ctx.KeyHandler('h'))
	g.SetKeybinding("", 'l', gocui.ModNone, ctx.KeyHandler('l'))
	g.SetKeybinding("", 'j', gocui.ModNone, ctx.KeyHandler('j'))
	g.SetKeybinding("", 'k', gocui.ModNone, ctx.KeyHandler('k'))
	g.SetKeybinding("", 'i', gocui.ModNone, ctx.KeyHandler('i'))
	g.SetKeybinding("", 'm', gocui.ModNone, ctx.KeyHandler('m'))
	g.SetKeybinding("", '+', gocui.ModNone, ctx.KeyHandler('+'))
	g.SetKeybinding("", '-', gocui.ModNone, ctx.KeyHandler('-'))

	r := raspi.NewAdaptor()
	srvBase := gpio.NewServoDriver(r, "15")
	srvBase.SetName("base")
	srvBase.Start()
	srvRight := gpio.NewServoDriver(r, "13")
	srvRight.SetName("right")
	srvRight.Start()
	srvLeft := gpio.NewServoDriver(r, "7")
	srvLeft.SetName("left")
	srvLeft.Start()
	srvClaw := gpio.NewServoDriver(r, "11")
	srvClaw.SetName("claw")
	srvClaw.Start()

	len1 := 0.3
	len2 := 0.8
	len3 := 0.8
	x := ik.MakeRootSegment(ik.MakeVector3(0, 0, 0))
	a := ik.MakeSegment(x, ik.Euler(0, -70, 0), ik.Euler(0, 70, 0), ik.MakeVector3(0, len1, 0))
	b := ik.MakeSegment(a, ik.Euler(0, 0, -70), ik.Euler(0, 0, 0), ik.MakeVector3(0, len2, 0))
	c := ik.MakeSegment(b, ik.Euler(0, 0, -120), ik.Euler(0, 0, -90), ik.MakeVector3(0, len3, 0))

	ctx.targets <- ctx.target

	work := func() {
		for {
			select {
			case t := <-ctx.targets:
				fmt.Printf("x: %.2f\n", t.x)
				fmt.Printf("y: %.2f\n", t.y)
				fmt.Printf("z: %.2f\n", t.z)
				fmt.Printf("claw: %.2f\n", t.claw)

				target := ik.MakeVector3(t.x, t.y, t.z)
				_, bestAngles := ik.Solve(x, target)

				// Restore the best Rotation
				a.SetRotation(&bestAngles[0])
				b.SetRotation(&bestAngles[1])
				c.SetRotation(&bestAngles[2])

				base := 90 + (float32(bestAngles[0].Pitch) * 180 / math.Pi)
				//baseP := base / 180
				srvBase.Move(uint8(base))
				fmt.Printf("base: %v\n", base)

				right := 180 - (90 + (float32(bestAngles[1].Bank) * 180 / math.Pi))
				//rightP := right / 180
				srvRight.Move(uint8(right))
				fmt.Printf("right: %v\n", right)

				left := -1 * (180 - (right + (-1 * float32(bestAngles[2].Bank) * 180 / math.Pi)))
				//leftP := left / 180
				srvLeft.Move(uint8(180 - left))
				fmt.Printf("left: %v\n", 180 - left)

/*
				srvBase.Move(uint8(t.x))
				srvRight.Move(uint8(t.y))
				srvLeft.Move(uint8(t.z))
*/
				srvClaw.Move(uint8(t.claw))
			}
		}
	}

	robot := gobot.NewRobot("srvbot",
		[]gobot.Connection{r},
		[]gobot.Device{srvBase, srvLeft, srvRight, srvClaw},
		work,
	)

	go robot.Start()
	defer robot.Stop()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "Hello world!")
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func (ctx *Context) KeyHandler(c rune) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		ctx.Lock()
		defer ctx.Unlock()
                delta := 0.02

		switch c {
		case 'h':
			ctx.target.x -= delta
		case 'l':
			ctx.target.x += delta
		case 'j':
			ctx.target.y -= delta
		case 'k':
			ctx.target.y += delta
		case 'i':
			ctx.target.z -= delta
		case 'm':
			ctx.target.z += delta
		case '-':
			ctx.target.claw -= delta
		case '+':
			ctx.target.claw += delta
		}
		ctx.targets <- ctx.target
		return nil
	}
}
