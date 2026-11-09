package game

import (
	"errors"
	"math"
	"time"
)

const (
	MaxPlayerHealth = 100
	CombatCooldown = time.Second
)

var (
	ErrUnknownEnemy = errors.New("unknown enemy")
	ErrEnemyOutOfRange = errors.New("enemy is out of range")
	ErrEnemyRespawning = errors.New("enemy is respawning")
	ErrCombatCooldown = errors.New("combat action is cooling down")
	ErrPlayerDefeated = errors.New("player is defeated")
	ErrRestOutOfRange = errors.New("resting is only possible near a safe spawn")
)

type EnemyDefinition struct {
	ID string `json:"id"`
	Name string `json:"name"`
	MaxHealth int `json:"maxHealth"`
	Attack int `json:"attack"`
	Defense int `json:"defense"`
	Experience int `json:"experience"`
	LootItemID string `json:"lootItemId"`
	LootQuantity int `json:"lootQuantity"`
	Position WorldPosition `json:"position"`
	InteractionRange float64 `json:"interactionRange"`
	RespawnSeconds int `json:"respawnSeconds"`
}

type EnemyState struct {
	Health int
	RespawnAt time.Time
	Initialized bool
}

type CombatResult struct {
	EnemyID string `json:"enemyId"`
	SkillID string `json:"skillId"`
	ParentSkillID string `json:"parentSkillId,omitempty"`
	EnemyHealth int `json:"enemyHealth"`
	PlayerHealth int `json:"playerHealth"`
	DamageDealt int `json:"damageDealt"`
	DamageReceived int `json:"damageReceived"`
	EnemyDefeated bool `json:"enemyDefeated"`
	PlayerDefeated bool `json:"playerDefeated"`
	Experience int `json:"experience"`
	LootItemID string `json:"lootItemId,omitempty"`
	LootQuantity int `json:"lootQuantity,omitempty"`
}

var enemyDefinitions = []EnemyDefinition{
	{ID:"frontier-rat-001",Name:"Frontier Rat",MaxHealth:12,Attack:2,Defense:0,Experience:25,LootItemID:"rat_pelt",LootQuantity:1,Position:WorldPosition{ZoneID:DefaultWorldZoneID,X:8,Z:0},InteractionRange:3,RespawnSeconds:30},
	{ID:"frontier-boar-001",Name:"Wild Boar",MaxHealth:35,Attack:6,Defense:2,Experience:60,LootItemID:"boar_hide",LootQuantity:1,Position:WorldPosition{ZoneID:DefaultWorldZoneID,X:-8,Z:0},InteractionRange:3,RespawnSeconds:60},
}

func EnemyDefinitions() []EnemyDefinition {
	result:=make([]EnemyDefinition,len(enemyDefinitions));copy(result,enemyDefinitions);return result
}

func FindEnemy(id string)(EnemyDefinition,bool){
	for _,enemy:=range enemyDefinitions{if enemy.ID==id{return enemy,true}}
	return EnemyDefinition{},false
}

func (p *Player) CanAct() error {
	if p.Health <= 0 { return ErrPlayerDefeated }
	return nil
}

