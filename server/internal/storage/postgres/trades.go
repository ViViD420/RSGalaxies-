package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math"

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
