package game

import (
	"errors"
	"math"
	"time"
)

const (
	MaxPlayerHealth = 100
	CombatCooldown  = time.Second
)

var (
	ErrUnknownEnemy      = errors.New("unknown enemy")
	ErrEnemyOutOfRange   = errors.New("enemy is out of range")
	ErrEnemyRespawning   = errors.New("enemy is respawning")
	ErrCombatCooldown    = errors.New("combat action is cooling down")
	ErrPlayerDefeated    = errors.New("player is defeated")
	ErrRestOutOfRange    = errors.New("resting is only possible near a safe spawn")
	ErrInvalidBossAttack = errors.New("invalid boss attack parameters")
)

type EnemyDefinition struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	MaxHealth        int           `json:"maxHealth"`
	Attack           int           `json:"attack"`
	Defense          int           `json:"defense"`
	Experience       int           `json:"experience"`
	LootItemID       string        `json:"lootItemId"`
	LootQuantity     int           `json:"lootQuantity"`
	LootTableID      string        `json:"lootTableId,omitempty"`
	Position         WorldPosition `json:"position"`
	InteractionRange float64       `json:"interactionRange"`
	RespawnSeconds   int           `json:"respawnSeconds"`
}

type EnemyState struct {
	Health      int
	RespawnAt   time.Time
	Initialized bool
}

type CombatResult struct {
	EnemyID          string `json:"enemyId"`
	SkillID          string `json:"skillId"`
	ParentSkillID    string `json:"parentSkillId,omitempty"`
	ParentExperience int    `json:"parentExperience,omitempty"`
	EnemyHealth      int    `json:"enemyHealth"`
	PlayerHealth     int    `json:"playerHealth"`
	DamageDealt      int    `json:"damageDealt"`
	DamageReceived   int    `json:"damageReceived"`
	EnemyDefeated    bool   `json:"enemyDefeated"`
	PlayerDefeated   bool   `json:"playerDefeated"`
	Experience       int    `json:"experience"`
	LootItemID       string `json:"lootItemId,omitempty"`
	LootQuantity     int    `json:"lootQuantity,omitempty"`
}

