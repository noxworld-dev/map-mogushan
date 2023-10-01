package stoneguard

import (
	"fmt"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
	"github.com/noxworld-dev/opennox-lib/types"

	"mogushan/ui"
)

// NewGuard creates a new guard boss unit with a given color and position.
func (s *State) NewGuard(color Element, pos types.Pointf) *Guard {
	fmt.Printf("NewGuard: %s (%d), ench=%v, room=%v; Red=%d, Green=%d, Blue=%d\n",
		color.String(), int(color), color.Enchant(), color.RoomEffect(),
		int(Red), int(Green), int(Blue))
	g := &Guard{s: s, color: color}
	// Create an actual boss unit and set it up.
	g.unit = ns4.CreateObject(BossModel, pos)
	g.prevPos = pos
	g.unit.LookWithAngle(32)
	// We set the health to the value of the common health pool.
	// Individual unit health be adjusted separately for each unit by the script, so that it's shared.
	g.unit.SetMaxHealth(BossHealth)
	g.prevHP = g.unit.CurrentHealth()
	// Set ability and enchant based on color/element.
	g.abil = color.NewAbility()
	g.unit.Enchant(color.Enchant(), ns4.Infinite())
	// Freeze the boss initially.
	g.unit.Enchant(enchant.FREEZE, ns4.Infinite())
	g.unit.Freeze(true)
	// Set other unit parameters.
	if Debug {
		// Disable aggression when testing the map.
		g.unit.AggressionLevel(0)
	}
	//fmt.Printf("speed: %v\n", obj.BaseSpeed())
	g.unit.SetBaseSpeed(BossSpeed)
	//fmt.Printf("mass: %v\n", g.obj.Mass())
	g.unit.SetMass(BossMass)
	// Add boss to the list.
	s.bosses = append(s.bosses, g)
	return g
}

// Guard contains state for one Stone Guard boss unit.
type Guard struct {
	s       *State
	color   Element
	unit    ns4.Obj
	hp      *ui.HealthBar
	ep      *ui.EnergyBar
	prevHP  int
	prevPos ns4.Pointf
	frame   int

	hitByDeathBall bool

	energy     int
	forceField ns4.Obj

	abil Ability
}

// Delete the unit and all its state.
func (g *Guard) Delete() {
	// delete health and energy meters
	if g.hp != nil {
		g.hp.Delete()
		g.hp = nil
	}
	if g.ep != nil {
		g.ep.Delete()
		g.ep = nil
	}
	// delete force field
	if g.forceField != nil {
		g.forceField.Delete()
		g.forceField = nil
	}
	// delete ability
	if g.abil != nil {
		g.abil.Delete()
		g.abil = nil
	}
	// finally, delete the actual unit
	g.unit.Delete()
}

// Start unfreezes the boss and makes it start fighting.
func (g *Guard) Start() {
	g.hp = ui.NewHealthBar(g.unit)
	g.ep = ui.NewEnergyBar(g.unit)
	g.unit.Freeze(false)
	g.unit.EnchantOff(enchant.FREEZE)
	ns4.CastSpell(spell.COUNTERSPELL, g.unit, g.unit)
}

// HealthDelta calculates the heal/damage delta for the current frame.
func (g *Guard) HealthDelta() int {
	return g.unit.CurrentHealth() - g.prevHP
}

// Update runs the main boss logic for a specific unit.
func (g *Guard) Update() {
	// show health and energy bars
	if g.hp != nil {
		g.hp.Update()
	}
	if g.ep != nil {
		g.ep.Update()
	}
	// prevent certain spells that are abusive
	g.antiSpell()
	// run the force field and energy logic
	g.gatherEnergyOrShield()
	// run the unique ability for the unit (if any)
	if g.abil != nil {
		g.abil.Update(g)
	}
	// some bookkeeping
	g.prevPos = g.unit.Pos()
	g.prevHP = g.unit.CurrentHealth()
	g.frame++
}

// antiSpell prevents certain abusive spells from affecting the boss unit.
func (g *Guard) antiSpell() {
	g.removeFlames()
	g.removeDeathBall()
	g.stopCharming()
}

