package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/200sc/pomodoro/internal/titlebar"
	"github.com/oakmound/grove/components/textinput"
	oak "github.com/oakmound/oak/v4"
	"github.com/oakmound/oak/v4/alg/floatgeom"
	"github.com/oakmound/oak/v4/audio"
	"github.com/oakmound/oak/v4/audio/synth"
	"github.com/oakmound/oak/v4/entities"
	"github.com/oakmound/oak/v4/entities/x/btn"
	"github.com/oakmound/oak/v4/event"
	"github.com/oakmound/oak/v4/mouse"
	"github.com/oakmound/oak/v4/render"
	"github.com/oakmound/oak/v4/scene"
	"golang.org/x/image/colornames"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	toggleCountdown := event.RegisterEvent[time.Duration]()

	doneSound := synth.Int16.Sin(synth.AtPitch(synth.A2), synth.Duration(2*time.Second))
	initPos := floatgeom.Point2{100, 100}

	oak.AddScene("pomodoro", scene.Scene{
		Start: func(ctx *scene.Context) {

			err := audio.InitDefault()
			if err != nil {
				panic("failed to initialize audio: " + err.Error())
			}

			tb := titlebar.New(ctx,
				titlebar.WithTitle("PMDR"),
				titlebar.WithButtons([]titlebar.Button{titlebar.ButtonClose}),
				titlebar.WithLayers([]int{0, 1}),
			)
			tb.DesktopPosition = initPos

			duration := time.Minute * 15
			var remaining time.Duration
			errtext := ""

			fnt := render.DefaultFont()
			ctx.Draw(fnt.NewStringerText(&remaining, 10, 62), 1)
			ctx.Draw(fnt.NewStrPtrText(&errtext, 10, 82), 1)

			input := textinput.New(ctx, textinput.And(
				textinput.WithFont(fnt),
				textinput.WithDims(80, 20),
				textinput.WithPlaceholder("duration"),
				textinput.WithFinalizer(func(s string) {
					d, err := time.ParseDuration(s)
					if err != nil {
						errtext = "invalid duration"
						ctx.DoAfter(3*time.Second, func() {
							errtext = ""
						})
						return
					}
					duration = d
				}),
				textinput.WithPosition(10, 42),
			))
			ctx.Draw(input.Renderable, 1)

			startText := "Start"
			btn.New(ctx,
				btn.Color(colornames.Green),
				btn.Font(fnt),
				btn.TextPtr(&startText),
				btn.Width(40),
				btn.Height(14),
				btn.Pos(100, 87),
				btn.Click(func(e1 *entities.Entity, e2 *mouse.Event) event.Response {
					if startText == "Start" {
						startText = "Stop"
					} else {
						startText = "Start"
					}
					<-event.TriggerOn(ctx, toggleCountdown, duration)
					return event.ResponseNone
				}),
			)

			var t *time.Ticker
			var tCancel func()
			event.GlobalBind(ctx, toggleCountdown, func(d time.Duration) event.Response {
				if t == nil {
					t = time.NewTicker(1 * time.Second)
				} else {
					tCancel()
					return event.ResponseNone
				}
				var tCtx context.Context
				remaining = d
				tCtx, tCancel = context.WithTimeout(ctx, d)
				go func() {
					for {
						select {
						case <-t.C:
							remaining -= time.Second
						case <-tCtx.Done():
							t.Stop()
							remaining = 0
							t = nil
							playCtx, _ := context.WithTimeout(ctx, 2*time.Second)
							audio.Play(playCtx, doneSound)
							return
						}
					}
				}()
				return event.ResponseNone
			})
		},
	})
	return oak.Init("pomodoro", func(c oak.Config) (oak.Config, error) {
		c.Screen.Width = 150
		c.Screen.Height = 102
		c.TopMost = true
		c.Borderless = true
		c.Screen.X = int(initPos.X())
		c.Screen.Y = int(initPos.Y())
		c.Title = "PMDR"
		return c, nil
	})
}
