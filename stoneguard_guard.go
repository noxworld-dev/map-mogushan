package mogushan

import (
	"fmt"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/damage"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
)

type GuardColor int

func (c GuardColor) String() string {
	switch c {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	}
	return fmt.Sprintf("GuardColor(%d)", int(c))
}

const (
	Red = GuardColor(iota)
	Green
	Blue
)

func (s *StoneGuards) NewGuard(color GuardColor) *Guard {
	g := &Guard{s: s, color: color}
	g.obj = ns4.CreateObject("Troll", startPos[color])
	g.obj.SetMaxHealth(BossHealth)
	g.prevPos = startPos[color]
	g.prevHP = g.obj.CurrentHealth()
	switch color {
	case Red:
		g.obj.Enchant(enchant.PROTECT_FROM_FIRE, ns4.Infinite())
	case Green:
		g.obj.Enchant(enchant.PROTECT_FROM_POISON, ns4.Infinite())
	case Blue:
		g.obj.Enchant(enchant.PROTECT_FROM_ELECTRICITY, ns4.Infinite())
	}
	g.obj.Enchant(enchant.FREEZE, ns4.Infinite())
	g.obj.Freeze(true)
	g.obj.LookWithAngle(32)
	g.obj.AggressionLevel(0) // TODO: remove
	// obj.SetBaseSpeed(5)
	//fmt.Printf("speed: %v\n", obj.BaseSpeed())
	g.obj.SetMass(20)
	//fmt.Printf("mass: %v\n", g.obj.Mass())
	s.bosses = append(s.bosses, g)
	return g
}

type Guard struct {
	s       *StoneGuards
	color   GuardColor
	obj     ns4.Obj
	hp      *HealthBar
	ep      *EnergyBar
	prevHP  int
	prevPos ns4.Pointf
	frame   int

	hitByDeathBall bool

	energy     int
	forceField ns4.Obj

	red abilitiesRed
}

func (g *Guard) Delete() {
	if g.hp != nil {
		g.hp.Delete()
		g.hp = nil
	}
	if g.ep != nil {
		g.ep.Delete()
		g.ep = nil
	}
	if g.forceField != nil {
		g.forceField.Delete()
		g.forceField = nil
	}
	g.red.Delete()
	g.obj.Delete()
}

func (g *Guard) Start() {
	g.hp = NewHealthBar(g.obj)
	g.ep = NewEnergyBar(g.obj)
	g.obj.Freeze(false)
	g.obj.EnchantOff(enchant.FREEZE)
	ns4.CastSpell(spell.COUNTERSPELL, g.obj, g.obj)
}

func (g *Guard) HealthDelta() int {
	return g.obj.CurrentHealth() - g.prevHP
}

func (g *Guard) Update() {
	if g.hp != nil {
		g.hp.Update()
	}
	if g.ep != nil {
		g.ep.Update()
	}
	g.antiSpell()
	g.maybeEnableForceField()
	g.updateAbility()
	g.prevPos = g.obj.Pos()
	g.prevHP = g.obj.CurrentHealth()
	g.frame++
}

func (g *Guard) removeFlames() {
	if g.obj.Pos() == g.prevPos {
		flames := ns4.FindObjects(nil, ns4.InCirclef{Center: g.obj, R: 5}, ns4.HasTypeName{
			"SmallFlame",
			"MediumFlame",
		})
		if flames >= 2 {
			barrel := ns4.CreateObject("WaterBarrel", g.obj.Pos())
			barrel.Damage(g.obj, 100, 1)
		}
	}
}

func (g *Guard) removeDeathBall() {
	dHP := -g.HealthDelta()
	if dHP >= 10 {
		if g.hitByDeathBall {
			g.obj.Enchant(enchant.INVULNERABLE, ns4.Seconds(1))
			ns4.CastSpell(spell.COUNTERSPELL, g.obj, g.obj)
		} else {
			g.hitByDeathBall = true
		}
	} else {
		g.hitByDeathBall = false
	}
	// TODO: fix DeathBall search

	// deaths := 0
	// ns4.FindObjects(
	// 	func(it ns4.Obj) bool {
	// 		deaths++
	// 		if !g.hitByDeathBall {
	// 			g.hitByDeathBall = true
	// 			return false
	// 		}
	// 		it.Delete()
	// 		return true
	// 	},
	// 	ns4.InCirclef{Center: g.obj, R: 100},
	// 	ns4.HasTypeName{"DeathBall"},
	// )
	// if deaths == 0 {
	// 	g.hitByDeathBall = false
	// }
}

func (g *Guard) antiSpell() {
	g.removeFlames()
	g.removeDeathBall()
	if g.obj.HasEnchant(enchant.CHARMING) {
		ns4.CastSpell(spell.COUNTERSPELL, g.obj, g.obj)
	}
}

func (g *Guard) maybeEnableForceField() {
	if g.frame < 4*30 { // 4 sec
		return
	}
	const fullChargeDur = 50 // sec
	hasAnother := false
	for _, boss := range g.s.bosses {
		if boss == g {
			continue
		}
		if g.obj.Pos().Sub(boss.obj.Pos()).Len() < 138 {
			hasAnother = true
		}
	}
	if hasAnother {
		if g.forceField != nil {
			g.forceField.Delete()
			g.forceField = nil
		}
		if g.frame%30 == 0 {
			g.energy++
		}
		if g.energy > fullChargeDur {
			g.triggerExplosion()
			g.energy = 0
		}
		g.ep.Set(float32(g.energy) / float32(fullChargeDur))
	} else {
		g.obj.Enchant(enchant.INVULNERABLE, ns4.Frames(1))
		if g.forceField == nil {
			g.forceField = ns4.CreateObject("MagicEnergy", g.obj)
		}
		g.forceField.SetPos(g.obj.Pos())
	}
}

// TODO: set unique damage type?
var damages = []damage.Type{
	int(Red):   damage.ZAP_RAY,
	int(Green): damage.ZAP_RAY,
	int(Blue):  damage.ZAP_RAY,
}

func (g *Guard) triggerExplosion() {
	dmg := 20 // if doesn't match
	if g.color == g.s.curEffect {
		dmg = 2 // if matches
		fmt.Println("Guard triggered effect!")
		g.s.resetEffect()
	}
	fmt.Printf("Guard %s dealing damage: %d\n",
		g.color.String(), dmg)
	typ := damages[g.color]
	g.s.allPlayersInRange(func(u ns4.Obj) {
		u.Damage(nil, dmg, typ)
	})
}

func (g *Guard) updateAbility() {
	switch g.color {
	case Red:
		g.red.Update(g)
	case Green:
		g.updateAbilityGreen()
	case Blue:
		g.updateAbilityBlue()
	}
}

func (g *Guard) updateAbilityGreen() {
}

func (g *Guard) updateAbilityBlue() {
}
