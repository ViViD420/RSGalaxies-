package game

import (
	"errors"
	"math"
	"time"
)

var (
	ErrUnknownGatheringNode = errors.New("unknown gathering node")
	ErrGatheringOutOfRange = errors.New("gathering node is out of range")
	ErrGatheringNodeDepleted = errors.New("gathering node is depleted")
	ErrInsufficientSkillLevel = errors.New("skill level is too low")
	ErrInventoryFull = errors.New("inventory stack is full")
)

type GatheringNodeDefinition struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	SkillID          string        `json:"skillId"`
	RequiredLevel    int           `json:"requiredLevel"`
	ResourceID       string        `json:"resourceId"`
	Experience       int           `json:"experience"`
	Position         WorldPosition `json:"position"`
	InteractionRange float64       `json:"interactionRange"`
	MaxCharges       int           `json:"maxCharges"`
	RespawnSeconds   int           `json:"respawnSeconds"`
}

type GatheringResult struct {
	NodeID     string        `json:"nodeId"`
	ItemID     string        `json:"itemId"`
	Quantity   int           `json:"quantity"`
	SkillID    string        `json:"skillId"`
	Experience int           `json:"experience"`
	Skill      SkillProgress `json:"skill"`
	LeveledUp  bool          `json:"leveledUp"`
}

type GatheringNodeState struct {
	Charges    int
	RespawnAt  time.Time
	Initialized bool
}

var gatheringNodes = []GatheringNodeDefinition{
	{ID: "tree-001", Name: "Young Frontier Tree", SkillID: "woodcutting", RequiredLevel: 1, ResourceID: "softwood_log", Experience: 25, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 2, Z: 0}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 30},
	{ID: "stone-001", Name: "Exposed Stone Deposit", SkillID: "mining", RequiredLevel: 1, ResourceID: "rough_stone", Experience: 20, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -2, Z: 0}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 30},
	{ID: "copper-001", Name: "Copper Vein", SkillID: "mining", RequiredLevel: 1, ResourceID: "copper_ore", Experience: 30, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: -8}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 45},
	{ID: "herb-001", Name: "Wild Herb Patch", SkillID: "foraging", RequiredLevel: 1, ResourceID: "wild_herb", Experience: 25, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 9, Z: 1}, InteractionRange: 3, MaxCharges: 4, RespawnSeconds: 35},
}

func GatheringNodes() []GatheringNodeDefinition {
	nodes := make([]GatheringNodeDefinition, len(gatheringNodes))
	copy(nodes, gatheringNodes)
	return nodes
}

func FindGatheringNode(id string) (GatheringNodeDefinition, bool) {
	for _, node := range gatheringNodes {
		if node.ID == id {
			return node, true
		}
	}
	return GatheringNodeDefinition{}, false
}

func (p *Player) Gather(node GatheringNodeDefinition, state GatheringNodeState, now time.Time) (GatheringResult, GatheringNodeState, error) {
	if err:=p.CanAct();err!=nil{return GatheringResult{},state,err}
	if p.Position.ZoneID != node.Position.ZoneID {
		return GatheringResult{}, state, ErrGatheringOutOfRange
	}
	distance := math.Hypot(p.Position.X-node.Position.X, p.Position.Z-node.Position.Z)
	if distance > node.InteractionRange {
		return GatheringResult{}, state, ErrGatheringOutOfRange
	}
	progress := p.Skills[node.SkillID]
	if progress.Level < node.RequiredLevel {
		return GatheringResult{}, state, ErrInsufficientSkillLevel
	}
	item, ok := FindItem(node.ResourceID)
	if !ok {
		return GatheringResult{}, state, ErrUnknownGatheringNode
	}
	if !state.Initialized {
		state.Charges = node.MaxCharges
		state.Initialized = true
	}
	if state.Charges <= 0 {
		if now.Before(state.RespawnAt) {
			return GatheringResult{}, state, ErrGatheringNodeDepleted
		}
		state.Charges = node.MaxCharges
		state.RespawnAt = time.Time{}
	}
	if p.Inventory == nil {
		p.Inventory = make(map[string]int)
	}
	if p.Inventory[item.ID] >= item.MaxStack {
		return GatheringResult{}, state, ErrInventoryFull
	}
	leveled, err := p.AwardExperience(node.SkillID, node.Experience)
	if err != nil {
		return GatheringResult{}, state, err
	}
	p.Inventory[item.ID]++
	state.Charges--
	if state.Charges == 0 {
		state.RespawnAt = now.Add(time.Duration(node.RespawnSeconds) * time.Second)
	}
	return GatheringResult{
		NodeID: node.ID, ItemID: item.ID, Quantity: 1, SkillID: node.SkillID,
		Experience: node.Experience, Skill: p.Skills[node.SkillID], LeveledUp: leveled,
	}, state, nil
}
