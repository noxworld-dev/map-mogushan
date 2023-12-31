package stoneguard

// This file contains all important constants that influence boss balance.

// Debug enables debugging mode for the boss.
// This may change boss behavior to help test it faster and/or enable debug messages.
var Debug = true

// General boss balance.
const (
	// BossModel is a unit model for the boss.
	BossModel = "Troll"
	// BossHealth is a value of a shared boss health pool.
	BossHealth = 1000
	// BossMass is a mass of a boss unit.
	BossMass = 20
	// BossSpeed is a base speed of a boss unit.
	BossSpeed = 1
	// BossAggression sets default boss aggression level.
	BossAggression = 1

	// BossStartFightDist is a distance from a boss to a player when the fight starts.
	BossStartFightDist = 138

	// BossFlamesR is a radius around in which the boss will react to flames. Usually corresponds to the unit model size.
	BossFlamesR = 5
	// BossFlamesCnt is the number of flames under that boss that triggers a defence reaction.
	BossFlamesCnt = 2
)

// Balance values for energy and shield.
const (
	// EnergyDelay is a delay before boss units will start gather energy or will enable shield.
	EnergyDelay = 4 // sec
	// EnergyDist is a distance between two boss units when they start gathering energy.
	EnergyDist = 138
	// EnergyShieldModel is an object that represents an enabled force field for the boss.
	EnergyShieldModel = "MagicEnergy"

	// EnergyExplosionChargeDur is a duration after which a boss unit will charge to 100% and trigger explosion.
	EnergyExplosionChargeDur = 50 // sec
	// EnergyExplosionDamage is damage dealt by elemental explosion (when room effect doesn't match).
	EnergyExplosionDamage = 20
	// EnergyExplosionDamageWeak is damage dealt by elemental explosion (when room effect matches).
	EnergyExplosionDamageWeak = 2
)

// Room effect balance values.
const (
	// RoomEffectDelay is a delay for the first global boss room effect.
	RoomEffectDelay = 2 // sec

	// RoomEffectTimeout is a duration when room effect switches, confusing players.
	RoomEffectTimeout      = 60 // sec
	RoomEffectFirstTimeout = 80 // sec
	// RoomEffectTimeoutConfuse is a duration of confuse effect cast on players after room effect timeout.
	RoomEffectTimeoutConfuse = 10 // sec

	// RoomEffectPowerInterval is an interval after which the room effect increases in power.
	RoomEffectPowerInterval = 15 // sec
	// RoomEffectPowerReport changes the interval at which current effect power will be printed to console.
	RoomEffectPowerReport = 5 // sec
)

const (
	// DemoEffectTimeout is a duration of a demo room effect.
	DemoEffectTimeout = 20 // sec
	// DemoEffectPowerInterval is an interval of a demo room effect power increase.
	DemoEffectPowerInterval = 5 // sec
	// DemoBossPlayersFreeze is a duration of players freeze effect.
	DemoBossPlayersFreeze = 10 // sec
	// DemoBossUnfreeze is a duration of boss freeze.
	DemoBossUnfreeze = 8 // sec
)

// Red ability balance values.
const (
	// RedCooldown sets how frequently the boss will cast the Red ability.
	RedCooldown = 48 // sec
	// RedCharge sets how long it will take for Red ability to charge (ray effect switching to flame line).
	RedCharge = 4 // sec
	// RedAfter sets a delay before the first Red ability is fired. After that, it will fire according to RedCooldown.
	RedAfter = 26
	// RedOnlyOne limits Red ability to a single target.
	RedOnlyOne = true

	// RedLineCnt sets a number of flames between the boss and the target.
	RedLineCnt = 3
	// RedLineMinDist sets a minimal distance between boss and target when the flame line disappears.
	RedLineMinDist = 34
	// RedLineModel sets an object model for flames between the boss and the target.
	RedLineModel = "SmallFlame"

	// RedTargetMinDist sets minimal distance at which the target effect starts to wear off.
	RedTargetMinDist = 184
	// RedTargetMaxDist sets maximal distance at which the target effect starts approaching the target.
	RedTargetMaxDist = 210
	// RedTargetReduceInterval sets time interval after which the target effect will be reduced by 1 level.
	RedTargetReduceInterval = 2 // sec
	// RedTargetWeakR sets the radius in which weak flames will circle the target.
	RedTargetWeakR = 42
	// RedTargetWeakModel sets an object model for weak flames spinning around target (when room effect matches).
	RedTargetWeakModel = "SmallFlame"
	// RedTargetWeakSpeed sets a spin speed for weak flames.
	RedTargetWeakSpeed = 0.05
)

