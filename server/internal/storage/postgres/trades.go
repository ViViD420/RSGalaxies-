package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

// validateTradeReservations preserves enough capacity to return escrowed assets.
func validateTradeReservations(ctx context.Context, tx *sql.Tx, accountID string, player game.Player) error {
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM player_trade_active_characters WHERE character_id=$1 AND account_id=$2)`, string(player.ID), accountID).Scan(&active); err != nil { return fmt.Errorf("check active trade reservation: %w", err) }
	if !active { return nil }
	rows, err := tx.QueryContext(ctx, `SELECT e.item_id, SUM(e.quantity) FROM player_trade_escrow_items e JOIN player_trade_participants p ON p.trade_id=e.trade_id AND p.character_id=e.owner_character_id WHERE e.owner_character_id=$1 AND p.account_id=$2 GROUP BY e.item_id`, string(player.ID), accountID)
	if err != nil { return fmt.Errorf("load reserved trade inventory: %w", err) }
	for rows.Next() {
		var itemID string; var reserved int64
		if err := rows.Scan(&itemID, &reserved); err != nil { _ = rows.Close(); return fmt.Errorf("scan reserved trade inventory: %w", err) }
		item, ok := game.FindItem(itemID)
		if !ok || reserved < 0 || reserved > int64(item.MaxStack) || player.Inventory[itemID] < 0 || int64(player.Inventory[itemID]) > int64(item.MaxStack)-reserved { _ = rows.Close(); return game.ErrTradeOfferCapacity }
	}
	if err := rows.Err(); err != nil { _ = rows.Close(); return fmt.Errorf("iterate reserved trade inventory: %w", err) }
	if err := rows.Close(); err != nil { return fmt.Errorf("close reserved trade inventory: %w", err) }
	var reservedCoins int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(o.offered_coins),0) FROM player_trade_offers o JOIN player_trade_participants p ON p.trade_id=o.trade_id AND p.character_id=o.character_id WHERE p.character_id=$1 AND p.account_id=$2`, string(player.ID), accountID).Scan(&reservedCoins); err != nil { return fmt.Errorf("load reserved trade currency: %w", err) }
	if player.Coins < 0 || reservedCoins < 0 || player.Coins > math.MaxInt64-reservedCoins { return game.ErrTradeOfferCapacity }
	return nil
}


type lockedTradePlayer struct {
	AccountID string
	Player game.Player
}

func rLockOwnedPlayer(ctx context.Context,tx *sql.Tx,accountID string,id game.PlayerID)(game.Player,error){
	return scanPlayer(tx.QueryRowContext(ctx,"SELECT "+playerColumns+" FROM players WHERE account_id=$1 AND id=$2 FOR UPDATE",accountID,string(id)))
}

func updateTradePlayer(ctx context.Context,tx *sql.Tx,ref lockedTradePlayer)error{
	return updateOwnedPlayer(ctx,tx,ref.AccountID,ref.Player)
}

type tradeParticipantRef struct {
	CharacterID game.PlayerID
	AccountID string
}

