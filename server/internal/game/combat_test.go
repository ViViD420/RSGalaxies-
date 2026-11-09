package game

import (
	"errors"
	"testing"
	"time"
)

func TestPlayerCombatDefeatsEnemyAwardsXPAndLoot(t *testing.T) {
	now:=time.Now().UTC()
	player,err:=NewPlayer("combat-player","CombatWayfarer",now.Add(-time.Second));if err!=nil{t.Fatal(err)}
	player.Position.X=6
	player.Inventory["training_sword"]=1
	if err:=player.EquipItem("training_sword");err!=nil{t.Fatal(err)}
	enemy,ok:=FindEnemy("frontier-rat-001");if !ok{t.Fatal("frontier rat missing")}
	state:=EnemyState{}
	for hit:=0;hit<3;hit++{
		result,next,err:=player.AttackEnemy(enemy,state,now.Add(time.Duration(hit)*time.Second))
		if err!=nil{t.Fatalf("attack %d: %v",hit,err)}
		state=next
		if hit<2&&result.EnemyDefeated{t.Fatalf("enemy defeated too early on hit %d",hit+1)}
		if hit==2{
			if !result.EnemyDefeated||result.Experience!=25||result.SkillID!="blades"||result.ParentSkillID!="melee"||result.ParentExperience!=6||result.LootItemID!="rat_pelt"||player.Inventory["rat_pelt"]!=1{t.Fatalf("combat result=%+v player=%+v",result,player)}
			if player.Skills["blades"].Experience != 25 || player.Skills["melee"].Experience != 6 || player.Skills["strength"].Experience != 3 || player.Skills["defense"].Experience != 2 || state.RespawnAt.IsZero() { t.Fatalf("combat progression/state blades=%+v melee=%+v strength=%+v defense=%+v state=%+v", player.Skills["blades"], player.Skills["melee"], player.Skills["strength"], player.Skills["defense"], state) }
		}
	}
	if _,_,err:=player.AttackEnemy(enemy,state,now.Add(3*time.Second));!errors.Is(err,ErrEnemyRespawning){t.Fatalf("respawning enemy error=%v",err)}
	if _,next,err:=player.AttackEnemy(enemy,state,now.Add(33*time.Second));err!=nil||next.Health!=enemy.MaxHealth-4{t.Fatalf("enemy respawn attack state=%+v err=%v",next,err)}
}

func TestCombatCooldownDefeatAndSafeZoneRest(t *testing.T) {
	now:=time.Now().UTC()
	player,err:=NewPlayer("combat-player-2","CombatWayfarer",now);if err!=nil{t.Fatal(err)}
	player.Position.X=6
	enemy,_:=FindEnemy("frontier-rat-001")
	if _,_,err:=player.AttackEnemy(enemy,EnemyState{},now);!errors.Is(err,ErrCombatCooldown){t.Fatalf("initial cooldown error=%v",err)}
	player.LastCombatAt=now.Add(-2*time.Second)
	player.Health=1
	enemy.Position=player.Position
	result,_,err:=player.AttackEnemy(enemy,EnemyState{},now)
	if err!=nil{t.Fatal(err)}
	if !result.PlayerDefeated||player.Health!=0{t.Fatalf("player defeat result=%+v health=%d",result,player.Health)}
	if _,_,err:=player.AttackEnemy(enemy,EnemyState{},now.Add(2*time.Second));!errors.Is(err,ErrPlayerDefeated){t.Fatalf("defeated player attack error=%v",err)}
	player.Health=10
	if err:=player.Rest();!errors.Is(err,ErrRestOutOfRange){t.Fatalf("rest while alive away from spawn error=%v",err)}
	player.Health=0
	if err:=player.Rest();err!=nil{t.Fatalf("death respawn: %v",err)}
	if player.Position.X!=0||player.Position.Z!=0{t.Fatalf("defeated player did not respawn at spawn: %+v",player.Position)}
	if player.Health!=MaxPlayerHealth{t.Fatalf("rest health=%d want=%d",player.Health,MaxPlayerHealth)}
}

func TestStrengthAndDefenseCoSkillsAffectCombat(t *testing.T) {
	now := time.Now().UTC()
	player, err := NewPlayer("combat-coskills", "CoSkillWayfarer", now)
	if err != nil { t.Fatal(err) }
	player.Position.X = 6
	player.Skills["strength"] = SkillProgress{Level: 11}
	player.Skills["defense"] = SkillProgress{Level: 11}
	enemy, ok := FindEnemy("frontier-rat-001")
	if !ok { t.Fatal("frontier rat missing") }
	enemy.Position = player.Position

	result, _, err := player.AttackEnemy(enemy, EnemyState{}, now.Add(2*time.Second))
	if err != nil { t.Fatal(err) }
	if result.SkillID != "unarmed" || result.ParentSkillID != "melee" {
		t.Fatalf("unarmed action attribution = %s/%s", result.SkillID, result.ParentSkillID)
	}
	if result.DamageDealt != 3 { t.Fatalf("strength-adjusted damage = %d, want 3", result.DamageDealt) }
	if result.DamageReceived != 1 { t.Fatalf("defense-adjusted incoming damage = %d, want 1", result.DamageReceived) }
	if player.Skills["strength"].Experience != 1 || player.Skills["defense"].Experience != 1 {
		t.Fatalf("co-skill XP not awarded: strength=%+v defense=%+v", player.Skills["strength"], player.Skills["defense"])
	}
}