// Blue ability balance values.
const (
	// BlueCooldown sets how frequently the boss will cast the Blue ability.
	BlueCooldown = 48 // sec
	// BlueCharge sets how long it will take for Blue ability to charge (direct lightning switching to circle).
	BlueCharge = 4 // sec
	// BlueAfter sets a delay before the first Blue ability is fired. After that, it will fire according to BlueCooldown.
	BlueAfter = 10

	// BlueDangerModel is a model that indicates a danger of a Blue spell area.
	BlueDangerModel = "BlueFlame"

	// BlueOuterR sets a radius of outer circle for Blue spell.
	BlueOuterR = 138
	// BlueOuterCnt sets a number of outer circle orbs for Blue spell.
	BlueOuterCnt = 10
	// BlueOuterDamage sets per-frame damage from an outer circle for Blue spell (as long as the player is in it).
	BlueOuterDamage = 2
	// BlueOuterModel sets a object model for outer circle orbs.
	BlueOuterModel = "DrainManaOrb"
	// BlueOuterSpeed sets a spin speed for outer circle orbs.
	BlueOuterSpeed = 0.05

	// BlueInnerR is a radius of inner circle for Blue spell.
	BlueInnerR = 46
	// BlueInnerCnt sets a number of inner circle orbs for Blue spell.
	BlueInnerCnt = 10
	// BlueInnerDamage sets damage done once to the player that enters inner circle (when room effect doesn't match).
	BlueInnerDamage = 20
	// BlueInnerDamageWeak sets damage done once to the player that enters inner circle (when room effect matches).
	BlueInnerDamageWeak = 2
	// BlueInnerStun sets stun duration when player enters inner circle (when room effect doesn't match).
	BlueInnerStun = 20 // sec
	// BlueInnerStunWeak sets stun duration when player enters inner circle (when room effect matches).
	BlueInnerStunWeak = 2 // sec
	// BlueInnerModel sets a object model for inner circle orbs.
	BlueInnerModel = "WhiteOrb"
	// BlueInnerSpeed sets a spin speed for inner circle orbs.
	BlueInnerSpeed = 0.05
)

// Green ability balance values.
const (
	// GreenCooldown sets how frequently the boss will cast the Green ability.
	GreenCooldown = 48 // sec
	// GreenAfter sets a delay before the first Green ability is fired. After that, it will fire according to GreenCooldown.
	GreenAfter = 42
	// GreenCharge sets how long it will take for Green ability to charge (FoN effect to projectile).
	GreenCharge = 4 // sec

	// GreenProjMax sets maximal amount of Green spell projectiles.
	GreenProjMax = 4
	// GreenProjSpeed sets the speed of small Green projectile.
	GreenProjSpeed = 2
	// GreenProjSpeedDeath sets the speed of large Green projectile.
	GreenProjSpeedDeath = 8
	// GreenProjKickInterval sets a minimal interval at which the Green projectile can be kicked around.
	GreenProjKickInterval = 1 // sec
	// GreenProjKickDist sets a distance at which green projectile is kicked.
	GreenProjKickDist = 23
	// GreenProjModel sets an object model for small Green projectile.
	GreenProjModel = "CurePoisonPotion"
)
