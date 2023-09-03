package mogushan

import (
	"math/rand"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/noxscript/ns/v4/spell"
	"github.com/noxworld-dev/opennox-lib/object"
)

var startPos = []ns4.Pointf{
	{4151, 4473},
	{4128, 4358},
	{4358, 4128},
	{4473, 4151},
}

const (
	BossWaiting = iota
	BossFighting
	BossDead
)

var stoneGuard StoneGuards

type StoneGuards struct {
	state  int
	bosses []*Guard
}

func (s *StoneGuards) NewGuard(i int) *Guard {
	g := &Guard{}
	g.obj = ns4.CreateObject("Troll", startPos[i])
	g.prevPos = startPos[i]
	switch i {
	case 0:
		g.obj.Enchant(enchant.PROTECT_FROM_ELECTRICITY, infinite())
	case 1:
		g.obj.Enchant(enchant.PROTECT_FROM_FIRE, infinite())
	case 2:
		g.obj.Enchant(enchant.PROTECT_FROM_POISON, infinite())
	}
	g.obj.Enchant(enchant.FREEZE, infinite())
	g.obj.Freeze(true)
	g.obj.LookWithAngle(32)
	g.obj.AggressionLevel(0) // TODO: remove
	// obj.SetBaseSpeed(5)
	//fmt.Printf("speed: %v\n", obj.BaseSpeed())
	s.bosses = append(s.bosses, g)
	return g
}

func (s *StoneGuards) eachGuard(fnc func(obj ns4.Obj)) {
	for _, g := range s.bosses {
		fnc(g.obj)
	}
}

func (s *StoneGuards) spawnBoss() {
	s.state = BossWaiting
	s.bosses = nil
	rand.Shuffle(len(startPos), func(i, j int) {
		startPos[i], startPos[j] = startPos[j], startPos[i]
	})
	for i := 0; i < 3; i++ {
		s.NewGuard(i)
	}
}

func (s *StoneGuards) Reset() {
	for _, g := range s.bosses {
		g.obj.Delete()
	}
	for _, pos := range wallPos {
		ns4.Wall(pos[0], pos[1]).Enable(false)
	}
	s.spawnBoss()
}

func (s *StoneGuards) arePlayersAlive() bool {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		if s.inBossRange(u) && u.CurrentHealth() > 0 {
			return true
		}
	}
	return false
}

func (s *StoneGuards) Update() {
	//println("state:", bossState)
	switch s.state {
	case BossWaiting:
		s.waitingUpdate()
	case BossFighting:
		s.fightingUpdate()
	}
}

func (s *StoneGuards) waitingUpdate() {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		for _, g := range s.bosses {
			if g.obj.HasOwner(u) {
				s.Reset()
				return
			}
		}
	}
	tooClose := false
	for _, g := range s.bosses {
		pl := ns4.FindClosestObject(g.obj, ns4.HasClass(object.ClassPlayer), ns4.ObjCondFunc(func(obj ns4.Obj) bool {
			return obj.CurrentHealth() > 0
		}))
		if pl != nil && g.obj.Pos().Sub(pl.Pos()).Len() < 138 {
			tooClose = true
			break
		}
	}
	if tooClose {
		s.startFight()
	}
}

func (s *StoneGuards) inBossRange(pl ns4.Obj) bool {
	pos := pl.Pos()
	return pos.X+pos.Y < 9820
}

var wallPos = [][2]int{
	{205, 209},
	{206, 208},
	{207, 207},
	{208, 206},
	{209, 205},
}

var playerPos = ns4.Ptf(4726, 4726)

func (s *StoneGuards) startFight() {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		if !s.inBossRange(u) {
			u.SetPos(playerPos)
		}
	}
	for _, g := range s.bosses {
		g.obj.Freeze(false)
		g.obj.EnchantOff(enchant.FREEZE)
		g.obj.SetMaxHealth(1000)
		ns4.CastSpell(spell.COUNTERSPELL, g.obj, g.obj)
	}
	for _, pos := range wallPos {
		ns4.Wall(pos[0], pos[1]).Enable(true)
	}
	s.state = BossFighting
}

func (s *StoneGuards) bossDead() {
	s.state = BossDead
	for _, g := range s.bosses {
		g.obj.Delete()
	}
	// TODO: prize!
}

func (s *StoneGuards) areAllAlive() bool {
	for _, g := range s.bosses {
		if g.obj.CurrentHealth() > 0 {
			return true
		}
	}
	return false
}

func (s *StoneGuards) fightingUpdate() {
	if !s.arePlayersAlive() {
		s.Reset()
		return
	}
	if !s.areAllAlive() {
		s.bossDead()
		return
	}
	for _, g := range s.bosses {
		g.Update()
	}
}

type Guard struct {
	obj     ns4.Obj
	prevPos ns4.Pointf
	frame   int
}

func (g *Guard) Update() {
	g.antiSpell()
	g.prevPos = g.obj.Pos()
	g.frame++
}

func (g *Guard) antiSpell() {
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
	if g.obj.HasEnchant(enchant.CHARMING) {
		ns4.CastSpell(spell.COUNTERSPELL, g.obj, g.obj)
	}
}
