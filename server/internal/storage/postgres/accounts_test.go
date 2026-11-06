package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"sync"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
)

func cleanupTestAccounts(db *sql.DB, accountIDs ...string) {
	for _, accountID := range accountIDs {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM players WHERE account_id=$1", accountID)
		_, _ = db.ExecContext(context.Background(), "DELETE FROM accounts WHERE id=$1", accountID)
	}
}

func TestPostgresAccountRegistrationLoginAndSessionRevocation(t *testing.T) {
	db:=integrationDatabase(t)
	repo:=NewAccountRepository(db)
	ctx:=context.Background()
	now:=time.Now().UTC().Truncate(time.Microsecond)
	email:=fmt.Sprintf("account-%d@example.test",now.UnixNano())
	account,token,err:=repo.Register(ctx,email,"correct-horse-battery-staple",now)
	if err!=nil{t.Fatalf("register: %v",err)}
	if account.ID=="" || account.Email!=email || token==""{t.Fatalf("invalid registration result: %+v",account)}
	t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})

	authenticated,err:=repo.AuthenticateSession(ctx,token,now.Add(time.Minute))
	if err!=nil || authenticated.ID!=account.ID{t.Fatalf("authenticate registration session: account=%+v err=%v",authenticated,err)}
	var lastUsed time.Time
	if err:=db.QueryRowContext(ctx,"SELECT last_used_at FROM account_sessions WHERE token_hash=(SELECT token_hash FROM account_sessions WHERE account_id=$1 ORDER BY created_at LIMIT 1)",account.ID).Scan(&lastUsed);err!=nil{t.Fatal(err)}
	if !lastUsed.Equal(now.Add(time.Minute)){t.Fatalf("session last-used timestamp=%s want=%s",lastUsed,now.Add(time.Minute))}
	// Repeated API authentication within the write-coalescing window must not
	// update the session row again.
	if _,err:=repo.AuthenticateSession(ctx,token,now.Add(90*time.Second));err!=nil{t.Fatalf("repeat session authentication: %v",err)}
	var unchanged time.Time
	if err:=db.QueryRowContext(ctx,"SELECT last_used_at FROM account_sessions WHERE token_hash=(SELECT token_hash FROM account_sessions WHERE account_id=$1 ORDER BY created_at LIMIT 1)",account.ID).Scan(&unchanged);err!=nil{t.Fatal(err)}
	if !unchanged.Equal(lastUsed){t.Fatalf("session activity write was not coalesced: before=%s after=%s",lastUsed,unchanged)}
	if _,err:=repo.AuthenticateSession(ctx,token,now.Add(2*time.Minute+time.Second));err!=nil{t.Fatalf("session authentication after coalescing window: %v",err)}
	var refreshed time.Time
	if err:=db.QueryRowContext(ctx,"SELECT last_used_at FROM account_sessions WHERE token_hash=(SELECT token_hash FROM account_sessions WHERE account_id=$1 ORDER BY created_at LIMIT 1)",account.ID).Scan(&refreshed);err!=nil{t.Fatal(err)}
	if !refreshed.Equal(now.Add(2*time.Minute+time.Second)){t.Fatalf("session activity not refreshed: got=%s",refreshed)}
	loggedIn,loginToken,err:=repo.Login(ctx,"  "+email+"  ","correct-horse-battery-staple",now.Add(2*time.Minute))
	if err!=nil || loggedIn.ID!=account.ID || loginToken==""{t.Fatalf("login: account=%+v err=%v",loggedIn,err)}
	if _,_,err:=repo.Login(ctx,email,"incorrect-password",now);!errors.Is(err,ErrInvalidCredentials){t.Fatalf("wrong password error=%v",err)}
	if err:=repo.RevokeSession(ctx,token,now.Add(3*time.Minute));err!=nil{t.Fatal(err)}
	if _,err:=repo.AuthenticateSession(ctx,token,now.Add(4*time.Minute));!errors.Is(err,ErrSessionNotFound){t.Fatalf("revoked session error=%v",err)}
	if _,err:=repo.AuthenticateSession(ctx,loginToken,now.Add(4*time.Minute));err!=nil{t.Fatalf("second session should remain valid: %v",err)}
	if _,err:=repo.AuthenticateSession(ctx,loginToken,now.Add(SessionLifetime+3*time.Minute));!errors.Is(err,ErrSessionNotFound){t.Fatalf("expired session error=%v",err)}
}