var enemyDefinitions = []EnemyDefinition{
	{ID: "frontier-rat-001", Name: "Frontier Rat", MaxHealth: 12, Attack: 2, Defense: 0, Experience: 25, LootTableID: "frontier-rat-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 8, Z: 0}, InteractionRange: 3, RespawnSeconds: 30},
	{ID: "frontier-boar-001", Name: "Wild Boar", MaxHealth: 35, Attack: 6, Defense: 2, Experience: 60, LootTableID: "frontier-boar-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -8, Z: 0}, InteractionRange: 3, RespawnSeconds: 60},
	{ID: "stoneback-colossus", Name: "Stoneback Colossus", MaxHealth: 100, Attack: 0, Defense: 2, Experience: 500, LootTableID: "stoneback-colossus-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 32, Z: 32}, InteractionRange: 3, RespawnSeconds: 300},
	// 6 Additional Enemies & Monsters across the Frontier
	{ID: "canyon-stalker-001", Name: "Canyon Stalker", MaxHealth: 24, Attack: 4, Defense: 1, Experience: 45, LootTableID: "canyon-stalker-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 18, Z: 12}, InteractionRange: 3, RespawnSeconds: 45},
	{ID: "mire-drake-001", Name: "Mire Drake", MaxHealth: 38, Attack: 5, Defense: 2, Experience: 70, LootTableID: "mire-drake-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -18, Z: 20}, InteractionRange: 3, RespawnSeconds: 55},
	{ID: "ash-hound-001", Name: "Ash Hound", MaxHealth: 45, Attack: 7, Defense: 2, Experience: 85, LootTableID: "ash-hound-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -20, Z: -16}, InteractionRange: 3, RespawnSeconds: 65},
	{ID: "frost-prowler-001", Name: "Frost Prowler", MaxHealth: 52, Attack: 8, Defense: 3, Experience: 105, LootTableID: "frost-prowler-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 22, Z: -8}, InteractionRange: 3, RespawnSeconds: 75},
	{ID: "ruin-sentinel-001", Name: "Ruin Sentinel", MaxHealth: 65, Attack: 9, Defense: 4, Experience: 130, LootTableID: "ruin-sentinel-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 24, Z: -20}, InteractionRange: 3, RespawnSeconds: 90},
	{ID: "aether-wraith-001", Name: "Aether Wraith", MaxHealth: 78, Attack: 11, Defense: 3, Experience: 165, LootTableID: "aether-wraith-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -28, Z: 10}, InteractionRange: 3, RespawnSeconds: 105},
	{ID: "feral-drone-001", Name: "Feral Scavenger Drone", MaxHealth: 30, Attack: 5, Defense: 2, Experience: 55, LootTableID: "feral-drone-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 12, Z: -18}, InteractionRange: 3, RespawnSeconds: 50},
	{ID: "thornback-behemoth-001", Name: "Thornback Behemoth", MaxHealth: 88, Attack: 12, Defense: 4, Experience: 190, LootTableID: "thornback-behemoth-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 32, Z: -12}, InteractionRange: 3, RespawnSeconds: 115},
	{ID: "obsidian-wyrm-001", Name: "Obsidian Wyrm", MaxHealth: 95, Attack: 13, Defense: 5, Experience: 220, LootTableID: "obsidian-wyrm-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -36, Z: -20}, InteractionRange: 3, RespawnSeconds: 125},
	{ID: "void-reaver-001", Name: "Void Reaver", MaxHealth: 110, Attack: 14, Defense: 4, Experience: 260, LootTableID: "void-reaver-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 36, Z: 26}, InteractionRange: 3, RespawnSeconds: 135},
	{ID: "star-golem-001", Name: "Star-Forged Construct", MaxHealth: 125, Attack: 15, Defense: 6, Experience: 310, LootTableID: "star-golem-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -38, Z: 28}, InteractionRange: 3, RespawnSeconds: 150},
	{ID: "dune-scorpion-001", Name: "Armored Dune Scorpion", MaxHealth: 60, Attack: 8, Defense: 3, Experience: 115, LootTableID: "dune-scorpion-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 28, Z: 6}, InteractionRange: 3, RespawnSeconds: 80},
	{ID: "crystalline-mantis-001", Name: "Crystalline Mantis", MaxHealth: 72, Attack: 10, Defense: 3, Experience: 145, LootTableID: "crystalline-mantis-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -24, Z: -28}, InteractionRange: 3, RespawnSeconds: 95},
	{ID: "magma-crusher-001", Name: "Magma Crusher", MaxHealth: 105, Attack: 13, Defense: 6, Experience: 240, LootTableID: "magma-crusher-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 40, Z: -24}, InteractionRange: 3, RespawnSeconds: 130},
	{ID: "rift-specter-001", Name: "Rift Specter", MaxHealth: 118, Attack: 15, Defense: 4, Experience: 280, LootTableID: "rift-specter-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -42, Z: 18}, InteractionRange: 3, RespawnSeconds: 140},
	{ID: "apex-devourer-001", Name: "Apex Rift Devourer", MaxHealth: 145, Attack: 17, Defense: 6, Experience: 360, LootTableID: "apex-devourer-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 42, Z: 42}, InteractionRange: 3, RespawnSeconds: 165},
	// Tier 6-8 Frontier Monsters (+5 enemies -> 24 total)
	{ID: "ironbark-treant-001", Name: "Ironbark Ancient", MaxHealth: 92, Attack: 11, Defense: 5, Experience: 205, LootTableID: "ironbark-treant-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 34, Z: -10}, InteractionRange: 3, RespawnSeconds: 115},
	{ID: "cryo-stalker-001", Name: "Glacial Frost-Warden", MaxHealth: 108, Attack: 13, Defense: 5, Experience: 245, LootTableID: "cryo-stalker-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -30, Z: -30}, InteractionRange: 3, RespawnSeconds: 125},
	{ID: "solar-drake-001", Name: "Solaris Sun-Drake", MaxHealth: 122, Attack: 15, Defense: 6, Experience: 295, LootTableID: "solar-drake-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 14, Z: 36}, InteractionRange: 3, RespawnSeconds: 140},
	{ID: "chrono-stalker-001", Name: "Temporal Phase-Hound", MaxHealth: 136, Attack: 16, Defense: 6, Experience: 335, LootTableID: "chrono-stalker-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -18, Z: -38}, InteractionRange: 3, RespawnSeconds: 155},
	{ID: "abyssal-goliath-001", Name: "Abyssal Chasm Goliath", MaxHealth: 160, Attack: 18, Defense: 7, Experience: 395, LootTableID: "abyssal-goliath-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -40, Z: -36}, InteractionRange: 3, RespawnSeconds: 175},
	// Tier 9-10 Elite Frontier Creatures & Level 10 Multi-Phase Apex Boss (+5 enemies -> 29 total)
	{ID: "dune-krayt-drake-001", Name: "Canyon Krayt-Wyrm", MaxHealth: 175, Attack: 19, Defense: 7, Experience: 450, LootTableID: "dune-krayt-drake-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 46, Z: 10}, InteractionRange: 3, RespawnSeconds: 185},
	{ID: "dathomir-hexweaver-001", Name: "Night-Coven Hexweaver", MaxHealth: 168, Attack: 20, Defense: 6, Experience: 460, LootTableID: "dathomir-hexweaver-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -44, Z: -8}, InteractionRange: 3, RespawnSeconds: 185},
	{ID: "acklay-reaper-001", Name: "Geonosian Chitin-Reaper", MaxHealth: 182, Attack: 20, Defense: 7, Experience: 480, LootTableID: "acklay-reaper-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 22, Z: 44}, InteractionRange: 3, RespawnSeconds: 190},
	{ID: "gorax-behemoth-001", Name: "Endor Crag-Gorax", MaxHealth: 198, Attack: 21, Defense: 8, Experience: 530, LootTableID: "gorax-behemoth-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -44, Z: 42}, InteractionRange: 3, RespawnSeconds: 200},
	{ID: "krayt-jad-sovereign", Name: "Ancient Krayt-Jad Sovereign", MaxHealth: 320, Attack: 0, Defense: 8, Experience: 1200, LootTableID: "krayt-jad-sovereign-harvest", Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: -44}, InteractionRange: 4, RespawnSeconds: 300},
}

func EnemyDefinitions() []EnemyDefinition {
	result := make([]EnemyDefinition, len(enemyDefinitions))
	copy(result, enemyDefinitions)
	return result
}

func FindEnemy(id string) (EnemyDefinition, bool) {
	for _, enemy := range enemyDefinitions {
		if enemy.ID == id {
			return enemy, true
		}
	}
	return EnemyDefinition{}, false
}

func (p *Player) CanAct() error {
	if p.Health <= 0 {
		return ErrPlayerDefeated
	}
	return nil
}

func (p *Player) AttackEnemy(enemy EnemyDefinition, state EnemyState, now time.Time) (CombatResult, EnemyState, error) {
	if err := p.CanAct(); err != nil {
		return CombatResult{}, state, err
	}
	p.EnsureVitalDefaults()
	p.ExpireBuffsIfNeeded(now)
	if p.Position.ZoneID != enemy.Position.ZoneID || math.Hypot(p.Position.X-enemy.Position.X, p.Position.Z-enemy.Position.Z) > enemy.InteractionRange {
		return CombatResult{}, state, ErrEnemyOutOfRange
	}
	if !p.LastCombatAt.IsZero() && now.Sub(p.LastCombatAt) < CombatCooldown {
		return CombatResult{}, state, ErrCombatCooldown
	}
	if !state.Initialized {
		state.Health = enemy.MaxHealth
		state.Initialized = true
	}
	if state.Health > enemy.MaxHealth {
		state.Health = enemy.MaxHealth
	}
	if state.Health <= 0 {
		if now.Before(state.RespawnAt) {
			return CombatResult{}, state, ErrEnemyRespawning
		}
		state.Health = enemy.MaxHealth
		state.RespawnAt = time.Time{}
	}
	attackBonus, _ := p.EquipmentBonuses()
	styleSkillID := "unarmed"
	if weaponID := p.Equipment["main_hand"]; weaponID != "" {
		if weapon, exists := FindItem(weaponID); exists {
			if weaponSkill, known := skillDefinition(weapon.RequiredSkillID); known {
				if weaponSkill.ParentID != "" {
					styleSkillID = weapon.RequiredSkillID
				} else if weapon.RequiredSkillID == "ranged" {
					styleSkillID = "rifles"
				} else if weapon.RequiredSkillID == "melee" {
					styleSkillID = "blades"
				}
			}
		}
	}
	parentSkillID := skillParentID(styleSkillID)
	if parentSkillID == "" {
		parentSkillID = "melee"
	}
	styleLevel := p.Skills[styleSkillID].Level
	parentLevel := p.Skills[parentSkillID].Level
	strengthLevel := p.Skills["strength"].Level
	defenseLevel := p.Skills["defense"].Level
	strengthBonus := 0
	if parentSkillID == "melee" {
		strengthBonus = (strengthLevel - 1) / 10
	}
	damage := 2 + attackBonus + strengthBonus + (styleLevel-1)/20 + (parentLevel-1)/20 - enemy.Defense
	if damage < 1 {
		damage = 1
	}
	if damage > state.Health {
		damage = state.Health
	}
	if styleSkillID == "magic" {
		if p.Mind > 2 {
			p.Mind -= 2
		}
	} else if p.Action > 2 {
		p.Action -= 2
	}
	if parentSkillID == "melee" {
		if _, err := p.AwardExperience("strength", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) {
			return CombatResult{}, state, err
		}
	}
	state.Health -= damage
	result := CombatResult{EnemyID: enemy.ID, SkillID: styleSkillID, ParentSkillID: parentSkillID, EnemyHealth: state.Health, DamageDealt: damage}
	if state.Health == 0 {
		state.RespawnAt = now.Add(time.Duration(enemy.RespawnSeconds) * time.Second)
		result.EnemyDefeated = true
		// Kill rewards are settled by the PostgreSQL encounter transaction.
		// The domain combat calculation must never grant them directly to the
		// final attacker, or group rewards could be duplicated or bypass claims.
	} else {
		_, defenseBonus := p.EquipmentBonuses()
		incoming := enemy.Attack - defenseBonus - (defenseLevel-1)/10
		if incoming < 0 {
			incoming = 0
		}
		if incoming > p.Health {
			incoming = p.Health
		}
		if incoming > 0 {
			if _, err := p.AwardExperience("defense", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) {
				return CombatResult{}, state, err
			}
		}
		p.Health -= incoming
		result.DamageReceived = incoming
		result.PlayerDefeated = p.Health == 0
	}
	p.LastCombatAt = now
	result.PlayerHealth = p.Health
	return result, state, nil
}

// EnemyAttackResult contains server-calculated damage from one completed enemy
// telegraph. The persistence layer must commit this with the enemy AI state.
type EnemyAttackResult struct {
	DamageReceived int
	PlayerDefeated bool
}

// ReceiveEnemyAttack resolves a completed enemy telegraph using the player's
// current equipment and Defense level. enemyPosition is the persisted live
// enemy position, not the static spawn definition.
func (p *Player) ReceiveEnemyAttack(enemy EnemyDefinition, enemyPosition WorldPosition, now time.Time) (EnemyAttackResult, error) {
	if err := p.CanAct(); err != nil {
		return EnemyAttackResult{}, err
	}
	if p.Position.ZoneID != enemyPosition.ZoneID ||
		math.Hypot(p.Position.X-enemyPosition.X, p.Position.Z-enemyPosition.Z) > enemy.InteractionRange {
		return EnemyAttackResult{}, ErrEnemyOutOfRange
	}
	_, defenseBonus := p.EquipmentBonuses()
	defenseLevel := p.Skills["defense"].Level
	incoming := enemy.Attack - defenseBonus - (defenseLevel-1)/10
	if incoming < 0 {
		incoming = 0
	}
	if incoming > p.Health {
		incoming = p.Health
	}
	if incoming > 0 {
		if _, err := p.AwardExperience("defense", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) {
			return EnemyAttackResult{}, err
		}
	}
	p.Health -= incoming
	p.LastCombatAt = now.UTC()
	return EnemyAttackResult{DamageReceived: incoming, PlayerDefeated: p.Health == 0}, nil
}

// ReceiveBossAttack resolves one completed, server-owned boss ability.
// The ability supplies its own range because large-area boss attacks may reach
// farther than an ordinary enemy's melee interaction radius. The caller must
// persist this player mutation atomically with the boss telegraph transition.
func (p *Player) ReceiveBossAttack(damage int, bossPosition WorldPosition, minRange, maxRange float64, now time.Time) (EnemyAttackResult, error) {
	if damage <= 0 || bossPosition.ZoneID == "" || minRange < 0 || maxRange < minRange ||
		math.IsNaN(minRange) || math.IsNaN(maxRange) ||
		math.IsInf(minRange, 0) || math.IsInf(maxRange, 0) {
		return EnemyAttackResult{}, ErrInvalidBossAttack
	}
	if err := p.CanAct(); err != nil {
		return EnemyAttackResult{}, err
	}
	if p.Position.ZoneID != bossPosition.ZoneID {
		return EnemyAttackResult{}, ErrEnemyOutOfRange
	}
	distance := math.Hypot(p.Position.X-bossPosition.X, p.Position.Z-bossPosition.Z)
	if distance < minRange || distance > maxRange {
		return EnemyAttackResult{}, ErrEnemyOutOfRange
	}

	_, defenseBonus := p.EquipmentBonuses()
	defenseLevel := p.Skills["defense"].Level
	incoming := damage - defenseBonus - (defenseLevel-1)/10
	if incoming < 0 {
		incoming = 0
	}
	if incoming > p.Health {
		incoming = p.Health
	}
	if incoming > 0 {
		if _, err := p.AwardExperience("defense", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) {
			return EnemyAttackResult{}, err
		}
	}
	p.Health -= incoming
	p.LastCombatAt = now.UTC()
	return EnemyAttackResult{DamageReceived: incoming, PlayerDefeated: p.Health == 0}, nil
}

func (p *Player) Rest() error {
	zone, ok := FindWorldZone(p.Position.ZoneID)
	if !ok {
		return ErrUnknownWorldZone
	}
	if p.Health > 0 && math.Hypot(p.Position.X-zone.Spawn.X, p.Position.Z-zone.Spawn.Z) > 4 {
		return ErrRestOutOfRange
	}
	if p.Health <= 0 {
		p.Position = zone.Spawn
	}
	p.EnsureVitalDefaults()
	p.Health = p.EffectiveMaxHealth()
	p.Action = p.EffectiveMaxAction()
	p.Mind = p.EffectiveMaxMind()
	p.LastCombatAt = time.Time{}
	return nil
}