func loadTradeSnapshot(ctx context.Context,q tradeQueryer,tradeID string)(game.TradeSnapshot,error){
	var snapshot game.TradeSnapshot
	var status string
	err:=q.QueryRowContext(ctx,`SELECT id,status,revision,created_at,expires_at FROM player_trades WHERE id=$1`,tradeID).Scan(&snapshot.ID,&status,&snapshot.Revision,&snapshot.CreatedAt,&snapshot.ExpiresAt)
	if errors.Is(err,sql.ErrNoRows){return game.TradeSnapshot{},game.ErrTradeNotFound}
	if err!=nil{return game.TradeSnapshot{},fmt.Errorf("load trade: %w",err)}
	snapshot.Status=game.TradeStatus(status);snapshot.CreatedAt=snapshot.CreatedAt.UTC();snapshot.ExpiresAt=snapshot.ExpiresAt.UTC()
	rows,err:=q.QueryContext(ctx,`SELECT tp.character_id,p.name,tp.role,o.offered_coins,tp.confirmed_revision
		FROM player_trade_participants tp JOIN players p ON p.id=tp.character_id
		JOIN player_trade_offers o ON o.trade_id=tp.trade_id AND o.character_id=tp.character_id
		WHERE tp.trade_id=$1 ORDER BY CASE tp.role WHEN 'initiator' THEN 0 ELSE 1 END`,tradeID)
	if err!=nil{return game.TradeSnapshot{},fmt.Errorf("load trade participants: %w",err)}
	for rows.Next(){
		var p game.TradeParticipant;var characterID,role string;var confirmed sql.NullInt64
		if err:=rows.Scan(&characterID,&p.Name,&role,&p.Offer.Coins,&confirmed);err!=nil{rows.Close();return game.TradeSnapshot{},fmt.Errorf("scan trade participant: %w",err)}
		p.CharacterID=game.PlayerID(characterID);p.Role=role;p.Offer.Items=make(map[string]int);p.Confirmed=confirmed.Valid&&confirmed.Int64==snapshot.Revision
		snapshot.Participants=append(snapshot.Participants,p)
	}
	if err:=rows.Err();err!=nil{rows.Close();return game.TradeSnapshot{},fmt.Errorf("iterate trade participants: %w",err)}
	if err:=rows.Close();err!=nil{return game.TradeSnapshot{},fmt.Errorf("close trade participants: %w",err)}
	if len(snapshot.Participants)!=2{return game.TradeSnapshot{},fmt.Errorf("trade %s has invalid participant count %d",tradeID,len(snapshot.Participants))}
	for i:=range snapshot.Participants{
		p:=&snapshot.Participants[i]
		itemRows,err:=q.QueryContext(ctx,`SELECT item_id,quantity FROM player_trade_escrow_items WHERE trade_id=$1 AND owner_character_id=$2 ORDER BY item_id`,tradeID,string(p.CharacterID))
		if err!=nil{return game.TradeSnapshot{},fmt.Errorf("load escrow items: %w",err)}
		for itemRows.Next(){var itemID string;var quantity int;if err:=itemRows.Scan(&itemID,&quantity);err!=nil{itemRows.Close();return game.TradeSnapshot{},fmt.Errorf("scan escrow item: %w",err)};p.Offer.Items[itemID]=quantity}
		if err:=itemRows.Err();err!=nil{itemRows.Close();return game.TradeSnapshot{},fmt.Errorf("iterate escrow items: %w",err)}
		if err:=itemRows.Close();err!=nil{return game.TradeSnapshot{},fmt.Errorf("close escrow items: %w",err)}
	}
	return snapshot,nil
}

func lockTrade(ctx context.Context,tx *sql.Tx,tradeID string)(game.TradeSnapshot,error){
	var id string
	err:=tx.QueryRowContext(ctx,`SELECT id FROM player_trades WHERE id=$1 FOR UPDATE`,tradeID).Scan(&id)
	if errors.Is(err,sql.ErrNoRows){return game.TradeSnapshot{},game.ErrTradeNotFound}
	if err!=nil{return game.TradeSnapshot{},fmt.Errorf("lock trade: %w",err)}
	return loadTradeSnapshot(ctx,tx,tradeID)
}

func lockTradePlayers(ctx context.Context,tx *sql.Tx,tradeID string)(map[game.PlayerID]lockedTradePlayer,error){
	rows,err:=tx.QueryContext(ctx,`SELECT character_id,account_id FROM player_trade_participants WHERE trade_id=$1 ORDER BY character_id FOR UPDATE`,tradeID)
	if err!=nil{return nil,fmt.Errorf("lock trade participants: %w",err)}
	var refs []tradeParticipantRef
	for rows.Next(){var ref tradeParticipantRef;var id string;if err:=rows.Scan(&id,&ref.AccountID);err!=nil{rows.Close();return nil,err};ref.CharacterID=game.PlayerID(id);refs=append(refs,ref)}
	if err:=rows.Err();err!=nil{rows.Close();return nil,err}
	if err:=rows.Close();err!=nil{return nil,err}
	if len(refs)!=2{return nil,game.ErrTradeNotFound}
	locked:=make(map[game.PlayerID]lockedTradePlayer,2)
	for _,ref:=range refs{
		player,err:=rLockOwnedPlayer(ctx,tx,ref.AccountID,ref.CharacterID);if err!=nil{return nil,err}
		locked[ref.CharacterID]=lockedTradePlayer{AccountID:ref.AccountID,Player:player}
	}
	return locked,nil
}

