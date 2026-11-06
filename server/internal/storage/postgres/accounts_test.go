package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
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
	if _,err:=repo.Login(ctx,email,"incorrect-password",now);!errors.Is(err,ErrInvalidCredentials){t.Fatalf("wrong password error=%v",err)}
	if err:=repo.RevokeSession(ctx,token,now.Add(3*time.Minute));err!=nil{t.Fatal(err)}
	if _,err:=repo.AuthenticateSession(ctx,token,now.Add(4*time.Minute));!errors.Is(err,ErrSessionNotFound){t.Fatalf("revoked session error=%v",err)}
	if _,err:=repo.AuthenticateSession(ctx,loginToken,now.Add(4*time.Minute));err!=nil{t.Fatalf("second session should remain valid: %v",err)}
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
