package ui

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
)

func NewHealthBar(obj ns4.Obj) *HealthBar {
	hp := &HealthBar{targ: obj}
	hp.left = ns4.CreateObject("HealOrb", obj)
	hp.mover = ns4.CreateObject("HealOrb", obj)
	hp.cur = ns4.CreateObject("HealOrb", obj)
	hp.right = ns4.CreateObject("DrainManaOrb", obj)
	hp.Update()
	return hp
}

type HealthBar struct {
	targ     ns4.Obj
	moverPos float32
	left     ns4.Obj
	mover    ns4.Obj
	cur      ns4.Obj
	right    ns4.Obj
}

func (hp *HealthBar) Delete() {
	hp.left.Delete()
	hp.mover.Delete()
	hp.cur.Delete()
	hp.right.Delete()
}

func (hp *HealthBar) Update() {
	const (
		offs  = 46
		width = 56
	)
	pos := hp.targ.Pos()
	cur, max := hp.targ.CurrentHealth(), hp.targ.MaxHealth()
	perc := float32(cur) / float32(max)
	speed := float32(1)
	hp.moverPos += speed
	moveWidth := perc * width
	movePerc := float32(int(hp.moverPos)%width) / width
	y := pos.Y - offs
	hp.left.SetPos(ns4.Pointf{X: pos.X - width/2, Y: y})
	hp.mover.SetPos(ns4.Pointf{X: pos.X - width/2 + movePerc*moveWidth, Y: y})
	hp.cur.SetPos(ns4.Pointf{X: pos.X - width/2 + perc*width, Y: y})
	hp.right.SetPos(ns4.Pointf{X: pos.X + width/2, Y: y})
}
