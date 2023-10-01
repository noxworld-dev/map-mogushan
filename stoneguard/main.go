package stoneguard

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/opennox-lib/object"
)

// state contains all state of the boss zone
var state State

func init() {
	// register map events
	ns4.OnMapEvent(ns4.MapInitialize, state.Reset)
	ns4.OnFrame(state.Update)
}

const (
	BossWaiting = iota
	BossFighting
	BossDead
)

// BossState is an enum for boss state.
type BossState int

// State contains all state of the Stone Guard boss zone.
type State struct {
	state           BossState
	frame           int
	health          int
	curEffect       Element // current room effect
	roomEffectStart int
	bosses          []*Guard
}

// IsAlive checks if boss is still alive.
func (s *State) IsAlive() bool {
	for _, g := range s.bosses {
		if g.unit.CurrentHealth() > 0 {
			return true
		}
	}
	return false
}

// Delete old boss units with all their state.
func (s *State) Delete() {
	for _, g := range s.bosses {
		g.Delete()
	}
}

// Reset the boss to the starting state.
func (s *State) Reset() {
	// delete old boss
	s.Delete()
	// open entrance
	s.switchEntrance(true)
	// respawn the boss
	s.spawnBoss()
}

// spawnBoss sets boss state to BossWaiting and respawns the bosses.
func (s *State) spawnBoss() {
	// set initial state
	s.curEffect = -1
	s.frame = 0
	s.state = BossWaiting
	s.bosses = nil
	// spawn the bosses
	spawns := s.randomBossPos()
	for i := 0; i < 3; i++ {
		s.NewGuard(Element(i)%colorMax, spawns[i])
	}
}

// EachPlayerInRoom is a helper that iterates over all alive players in the boss room.
func (s *State) EachPlayerInRoom(fnc func(u ns4.Obj)) {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		// TODO: check for observer mode
		if s.InRoom(u) && u.CurrentHealth() > 0 /* && !u.HasEnchant(enchant.ETHEREAL) */ {
			fnc(u)
		}
	}
}

// ArePlayersAlive checks if there are any alive players in the boss room.
func (s *State) ArePlayersAlive() bool {
	ok := false
	s.EachPlayerInRoom(func(u ns4.Obj) {
		ok = true
	})
	return ok
}

// Update the boss state. This is the main script function.
func (s *State) Update() {
	switch s.state {
	case BossWaiting:
		s.waitingUpdate()
	case BossFighting:
		s.fightingUpdate()
	case BossDead:
		// nothing to update
	}
}

// waitingUpdate is the update function for the BossWaiting state.
func (s *State) waitingUpdate() {
	// check if player attempts to charm the boss
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		for _, g := range s.bosses {
			if g.unit.HasOwner(u) {
				// reset the fight immediately
				s.Reset()
				return
			}
		}
	}
	// check if players are close enough to start a fight
	tooClose := false
	for _, g := range s.bosses {
		pl := ns4.FindClosestObject(g.unit, ns4.HasClass(object.ClassPlayer), ns4.ObjCondFunc(func(obj ns4.Obj) bool {
			return obj.CurrentHealth() > 0
		}))
		if pl != nil && g.unit.Pos().Sub(pl.Pos()).Len() < BossStartFightDist {
			tooClose = true
			break
		}
	}
	if tooClose {
		s.startFight()
	}
}

// startFight starts the boss fight. It switches boss state to BossFighting.
func (s *State) startFight() {
	// set shared boss health pool
	s.health = BossHealth
	// teleport players that are not in the room already
	s.teleportPlayersToRoom()
	// start the bosses
	for _, g := range s.bosses {
		g.Start()
	}
	// close the entrance
	s.switchEntrance(false)
	// init other state
	s.frame = 0
	s.curEffect = -1
	s.state = BossFighting
}

// bossDead ends the boss fight with boss death. Switches state to BossDead.
func (s *State) bossDead() {
	s.state = BossDead
	// delete all remaining state
	for _, g := range s.bosses {
		g.Delete()
	}
	// TODO: prize!
}

// fightingUpdate is the update function for the BossFighting state.
func (s *State) fightingUpdate() {
	if !s.ArePlayersAlive() {
		s.Reset()
		return
	}
	if !s.IsAlive() {
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
		g.unit.SetHealth(s.health)
		g.prevHP = s.health
	}
	s.roomEffectUpdate()
	s.frame++
}
