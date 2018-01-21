package main

import (
	"fmt"
	"math"
	"sync"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"

	"github.com/adammck/ik"
	"github.com/llgcode/draw2d/draw2dkit"

	"github.com/tcolgate/mearm/bot/raspi"
	"github.com/tcolgate/mearm/draw2dmearm"
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

type arm struct {
	r *raspi.Adaptor

	base  *gpio.ServoDriver
	left  *gpio.ServoDriver
	right *gpio.ServoDriver
	claw  *gpio.ServoDriver

	baseLen float64
	seg1Len float64
	seg2Len float64

	rootSeg *ik.Segment
	baseSeg *ik.Segment
	seg1    *ik.Segment
	seg2    *ik.Segment

	drawz     float64
	drawLift  float64
	drawSleep time.Duration
}

func New() *arm {
	r := raspi.NewAdaptor()
	a := &arm{
		r:        r,
		drawz:    0.8,
		drawLift: 0.1,
	}

	a.base = gpio.NewServoDriver(r, "15")
	a.base.SetName("base")
	a.base.Start()
	a.right = gpio.NewServoDriver(r, "13")
	a.right.SetName("right")
	a.right.Start()
	a.left = gpio.NewServoDriver(r, "7")
	a.left.SetName("left")
	a.left.Start()
	a.claw = gpio.NewServoDriver(r, "11")
	a.claw.SetName("claw")
	a.claw.Start()

	a.baseLen = 0.3
	a.seg1Len = 0.8
	a.seg2Len = 0.8

	a.rootSeg = ik.MakeRootSegment(ik.MakeVector3(0, 0, 0))
	a.baseSeg = ik.MakeSegment(a.rootSeg, ik.Euler(0, -70, 0), ik.Euler(0, 70, 0), ik.MakeVector3(0, a.baseLen, 0))
	a.seg1 = ik.MakeSegment(a.baseSeg, ik.Euler(0, 0, -70), ik.Euler(0, 0, 0), ik.MakeVector3(0, a.seg1Len, 0))
	a.seg2 = ik.MakeSegment(a.seg1, ik.Euler(0, 0, -120), ik.Euler(0, 0, -90), ik.MakeVector3(0, a.seg2Len, 0))

	return a
}

func (a *arm) Target(x, y, z float64) {
	fmt.Printf("Target: %v,%v,%v\n", x, y, z)
	target := ik.MakeVector3(x, y, z)
	_, bestAngles := ik.Solve(a.rootSeg, target)

	// Restore the best Rotation
	a.baseSeg.SetRotation(&bestAngles[0])
	a.seg1.SetRotation(&bestAngles[1])
	a.seg2.SetRotation(&bestAngles[2])

	base := 90 + (float32(bestAngles[0].Pitch) * 180 / math.Pi)
	a.base.Move(uint8(base))

	right := 180 - (90 + (float32(bestAngles[1].Bank) * 180 / math.Pi))
	a.right.Move(uint8(right))

	left := -1 * (180 - (right + (-1 * float32(bestAngles[2].Bank) * 180 / math.Pi)))
	a.left.Move(uint8(180 - left))

	fmt.Printf("Angles %v,%v,%v\n", base, right, 180-left)

	time.Sleep(500 * time.Millisecond)
	//	a.base.Move(uint8(x))
	//	a.right.Move(uint8(y))
	//	a.left.Move(uint8(z))
	// srvClaw.Move(uint8(t.claw))
}

func (a *arm) clawOpen(open float64) {
	a.claw.Move(uint8(open))
}

func (a *arm) MoveTo(x, y float64) {
	a.Target(a.drawz-a.drawLift, x, y)
}

func (a *arm) LineTo(x, y float64) {
	a.Target(a.drawz, x, y)
}

func (a *arm) End() {
}

func main() {
	arm := New()

	work := func() {
		fn(arm)
	}

	robot := gobot.NewRobot("srvbot",
		[]gobot.Connection{arm.r},
		[]gobot.Device{arm.base, arm.left, arm.right, arm.claw},
		work,
	)

	robot.Start()
	defer robot.Stop()
}

func fn(a *arm) {
	gc := draw2dmearm.NewGraphicContext(a)

	gc.BeginPath() // Initialize a new path
	draw2dkit.Circle(gc, 0, 0.50, 0.20)
	gc.Stroke()

	/*
		gc.SetLineWidth(0)
		gc.BeginPath() // Initialize a new path
		gc.MoveTo(0, 0.3)
		gc.LineTo(0, 0.6)
		gc.LineTo(0, 0.8)
		gc.LineTo(0.2, 0.8)
		gc.Stroke()
	*/
}
