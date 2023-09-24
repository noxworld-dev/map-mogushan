package mogushan

import (
	"math"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/damage"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
)

// Blue ability balance.
const (
	// blueCooldown sets how frequently the boss will cast the ability.
	blueCooldown = 24 // sec
	// blueCharge sets how long it will take for ability to charge (lightning to circle).
	blueCharge = 4 // sec

	// blueBlueR sets a radius of blue circle.
	blueBlueR = 138
	// blueBlueCnt sets a number of blue orbs.
	blueBlueCnt = 10
	// blueBlueDamage sets per-frame damage of blue circle (as long as the player is in it).
	blueBlueDamage = 2

	// blueWhiteR is a radius of white circle.
	blueWhiteR = 46
	// blueWhiteCnt sets a number of blue orbs.
	blueWhiteCnt = 10
	// blueWhiteDamage sets damage done once to the player enters white circle (when zone effect doesn't match).
	blueWhiteDamage = 20
	// blueWhiteDamageWeak sets damage done once to the player enters white circle (when zone effect matches).
	blueWhiteDamageWeak = 2
	// blueWhiteStun sets stun duration when player enters white circle (when zone effect doesn't match).
	blueWhiteStun = 20 // sec
	// blueWhiteStunWeak sets stun duration when player enters white circle (when zone effect matches).
	blueWhiteStunWeak = 2 // sec
)

// abilitiesBlue stores all active Blue abilities.
type abilitiesBlue struct {
	charge int
	active []*abilityBlue
}

// Delete all Blue abilities.
func (g *abilitiesBlue) Delete() {
	for _, a := range g.active {
		a.Delete()
	}
	g.active = nil
}

// Update all Blue abilities, starting new ones and removing stopped ones.
func (g *abilitiesBlue) Update(b *Guard) {
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
	const max = blueCooldown * 30
	g.charge++
	if g.charge < max {
		return // not charged yet
	}
	// ability charged - create new one and reset charge
	g.charge = 0

	// pick random player in boss zone
	var players []ns4.Obj
	b.s.allPlayersInRange(func(u ns4.Obj) {
		players = append(players, u)
	})
	if len(players) == 0 {
		return // no players in zone
	}
	ind := ns4.Random(0, len(players)-1)
	targ := players[ind].Pos()
	// add new ability to active ones
	g.active = append(g.active, &abilityBlue{
		target: targ,
	})
}

// abilityBlue stores variables of a single Blue ability.
type abilityBlue struct {
	target ns4.Pointf
	frame  int
	blue   []ns4.Obj // blue orbs
	white  []ns4.Obj // white orbs
	flame  ns4.Obj   // central flame
	stop   bool
}

// Delete a single Blue ability.
func (g *abilityBlue) Delete() {
	for _, a := range g.blue {
		a.Delete()
	}
	for _, b := range g.white {
		b.Delete()
	}
	g.blue = nil
	g.white = nil
	if g.flame != nil {
		g.flame.Delete()
		g.flame = nil
	}
}

func (g *abilityBlue) Update(b *Guard) {
	if g.stop {
		return
	}
	g.frame++
	boss, targ := b.obj, g.target
	if g.frame < blueCharge*30 {
		ns4.Effect(effect.LIGHTNING, boss, targ)
		return
	}
	if g.flame == nil {
		g.flame = ns4.CreateObject("BlueFlame", boss)
		g.flame.SetOwner(boss)
		g.flame.SetPos(targ)
		for i := 0; i < blueBlueCnt; i++ {
			o := ns4.CreateObject("DrainManaOrb", boss)
			o.SetOwner(boss)
			o.SetPos(targ)
			g.blue = append(g.blue, o)
		}
		for i := 0; i < blueWhiteCnt; i++ {
			o := ns4.CreateObject("WhiteOrb", boss)
			o.SetOwner(boss)
			o.SetPos(targ)
			g.white = append(g.white, o)
		}
	}

	hit := false
	blueHit := false
	b.s.allPlayersInRange(func(u ns4.Obj) {
		d := u.Pos().Sub(targ).Len()
		if !hit && d < blueWhiteR {
			hit = true
			if b.color == b.s.curEffect {
				u.Damage(nil, blueWhiteDamageWeak, damage.ELECTRIC)
				u.Enchant(enchant.HELD, ns4.Seconds(blueWhiteStunWeak))
			} else {
				u.Damage(nil, blueWhiteDamage, damage.ELECTRIC)
				u.Enchant(enchant.HELD, ns4.Seconds(blueWhiteStun))
			}
		} else if d < blueBlueR {
			if b.color != b.s.curEffect {
				u.Damage(nil, blueBlueDamage, damage.ELECTRIC)
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
	light := ns4.Random(0, len(g.blue)-1)
	for i, o := range g.blue {
		ph := float64(i)*2*math.Pi/float64(len(g.blue)) + float64(g.frame)/20
		dx, dy := float32(blueBlueR*math.Cos(ph)), float32(blueBlueR*math.Sin(ph))
		pos := targ.Add(ns4.Ptf(dx, dy))
		o.SetPos(pos)
		if !blueHit && g.frame%4 == 0 && i == light {
			ns4.Effect(effect.LIGHTNING, targ, pos)
		}
	}
	for i, o := range g.white {
		ph := float64(i)*2*math.Pi/float64(len(g.white)) + float64(g.frame)/20
		dx, dy := float32(blueWhiteR*math.Cos(ph)), float32(blueWhiteR*math.Sin(ph))
		o.SetPos(targ.Add(ns4.Ptf(dx, dy)))
	}
}
