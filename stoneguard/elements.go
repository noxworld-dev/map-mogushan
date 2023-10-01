package stoneguard

import (
	"fmt"

	"github.com/noxworld-dev/noxscript/ns/v4/damage"
	"github.com/noxworld-dev/noxscript/ns/v4/effect"
	"github.com/noxworld-dev/noxscript/ns/v4/enchant"
)

// Element is a color/element of a boss unit. It is also used for the room effects.
type Element int

const (
	Red      = Element(0)
	Green    = Element(1)
	Blue     = Element(2)
	colorMax = Element(3)
)

func (c Element) String() string {
	switch c {
	case Red:
		return "Red"
	case Green:
		return "Green"
	case Blue:
		return "Blue"
	}
	return fmt.Sprintf("GuardColor(%d)", int(c))
}

// Enchant returns an enchant that corresponds to the color/element.
func (c Element) Enchant() enchant.Enchant {
	switch c {
	case Red:
		return enchant.PROTECT_FROM_FIRE
	case Green:
		return enchant.PROTECT_FROM_POISON
	case Blue:
		return enchant.PROTECT_FROM_ELECTRICITY
	}
	return ""
}

// RoomEffect returns a room effect that corresponds to the color/element.
func (c Element) RoomEffect() effect.Effect {
	switch c {
	case Red:
		return effect.GREATER_HEAL
	case Green:
		return effect.CHARM
	case Blue:
		return effect.DRAIN_MANA
	}
	return ""
}

// DamageType returns a damage type that corresponds to the color/element.
func (c Element) DamageType() damage.Type {
	// TODO: set unique damage type?
	switch c {
	case Red:
		return damage.ZAP_RAY
	case Green:
		return damage.ZAP_RAY
	case Blue:
		return damage.ZAP_RAY
	}
	return damage.ZAP_RAY
}

// Ability is a unique boss ability.
type Ability interface {
	Update(g *Guard)
	Delete()
}

// NewAbility creates ability that corresponds a given color/element.
func (c Element) NewAbility() Ability {
	switch c {
	case Red:
		return &RedAbility{}
	case Green:
		return &abilitiesGreen{}
	case Blue:
		return &BlueAbility{}
	}
	return nil
}
