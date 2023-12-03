package stoneguard

import (
	"math"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/damage"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
)

// BlueAbility is an ability for Blue color/element.
// It stores all active spells of this kind for a given boss unit.
type BlueAbility struct {
	frame    int
	charge   int
	active   []*blueSpell
	notFirst bool
}

// Delete all active Blue spells.
func (g *BlueAbility) Delete() {
	for _, a := range g.active {
		a.Delete()
	}
	g.active = nil
}

// Update all Blue spells for a Guard, starting new ones and removing ended ones.
func (g *BlueAbility) Update(b *Guard) {
	g.frame++
	if g.frame < BlueAfter*ns4.FrameRate() {
		return
	}
	// delete stopped abilities
	for i := 0; i < len(g.active); i++ {
		if g.active[i].stop {
			g.active[i].Delete()
			g.active = append(g.active[:i], g.active[i+1:]...)
			i--
		}
	}

	// update active abilities
	for _, a := range g.active {
		a.Update(b)
	}

	// charge the ability for this number of frames
	g.charge++
	if g.notFirst && g.charge < BlueCooldown*ns4.FrameRate() {
		return // not charged yet
	}
	// ability charged - create new spell and reset charge
	g.charge = 0

	// pick random player in boss room
	var players []ns4.Obj
	b.s.EachPlayerInRoom(func(u ns4.Obj) {
		players = append(players, u)
	})
	if len(players) == 0 {
		return // no players in room
	}
	ind := ns4.Random(0, len(players)-1)
	targ := players[ind].Pos()
	// add new spell to active ones
	g.active = append(g.active, &blueSpell{
		target: targ,
	})
	g.notFirst = true
}

// blueSpell stores state of a single Blue spell.
type blueSpell struct {
	target ns4.Pointf
	frame  int
	outer  ns4.Objects // outer orbs
	inner  ns4.Objects // inner orbs
	flame  ns4.Obj     // central flame
	stop   bool
}

// Delete a single Blue spell.
func (g *blueSpell) Delete() {
	g.outer.Delete()
	g.outer = nil
	g.inner.Delete()
	g.inner = nil
	if g.flame != nil {
		g.flame.Delete()
		g.flame = nil
	}
}

// Update runs logic for a single Blue spell for a Guard.
func (g *blueSpell) Update(b *Guard) {
	if g.stop {
		return
	}
	g.frame++
	boss, targ := b.unit, g.target
	if g.frame < BlueCharge*ns4.FrameRate() {
		ns4.Effect(effect.LIGHTNING, boss, targ)
		return
	}
	if g.flame == nil {
		g.flame = ns4.CreateObject(BlueDangerModel, boss)
		g.flame.SetOwner(boss)
		g.flame.SetPos(targ)
		for i := 0; i < BlueOuterCnt; i++ {
			o := ns4.CreateObject(BlueOuterModel, boss)
			o.SetOwner(boss)
			o.SetPos(targ)
			g.outer = append(g.outer, o)
		}
		for i := 0; i < BlueInnerCnt; i++ {
			o := ns4.CreateObject(BlueInnerModel, boss)
			o.SetOwner(boss)
			o.SetPos(targ)
			g.inner = append(g.inner, o)
		}
	}

	hit := false
	blueHit := false
	b.s.EachPlayerInRoom(func(u ns4.Obj) {
		d := u.Pos().Sub(targ).Len()
		if !hit && d < BlueInnerR {
			hit = true
			if b.color == b.s.curEffect {
				u.Damage(nil, BlueInnerDamageWeak, damage.ELECTRIC)
				u.Enchant(enchant.HELD, ns4.Seconds(BlueInnerStunWeak))
			} else {
				u.Damage(nil, BlueInnerDamage, damage.ELECTRIC)
				u.Enchant(enchant.HELD, ns4.Seconds(BlueInnerStun))
			}
		} else if d < BlueOuterR {
			if b.color != b.s.curEffect {
				u.Damage(nil, BlueOuterDamage, damage.ELECTRIC)
			}
			ns4.Effect(effect.LIGHTNING, targ, u.Pos())
			blueHit = true
		}
	})
	if hit {
		g.stop = true
		return
	}

	if b.color == b.s.curEffect {
		g.flame.Enable(false)
	} else {
		g.flame.Enable(true)
	}
	light := ns4.Random(0, len(g.outer)-1)
	for i, o := range g.outer {
		ph := float64(i)*2*math.Pi/float64(len(g.outer)) + float64(g.frame)*BlueOuterSpeed
		dx, dy := float32(BlueOuterR*math.Cos(ph)), float32(BlueOuterR*math.Sin(ph))
		pos := targ.Add(ns4.Ptf(dx, dy))
		o.SetPos(pos)
		if !blueHit && g.frame%4 == 0 && i == light {
			ns4.Effect(effect.LIGHTNING, targ, pos)
		}
	}
	for i, o := range g.inner {
		ph := float64(i)*2*math.Pi/float64(len(g.inner)) + float64(g.frame)*BlueInnerSpeed
		dx, dy := float32(BlueInnerR*math.Cos(ph)), float32(BlueInnerR*math.Sin(ph))
		o.SetPos(targ.Add(ns4.Ptf(dx, dy)))
	}
}