func TestPostgresAccountEmailUniquenessAndValidation(t *testing.T) {
	db:=integrationDatabase(t)
	repo:=NewAccountRepository(db)
	ctx:=context.Background()
	now:=time.Now().UTC()
	email:=fmt.Sprintf("unique-%d@example.test",now.UnixNano())
	account,_,err:=repo.Register(ctx,email,"correct-horse-battery-staple",now)
	if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})
	if _,_,err:=repo.Register(ctx,"  "+email+" ","correct-horse-battery-staple",now);!errors.Is(err,ErrAccountExists){t.Fatalf("duplicate email error=%v",err)}
	if _,_,err:=repo.Register(ctx,"not-an-email","correct-horse-battery-staple",now);!errors.Is(err,ErrInvalidEmail){t.Fatalf("invalid email error=%v",err)}
}

func TestPostgresCharactersAreScopedToOwningAccount(t *testing.T) {
	db:=integrationDatabase(t)
	repo:=NewAccountRepository(db)
	ctx:=context.Background()
	now:=time.Now().UTC()
	first,_,err:=repo.Register(ctx,fmt.Sprintf("owner-a-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now)
	if err!=nil{t.Fatal(err)}
	second,_,err:=repo.Register(ctx,fmt.Sprintf("owner-b-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now)
	if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID)})
	player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("owned-%d",now.UnixNano())),"OwnedWayfarer",now)
	if err!=nil{t.Fatal(err)}
	if err:=repo.CreateCharacter(ctx,first.ID,player);err!=nil{t.Fatalf("create character: %v",err)}
	owned,err:=repo.GetCharacter(ctx,first.ID,player.ID)
	if err!=nil || owned.ID!=player.ID{t.Fatalf("owner lookup failed: %+v err=%v",owned,err)}
	if _,err:=repo.GetCharacter(ctx,second.ID,player.ID);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("cross-account lookup error=%v",err)}
	other,err:=repo.ListCharacters(ctx,second.ID)
	if err!=nil || len(other)!=0{t.Fatalf("other account list=%v err=%v",other,err)}
	own,err:=repo.ListCharacters(ctx,first.ID)
	if err!=nil || len(own)!=1{t.Fatalf("owner list=%v err=%v",own,err)}
}

