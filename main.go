// +build darwin linux

package main

import (
	"image"
	"log"
	"math"
	"time"

	"github.com/golang/freetype/truetype"

	"image/color"
	"image/draw"
	_ "image/jpeg"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"golang.org/x/mobile/app"
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
	"golang.org/x/mobile/geom"
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

const (
	spriteWidth  = 250
	spriteHeight = 40
	spriteX      = 200
	spriteY      = 200
)

var sz size.Event

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
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
	text := "Hello, Gopher!"
	fontsize := float64(30)
	dpi := float64(72)
	img := images.NewImage(spriteWidth, spriteHeight)

	c := color.RGBA{255, 0, 0, 255}
	fg, bg := image.NewUniform(c), image.Black
	draw.Draw(img.RGBA, img.RGBA.Bounds(), bg, image.Point{}, draw.Src)

	// Draw the text.
	h := font.HintingNone

	gofont, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic("failed to load font!")
	}

	d := &font.Drawer{
		Dst: img.RGBA,
		Src: fg,
		Face: truetype.NewFace(gofont, &truetype.Options{
			Size:    fontsize,
			DPI:     dpi,
			Hinting: h,
		}),
	}

	textWidth := d.MeasureString(text)

	d.Dot = fixed.Point26_6{
		X: fixed.I(spriteWidth/2) - textWidth/2,
		Y: fixed.I(int(fontsize * dpi / 72)),
	}
	d.DrawString(text)

	img.Upload()

	scale := geom.Pt(4)
	img.Draw(
		sz,
		geom.Point{X: 0, Y: (sz.HeightPt - geom.Pt(spriteHeight)/scale)},
		geom.Point{X: geom.Pt(spriteWidth) / scale, Y: (sz.HeightPt - geom.Pt(spriteHeight)/scale)},
		geom.Point{X: 0, Y: (sz.HeightPt - geom.Pt(spriteHeight)/scale)},
		img.RGBA.Bounds().Inset(1),
	)

	t, err := eng.LoadTexture(img.RGBA)
	if err != nil {
		log.Fatal(err)
	}

	return []sprite.SubTex{
		texGopherR: sprite.SubTex{T: t, R: image.Rect(0, 0, spriteWidth, spriteHeight)},
	}
}

type arrangerFunc func(e sprite.Engine, n *sprite.Node, t clock.Time)

func (a arrangerFunc) Arrange(e sprite.Engine, n *sprite.Node, t clock.Time) { a(e, n, t) }
