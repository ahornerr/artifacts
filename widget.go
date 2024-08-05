package main

import (
	"github.com/gizak/termui/v3/widgets"
)

type CharacterWidget struct {
	Status   *widgets.Paragraph
	Action   *widgets.Paragraph
	Task     *widgets.Paragraph
	Cooldown *widgets.Gauge
	Levels   map[string]*widgets.Gauge
}

var characterWidgets = map[string]*CharacterWidget{}
