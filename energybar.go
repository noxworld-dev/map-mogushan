package mogushan

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
)

func NewEnergyBar(obj ns4.Obj) *EnergyBar {
	bar := &EnergyBar{targ: obj}
	for i := range bar.slots {
		bar.slots[i] = ns4.CreateObject("WhiteOrb", obj)
	}
	bar.Update()
	return bar
}

type EnergyBar struct {
	targ  ns4.Obj
	val   float32
	slots [10]ns4.Obj
}

func (bar *EnergyBar) Set(val float32) {
	bar.val = val
}

func (bar *EnergyBar) Delete() {
	for i, o := range bar.slots {
		o.Delete()
		bar.slots[i] = nil
	}
}

func (bar *EnergyBar) Update() {
	const (
		offs   = 46
		height = 56
	)
	pos := bar.targ.Pos()
	x := pos.X + offs
	for i := range bar.slots {
		hperc := float32(i) / float32(len(bar.slots))
		bar.slots[i].SetPos(ns4.Pointf{X: x, Y: pos.Y + height/2 - hperc*height})
	}
	perc := bar.val
	if perc < 0 {
		perc = 0
	}
	if perc > 1 {
		perc = 1
	}
	enabled := int(perc * float32(len(bar.slots)))
	for i := range bar.slots {
		if i > enabled {
			bar.slots[i].SetPos(ns4.Pointf{})
		}
	}
}
