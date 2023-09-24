package mogushan

import (
	"math"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
)

type abilitiesRed struct {
	charge int
	active []*abilityRed
}

type abilityRed struct {
	target    ns4.Obj
	frame     int
	line      []ns4.Obj
	switchDur int
	large     [4]ns4.Obj
	small     [4]ns4.Obj
	stop      bool
}

func (g *abilitiesRed) Delete() {
	for _, a := range g.active {
		a.Delete()
	}
	g.active = nil
}

func (g *abilityRed) Delete() {
	for _, a := range g.line {
		a.Delete()
	}
	for i, a := range g.large {
		a.Delete()
		g.large[i] = nil
	}
	for i, a := range g.small {
		a.Delete()
		g.small[i] = nil
	}
	g.line = nil
}

func (g *abilitiesRed) hasTarget(u ns4.Obj) bool {
	for _, a := range g.active {
		if a.target == u {
			return true
		}
	}
	return false
}

func (g *abilitiesRed) Update(b *Guard) {
	//const max = 10 * 30 // 10 sec
	const max = 24 * 30 // 24 sec
	g.charge++
	for i := 0; i < len(g.active); i++ {
		if g.active[i].stop {
			g.active[i].Delete()
			g.active = append(g.active[:i], g.active[i+1:]...)
			i--
		}
	}
	for _, a := range g.active {
		a.Update(b)
	}
	if g.charge < max {
		return
	}
	g.charge = 0
	var players []ns4.Obj
	b.s.allPlayersInRange(func(u ns4.Obj) {
		if !g.hasTarget(u) {
			players = append(players, u)
		}
	})
	if len(players) == 0 {
		return
	}
	ind := ns4.Random(0, len(players)-1)
	a := &abilityRed{
		target: players[ind],
	}
	g.active = append(g.active, a)
}

func (g *abilityRed) Update(b *Guard) {
	if g.stop {
		return
	}
	g.frame++
	boss, targ := b.obj, g.target
	if g.frame < 4*30 {
		ns4.Effect(effect.GREATER_HEAL, boss, targ)
		ns4.Effect(effect.GREATER_HEAL, targ, boss)
		return
	}
	if g.line == nil {
		for i := 0; i < 3; i++ {
			flame := ns4.CreateObject("SmallFlame", boss)
			flame.SetOwner(boss)
			g.line = append(g.line, flame)
		}
	}
	if g.large[0] == nil {
		g.large[0] = ns4.CreateObject("LargeFlame", boss)
		g.large[1] = ns4.CreateObject("Flame", boss)
		g.large[2] = ns4.CreateObject("MediumFlame", boss)
		g.large[3] = ns4.CreateObject("SmallFlame", boss)
		for _, a := range g.large {
			a.SetOwner(boss)
		}
		for i := range g.small {
			g.small[i] = ns4.CreateObject("SmallFlame", boss)
			g.small[i].SetOwner(boss)
		}
	}
	p1, p2 := boss.Pos(), targ.Pos()
	vec := p2.Sub(p1)
	dir := vec.Normalize()
	d := float32(vec.Len())

	const dist = 34
	if l := d; l >= dist*2 {
		l -= dist * 2
		fpos := []ns4.Pointf{
			p1.Add(dir.Mul(dist + 0.25*l)),
			p1.Add(dir.Mul(dist + 0.50*l)),
			p1.Add(dir.Mul(dist + 0.75*l)),
		}
		for i := 0; i < 3; i++ {
			g.line[i].Enable(true)
			g.line[i].SetPos(fpos[i])
		}
	} else {
		for i := 0; i < 3; i++ {
			g.line[i].Enable(false)
		}
	}

	const lmin, lmax = 184, 210
	var ldist float32
	if d < lmax {
		ldist = lmax - d
	}

	lpos := p2.Add(dir.Mul(ldist))

	lactive := g.switchDur / (2 * 30)
	if b.s.curEffect == b.color {
		// same color - 4 small flames around
		for _, a := range g.large {
			a.Enable(false)
			a.SetPos(lpos)
		}
		if lactive >= len(g.small) {
			g.stop = true
			return
		}
		const sdist = 42
		for i, a := range g.small {
			ph := float64(i)*math.Pi/2 + float64(g.frame)/20
			dx, dy := float32(sdist*math.Cos(ph)), float32(sdist*math.Sin(ph))
			a.SetPos(lpos.Add(ns4.Ptf(dx, dy)))
			if i >= lactive {
				a.Enable(true)
			} else {
				a.Enable(false)
			}
		}
	} else {
		// different colors - large flame that hits when too far
		for _, a := range g.small {
			a.Enable(false)
			a.SetPos(lpos)
		}
		if lactive >= len(g.large) {
			g.stop = true
			return
		}
		for i, a := range g.large {
			a.SetPos(lpos)
			if i == lactive {
				a.Enable(true)
			} else {
				a.Enable(false)
			}
		}
	}
	if d > lmin {
		g.switchDur++
	}
}