func TestPostgresOwnedGameplayMutationsEnforceOwnership(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 first,_,err:=repo.Register(ctx,fmt.Sprintf("game-owner-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 second,_,err:=repo.Register(ctx,fmt.Sprintf("game-other-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID)})
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("game-char-%d",now.UnixNano())),"GameWayfarer",now);if err!=nil{t.Fatal(err)}
 player.Position.Z=2
 player.Inventory["softwood_log"]=2
 if err:=repo.CreateCharacter(ctx,first.ID,player);err!=nil{t.Fatal(err)}
 if _,_,err:=repo.AwardExperienceForAccount(ctx,second.ID,player.ID,"mining",50);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign XP mutation error=%v",err)}
 if _,_,err:=repo.CraftForAccount(ctx,second.ID,player.ID,"basic-plank",now.Add(time.Second));!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign craft error=%v",err)}
 if _,err:=repo.TransferBankItemForAccount(ctx,second.ID,player.ID,"softwood_log",1,true);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign bank deposit error=%v",err)}
 if deposited,err:=repo.TransferBankItemForAccount(ctx,first.ID,player.ID,"softwood_log",1,true);err!=nil||deposited.Bank["softwood_log"]!=1||deposited.Inventory["softwood_log"]!=1{t.Fatalf("owner bank deposit player=%+v err=%v",deposited,err)}
 persistedBank,err:=repo.GetCharacter(ctx,first.ID,player.ID);if err!=nil||persistedBank.Bank["softwood_log"]!=1{t.Fatalf("bank persistence player=%+v err=%v",persistedBank,err)}
 if withdrawn,err:=repo.TransferBankItemForAccount(ctx,first.ID,player.ID,"softwood_log",1,false);err!=nil||withdrawn.Bank["softwood_log"]!=0||withdrawn.Inventory["softwood_log"]!=2{t.Fatalf("owner bank withdrawal player=%+v err=%v",withdrawn,err)}
 if _,_,err:=repo.TradeItemForAccount(ctx,second.ID,player.ID,"rough_stone",1,true);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign market trade error=%v",err)}
 bought,cost,err:=repo.TradeItemForAccount(ctx,first.ID,player.ID,"rough_stone",2,true);if err!=nil||cost!=6||bought.Coins!=94||bought.Inventory["rough_stone"]!=2{t.Fatalf("owner market buy player=%+v cost=%d err=%v",bought,cost,err)}
 sold,revenue,err:=repo.TradeItemForAccount(ctx,first.ID,player.ID,"rough_stone",1,false);if err!=nil||revenue!=1||sold.Coins!=95||sold.Inventory["rough_stone"]!=1{t.Fatalf("owner market sell player=%+v revenue=%d err=%v",sold,revenue,err)}
 if _,err:=repo.EquipItemForAccount(ctx,second.ID,player.ID,"training_sword");!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign equip error=%v",err)}
 if _,_,err:=repo.TradeItemForAccount(ctx,first.ID,player.ID,"training_sword",1,true);err!=nil{t.Fatalf("buy training sword: %v",err)}
 equipped,err:=repo.EquipItemForAccount(ctx,first.ID,player.ID,"training_sword");if err!=nil||equipped.Equipment["main_hand"]!="training_sword"||equipped.Inventory["training_sword"]!=0{t.Fatalf("equip owned item player=%+v err=%v",equipped,err)}
 persistedEquipment,err:=repo.GetCharacter(ctx,first.ID,player.ID);if err!=nil||persistedEquipment.Equipment["main_hand"]!="training_sword"{t.Fatalf("equipment persistence player=%+v err=%v",persistedEquipment,err)}
 unequipped,err:=repo.UnequipItemForAccount(ctx,first.ID,player.ID,"main_hand");if err!=nil||unequipped.Equipment["main_hand"]!=""||unequipped.Inventory["training_sword"]!=1{t.Fatalf("unequip owned item player=%+v err=%v",unequipped,err)}
 if _,_,err:=repo.AwardExperienceForAccount(ctx,first.ID,player.ID,"mining",50);err!=nil{t.Fatalf("owner XP award: %v",err)}
 if _,err:=repo.MovePlayerForAccount(ctx,first.ID,player.ID,0.1,2,now.Add(time.Second));err!=nil{t.Fatalf("owner movement: %v",err)}
 crafted,result,err:=repo.CraftForAccount(ctx,first.ID,player.ID,"basic-plank",now.Add(2*time.Second));if err!=nil{t.Fatalf("owner craft: %v",err)}
 if result.Experience!=30||crafted.Inventory["wooden_plank"]!=1||crafted.Inventory["softwood_log"]!=0{t.Fatalf("owner craft result=%+v player=%+v",result,crafted)}
 persisted,err:=repo.GetCharacter(ctx,first.ID,player.ID);if err!=nil{t.Fatal(err)}
 if persisted.Skills["crafting"].Experience!=30{t.Fatalf("owned crafting XP not persisted: %+v",persisted.Skills["crafting"])}
}

func TestPostgresMigrationsAreIdempotent(t *testing.T) {
 db:=integrationDatabase(t)
 ctx:=context.Background()
 if err:=Migrate(ctx,db);err!=nil{t.Fatalf("second migration: %v",err)}
 var count int
 if err:=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM schema_migrations WHERE version IN (1,2,3,4,5,6,7,8,9)").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=9{t.Fatalf("migration versions applied=%d want=9",count)}
}

func TestPostgresOwnedGatheringLocksSharedNodeAndChecksOwner(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 first,_,err:=repo.Register(ctx,fmt.Sprintf("gather-owner-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 second,_,err:=repo.Register(ctx,fmt.Sprintf("gather-other-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID);_,_=db.ExecContext(context.Background(),"DELETE FROM gathering_node_state WHERE node_id=$1","tree-001")})
 _,_=db.ExecContext(ctx,"DELETE FROM gathering_node_state WHERE node_id=$1","tree-001")
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("gather-owned-%d",now.UnixNano())),"GatherOwned",now);if err!=nil{t.Fatal(err)}
 player.Position.X=2
 if err:=repo.CreateCharacter(ctx,first.ID,player);err!=nil{t.Fatal(err)}
 if _,_,err:=repo.GatherForAccount(ctx,second.ID,player.ID,"tree-001",now.Add(time.Second));!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign gather error=%v",err)}
 gathered,result,err:=repo.GatherForAccount(ctx,first.ID,player.ID,"tree-001",now.Add(time.Second));if err!=nil{t.Fatalf("owner gather: %v",err)}
 if result.ItemID!="softwood_log"||gathered.Inventory["softwood_log"]!=1||gathered.Skills["woodcutting"].Experience!=25{t.Fatalf("owned gather result=%+v player=%+v",result,gathered)}
}

