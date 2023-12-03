package stoneguard

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
	"github.com/noxworld-dev/opennox-lib/types"
)

var (
	urchinPos = [6]types.Pointf{
		{4887, 4979},
		{4979, 4887},
		{4841, 5025},
		{5025, 4841},
		{4795, 5071},
		{5071, 4795},
	}
	urchinBossPos = types.Pointf{4933, 4933}
	demoAxisStart = types.Pointf{4910, 4910}
)

const (
	// demoLength is a length or the demo room, starting from demoAxisStart, and following a diagonal.
	demoLength = 217
	// demoWidth is a width of a demo room, starting from demoAxisStart as a center and perpendicular to diagonal.
	demoWidth = 368
)

// demoState contains all state of the demo scene
var demoState DemoState

func init() {
	// register map events
	ns4.OnMapEvent(ns4.MapInitialize, demoState.Reset)
	ns4.OnFrame(demoState.Update)
}

type DemoStatus int

const (
	DemoWaiting = DemoStatus(iota)
	DemoEffect
	DemoBoss
	DemoEnd
)

type DemoState struct {
	urchins ns4.Objects
	boss    ns4.Obj
	shield  ns4.Obj
	status  DemoStatus
	effect  Element
	frame   int
}

func (d *DemoState) Delete() {
	d.urchins.Delete()
	d.urchins = nil
	if d.boss != nil {
		d.boss.Delete()
		d.boss = nil
	}
	if d.shield != nil {
		d.shield.Delete()
		d.shield = nil
	}
}

func (d *DemoState) Reset() {
	d.Delete()
	d.status = DemoWaiting
	d.frame = -1
	for _, pos := range urchinPos {
		obj := ns4.CreateObject("Urchin", pos)
		obj.LookWithAngle(32)
		d.urchins = append(d.urchins, obj)
	}

	d.boss = ns4.CreateObject("Urchin", urchinBossPos)
	d.boss.AggressionLevel(0)
	d.boss.Enchant(enchant.INVULNERABLE, ns4.Infinite())

	d.shield = ns4.CreateObject(EnergyShieldModel, urchinBossPos)
	d.shield.Freeze(true)
}

func (d *DemoState) Update() {
	switch d.status {
	case DemoWaiting: // not started, check player coords
		hit := false
		for _, pl := range ns4.Players() {
			u := pl.Unit()
			pos := u.Pos()
			if pos.X+pos.Y < 10280 {
				hit = true
				break
			}
		}
		if !hit {
			return
		}
		d.startEffect()
	case DemoEffect:
		d.frame++
		d.updateEffect()
	case DemoBoss:
		d.frame++
		d.updateBoss()
	}
}

func (d *DemoState) startEffect() {
	d.status = DemoEffect
	d.effect = Element(ns4.Random(0, int(colorMax)))
	d.boss.Enchant(d.effect.Enchant(), ns4.Infinite())
}

func (d *DemoState) updateEffect() {
	// Check current effect and its duration.
	df := d.frame

	// Check if effect should timeout.
	if df > DemoEffectTimeout*ns4.FrameRate() {
		d.startBoss()
		return
	}
	// Power rises as the time passes.
	// Due to integer division, it will rise in steps.
	power := df / (DemoEffectPowerInterval * ns4.FrameRate())
	drawRoomEffect(d.effect, df, power, demoAxisStart, demoLength, demoWidth)
}

func (d *DemoState) startBoss() {
	d.frame = 0
	d.status = DemoBoss
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		pos := u.Pos()
		if pos.X+pos.Y > 10280 {
			continue
		}
		u.Enchant(enchant.INVULNERABLE, ns4.Seconds(DemoBossPlayersFreeze))
		u.Enchant(enchant.FREEZE, ns4.Seconds(DemoBossPlayersFreeze))
	}
	d.boss.Freeze(false)
}

func (d *DemoState) updateBoss() {
	if d.frame > DemoBossUnfreeze*ns4.FrameRate() {
		d.frame = 0
		d.status = DemoEnd
		d.boss.AggressionLevel(1)
		d.boss.EnchantOff(enchant.INVULNERABLE)
		ns4.CastSpell(spell.TURN_UNDEAD, d.boss, d.boss)
		if d.shield != nil {
			d.shield.Delete()
			d.shield = nil
		}
	}
}
