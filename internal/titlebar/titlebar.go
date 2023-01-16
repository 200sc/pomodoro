package titlebar

import (
	"image"
	"image/color"
	"strconv"
	"time"

	oak "github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/alg/intgeom"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/entities/x/btn"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/mouse"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/render/mod"
	"github.com/oakmound/oak/v4/scene"
	"github.com/oakmound/oak/v4/shape"
)

type TitleBar struct {
	lastPressAt        time.Time
	draggingStartPos   floatgeom.Point2
	draggingWindow     bool
	buttons            map[Button]*entities.Entity
	startingDimensions intgeom.Point2
	maximized          bool
	// TODO: restore desktop position getter
	DesktopPosition floatgeom.Point2
}

type Constructor struct {
	Color          color.Color
	HighlightColor color.Color
	MouseDownColor color.Color
	Height         float64
	Layers         []int

	Title          string
	TitleFontSize  int
	TitleXOffset   int
	TitleTextColor color.Color

	Buttons              []Button
	ButtonWidth          float64
	DoubleClickThreshold time.Duration
}

type Button uint8

// Buttons to show on the title bar
const (
	ButtonMinimize Button = iota
	ButtonMaximize Button = iota
	ButtonClose    Button = iota
)

var DefaultConstructor = Constructor{
	Color:  color.RGBA{128, 128, 128, 255},
	Height: 32,
	Layers: []int{},
	Buttons: []Button{
		ButtonMinimize,
		ButtonMaximize,
		ButtonClose,
	},
	ButtonWidth:          32,
	TitleFontSize:        17,
	TitleXOffset:         10,
	TitleTextColor:       color.RGBA{255, 255, 255, 255},
	DoubleClickThreshold: 200 * time.Millisecond,
}

