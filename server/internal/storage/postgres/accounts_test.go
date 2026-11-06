package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
)

func TestPostgresAccountRegistrationLoginAndSessionRevocation(t *testing.T) {
	db:=integrationDatabase(t)
	repo:=NewAccountRepository(db)
	ctx:=context.Background()
	now:=time.Now().UTC().Truncate(time.Microsecond)
	email:=fmt.Sprintf("account-%d@example.test",now.UnixNano())
	account,token,err:=repo.Register(ctx,email,"correct-horse-battery-staple",now)
	if err!=nil{t.Fatalf("register: %v",err)}
	if account.ID=="" || account.Email!=email || token==""{t.Fatalf("invalid registration result: %+v",account)}
	t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id=$1",account.ID)})

	authenticated,err:=repo.AuthenticateSession(ctx,token,now.Add(time.Minute))
	if err!=nil || authenticated.ID!=account.ID{t.Fatalf("authenticate registration session: account=%+v err=%v",authenticated,err)}
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
	t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id=$1",account.ID)})
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
	t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id IN ($1,$2)",first.ID,second.ID)})
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
 t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id IN ($1,$2)",first.ID,second.ID)})
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
 if err:=db.QueryRowContext(ctx,"SELECT COUNT(*) FROM schema_migrations WHERE version IN (1,2)").Scan(&count);err!=nil{t.Fatal(err)}
 if count!=2{t.Fatalf("migration versions applied=%d want=2",count)}
}

func TestPostgresOwnedGatheringLocksSharedNodeAndChecksOwner(t *testing.T) {
 db:=integrationDatabase(t);repo:=NewAccountRepository(db);ctx:=context.Background();now:=time.Now().UTC()
 first,_,err:=repo.Register(ctx,fmt.Sprintf("gather-owner-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 second,_,err:=repo.Register(ctx,fmt.Sprintf("gather-other-%d@example.test",now.UnixNano()),"correct-horse-battery-staple",now);if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id IN ($1,$2)",first.ID,second.ID);_,_=db.ExecContext(context.Background(),"DELETE FROM gathering_node_state WHERE node_id=$1","tree-001")})
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
 t.Cleanup(func(){_,_=db.ExecContext(context.Background(),"DELETE FROM accounts WHERE id=$1",account.ID)})
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
