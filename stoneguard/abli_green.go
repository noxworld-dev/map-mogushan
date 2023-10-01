package stoneguard

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
	"github.com/noxworld-dev/opennox-lib/object"
)

type GreenAbility struct {
	frame    int
	charge   int
	active   []*greenSpell
	notFirst bool
}

// Delete all active Green spells.
func (g *GreenAbility) Delete() {
	for _, a := range g.active {
		a.Delete()
	}
	g.active = nil
}

// Update all Green spells for a Guard, starting new ones and removing ended ones.
func (g *GreenAbility) Update(b *Guard) {
	g.frame++
	if g.frame < GreenAfter*ns4.FrameRate() {
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
	if len(g.active) >= GreenProjMax {
		return
	}

	// charge the ability for this number of frames
	g.charge++
	if g.notFirst && g.charge < GreenCooldown*ns4.FrameRate() {
		return // not charged yet
	}
	// ability charged - create new spell and reset charge
	g.charge = 0

	b.unit.AggressionLevel(0)
	b.unit.WalkTo(b.unit.Pos())
	charge := ns4.CreateObject("ForceOfNatureCharge", b.unit.Pos())
	charge.SetOwner(b.unit)

	// add new spell to active ones
	g.active = append(g.active, &greenSpell{charge: charge})
	g.notFirst = true
}

// greenSpell stores state of a single Green spell.
type greenSpell struct {
	charge  ns4.Obj
	proj    ns4.Obj
	ball    ns4.Obj
	pos     ns4.Pointf
	vec     ns4.Pointf
	frame   int
	lastHit int
	stop    bool
}

// Delete a single Blue spell.
func (g *greenSpell) Delete() {
	if g.charge != nil {
		g.charge.Delete()
		g.charge = nil
	}
	if g.ball != nil {
		g.ball.Delete()
		g.ball = nil
	}
	if g.proj != nil {
		g.proj.Delete()
		g.proj = nil
	}
}

// Update runs logic for a single Green spell for a Guard.
func (g *greenSpell) Update(b *Guard) {
	if g.stop {
		return
	}
	g.frame++
	boss := b.unit
	if g.frame < GreenCharge*ns4.FrameRate() {
		return
	}
	if g.charge != nil {
		g.charge.Delete()
		g.charge = nil
		b.unit.AggressionLevel(1)
	}
	if g.ball != nil && g.ball.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		g.ball = nil
		g.proj.Enable(true)
		ns4.CastSpell(spell.TOXIC_CLOUD, g.pos, g.pos)
	}
	if g.proj == nil {
		// pick random player in boss room
		var players []ns4.Obj
		b.s.EachPlayerInRoom(func(u ns4.Obj) {
			players = append(players, u)
		})
		if len(players) == 0 {
			g.stop = true
			return // no players in room
		}
		ind := ns4.Random(0, len(players)-1)
		targ := players[ind].Pos()

		g.pos = boss.Pos()
		g.vec = targ.Sub(boss.Pos()).Normalize()

		g.proj = ns4.CreateObject("CurePoisonPotion", g.pos)
		if g.proj == nil {
			panic("cannot create!")
		}
		g.proj.SetOwner(boss)
	}
	speed := GreenProjSpeed
	if g.ball != nil {
		speed = GreenProjSpeedDeath
	}
	g.pos = g.pos.Add(g.vec.Mul(float32(speed)))
	g.proj.SetPos(g.pos)
	if g.ball != nil {
		g.ball.SetPos(g.pos)
	}
	var hit bool
	g.vec, hit = hitsWall(g.pos, g.vec)
	if dt := g.frame - g.lastHit; dt >= 1*ns4.FrameRate() {
		b.s.EachPlayerInRoom(func(u ns4.Obj) {
			if dt == 0 {
				return
			}
			if sub := u.Pos().Sub(g.pos); sub.Len() < 23 {
				g.vec = sub.Normalize().Mul(-1)
				g.lastHit = g.frame
				dt = 0
			}
		})
		for _, b2 := range b.s.bosses {
			if sub := b2.unit.Pos().Sub(g.pos); sub.Len() < 23 {
				g.vec = sub.Normalize().Mul(-1)
				g.lastHit = g.frame
				break
			}
		}
	}
	if g.ball == nil && hit {
		if b.s.curEffect == b.color {
			ns4.CastSpell(spell.TOXIC_CLOUD, g.pos, g.pos)
			g.stop = true
			return
		}

		g.proj.Enable(false)
		g.ball = ns4.CreateObject("DeathBall", g.pos)
		g.ball.SetOwner(b.unit)
	}
}