// New constructs a new TitleBar
func New(ctx *scene.Context, opts ...Option) *TitleBar {

	construct := DefaultConstructor
	for _, opt := range opts {
		construct = opt(construct)
	}
	if construct.HighlightColor == nil {
		construct.HighlightColor = mod.Lighter(construct.Color, .10)
	}
	if construct.MouseDownColor == nil {
		construct.MouseDownColor = mod.Lighter(construct.Color, .20)
	}

	dims := ctx.Window.Bounds()
	screenWidth := dims.X()
	screenHeight := dims.Y()

	font, _ := render.DefaultFont().RegenerateWith(func(fg render.FontGenerator) render.FontGenerator {
		fg.Size = float64(construct.TitleFontSize)
		fg.Color = image.NewUniform(construct.TitleTextColor)
		return fg
	})

	dragBarWidth := float64(screenWidth)

	totalButtonsSize := construct.ButtonWidth * float64(len(construct.Buttons))
	dragBarWidth -= totalButtonsSize

	hdr := &TitleBar{
		lastPressAt:        time.Now(),
		buttons:            make(map[Button]*entities.Entity),
		startingDimensions: intgeom.Point2{screenWidth, screenHeight},
	}

	for i, button := range construct.Buttons {
		i := i
		button := button
		var r render.Modifiable = render.NewColorBox(int(construct.ButtonWidth), int(construct.Height), construct.Color)
		txt := strconv.Itoa(i)
		var clickBinding = func(_ *entities.Entity, _ *mouse.Event) event.Response {
			return 0
		}

		switch button {
		case ButtonMinimize:
			r = render.NewSwitch("nohover", map[string]render.Modifiable{
				"nohover": spriteFromShape(minimizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.Color),
				"hover":   spriteFromShape(minimizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.HighlightColor),
				"onpress": spriteFromShape(minimizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.MouseDownColor),
			})
			txt = ""
			clickBinding = func(_ *entities.Entity, _ *mouse.Event) event.Response {
				// TODO: restore minimize functionality
				//ctx.Window.
				return 0
			}
		case ButtonClose:
			r = render.NewSwitch("nohover", map[string]render.Modifiable{
				"nohover": spriteFromShape(closeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.Color),
				"hover":   spriteFromShape(closeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.HighlightColor),
				"onpress": spriteFromShape(closeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.MouseDownColor),
			})
			txt = ""
			clickBinding = func(_ *entities.Entity, _ *mouse.Event) event.Response {
				ctx.Window.Quit()
				return 0
			}
		case ButtonMaximize:
			r = render.NewSwitch("nohover", map[string]render.Modifiable{
				"nohover":        spriteFromShape(maximizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.Color),
				"hover":          spriteFromShape(maximizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.HighlightColor),
				"onpress":        spriteFromShape(maximizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.MouseDownColor),
				"nohover-revert": spriteFromShape(normalizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.Color),
				"hover-revert":   spriteFromShape(normalizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.HighlightColor),
				"onpress-revert": spriteFromShape(normalizeIcon, int(construct.ButtonWidth), int(construct.Height), color.RGBA{255, 255, 255, 255}, construct.MouseDownColor),
			})
			txt = ""

			clickBinding = func(e *entities.Entity, _ *mouse.Event) event.Response {
				hdr.maximized = toggleMaximize(ctx, e)
				return 0
			}
		}
		hdr.buttons[button] = btn.New(ctx,
			btn.Text(txt),
			btn.Pos(dragBarWidth+float64(i)*construct.ButtonWidth, 0),
			btn.Renderable(r),
			btn.Height(construct.Height),
			btn.Width(construct.ButtonWidth),
			btn.Layers(construct.Layers...),
			btn.Binding(mouse.Start, func(e *entities.Entity, _ *mouse.Event) event.Response {
				if sw, ok := e.Renderable.(*render.Switch); ok {
					suffix, _ := e.Metadata("switch-suffix")
					sw.Set("hover" + suffix)
				}
				return 0
			}),
			btn.Binding(mouse.Stop, func(e *entities.Entity, _ *mouse.Event) event.Response {
				if sw, ok := e.Renderable.(*render.Switch); ok {
					suffix, _ := e.Metadata("switch-suffix")
					sw.Set("nohover" + suffix)
				}
				return 0
			}),
			btn.Binding(mouse.PressOn, func(e *entities.Entity, _ *mouse.Event) event.Response {
				if sw, ok := e.Renderable.(*render.Switch); ok {
					suffix, _ := e.Metadata("switch-suffix")
					sw.Set("onpress" + suffix)
				}
				return 0
			}),
			btn.Click(clickBinding),
			btn.Binding(oak.ViewportUpdate, func(e *entities.Entity, pt intgeom.Point2) event.Response {
				e.SetPos(floatgeom.Point2{float64(pt.X()) - totalButtonsSize + float64(i)*construct.ButtonWidth, 0})
				return 0
			}),
		)
	}

	btn.New(ctx,
		btn.Font(font),
		btn.Text(construct.Title),
		btn.TxtOff(10, construct.Height/2-float64(construct.TitleFontSize)/2),
		btn.Layers(construct.Layers...),
		btn.Width(dragBarWidth),
		btn.Height(construct.Height),
		btn.Color(construct.Color),
		btn.Binding(mouse.PressOn, func(_ *entities.Entity, _ *mouse.Event) event.Response {
			if time.Since(hdr.lastPressAt) < construct.DoubleClickThreshold {
				if mxbtn, ok := hdr.buttons[ButtonMaximize]; ok {
					hdr.maximized = toggleMaximize(ctx, mxbtn)
				}
				// if this is not set, dragging can persist after the window shrinks
				hdr.draggingWindow = false
				return 0
			}
			hdr.lastPressAt = time.Now()
			hdr.draggingWindow = true
			x, y := oak.GetCursorPosition()
			hdr.draggingStartPos = floatgeom.Point2{float64(x), float64(y)}
			return 0
		}),
		// Q: Why not mouse.Drag?
		// A: mouse.Drag is only triggered for on-screen mouse events. If the mouse
		//    falls out of the window, as it likely will if you drag the window up,
		//    the window will freeze until you bring the mouse cursor back into the window.
		btn.Binding(event.Enter, func(_ *entities.Entity, _ event.EnterPayload) event.Response {
			if hdr.draggingWindow {
				x, y := oak.GetCursorPosition()
				pt := floatgeom.Point2{float64(x), float64(y)}
				delta := pt.Sub(hdr.draggingStartPos)
				if delta == (floatgeom.Point2{}) {
					return 0
				}
				newX := hdr.DesktopPosition.X()
				newY := hdr.DesktopPosition.Y()
				newX += delta.X()
				newY += delta.Y()
				hdr.DesktopPosition = floatgeom.Point2{newX, newY}
				if hdr.maximized {
					if mxbtn, ok := hdr.buttons[ButtonMaximize]; ok {
						hdr.maximized = toggleMaximize(ctx, mxbtn)
					}
				}
				ctx.Window.MoveWindow(int(newX), int(newY), screenWidth, screenHeight)
				if !floatgeom.NewRect2WH(0, 0, float64(screenWidth), float64(screenHeight)).Contains(hdr.draggingStartPos) {
					hdr.draggingStartPos = floatgeom.Point2{
						float64(screenWidth) / 2, 16,
					}
				}
			}
			return 0
		}),
		btn.Binding(mouse.Release, func(_ *entities.Entity, _ *mouse.Event) event.Response {
			if hdr.draggingWindow {
				hdr.draggingWindow = false
			}
			return 0
		}),
		btn.Binding(oak.ViewportUpdate, func(e *entities.Entity, pt intgeom.Point2) event.Response {
			ctx.Window.(*oak.Window).UpdateViewSize(pt.X(), pt.Y())
			b := e.Children[0]
			newW := float64(pt.X()) - totalButtonsSize
			b.Renderable.Undraw()
			b.Renderable = render.NewColorBox(int(newW), int(construct.Height), construct.Color)
			ctx.DrawStack.Draw(b.Renderable, construct.Layers...)
			ctx.MouseTree.UpdateSpace(0, 0, newW, construct.Height, b.Space)
			return 0
		}),
	)
	return hdr
}

