package stoneguard

import (
	"math"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
)

// RedAbility is an ability for Red color/element.
// It stores all active spells of this kind for a given boss unit.
type RedAbility struct {
	frame    int
	charge   int
	active   []*redSpell
	notFirst bool
}

// Delete all active Red spells.
func (g *RedAbility) Delete() {
	for _, a := range g.active {
		a.Delete()
	}
	g.active = nil
}

// hasTarget checks if ability already targets a given unit.
func (g *RedAbility) hasTarget(u ns4.Obj) bool {
	for _, a := range g.active {
		if a.target == u {
			return true
		}
	}
	return false
}

// Update all Red spells for a Guard, starting new ones and removing ended ones.
func (g *RedAbility) Update(b *Guard) {
	g.frame++
	if g.frame < RedAfter*ns4.FrameRate() {
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
	if g.notFirst && g.charge < RedCooldown*ns4.FrameRate() {
		return // not charged yet
	}
	// ability charged - create new spell and reset charge
	g.charge = 0

	// pick random player position in boss room
	var players []ns4.Obj
	b.s.EachPlayerInRoom(func(u ns4.Obj) {
		if !g.hasTarget(u) {
			players = append(players, u)
		}
	})
	if len(players) == 0 {
		return // no players in room
	}
	ind := ns4.Random(0, len(players)-1)
	targ := players[ind]
	// add new spell to active ones
	g.active = append(g.active, &redSpell{
		target: targ,
	})
	g.notFirst = true
}

// redSpell stores state of a single Red spell.
type redSpell struct {
	target ns4.Obj
	frame  int
	line   []ns4.Obj // flame line between the boss and target
	reduce int
	strong [4]ns4.Obj // flame variations for the strong flame that tries to hit the player (large to small)
	weak   [4]ns4.Obj // weak flames that will circle the target for the weak spell variant
	stop   bool
}

// Delete a single Red spell.
func (g *redSpell) Delete() {
	// delete the flame line
	for _, a := range g.line {
		a.Delete()
	}
	g.line = nil
	// delete strong target effect
	for i, a := range g.strong {
		a.Delete()
		g.strong[i] = nil
	}
	// delete weak target effect
	for i, a := range g.weak {
		a.Delete()
		g.weak[i] = nil
	}
}

// Update runs logic for a single Red spell for a Guard.
func (g *redSpell) Update(b *Guard) {
	if g.stop {
		return
	}
	g.frame++
	boss, targ := b.unit, g.target

	// If the spell is charging, show a ray effect between the boss that the target.
	if g.frame < RedCharge*ns4.FrameRate() {
		ns4.Effect(effect.GREATER_HEAL, boss, targ)
		ns4.Effect(effect.GREATER_HEAL, targ, boss)
		return
	}
	// Initialize the flame line if not done already.
	if g.line == nil {
		for i := 0; i < RedLineCnt; i++ {
			flame := ns4.CreateObject(RedLineModel, boss)
			flame.SetOwner(boss)
			g.line = append(g.line, flame)
		}
	}
	// Initialize strong target effect, if not done already.
	if g.strong[0] == nil {
		g.strong[0] = ns4.CreateObject("LargeFlame", boss)
		g.strong[1] = ns4.CreateObject("Flame", boss)
		g.strong[2] = ns4.CreateObject("MediumFlame", boss)
		g.strong[3] = ns4.CreateObject("SmallFlame", boss)
		for _, a := range g.strong {
			a.SetOwner(boss)
		}
	}
	// Initialize weak target effect, if not done already.
	if g.weak[0] == nil {
		for i := range g.weak {
			g.weak[i] = ns4.CreateObject(RedTargetWeakModel, boss)
			g.weak[i].SetOwner(boss)
		}
	}

	// Calculate the distance and direction vector between the boss and the target.
	p1, p2 := boss.Pos(), targ.Pos()
	vec := p2.Sub(p1)
	dir := vec.Normalize()
	dist := float32(vec.Len())

	// Show the flame line between the boss and the target.
	if lineDist := dist; lineDist >= RedLineMinDist*2 {
		// Calculate "pure" distance, without the boss/target model sizes.
		lineDist -= RedLineMinDist * 2
		// Put flames in between at an even intervals.
		for i, f := range g.line {
			perc := float32(i+1) / float32(len(g.line))
			fpos := p1.Add(dir.Mul(RedLineMinDist + perc*lineDist))
			f.Enable(true)
			f.SetPos(fpos)
		}
	} else {
		// Disable the line if too close. We don't want it to damage the target or the boss.
		for _, f := range g.line {
			f.Enable(false)
		}
	}

	// Calculate the spell target position.
	// If the actual target (player) is too far, the spell target will be closer.
	// Otherwise, the spell target will lock at a specific distance.
	var tdist float32
	if dist < RedTargetMaxDist {
		tdist = RedTargetMaxDist - dist
	}
	targPos := p2.Add(dir.Mul(tdist))

	reduceLvl := g.reduce / (RedTargetReduceInterval * ns4.FrameRate())
	// Check if the room effect matches the color/element of the boss.
	if b.s.curEffect == b.color {
		// Color/element matches - weak spell variant.

		// Disable strong effect first.
		for _, a := range g.strong {
			a.Enable(false)
			a.SetPos(targPos)
		}

		// Stop the spell if it died down.
		if reduceLvl >= len(g.weak) {
			g.stop = true
			return
		}

		// Put the weak flames around the target.
		for i, a := range g.weak {
			ph := float64(i)*math.Pi/2 + float64(g.frame)*RedTargetWeakSpeed
			dx, dy := float32(RedTargetWeakR*math.Cos(ph)), float32(RedTargetWeakR*math.Sin(ph))
			a.SetPos(targPos.Add(ns4.Ptf(dx, dy)))

			// Disable more weak flames if the effect is reduced.
			if i >= reduceLvl {
				a.Enable(true)
			} else {
				a.Enable(false)
			}
		}
	} else {
		// Color/element doesn't match - strong spell variant.

		// Disable weak effect first.
		for _, a := range g.weak {
			a.Enable(false)
			a.SetPos(targPos)
		}

		// Stop the spell if it died down.
		if reduceLvl >= len(g.strong) {
			g.stop = true
			return
		}

		//
		for i, a := range g.strong {
			a.SetPos(targPos)
			if i == reduceLvl {
				a.Enable(true)
			} else {
				a.Enable(false)
			}
		}
	}

	// If the distance between boss and the player is large enough - slowly reduce the target effect.
	if dist > RedTargetMinDist {
		g.reduce++
	}
}
