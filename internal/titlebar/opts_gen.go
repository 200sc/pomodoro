// Code generated by foptgen; DO NOT EDIT.

package titlebar

import (
	"image/color"
	"time"
)

type Option func(Constructor) Constructor

func WithColor(v color.Color) Option {
	return func(s Constructor) Constructor {
		s.Color = v
		return s
	}
}

func WithHighlightColor(v color.Color) Option {
	return func(s Constructor) Constructor {
		s.HighlightColor = v
		return s
	}
}

func WithMouseDownColor(v color.Color) Option {
	return func(s Constructor) Constructor {
		s.MouseDownColor = v
		return s
	}
}

func WithHeight(v float64) Option {
	return func(s Constructor) Constructor {
		s.Height = v
		return s
	}
}

func WithLayers(v []int) Option {
	return func(s Constructor) Constructor {
		s.Layers = v
		return s
	}
}

func WithTitle(v string) Option {
	return func(s Constructor) Constructor {
		s.Title = v
		return s
	}
}

func WithTitleFontSize(v int) Option {
	return func(s Constructor) Constructor {
		s.TitleFontSize = v
		return s
	}
}

func WithTitleXOffset(v int) Option {
	return func(s Constructor) Constructor {
		s.TitleXOffset = v
		return s
	}
}

func WithTitleTextColor(v color.Color) Option {
	return func(s Constructor) Constructor {
		s.TitleTextColor = v
		return s
	}
}

func WithButtons(v []Button) Option {
	return func(s Constructor) Constructor {
		s.Buttons = v
		return s
	}
}

func WithButtonWidth(v float64) Option {
	return func(s Constructor) Constructor {
		s.ButtonWidth = v
		return s
	}
}

func WithDoubleClickThreshold(v time.Duration) Option {
	return func(s Constructor) Constructor {
		s.DoubleClickThreshold = v
		return s
	}
}