func TestPostgresAccountSessionLimitPrunesOldestSessions(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,firstToken,err:=repo.Register(ctx,fmt.Sprintf("session-cap-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})
 for i:=1;i<=MaxActiveSessionsPerAccount;i++ {
  _,hash,err:=auth.NewSessionToken();if err!=nil{t.Fatal(err)}
  tx,err:=db.BeginTx(ctx,nil);if err!=nil{t.Fatal(err)}
  if err:=createSession(ctx,tx,account.ID,hash[:],now.Add(time.Duration(i)*time.Second));err!=nil{_ = tx.Rollback();t.Fatal(err)}
  if err:=tx.Commit();err!=nil{t.Fatal(err)}
 }
 if _,err:=repo.AuthenticateSession(ctx,firstToken,now.Add(time.Duration(MaxActiveSessionsPerAccount)*time.Second));!errors.Is(err,ErrSessionNotFound){t.Fatalf("oldest session should be pruned, got %v",err)}
 var active int
 if err:=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM account_sessions WHERE account_id=$1 AND revoked_at IS NULL AND expires_at>$2",account.ID,now.Add(time.Duration(MaxActiveSessionsPerAccount)*time.Second)).Scan(&active);err!=nil{t.Fatal(err)}
 if active!=MaxActiveSessionsPerAccount{t.Fatalf("active sessions=%d want=%d",active,MaxActiveSessionsPerAccount)}
}

func TestPostgresConcurrentBankDepositsDoNotLoseItems(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,_,err:=repo.Register(ctx,fmt.Sprintf("bank-race-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("bank-race-char-%d",now.UnixNano())),"BankRace",now);if err!=nil{t.Fatal(err)}
 player.Inventory["softwood_log"]=10
 if err:=repo.CreateCharacter(ctx,account.ID,player);err!=nil{t.Fatal(err)}
 var group sync.WaitGroup
 errorsSeen:=make(chan error,10)
 for i:=0;i<10;i++{group.Add(1);go func(){defer group.Done();_,err:=repo.TransferBankItemForAccount(ctx,account.ID,player.ID,"softwood_log",1,true);errorsSeen<-err}()}
 group.Wait();close(errorsSeen)
 for err:=range errorsSeen{if err!=nil{t.Fatalf("concurrent bank deposit: %v",err)}}
 persisted,err:=repo.GetCharacter(ctx,account.ID,player.ID);if err!=nil{t.Fatal(err)}
 if persisted.Inventory["softwood_log"]!=0||persisted.Bank["softwood_log"]!=10{t.Fatalf("concurrent transfer lost updates: inventory=%v bank=%v",persisted.Inventory,persisted.Bank)}
}

func TestPostgresOwnedCopperMiningPersistsNodeAndRewards(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,_,err:=repo.Register(ctx,fmt.Sprintf("copper-miner-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID);_,_=db.ExecContext(context.Background(),"DELETE FROM gathering_node_state WHERE node_id=$1","copper-001")})
 _,_=db.ExecContext(ctx,"DELETE FROM gathering_node_state WHERE node_id=$1","copper-001")
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("copper-miner-char-%d",now.UnixNano())),"CopperMiner",now);if err!=nil{t.Fatal(err)}
 player.Position.Z=-5
 if err:=repo.CreateCharacter(ctx,account.ID,player);err!=nil{t.Fatal(err)}
 for i:=1;i<=2;i++{gathered,result,err:=repo.GatherForAccount(ctx,account.ID,player.ID,"copper-001",now.Add(time.Duration(i)*time.Second));if err!=nil{t.Fatalf("copper gather %d: %v",i,err)};if result.ItemID!="copper_ore"||gathered.Inventory["copper_ore"]!=i{t.Fatalf("copper gather %d result=%+v player=%+v",i,result,gathered)}}
 persisted,err:=repo.GetCharacter(ctx,account.ID,player.ID);if err!=nil{t.Fatal(err)}
 if persisted.Inventory["copper_ore"]!=2||persisted.Skills["mining"].Experience!=60{t.Fatalf("copper rewards not persisted: %+v",persisted)}
 var charges int;var initialized bool
 if err:=db.QueryRowContext(ctx,"SELECT charges,initialized FROM gathering_node_state WHERE node_id=$1","copper-001").Scan(&charges,&initialized);err!=nil{t.Fatal(err)}
 if !initialized||charges!=0{t.Fatalf("copper node state initialized=%v charges=%d",initialized,charges)}
}

func TestPostgresOwnedCopperSmithingPersistsRecipes(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,_,err:=repo.Register(ctx,fmt.Sprintf("copper-smith-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("copper-smith-char-%d",now.UnixNano())),"CopperSmith",now);if err!=nil{t.Fatal(err)}
 player.Position.Z=8;player.Inventory["copper_ore"]=2
 if err:=repo.CreateCharacter(ctx,account.ID,player);err!=nil{t.Fatal(err)}
 smelted,_,err:=repo.CraftForAccount(ctx,account.ID,player.ID,"copper-bar",now.Add(time.Second));if err!=nil{t.Fatalf("smelt copper: %v",err)}
 if smelted.Inventory["copper_ore"]!=0||smelted.Inventory["copper_bar"]!=1||smelted.Skills["smithing"].Experience!=35{t.Fatalf("smelting state=%+v",smelted)}
 smelted.Inventory["copper_bar"]=3;smelted.Inventory["wooden_plank"]=1
 // Seed the additional recipe inputs as an explicit test fixture.
 if _,err:=db.ExecContext(ctx,"UPDATE players SET inventory=$3::jsonb WHERE account_id=$1 AND id=$2",account.ID,string(player.ID),`{"copper_bar":3,"wooden_plank":1}`);err!=nil{t.Fatal(err)}
 forged,_,err:=repo.CraftForAccount(ctx,account.ID,player.ID,"copper-sword",now.Add(2*time.Second));if err!=nil{t.Fatalf("forge training sword: %v",err)}
 persisted,err:=repo.GetCharacter(ctx,account.ID,player.ID);if err!=nil{t.Fatal(err)}
 if forged.Inventory["training_sword"]!=1||persisted.Inventory["training_sword"]!=1||persisted.Skills["smithing"].Experience!=95{t.Fatalf("forged sword state=%+v persisted=%+v",forged,persisted)}
}

func TestPostgresCombatPersistsEnemyPlayerLootAndRestState(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,_,err:=repo.Register(ctx,fmt.Sprintf("combat-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID);_,_=db.ExecContext(context.Background(),"DELETE FROM enemy_spawn_state WHERE enemy_id=$1","frontier-rat-001")})
 _,_=db.ExecContext(ctx,"DELETE FROM enemy_spawn_state WHERE enemy_id=$1","frontier-rat-001")
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("combat-char-%d",now.UnixNano())),"CombatWayfarer",now.Add(-2*time.Second));if err!=nil{t.Fatal(err)}
 player.Position.X=6;player.Equipment["main_hand"]="training_sword"
 if err:=repo.CreateCharacter(ctx,account.ID,player);err!=nil{t.Fatal(err)}
 var last game.Player
 var final game.CombatResult
 for hit:=0;hit<3;hit++{last,final,err=repo.AttackEnemyForAccount(ctx,account.ID,player.ID,"frontier-rat-001",now.Add(time.Duration(hit)*time.Second));if err!=nil{t.Fatalf("combat hit %d: %v",hit,err)}}
 if !final.EnemyDefeated||final.Experience!=25||final.LootItemID!="rat_pelt"||last.Inventory["rat_pelt"]!=1||last.Skills["combat"].Experience!=25{t.Fatalf("combat defeat result=%+v player=%+v",final,last)}
 persisted,err:=repo.GetCharacter(ctx,account.ID,player.ID);if err!=nil{t.Fatal(err)}
 if persisted.Health!=96||persisted.Inventory["rat_pelt"]!=1||persisted.Skills["combat"].Experience!=25{t.Fatalf("combat player state not persisted: %+v",persisted)}
 if _,err:=repo.RestForAccount(ctx,account.ID,player.ID);!errors.Is(err,game.ErrRestOutOfRange){t.Fatalf("rest outside safe zone error=%v",err)}
 if _,err:=db.ExecContext(ctx,"UPDATE players SET position_x=0,position_z=0 WHERE id=$1",string(player.ID));err!=nil{t.Fatal(err)}
 rested,err:=repo.RestForAccount(ctx,account.ID,player.ID);if err!=nil{t.Fatalf("safe-zone rest: %v",err)}
 if rested.Health!=game.MaxPlayerHealth{t.Fatalf("rest health=%d want=%d",rested.Health,game.MaxPlayerHealth)}
 var enemyHealth int;var initialized bool
 if err:=db.QueryRowContext(ctx,"SELECT health,initialized FROM enemy_spawn_state WHERE enemy_id=$1","frontier-rat-001").Scan(&enemyHealth,&initialized);err!=nil{t.Fatal(err)}
 if !initialized||enemyHealth!=0{t.Fatalf("persisted enemy state initialized=%v health=%d",initialized,enemyHealth)}
}

func TestPostgresLocalChatIsZoneScopedAndRateLimited(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 first,_,err:=repo.Register(ctx,fmt.Sprintf("chat-one-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 second,_,err:=repo.Register(ctx,fmt.Sprintf("chat-two-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID)})
 playerOne,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("chat-char-one-%d",now.UnixNano())),"ChatOne",now);if err!=nil{t.Fatal(err)}
 playerTwo,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("chat-char-two-%d",now.UnixNano())),"ChatTwo",now);if err!=nil{t.Fatal(err)}
 if err:=repo.CreateCharacter(ctx,first.ID,playerOne);err!=nil{t.Fatal(err)}
 if err:=repo.CreateCharacter(ctx,second.ID,playerTwo);err!=nil{t.Fatal(err)}
 sent,err:=repo.SendLocalChatForAccount(ctx,first.ID,playerOne.ID,"  Hello frontier!  ",now);if err!=nil{t.Fatalf("send local chat: %v",err)}
 if sent.ID==0||sent.Message!="Hello frontier!"||sent.CharacterName!="ChatOne"{t.Fatalf("sent chat message=%+v",sent)}
 visible,err:=repo.ListLocalChatForAccount(ctx,second.ID,playerTwo.ID,0,50);if err!=nil{t.Fatalf("read zone chat: %v",err)}
 if len(visible)!=1||visible[0].ID!=sent.ID||visible[0].Message!="Hello frontier!"{t.Fatalf("zone chat visibility=%+v",visible)}
 if _,err:=repo.SendLocalChatForAccount(ctx,first.ID,playerOne.ID,"too soon",now.Add(time.Second));!errors.Is(err,game.ErrChatCooldown){t.Fatalf("chat cooldown error=%v",err)}
 after,err:=repo.ListLocalChatForAccount(ctx,second.ID,playerTwo.ID,sent.ID,50);if err!=nil||len(after)!=0{t.Fatalf("chat cursor result=%+v err=%v",after,err)}
 if _,err:=repo.ListLocalChatForAccount(ctx,second.ID,playerOne.ID,0,50);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign chat read error=%v",err)}
	for index:=0;index<55;index++ {
		messageAt:=now.Add(time.Duration(index+1)*game.ChatCooldown)
		if _,err:=repo.SendLocalChatForAccount(ctx,second.ID,playerTwo.ID,fmt.Sprintf("history message %02d",index),messageAt);err!=nil{t.Fatalf("send history message %d: %v",index,err)}
	}
	latest,err:=repo.ListLocalChatForAccount(ctx,first.ID,playerOne.ID,0,50)
	if err!=nil{t.Fatalf("load initial bounded chat history: %v",err)}
	if len(latest)!=50||latest[0].ID<=sent.ID||latest[len(latest)-1].ID<=latest[0].ID{t.Fatalf("initial chat history should contain newest 50 ordered messages: count=%d first=%+v last=%+v",len(latest),latest[0],latest[len(latest)-1])}
	newer,err:=repo.ListLocalChatForAccount(ctx,first.ID,playerOne.ID,latest[len(latest)-1].ID,50)
	if err!=nil||len(newer)!=0{t.Fatalf("chat cursor after newest history result=%+v err=%v",newer,err)}

}

func TestPostgresOwnedConsumableUsePersistsHealthAndInventory(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 account,_,err:=repo.Register(ctx,fmt.Sprintf("potion-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){cleanupTestAccounts(db,account.ID)})
 player,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("potion-char-%d",now.UnixNano())),"PotionWayfarer",now);if err!=nil{t.Fatal(err)}
 player.Health=60;player.Inventory["minor_healing_potion"]=2
 if err:=repo.CreateCharacter(ctx,account.ID,player);err!=nil{t.Fatal(err)}
 used,healed,err:=repo.UseItemForAccount(ctx,account.ID,player.ID,"minor_healing_potion");if err!=nil{t.Fatalf("use potion: %v",err)}
 if healed!=25||used.Health!=85||used.Inventory["minor_healing_potion"]!=1{t.Fatalf("potion result healed=%d player=%+v",healed,used)}
 persisted,err:=repo.GetCharacter(ctx,account.ID,player.ID);if err!=nil{t.Fatal(err)}
 if persisted.Health!=85||persisted.Inventory["minor_healing_potion"]!=1{t.Fatalf("potion use not persisted: %+v",persisted)}
}

func TestPostgresZonePresenceHeartbeatLeaseAndOwnership(t *testing.T) {
	db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
	first,_,err:=repo.Register(ctx,fmt.Sprintf("presence-a-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
	second,_,err:=repo.Register(ctx,fmt.Sprintf("presence-b-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID)})
	one,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("presence-one-%d",now.UnixNano())),"PresenceOne",now);if err!=nil{t.Fatal(err)}
	two,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("presence-two-%d",now.UnixNano())),"PresenceTwo",now);if err!=nil{t.Fatal(err)}
	if err:=repo.CreateCharacter(ctx,first.ID,one);err!=nil{t.Fatal(err)}
	if err:=repo.CreateCharacter(ctx,second.ID,two);err!=nil{t.Fatal(err)}
	if err:=repo.HeartbeatPresenceForAccount(ctx,second.ID,two.ID,now);err!=nil{t.Fatalf("presence heartbeat: %v",err)}
	visible,err:=repo.ListZonePresenceForAccount(ctx,first.ID,one.ID,now.Add(time.Second));if err!=nil{t.Fatalf("list active presence: %v",err)}
	if len(visible)!=1||visible[0].CharacterID!=two.ID||visible[0].Name!=two.Name||visible[0].ZoneID!=two.Position.ZoneID{t.Fatalf("presence roster=%+v",visible)}
	expired,err:=repo.ListZonePresenceForAccount(ctx,first.ID,one.ID,now.Add(game.PresenceLease+time.Second));if err!=nil{t.Fatal(err)}
	if len(expired)!=0{t.Fatalf("expired presence remained visible: %+v",expired)}
	if _,err:=repo.ListZonePresenceForAccount(ctx,first.ID,two.ID,now);!errors.Is(err,game.ErrPlayerNotFound){t.Fatalf("foreign character presence error=%v",err)}
}


func TestPostgresPlayerTradeEscrowSchemaEnforcesReservationsAndPositiveBalances(t *testing.T) {
	db := integrationDatabase(t)
	repo := NewAccountRepository(db)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	first, _, err := repo.Register(ctx, fmt.Sprintf("trade-a-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { t.Fatal(err) }
	second, _, err := repo.Register(ctx, fmt.Sprintf("trade-b-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { cleanupTestAccounts(db, first.ID); t.Fatal(err) }
	third, _, err := repo.Register(ctx, fmt.Sprintf("trade-c-%d@example.test", now.UnixNano()), "correct-horse-battery-staple", now)
	if err != nil { cleanupTestAccounts(db, first.ID, second.ID); t.Fatal(err) }

	firstPlayer, err := game.NewPlayer(game.PlayerID(fmt.Sprintf("trade-char-a-%d", now.UnixNano())), "TradeWayfarerA", now)
	if err != nil { t.Fatal(err) }
	secondPlayer, err := game.NewPlayer(game.PlayerID(fmt.Sprintf("trade-char-b-%d", now.UnixNano())), "TradeWayfarerB", now)
	if err != nil { t.Fatal(err) }
	thirdPlayer, err := game.NewPlayer(game.PlayerID(fmt.Sprintf("trade-char-c-%d", now.UnixNano())), "TradeWayfarerC", now)
	if err != nil { t.Fatal(err) }
	if err := repo.CreateCharacter(ctx, first.ID, firstPlayer); err != nil { t.Fatal(err) }
	if err := repo.CreateCharacter(ctx, second.ID, secondPlayer); err != nil { t.Fatal(err) }
	if err := repo.CreateCharacter(ctx, third.ID, thirdPlayer); err != nil { t.Fatal(err) }

	tradeID := fmt.Sprintf("trade-%d", now.UnixNano())
	secondTradeID := tradeID + "-second"
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM player_trades WHERE id IN ($1,$2)", tradeID, secondTradeID)
		cleanupTestAccounts(db, first.ID, second.ID, third.ID)
	})
	if _, err := db.ExecContext(ctx, `INSERT INTO player_trades(id,status,revision,created_at,updated_at,expires_at)
		VALUES($1,'pending',1,$2,$2,$3)`, tradeID, now, now.Add(10*time.Minute)); err != nil { t.Fatal(err) }
	for _, participant := range []struct{ player game.Player; accountID, role string }{
		{firstPlayer, first.ID, "initiator"}, {secondPlayer, second.ID, "target"},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_participants(trade_id,character_id,account_id,role)
			VALUES($1,$2,$3,$4)`, tradeID, string(participant.player.ID), participant.accountID, participant.role); err != nil { t.Fatal(err) }
		if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_offers(trade_id,character_id,offered_coins)
			VALUES($1,$2,0)`, tradeID, string(participant.player.ID)); err != nil { t.Fatal(err) }
		if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_active_characters(character_id,trade_id,account_id)
			VALUES($1,$2,$3)`, string(participant.player.ID), tradeID, participant.accountID); err != nil { t.Fatal(err) }
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_escrow_items(trade_id,owner_character_id,item_id,quantity)
		VALUES($1,$2,'softwood_log',2)`, tradeID, string(firstPlayer.ID)); err != nil { t.Fatal(err) }

	if _, err := db.ExecContext(ctx, `INSERT INTO player_trades(id,status,revision,created_at,updated_at,expires_at)
		VALUES($1,'pending',1,$2,$2,$3)`, secondTradeID, now, now.Add(10*time.Minute)); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_participants(trade_id,character_id,account_id,role)
		VALUES($1,$2,$3,'initiator')`, secondTradeID, string(thirdPlayer.ID), first.ID); err == nil {
		t.Fatal("trade participant was allowed to claim a character owned by another account")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_participants(trade_id,character_id,account_id,role)
		VALUES($1,$2,$3,'initiator')`, secondTradeID, string(firstPlayer.ID), first.ID); err != nil { t.Fatal(err) }
	if _, err := db.ExecContext(ctx, `INSERT INTO player_trade_active_characters(character_id,trade_id,account_id)
		VALUES($1,$2,$3)`, string(firstPlayer.ID), secondTradeID, first.ID); err == nil {
		t.Fatal("character was allowed to reserve a second active trade")
	}
	if _, err := db.ExecContext(ctx, `UPDATE player_trade_offers SET offered_coins=-1 WHERE trade_id=$1 AND character_id=$2`, tradeID, string(firstPlayer.ID)); err == nil {
		t.Fatal("negative escrow currency was accepted")
	}
	if _, err := db.ExecContext(ctx, `UPDATE player_trade_escrow_items SET quantity=0 WHERE trade_id=$1 AND owner_character_id=$2 AND item_id='softwood_log'`, tradeID, string(firstPlayer.ID)); err == nil {
		t.Fatal("zero-quantity escrow item was accepted")
	}
}

func TestPostgresTradeInviteAcceptAndAccountPrivacy(t *testing.T) {
	db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
	first,_,err:=repo.Register(ctx,fmt.Sprintf("trade-invite-a-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
	second,_,err:=repo.Register(ctx,fmt.Sprintf("trade-invite-b-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
	outsider,_,err:=repo.Register(ctx,fmt.Sprintf("trade-invite-c-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){cleanupTestAccounts(db,first.ID,second.ID,outsider.ID)})
	initiator,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("trade-invite-a-%d",now.UnixNano())),"TradeInviteA",now);if err!=nil{t.Fatal(err)}
	target,err:=game.NewPlayer(game.PlayerID(fmt.Sprintf("trade-invite-b-%d",now.UnixNano())),"TradeInviteB",now);if err!=nil{t.Fatal(err)}
	if err:=repo.CreateCharacter(ctx,first.ID,initiator);err!=nil{t.Fatal(err)}
	if err:=repo.CreateCharacter(ctx,second.ID,target);err!=nil{t.Fatal(err)}
	if err:=repo.HeartbeatPresenceForAccount(ctx,second.ID,target.ID,now);err!=nil{t.Fatal(err)}
	trade,err:=repo.CreateTradeInviteForAccount(ctx,first.ID,initiator.ID,target.ID,now);if err!=nil{t.Fatalf("create invite: %v",err)}
	if trade.Status!=game.TradeStatusPending||trade.Revision!=1||len(trade.Participants)!=2{t.Fatalf("trade invite snapshot=%+v",trade)}
	if _,err:=repo.GetTradeForAccount(ctx,outsider.ID,trade.ID);!errors.Is(err,game.ErrTradeNotFound){t.Fatalf("outsider read error=%v",err)}
	targetView,err:=repo.GetTradeForAccount(ctx,second.ID,trade.ID);if err!=nil||targetView.ID!=trade.ID{t.Fatalf("target trade read=%+v err=%v",targetView,err)}
	accepted,err:=repo.AcceptTradeInviteForAccount(ctx,second.ID,trade.ID,now.Add(time.Second));if err!=nil{t.Fatalf("accept invite: %v",err)}
	if accepted.Status!=game.TradeStatusOpen{t.Fatalf("accepted trade status=%s",accepted.Status)}
	if _,err:=repo.AcceptTradeInviteForAccount(ctx,first.ID,trade.ID,now.Add(2*time.Second));!errors.Is(err,game.ErrTradeNotParticipant){t.Fatalf("initiator accept error=%v",err)}
	if _,err:=repo.GetTradeForAccount(ctx,first.ID,trade.ID);err!=nil{t.Fatalf("initiator read after acceptance: %v",err)}
}