// removeFlames removes the flames under the boss if it's standing in them for too long.
func (g *Guard) removeFlames() {
	// If boss is standing on the same spot as the last frame.
	if g.unit.Pos() == g.prevPos {
		// Count the player-owned flames under it.
		flames := ns4.FindObjects(nil,
			// Check in certain radius, usually corresponding to the unit model size.
			ns4.InCirclef{Center: g.unit, R: BossFlamesR},
			// We are only interested in flames.
			ns4.HasTypeName{
				"SmallFlame",
				"MediumFlame",
				"Flame",
				"LargeFlame",
			},
			// Only consider player-owned flames, since boss itself may use flame for the unique abilities.
			ns4.ObjCondFunc(func(obj ns4.Obj) bool {
				playerOwn := false
				for _, pl := range ns4.Players() {
					u := pl.Unit()
					if obj.HasOwner(u) {
						playerOwn = true
					}
				}
				return playerOwn
			},
			))
		// If there are too many flames under it - trigger a breaking water barrel to put them out.
		if flames >= BossFlamesCnt {
			barrel := ns4.CreateObject("WaterBarrel", g.unit.Pos())
			barrel.Damage(g.unit, 100, 1)
		}
	}
}

// removeDeathBall removes a Death Ball (Force of Nature projectile) from the boss if it damage it too much.
func (g *Guard) removeDeathBall() {
	dHP := -g.HealthDelta()
	// If damage crosses a threshold for 2 frames in a row - counter-spell the ball and make unit temporarily invulnerable.
	if dHP >= 10 {
		if g.hitByDeathBall {
			g.unit.Enchant(enchant.INVULNERABLE, ns4.Seconds(1))
			ns4.CastSpell(spell.COUNTERSPELL, g.unit, g.unit)
		} else {
			g.hitByDeathBall = true
		}
	} else {
		g.hitByDeathBall = false
	}
	// TODO: fix DeathBall search in NS4

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

// stopCharming stops charming attempts toward the boss.
func (g *Guard) stopCharming() {
	if g.unit.HasEnchant(enchant.CHARMING) {
		ns4.CastSpell(spell.COUNTERSPELL, g.unit, g.unit)
	}
}

// gatherEnergyOrShield is responsible for boss energy logic and the force field.
func (g *Guard) gatherEnergyOrShield() {
	if g.frame < EnergyDelay*ns4.FrameRate() {
		return
	}
	// If there's at least a second boss unit around - these units will gather energy.
	hasAnother := false
	for _, boss := range g.s.bosses {
		if boss == g {
			continue
		}
		if g.unit.Pos().Sub(boss.unit.Pos()).Len() < EnergyDist {
			hasAnother = true
		}
	}
	if hasAnother {
		// Gather energy while at least one another boss is around.
		if g.forceField != nil {
			g.forceField.Delete()
			g.forceField = nil
		}
		// Energy is increased each second.
		if g.frame%ns4.FrameRate() == 0 {
			g.energy++
		}
		// When charged to 100% - trigger explosion.
		if g.energy > EnergyExplosionChargeDur {
			g.triggerExplosion()
			g.energy = 0
		}
		// Update energy bar on the unit.
		g.ep.Set(float32(g.energy) / float32(EnergyExplosionChargeDur))
	} else {
		// If no other boss is around - make unit invulnerable and show a force field.
		g.unit.Enchant(enchant.INVULNERABLE, ns4.Frames(1))
		if g.forceField == nil {
			g.forceField = ns4.CreateObject(EnergyShieldModel, g.unit)
		}
		g.forceField.SetPos(g.unit.Pos())
	}
}

// triggerExplosion creates an elemental explosion from the unit.
func (g *Guard) triggerExplosion() {
	var dmg int
	if g.color == g.s.curEffect {
		// If room effect matches the unit color/element - deal minor damage and switch room effect.
		dmg = EnergyExplosionDamageWeak
		fmt.Println("Guard triggered effect switch!")
		g.s.nextRoomEffect()
	} else {
		// If room effect doesn't match the unit color/element - deal major damage and keep the effect.
		// This will eventually allow the effect to timeout, confuse players and switch on its own.
		dmg = EnergyExplosionDamage
	}
	fmt.Printf("Guard %s dealing damage: %d\n", g.color.String(), dmg)
	typ := g.color.DamageType()
	g.s.EachPlayerInRoom(func(u ns4.Obj) {
		u.Damage(nil, dmg, typ)
	})
}