func (p *Player) AttackEnemy(enemy EnemyDefinition,state EnemyState,now time.Time)(CombatResult,EnemyState,error){
	if err:=p.CanAct();err!=nil{return CombatResult{},state,err}
	if p.Position.ZoneID!=enemy.Position.ZoneID||math.Hypot(p.Position.X-enemy.Position.X,p.Position.Z-enemy.Position.Z)>enemy.InteractionRange{return CombatResult{},state,ErrEnemyOutOfRange}
	if !p.LastCombatAt.IsZero()&&now.Sub(p.LastCombatAt)<CombatCooldown{return CombatResult{},state,ErrCombatCooldown}
	if !state.Initialized{state.Health=enemy.MaxHealth;state.Initialized=true}
	if state.Health>enemy.MaxHealth{state.Health=enemy.MaxHealth}
	if state.Health<=0{
		if now.Before(state.RespawnAt){return CombatResult{},state,ErrEnemyRespawning}
		state.Health=enemy.MaxHealth
		state.RespawnAt=time.Time{}
	}
	attackBonus,_:=p.EquipmentBonuses()
	styleSkillID := "unarmed"
	if weaponID := p.Equipment["main_hand"]; weaponID != "" {
		if weapon, exists := FindItem(weaponID); exists {
			if weaponSkill, known := skillDefinition(weapon.RequiredSkillID); known && weaponSkill.ParentID != "" { styleSkillID = weapon.RequiredSkillID }
		}
	}
	parentSkillID := skillParentID(styleSkillID)
	styleLevel := p.Skills[styleSkillID].Level
	meleeLevel := p.Skills["melee"].Level
	strengthLevel := p.Skills["strength"].Level
	defenseLevel := p.Skills["defense"].Level
	damage := 2 + attackBonus + (strengthLevel-1)/10 + (styleLevel-1)/20 + (meleeLevel-1)/20 - enemy.Defense
	if damage<1{damage=1}
	if damage>state.Health{damage=state.Health}
	if _, err := p.AwardExperience("strength", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) { return CombatResult{}, state, err }
	state.Health-=damage
	result:=CombatResult{EnemyID:enemy.ID,SkillID:styleSkillID,ParentSkillID:parentSkillID,EnemyHealth:state.Health,DamageDealt:damage}
	if state.Health==0{
		state.RespawnAt=now.Add(time.Duration(enemy.RespawnSeconds)*time.Second)
		result.EnemyDefeated=true
		result.Experience=enemy.Experience
		if _, err := p.AwardExperience(styleSkillID, enemy.Experience); err != nil && !errors.Is(err, ErrExperienceOverflow) { return CombatResult{}, state, err }
		if parentSkillID != "" {
			parentXP := enemy.Experience / 4
			if parentXP < 1 { parentXP = 1 }
			if _, err := p.AwardExperience(parentSkillID, parentXP); err != nil && !errors.Is(err, ErrExperienceOverflow) { return CombatResult{}, state, err }
		}
		if enemy.LootItemID!=""&&enemy.LootQuantity>0{
			item,ok:=FindItem(enemy.LootItemID)
			if ok&&p.Inventory[enemy.LootItemID]<=item.MaxStack-enemy.LootQuantity{
				if p.Inventory==nil{p.Inventory=make(map[string]int)}
				p.Inventory[enemy.LootItemID]+=enemy.LootQuantity
				result.LootItemID=enemy.LootItemID
				result.LootQuantity=enemy.LootQuantity
			}
		}
	}else{
		_,defenseBonus:=p.EquipmentBonuses()
		incoming:=enemy.Attack-defenseBonus-(defenseLevel-1)/10
		if incoming<0{incoming=0}
		if incoming>p.Health{incoming=p.Health}
		if incoming > 0 { if _, err := p.AwardExperience("defense", 1); err != nil && !errors.Is(err, ErrExperienceOverflow) { return CombatResult{}, state, err } }
		p.Health-=incoming
		result.DamageReceived=incoming
		result.PlayerDefeated=p.Health==0
	}
	p.LastCombatAt=now
	result.PlayerHealth=p.Health
	return result,state,nil
}

func (p *Player) Rest() error {
	zone,ok:=FindWorldZone(p.Position.ZoneID);if !ok{return ErrUnknownWorldZone}
	if p.Health>0&&math.Hypot(p.Position.X-zone.Spawn.X,p.Position.Z-zone.Spawn.Z)>4{return ErrRestOutOfRange}
	if p.Health<=0{p.Position=zone.Spawn}
	p.Health=MaxPlayerHealth
	p.LastCombatAt=time.Time{}
	return nil
}
