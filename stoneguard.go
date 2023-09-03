package mogushan

import (
	"math/rand"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
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

var stoneGuard StoneGuards

type StoneGuards struct {
	state  int
	health int
	bosses []*Guard
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
		g.Delete()
	}
	for _, pos := range wallPos {
		ns4.Wall(pos[0], pos[1]).Enable(false)
	}
	s.spawnBoss()
}

func (s *StoneGuards) arePlayersAlive() bool {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		// TODO: check for observer mode
		if s.inBossRange(u) && u.CurrentHealth() > 0 /* && !u.HasEnchant(enchant.ETHEREAL) */ {
			return true
		}
	}
	return false
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
}
