package game

import (
	"errors"
	"math"
	"time"
)

func ValidateTradeInvite(initiator,target Player) error {
	if initiator.ID==""||target.ID=="" { return ErrTradeParticipantUnavailable }
	if initiator.ID==target.ID { return ErrTradeSelf }
	if err:=initiator.CanAct();err!=nil{return err}
	if err:=target.CanAct();err!=nil{return ErrTradeParticipantUnavailable}
	if initiator.Position.ZoneID!=target.Position.ZoneID{return ErrTradeDifferentZone}
	if math.Hypot(initiator.Position.X-target.Position.X,initiator.Position.Z-target.Position.Z)>TradeInteractionRange{return ErrTradeOutOfRange}
	return nil
}