func participantRole(snapshot game.TradeSnapshot,id game.PlayerID)(game.TradeParticipant,bool){
	for _,participant:=range snapshot.Participants{if participant.CharacterID==id{return participant,true}}
	return game.TradeParticipant{},false
}

func persistTradeOffer(ctx context.Context,tx *sql.Tx,tradeID string,characterID game.PlayerID,offer game.TradeOffer)error{
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trade_offers SET offered_coins=$3 WHERE trade_id=$1 AND character_id=$2`,tradeID,string(characterID),offer.Coins);err!=nil{return fmt.Errorf("update escrow coins: %w",err)}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM player_trade_escrow_items WHERE trade_id=$1 AND owner_character_id=$2`,tradeID,string(characterID));err!=nil{return fmt.Errorf("replace escrow items: %w",err)}
	for itemID,quantity:=range offer.Items{if _,err:=tx.ExecContext(ctx,`INSERT INTO player_trade_escrow_items(trade_id,owner_character_id,item_id,quantity) VALUES($1,$2,$3,$4)`,tradeID,string(characterID),itemID,quantity);err!=nil{return fmt.Errorf("insert escrow item: %w",err)}}
	return nil
}

// CreateTradeInviteForAccount starts an empty two-party offer; both sides add
// assets only after the invitation has been accepted.
func (r *AccountRepository) CreateTradeInviteForAccount(ctx context.Context, accountID string, initiatorID, targetID game.PlayerID, now time.Time) (game.TradeSnapshot, error) {
	return r.CreateTradeForAccount(ctx, accountID, initiatorID, targetID, game.TradeOffer{Items: map[string]int{}}, now)
}

// AcceptTradeInviteForAccount resolves the invited target character from the
// account-scoped participant row and opens the trade with an empty initial offer.
func (r *AccountRepository) AcceptTradeInviteForAccount(ctx context.Context, accountID, tradeID string, now time.Time) (game.TradeSnapshot, error) {
	var targetID string
	err := r.db.QueryRowContext(ctx, `SELECT character_id FROM player_trade_participants
		WHERE trade_id=$1 AND account_id=$2 AND role='target'`, tradeID, accountID).Scan(&targetID)
	if errors.Is(err, sql.ErrNoRows) {
		return game.TradeSnapshot{}, game.ErrTradeNotFound
	}
	if err != nil {
		return game.TradeSnapshot{}, fmt.Errorf("resolve invited trade character: %w", err)
	}
	return r.AcceptTradeForAccount(ctx, accountID, game.PlayerID(targetID), tradeID, game.TradeOffer{Items: map[string]int{}}, now)
}

