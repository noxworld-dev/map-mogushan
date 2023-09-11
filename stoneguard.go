package mogushan

import (
	"fmt"
	"math/rand"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/opennox-lib/object"
)

const BossHealth = 1000

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

type BossState int

var stoneGuard StoneGuards

type StoneGuards struct {
	state       BossState
	frame       int
	health      int
	curEffect   GuardColor
	startEffect int
	bosses      []*Guard
}

func (s *StoneGuards) spawnBoss() {
	s.curEffect = -1
	s.frame = 0
	s.state = BossWaiting
	s.bosses = nil
	rand.Shuffle(len(startPos), func(i, j int) {
		startPos[i], startPos[j] = startPos[j], startPos[i]
	})
	for i := 0; i < 3; i++ {
		s.NewGuard(GuardColor(i))
	}
}

func (s *StoneGuards) Reset() {
	for _, g := range s.bosses {
		g.Delete()
	}
	for _, pos := range wallPos {
		ns4.Wall(pos[0], pos[1]).Enable(false)
	}
	s.spawnBoss()
}

func (s *StoneGuards) allPlayersInRange(fnc func(u ns4.Obj)) {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		// TODO: check for observer mode
		if s.inBossRange(u) && u.CurrentHealth() > 0 /* && !u.HasEnchant(enchant.ETHEREAL) */ {
			fnc(u)
		}
	}
}

func (s *StoneGuards) arePlayersAlive() bool {
	ok := false
	s.allPlayersInRange(func(u ns4.Obj) {
		ok = true
	})
	return ok
}

func (s *StoneGuards) Update() {
	switch s.state {
	case BossWaiting:
		s.waitingUpdate()
	case BossFighting:
		s.fightingUpdate()
	}
}

func (s *StoneGuards) waitingUpdate() {
	// check if player attemted to charm the boss
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		for _, g := range s.bosses {
			if g.obj.HasOwner(u) {
				// reset the fight
				s.Reset()
				return
			}
		}
	}
	// check if players are close enough to start a fight
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
	s.health = BossHealth
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		if !s.inBossRange(u) {
			u.SetPos(playerPos)
		}
	}
	for _, g := range s.bosses {
		g.Start()
	}
	for _, pos := range wallPos {
		ns4.Wall(pos[0], pos[1]).Enable(true)
	}
	s.frame = 0
	s.curEffect = -1
	s.state = BossFighting
}

func (s *StoneGuards) bossDead() {
	s.state = BossDead
	for _, g := range s.bosses {
		g.Delete()
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
	delta := 0
	for _, g := range s.bosses {
		delta += g.HealthDelta()
	}
	// if delta != 0 {
	// 	fmt.Printf("boss damage: %d+%d = %d\n", s.health, delta, s.health+delta)
	// }
	s.health += delta
	for _, g := range s.bosses {
		g.Update()
	}
	for _, g := range s.bosses {
		g.obj.SetHealth(s.health)
		g.prevHP = s.health
	}
	s.effectUpdate()
	s.frame++
}

var effects = []effect.Effect{
	int(Red):   effect.GREATER_HEAL,
	int(Green): effect.CHARM,
	int(Blue):  effect.DRAIN_MANA,
}

func (s *StoneGuards) resetEffect() {
	prev := s.curEffect
	for {
		s.curEffect = GuardColor(ns4.Random(0, 2))
		if prev != s.curEffect {
			break
		}
	}
	s.startEffect = s.frame
}

func (s *StoneGuards) effectUpdate() {
	if s.curEffect < 0 {
		if s.frame < 2*30 { // 2 sec
			return
		}
		s.resetEffect()
	}
	eff := effects[s.curEffect]
	df := s.frame - s.startEffect
	const max = 60 * 30
	if df > max { // 60 sec
		fmt.Println("Effect timeout!")
		s.resetEffect()
		for _, pl := range ns4.Players() {
			u := pl.Unit()
			if s.inBossRange(u) {
				u.Enchant(enchant.CONFUSED, ns4.Seconds(10))
			}
		}
		return
	}
	power := df / (15 * 30) // each 15 sec
	if s.frame%(5*30) == 0 {
		fmt.Printf("Effect power: %d\n", power)
	}

	var cnt int
	switch power {
	case 0:
		if df%4 == 0 {
			cnt = 1
		}
	case 1:
		if df%3 == 0 {
			cnt = 1
		}
	case 2:
		if df%2 == 0 {
			cnt = 1
		}
	default:
		cnt = 1
	}

	for i := 0; i < cnt; i++ {
		v := float32(ns4.Random(0, 690))
		p0 := ns4.Ptf(4197, 4197)
		p0 = p0.Add(ns4.Ptf(v, v))

		p1 := p0.Add(ns4.Ptf(-184, +184))
		p2 := p0.Add(ns4.Ptf(+184, -184))

		switch s.curEffect {
		case Red, Blue:
			ns4.Effect(eff, p1, p2)
			ns4.Effect(eff, p2, p1)
		case Green:
			ns4.Effect(eff, p1, p2)
		}
	}
}
