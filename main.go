// +build darwin linux

package main

import (
	"image"
	"log"
	"math"
	"time"

	_ "image/jpeg"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/app/debug"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/exp/gl/glutil"
	"golang.org/x/mobile/exp/sprite"
	"golang.org/x/mobile/exp/sprite/clock"
	"golang.org/x/mobile/exp/sprite/glsprite"
	"golang.org/x/mobile/gl"
)

var (
	startTime = time.Now()
	images    *glutil.Images
	eng       sprite.Engine
	scene     *sprite.Node
	fps       *debug.FPS
	degree    float32
	affine    *f32.Affine
)

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		var sz size.Event
		for e := range a.Events() {
			switch e := a.Filter(e).(type) {
			case lifecycle.Event:
				switch e.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					glctx, _ = e.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					onStop()
					glctx = nil
				}
			case size.Event:
				sz = e
			case paint.Event:
				if glctx == nil || e.External {
					continue
				}
				degree++
				onPaint(glctx, sz)
				a.Publish()
				a.Send(paint.Event{}) // keep animating
			case touch.Event:
				/* nop */
			}
		}
	})
}

func onStart(glctx gl.Context) {
	images = glutil.NewImages(glctx)
	fps = debug.NewFPS(images)
	eng = glsprite.Engine(images)
	loadScene()
}

func onStop() {
	eng.Release()
	fps.Release()
	images.Release()
}

func onPaint(glctx gl.Context, sz size.Event) {
	glctx.ClearColor(1, 1, 1, 1)
	glctx.Clear(gl.COLOR_BUFFER_BIT)
	now := clock.Time(time.Since(startTime) * 60 / time.Second)
	eng.Render(scene, now, sz)
	fps.Draw(sz)
}

func newNode() *sprite.Node {
	n := &sprite.Node{}
	eng.Register(n)
	scene.AppendChild(n)
	return n
}

func loadScene() {
	texs := loadTextures()
	scene = &sprite.Node{}
	eng.Register(scene)
	eng.SetTransform(scene, f32.Affine{
		{1, 0, 0},
		{0, 1, 0},
	})

	var n *sprite.Node

	n = newNode()
	eng.SetSubTex(n, texs[texGopherR])
	spriteWidth := float32(140)
	spriteHeight := float32(90)
	spriteX := float32(200)
	spriteY := float32(200)
	n.Arranger = arrangerFunc(func(eng sprite.Engine, n *sprite.Node, t clock.Time) {
		radian := float32(degree) * math.Pi / 180

		// initialize affine variable
		affine = &f32.Affine{
			{1, 0, 0},
			{0, 1, 0},
		}

		// move to specified position
		affine.Translate(affine, spriteX, spriteY)

		// rotation
		rotate(affine, radian, spriteWidth, spriteHeight)

		// apply affine transformation
		eng.SetTransform(n, *affine)
	})
}

func rotate(affine *f32.Affine, radian, spriteWidth, spriteHeight float32) {
	affine.Translate(affine, 0.5, 0.5)
	affine.Rotate(affine, radian)
	affine.Scale(affine, spriteWidth, spriteHeight)
	affine.Translate(affine, -0.5, -0.5)
}

const (
	texGopherR = iota
)

func loadTextures() []sprite.SubTex {
	a, err := asset.Open("waza-gophers.jpeg")
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	img, _, err := image.Decode(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := eng.LoadTexture(img)
	if err != nil {
		log.Fatal(err)
	}

	return []sprite.SubTex{
		texGopherR: sprite.SubTex{T: t, R: image.Rect(152, 10, 152+140, 10+90)},
	}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