func (r *AccountRepository) CreateTradeForAccount(ctx context.Context,accountID string,initiatorID,targetID game.PlayerID,offer game.TradeOffer,now time.Time)(game.TradeSnapshot,error){
	if initiatorID==targetID{return game.TradeSnapshot{},game.ErrTradeSelf}
	now=now.UTC();tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.TradeSnapshot{},err};defer tx.Rollback()
	var targetAccountID string
	err=tx.QueryRowContext(ctx,`SELECT p.account_id FROM players p JOIN zone_presence z ON z.character_id=p.id WHERE p.id=$1 AND z.last_seen_at>$2`,string(targetID),now.Add(-game.PresenceLease)).Scan(&targetAccountID)
	if errors.Is(err,sql.ErrNoRows){return game.TradeSnapshot{},game.ErrTradeParticipantUnavailable};if err!=nil{return game.TradeSnapshot{},fmt.Errorf("resolve active trade target: %w",err)}
	if targetAccountID==accountID{return game.TradeSnapshot{},game.ErrTradeSelf}
	refs:=[]tradeParticipantRef{{CharacterID:initiatorID,AccountID:accountID},{CharacterID:targetID,AccountID:targetAccountID}}
	sort.Slice(refs,func(i,j int)bool{return refs[i].CharacterID<refs[j].CharacterID})
	players:=make(map[game.PlayerID]lockedTradePlayer,2)
	for _,ref:=range refs{player,err:=rLockOwnedPlayer(ctx,tx,ref.AccountID,ref.CharacterID);if err!=nil{return game.TradeSnapshot{},err};players[ref.CharacterID]=lockedTradePlayer{AccountID:ref.AccountID,Player:player}}
	initiator,target:=players[initiatorID].Player,players[targetID].Player
	if err:=game.ValidateTradeInvite(initiator,target);err!=nil{return game.TradeSnapshot{},err}
	if err:=game.EscrowTradeOffer(&initiator,offer);err!=nil{return game.TradeSnapshot{},err}
	tradeID,err:=newAccountID();if err!=nil{return game.TradeSnapshot{},fmt.Errorf("generate trade ID: %w",err)}
	expires:=now.Add(game.TradeLifetime)
	if _,err:=tx.ExecContext(ctx,`INSERT INTO player_trades(id,status,revision,created_at,updated_at,expires_at) VALUES($1,'pending',1,$2,$2,$3)`,tradeID,now,expires);err!=nil{return game.TradeSnapshot{},fmt.Errorf("create trade: %w",err)}
	for _,ref:=range refs{
		role:="target";if ref.CharacterID==initiatorID{role="initiator"}
		if _,err:=tx.ExecContext(ctx,`INSERT INTO player_trade_participants(trade_id,character_id,account_id,role) VALUES($1,$2,$3,$4)`,tradeID,string(ref.CharacterID),ref.AccountID,role);err!=nil{return game.TradeSnapshot{},fmt.Errorf("insert participant: %w",err)}
		coins:=int64(0);if ref.CharacterID==initiatorID{coins=offer.Coins}
		if _,err:=tx.ExecContext(ctx,`INSERT INTO player_trade_offers(trade_id,character_id,offered_coins) VALUES($1,$2,$3)`,tradeID,string(ref.CharacterID),coins);err!=nil{return game.TradeSnapshot{},fmt.Errorf("insert participant offer: %w",err)}
	}
	if err:=persistTradeOffer(ctx,tx,tradeID,initiatorID,offer);err!=nil{return game.TradeSnapshot{},err}
	for _,ref:=range refs{if _,err:=tx.ExecContext(ctx,`INSERT INTO player_trade_active_characters(character_id,trade_id,account_id) VALUES($1,$2,$3)`,string(ref.CharacterID),tradeID,ref.AccountID);err!=nil{var pgErr *pgconn.PgError;if errors.As(err,&pgErr)&&pgErr.Code=="23505"{return game.TradeSnapshot{},game.ErrTradeAlreadyActive};return game.TradeSnapshot{},fmt.Errorf("reserve trade participant: %w",err)}}
	players[initiatorID]=lockedTradePlayer{AccountID:accountID,Player:initiator}
	if err:=updateOwnedPlayer(ctx,tx,accountID,initiator);err!=nil{return game.TradeSnapshot{},err}
	if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},fmt.Errorf("commit trade invitation: %w",err)}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) AcceptTradeForAccount(ctx context.Context,accountID string,id game.PlayerID,tradeID string,offer game.TradeOffer,now time.Time)(game.TradeSnapshot,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.TradeSnapshot{},err};defer tx.Rollback()
	snapshot,err:=lockTrade(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if snapshot.Status!=game.TradeStatusPending{return game.TradeSnapshot{},game.ErrTradeNotAccepted}
	if !now.Before(snapshot.ExpiresAt){return game.TradeSnapshot{},game.ErrTradeExpired}
	participant,ok:=participantRole(snapshot,id);if !ok||participant.Role!="target"{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	players,err:=lockTradePlayers(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if players[id].AccountID!=accountID{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	var initiatorID game.PlayerID
	for _,p:=range snapshot.Participants{if p.Role=="initiator"{initiatorID=p.CharacterID}}
	initiator,target:=players[initiatorID].Player,players[id].Player
	if err:=game.ValidateTradeInvite(initiator,target);err!=nil{return game.TradeSnapshot{},err}
	if err:=game.EscrowTradeOffer(&target,offer);err!=nil{return game.TradeSnapshot{},err}
	if err:=persistTradeOffer(ctx,tx,tradeID,id,offer);err!=nil{return game.TradeSnapshot{},err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,target);err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trades SET status='open',revision=revision+1,updated_at=$2 WHERE id=$1`,tradeID,now.UTC());err!=nil{return game.TradeSnapshot{},err}
	if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},fmt.Errorf("commit trade acceptance: %w",err)}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) expireTrade(ctx context.Context, tradeID string, now time.Time) error {
	_, err := r.finishTrade(ctx, "", "", tradeID, game.TradeStatusExpired, now)
	if errors.Is(err, game.ErrTradeNotAccepted) {
		return nil
	}
	return err
}

func (r *AccountRepository) UpdateTradeOfferForAccount(ctx context.Context,accountID string,id game.PlayerID,tradeID string,revision int64,offer game.TradeOffer,now time.Time)(game.TradeSnapshot,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.TradeSnapshot{},err};defer tx.Rollback()
	snapshot,err:=lockTrade(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if !now.UTC().Before(snapshot.ExpiresAt) {
		_ = tx.Rollback()
		if err:=r.expireTrade(ctx,tradeID,now);err!=nil{return game.TradeSnapshot{},err}
		return game.TradeSnapshot{},game.ErrTradeExpired
	}
	if snapshot.Status!=game.TradeStatusOpen{return game.TradeSnapshot{},game.ErrTradeNotAccepted}
	if snapshot.Revision!=revision{return game.TradeSnapshot{},game.ErrTradeRevisionConflict}
	_,ok:=participantRole(snapshot,id);if !ok{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	players,err:=lockTradePlayers(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	ref:=players[id];if ref.AccountID!=accountID{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	if err:=game.ApplyTradeOffer(&ref.Player,participant.Offer);err!=nil{return game.TradeSnapshot{},err}
	if err:=game.EscrowTradeOffer(&ref.Player,offer);err!=nil{return game.TradeSnapshot{},err}
	if err:=persistTradeOffer(ctx,tx,tradeID,id,offer);err!=nil{return game.TradeSnapshot{},err}
	if err:=updateOwnedPlayer(ctx,tx,accountID,ref.Player);err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trades SET revision=revision+1,updated_at=$2 WHERE id=$1`,tradeID,now.UTC());err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trade_participants SET confirmed_revision=NULL WHERE trade_id=$1`,tradeID);err!=nil{return game.TradeSnapshot{},err}
	if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},fmt.Errorf("commit trade offer update: %w",err)}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) ConfirmTradeForAccount(ctx context.Context,accountID string,id game.PlayerID,tradeID string,revision int64,now time.Time)(game.TradeSnapshot,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.TradeSnapshot{},err};defer tx.Rollback()
	snapshot,err:=lockTrade(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if !now.UTC().Before(snapshot.ExpiresAt) {
		_ = tx.Rollback()
		if err:=r.expireTrade(ctx,tradeID,now);err!=nil{return game.TradeSnapshot{},err}
		return game.TradeSnapshot{},game.ErrTradeExpired
	}
	if snapshot.Status!=game.TradeStatusOpen{return game.TradeSnapshot{},game.ErrTradeNotAccepted}
	if snapshot.Revision!=revision{return game.TradeSnapshot{},game.ErrTradeRevisionConflict}
	participant,ok:=participantRole(snapshot,id);if !ok{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	players,err:=lockTradePlayers(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	ref:=players[id];if ref.AccountID!=accountID{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trade_participants SET confirmed_revision=$3 WHERE trade_id=$1 AND character_id=$2`,tradeID,string(id),revision);err!=nil{return game.TradeSnapshot{},err}
	var confirmations int
	if err:=tx.QueryRowContext(ctx,`SELECT COUNT(*) FROM player_trade_participants WHERE trade_id=$1 AND confirmed_revision=$2`,tradeID,revision).Scan(&confirmations);err!=nil{return game.TradeSnapshot{},err}
	if confirmations<2{if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},err};return loadTradeSnapshot(ctx,r.db,tradeID)}
	var initiatorID,targetID game.PlayerID;var initiatorOffer,targetOffer game.TradeOffer
	for _,p:=range snapshot.Participants{if p.Role=="initiator"{initiatorID=p.CharacterID;initiatorOffer=p.Offer}else{targetID=p.CharacterID;targetOffer=p.Offer}}
	initiator,target:=players[initiatorID],players[targetID]
	if err:=game.ApplyTradeOffer(&initiator.Player,targetOffer);err!=nil{return game.TradeSnapshot{},err}
	if err:=game.ApplyTradeOffer(&target.Player,initiatorOffer);err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trades SET status='completed',updated_at=$2,terminal_at=$2 WHERE id=$1`,tradeID,now.UTC());err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM player_trade_active_characters WHERE trade_id=$1`,tradeID);err!=nil{return game.TradeSnapshot{},err}
	if err:=updateTradePlayer(ctx,tx,initiator);err!=nil{return game.TradeSnapshot{},err}
	if err:=updateTradePlayer(ctx,tx,target);err!=nil{return game.TradeSnapshot{},err}
	if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},fmt.Errorf("commit trade settlement: %w",err)}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) CancelTradeForAccount(ctx context.Context,accountID string,id game.PlayerID,tradeID string,now time.Time)(game.TradeSnapshot,error){
	return r.finishTrade(ctx,accountID,id,tradeID,game.TradeStatusCancelled,now)
}

