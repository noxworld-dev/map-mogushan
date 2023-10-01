package stoneguard

// This file contains room-related constants and functions, for example boss spawn positions, wall positions, etc.
// It is also responsible for the global room effects.

import (
	"fmt"
	"math"
	"math/rand"

	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
	"github.com/noxworld-dev/opennox-lib/types"
)

// startPos is an array of boss starting positions.
// Each boss will pick one random position from this list.
var startPos = []ns4.Pointf{
	{4151, 4473},
	{4128, 4358},
	{4358, 4128},
	{4473, 4151},
}

// roomAxisStart is a start coordinate for the main room axis of symmetry.
var roomAxisStart = ns4.Ptf(4197, 4197)

const (
	// roomLength is a length or the room, starting from roomAxisStart, and following a diagonal.
	roomLength = 690
	// roomWidth is a width of a room, starting from roomAxisStart as a center and perpendicular to diagonal.
	roomWidth = 368
)

// entranceWalls is an array of wall coordinates for the entrance.
var entranceWalls = [][2]int{
	{205, 209},
	{206, 208},
	{207, 207},
	{208, 206},
	{209, 205},
}

// playerPos is a default positions where players will be teleported to when the fight starts.
var playerPos = ns4.Ptf(4726, 4726)

// switchEntrance switches boss zone entrance on/off.
func (s *State) switchEntrance(open bool) {
	for _, pos := range entranceWalls {
		ns4.Wall(pos[0], pos[1]).Enable(!open)
	}
}

// InRoom checks if object is in the boss room.
func (s *State) InRoom(pl ns4.Obj) bool {
	if pl == nil {
		return false
	}
	pos := pl.Pos()
	return pos.X+pos.Y < 9820
}

// teleportPlayersToRoom teleports players that are not in the room already to playerPos.
func (s *State) teleportPlayersToRoom() {
	for _, pl := range ns4.Players() {
		u := pl.Unit()
		if u != nil && !s.InRoom(u) {
			u.SetPos(playerPos)
		}
	}
}

// randomBossPos selects 3 random boss spawn positions.
func (s *State) randomBossPos() [3]types.Pointf {
	rand.Shuffle(len(startPos), func(i, j int) {
		startPos[i], startPos[j] = startPos[j], startPos[i]
	})
	return [3]types.Pointf(startPos[:3])
}

// nextRoomEffect sets a new global room effect.
func (s *State) nextRoomEffect() {
	// do not allow the same effect to play twice
	prev := s.curEffect
	for {
		s.curEffect = Element(ns4.Random(0, 3)) % colorMax
		if prev != s.curEffect {
			break
		}
	}
	s.roomEffectStart = s.frame
}

// roomEffectUpdate updates the global boss room effect.
func (s *State) roomEffectUpdate() {
	if s.curEffect < 0 {
		// Start the first effect only after a delay.
		if s.frame < RoomEffectDelay*ns4.FrameRate() {
			return
		}
		s.nextRoomEffect()
	}
	// Check current effect and its duration.
	eff := s.curEffect.RoomEffect()
	df := s.frame - s.roomEffectStart

	// Check if effect should timeout.
	if df > RoomEffectTimeout*ns4.FrameRate() {
		// Switch effect and confuse players.
		fmt.Println("Effect timeout!")
		s.nextRoomEffect()
		s.EachPlayerInRoom(func(u ns4.Obj) {
			u.Enchant(enchant.CONFUSED, ns4.Seconds(RoomEffectTimeoutConfuse))
		})
		return
	}
	// Power rises as the time passes.
	// Due to integer division, it will rise in steps.
	power := df / (RoomEffectPowerInterval * ns4.FrameRate())

	// Print effect power for debugging.
	if Debug && s.frame%(RoomEffectPowerReport*ns4.FrameRate()) == 0 {
		fmt.Printf("Effect power: %d\n", power)
	}

	// Choose effect density based on the current effect power.
	var cnt int
	switch power {
	case 0:
		if df%4 == 0 { // once per 4 frames
			cnt = 1
		}
	case 1:
		if df%3 == 0 { // once per 3 frames
			cnt = 1
		}
	case 2:
		if df%2 == 0 { // once per 2 frames
			cnt = 1
		}
	default:
		cnt = 1 // every frame
	}

	// Spawn 'cnt' effects across the room.
	for i := 0; i < cnt; i++ {
		// Room is placed right on the diagonal.
		p0 := roomAxisStart

		// Pick random point across the diagonal.
		v := float32(ns4.Random(0, roomLength))
		p0 = p0.Add(ns4.Ptf(v, v))

		// Effect crosses the whole room, perpendicular to diagonal.
		p1 := p0.Add(ns4.Ptf(-roomWidth/2, +roomWidth/2))
		p2 := p0.Add(ns4.Ptf(+roomWidth/2, -roomWidth/2))

		// Display the actual effect.
		switch s.curEffect {
		case Red, Blue:
			ns4.Effect(eff, p1, p2)
			ns4.Effect(eff, p2, p1)
		case Green:
			// it's already bidirectional
			ns4.Effect(eff, p1, p2)
		}
	}
}

var wallPoints = []ns4.Pointf{
	{3990, 4381}, // left
	{4381, 3990}, // top
	{5094, 4703}, // right
	{4703, 5094}, // bottom
}

var walls = []ns4.Pointf{
	wallPoints[3].Sub(wallPoints[0]), // left
	wallPoints[0].Sub(wallPoints[1]), // far
	wallPoints[1].Sub(wallPoints[2]), // right
	wallPoints[2].Sub(wallPoints[3]), // close
}

var (
	phiLeft  = math.Atan2(float64(walls[0].Y), float64(walls[0].X))
	phiRight = math.Atan2(float64(walls[2].Y), float64(walls[2].X))
	phiFar   = math.Atan2(float64(walls[1].Y), float64(walls[1].X))
	phiClose = math.Atan2(float64(walls[3].Y), float64(walls[3].X))
)

func hitsWall(pos, vec ns4.Pointf) (ns4.Pointf, bool) {
	phi := math.Atan2(float64(vec.Y), float64(vec.X))
	var wph float64
	if pos.X+pos.Y < 8371 {
		wph = phiFar
	} else if pos.X+pos.Y > 9797 {
		wph = phiClose
	} else if pos.X-pos.Y > 391 {
		wph = phiRight
	} else if pos.X-pos.Y < -391 {
		wph = phiLeft
	} else {
		return vec, false
	}
	dphi := wph - phi
	phi = wph + dphi
	return ns4.Pointf{X: float32(math.Cos(phi)), Y: float32(math.Sin(phi))}, true
}