var closeIcon = shape.JustIn(shape.AndIn(
	shape.XRange(.35, .65),
	func(x, y int, sizes ...int) bool {
		size := sizes[0]
		return x == y || y == (size-x)
	},
))

var minimizeIcon = shape.JustIn(shape.AndIn(
	shape.XRange(.35, .65),
	func(x, y int, sizes ...int) bool {
		return y == sizes[0]/2
	},
))

var maximizeIcon = shape.JustIn(squarePercent(.35, .65))

var normalizeIcon = shape.JustIn(shape.OrIn(
	squarePercent(.35, .65),
	squarePercent(.45, .55),
))

func squarePercent(minPerc, maxPerc float64) shape.In {
	return shape.AndIn(
		shape.XRange(minPerc-.03, maxPerc),
		func(x, y int, sizes ...int) bool {
			yf := float64(y)
			sf := float64(sizes[0])
			return (yf >= sf*(minPerc-.03)) && (yf <= sf*maxPerc)
		},
		func(x, y int, sizes ...int) bool {
			size := sizes[0]
			return x == int(float64(size)*minPerc) ||
				x == int(float64(size)*maxPerc) ||
				y == int(float64(size)*minPerc) ||
				y == int(float64(size)*maxPerc)
		},
	)
}

func toggleMaximize(ctx *scene.Context, b *entities.Entity) bool {
	// TODO
	panic("unimplemented")
	// if sfx, _ := b.Metadata("switch-suffix"); sfx != "" {
	// 	ctx.Window.(*oak.Window).Normalize()
	// 	b.SetMetadata("switch-suffix", "")
	// 	if sw, ok := b.GetRenderable().(*render.Switch); ok {
	// 		sw.Set("nohover")
	// 	}
	// 	return false
	// }
	// ctx.Window.(*oak.Window).Maximize()
	// b.SetMetadata("switch-suffix", "-revert")
	// if sw, ok := b.GetRenderable().(*render.Switch); ok {
	// 	sw.Set("nohover-revert")
	// }
	// return true
}

func spriteFromShape(sh shape.Shape, w, h int, on, off color.Color) *render.Sprite {
	rect := sh.Rect(w, h)
	rgba := image.NewRGBA(image.Rect(0, 0, len(rect), len(rect[0])))
	sp := render.NewSprite(0, 0, rgba)
	for x := 0; x < len(rect); x++ {
		for y := 0; y < len(rect[0]); y++ {
			if rect[x][y] {
				sp.Set(x, y, on)
			} else {
				sp.Set(x, y, off)
			}
		}
	}
	return sp
}