func (r *AccountRepository) finishTrade(ctx context.Context,accountID string,id game.PlayerID,tradeID string,status game.TradeStatus,now time.Time)(game.TradeSnapshot,error){
	tx,err:=r.db.BeginTx(ctx,nil);if err!=nil{return game.TradeSnapshot{},err};defer tx.Rollback()
	snapshot,err:=lockTrade(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if snapshot.Status==game.TradeStatusCompleted||snapshot.Status==game.TradeStatusCancelled||snapshot.Status==game.TradeStatusExpired{return game.TradeSnapshot{},game.ErrTradeNotAccepted}
	players,err:=lockTradePlayers(ctx,tx,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if accountID!=""&&players[id].AccountID!=accountID{return game.TradeSnapshot{},game.ErrTradeNotParticipant}
	if accountID!=""{if _,ok:=participantRole(snapshot,id);!ok{return game.TradeSnapshot{},game.ErrTradeNotParticipant}}
	for _,participant:=range snapshot.Participants{ref:=players[participant.CharacterID];if err:=game.ApplyTradeOffer(&ref.Player,participant.Offer);err!=nil{return game.TradeSnapshot{},fmt.Errorf("release trade escrow: %w",err)};players[participant.CharacterID]=ref}
	if _,err:=tx.ExecContext(ctx,`UPDATE player_trades SET status=$2,updated_at=$3,terminal_at=$3 WHERE id=$1`,tradeID,string(status),now.UTC());err!=nil{return game.TradeSnapshot{},err}
	if _,err:=tx.ExecContext(ctx,`DELETE FROM player_trade_active_characters WHERE trade_id=$1`,tradeID);err!=nil{return game.TradeSnapshot{},err}
	for _,participant:=range snapshot.Participants{if err:=updateTradePlayer(ctx,tx,players[participant.CharacterID]);err!=nil{return game.TradeSnapshot{},err}}
	if err:=tx.Commit();err!=nil{return game.TradeSnapshot{},fmt.Errorf("commit trade closure: %w",err)}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) GetActiveTradeForAccount(ctx context.Context,accountID string,id game.PlayerID,now time.Time)(game.TradeSnapshot,error){
	var tradeID string
	err:=r.db.QueryRowContext(ctx,`SELECT trade_id FROM player_trade_active_characters WHERE account_id=$1 AND character_id=$2`,accountID,string(id)).Scan(&tradeID)
	if errors.Is(err,sql.ErrNoRows){return game.TradeSnapshot{},game.ErrTradeNotFound};if err!=nil{return game.TradeSnapshot{},fmt.Errorf("load active trade: %w",err)}
	snapshot,err:=loadTradeSnapshot(ctx,r.db,tradeID);if err!=nil{return game.TradeSnapshot{},err}
	if !now.Before(snapshot.ExpiresAt)&&(snapshot.Status==game.TradeStatusPending||snapshot.Status==game.TradeStatusOpen){return r.finishTrade(ctx,"", "",tradeID,game.TradeStatusExpired,now)}
	return snapshot,nil
}

func (r *AccountRepository) ExpireTrades(ctx context.Context,now time.Time,limit int)(int,error){
	if limit<1{limit=50};if limit>500{limit=500}
	rows,err:=r.db.QueryContext(ctx,`SELECT id FROM player_trades WHERE status IN ('pending','open') AND expires_at<=$1 ORDER BY expires_at,id LIMIT $2`,now.UTC(),limit)
	if err!=nil{return 0,fmt.Errorf("find expired trades: %w",err)}
	var ids []string;for rows.Next(){var id string;if err:=rows.Scan(&id);err!=nil{rows.Close();return 0,err};ids=append(ids,id)}
	if err:=rows.Err();err!=nil{rows.Close();return 0,err};if err:=rows.Close();err!=nil{return 0,err}
	count:=0
	for _,id:=range ids{if _,err:=r.finishTrade(ctx,"","",id,game.TradeStatusExpired,now);err!=nil{if errors.Is(err,game.ErrTradeNotAccepted)||errors.Is(err,game.ErrTradeNotFound){continue};return count,err};count++}
	return count,nil
}

func (r *AccountRepository) GetTradeForAccount(ctx context.Context,accountID,tradeID string)(game.TradeSnapshot,error){
	var allowed bool
	if err:=r.db.QueryRowContext(ctx,`SELECT EXISTS(SELECT 1 FROM player_trade_participants WHERE trade_id=$1 AND account_id=$2)`,tradeID,accountID).Scan(&allowed);err!=nil{return game.TradeSnapshot{},fmt.Errorf("authorize trade read: %w",err)}
	if !allowed{return game.TradeSnapshot{},game.ErrTradeNotFound}
	return loadTradeSnapshot(ctx,r.db,tradeID)
}

func (r *AccountRepository) CreateTradeInviteForAccount(ctx context.Context,accountID string,initiatorID,targetID game.PlayerID,now time.Time)(game.TradeSnapshot,error){
	return r.CreateTradeForAccount(ctx,accountID,initiatorID,targetID,game.TradeOffer{Items:map[string]int{}},now)
}

func (r *AccountRepository) AcceptTradeInviteForAccount(ctx context.Context,accountID,tradeID string,now time.Time)(game.TradeSnapshot,error){
	var targetID string
	err:=r.db.QueryRowContext(ctx,`SELECT character_id FROM player_trade_participants WHERE trade_id=$1 AND account_id=$2 AND role='target'`,tradeID,accountID).Scan(&targetID)
	if errors.Is(err,sql.ErrNoRows){return game.TradeSnapshot{},game.ErrTradeNotFound}
	if err!=nil{return game.TradeSnapshot{},fmt.Errorf("resolve accepting trade character: %w",err)}
	return r.AcceptTradeForAccount(ctx,accountID,game.PlayerID(targetID),tradeID,game.TradeOffer{Items:map[string]int{}},now)
}

