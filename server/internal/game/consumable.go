package game

import (
	"errors"
	"time"
)

const DefaultAttributeBuffDuration = 10 * time.Minute

var (
	ErrItemNotUsable = errors.New("item has no use effect")
	ErrHealthFull    = errors.New("health is already full")
)

func (p *Player) UseItem(itemID string) (int, error) {
	if err := p.CanAct(); err != nil {
		return 0, err
	}
	p.EnsureVitalDefaults()
	item, ok := FindItem(itemID)
	if !ok || !item.Consumable {
		return 0, ErrItemNotUsable
	}
	hasEffect := item.HealAmount > 0 || item.ActionRestoreAmount > 0 || item.MindRestoreAmount > 0 ||
		item.BuffHealthAmount > 0 || item.BuffActionAmount > 0 || item.BuffMindAmount > 0 ||
		item.ConstitutionBonus > 0 || item.StaminaBonus > 0 || item.WillpowerBonus > 0
	if !hasEffect {
		return 0, ErrItemNotUsable
	}
	if p.Inventory[itemID] < 1 {
		return 0, ErrInsufficientItems
	}

	isBuffItem := item.BuffHealthAmount > 0 || item.BuffActionAmount > 0 || item.BuffMindAmount > 0 ||
		item.ConstitutionBonus > 0 || item.StaminaBonus > 0 || item.WillpowerBonus > 0
	maxHealthBefore := p.EffectiveMaxHealth()
	maxActionBefore := p.EffectiveMaxAction()
	maxMindBefore := p.EffectiveMaxMind()

	if !isBuffItem {
		healthNeedsRestore := item.HealAmount > 0 && p.Health < maxHealthBefore
		actionNeedsRestore := item.ActionRestoreAmount > 0 && p.Action < maxActionBefore
		mindNeedsRestore := item.MindRestoreAmount > 0 && p.Mind < maxMindBefore
		if !healthNeedsRestore && !actionNeedsRestore && !mindNeedsRestore {
			return 0, ErrHealthFull
		}
	}

	if item.BuffHealthAmount > 0 {
		p.MaxHealthBuff = item.BuffHealthAmount
	}
	if item.BuffActionAmount > 0 {
		p.MaxActionBuff = item.BuffActionAmount
	}
	if item.BuffMindAmount > 0 {
		p.MaxMindBuff = item.BuffMindAmount
	}
	if isBuffItem {
		if item.ConstitutionBonus > 0 && p.Constitution < BasePlayerConstitution+item.ConstitutionBonus {
			p.Constitution = BasePlayerConstitution + item.ConstitutionBonus
		}
		if item.StaminaBonus > 0 && p.Stamina < BasePlayerStamina+item.StaminaBonus {
			p.Stamina = BasePlayerStamina + item.StaminaBonus
		}
		if item.WillpowerBonus > 0 && p.Willpower < BasePlayerWillpower+item.WillpowerBonus {
			p.Willpower = BasePlayerWillpower + item.WillpowerBonus
		}
		duration := DefaultAttributeBuffDuration
		if item.BuffDurationSeconds > 0 {
			duration = time.Duration(item.BuffDurationSeconds) * time.Second
		}
		p.BuffExpiresAt = time.Now().UTC().Add(duration)
	}

	maxHealth := p.EffectiveMaxHealth()
	maxAction := p.EffectiveMaxAction()
	maxMind := p.EffectiveMaxMind()

	// Buff*Amount fields only raise the effective pool caps (applied above via
	// Max*Buff). Restoration comes solely from HealAmount / ActionRestoreAmount /
	// MindRestoreAmount so a buff item never restores more than it documents.
	healed := 0
	healTarget := item.HealAmount
	if healTarget > 0 && p.Health < maxHealth {
		healed = healTarget
		if healed > maxHealth-p.Health {
			healed = maxHealth - p.Health
		}
		p.Health += healed
	}
	actionRestore := item.ActionRestoreAmount
	if actionRestore > 0 && p.Action < maxAction {
		if actionRestore > maxAction-p.Action {
			actionRestore = maxAction - p.Action
		}
		p.Action += actionRestore
		if healed == 0 {
			healed = actionRestore
		}
	}
	mindRestore := item.MindRestoreAmount
	if mindRestore > 0 && p.Mind < maxMind {
		if mindRestore > maxMind-p.Mind {
			mindRestore = maxMind - p.Mind
		}
		p.Mind += mindRestore
		if healed == 0 {
			healed = mindRestore
		}
	}
	if isBuffItem && healed == 0 {
		healed = item.BuffHealthAmount + item.BuffActionAmount + item.BuffMindAmount + item.ConstitutionBonus
	}

	p.Inventory[itemID]--
	if p.Inventory[itemID] == 0 {
		delete(p.Inventory, itemID)
	}
	return healed, nil
}
