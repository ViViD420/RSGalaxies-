package game

import (
	"errors"
	"math"
	"time"
)

var (
	ErrUnknownGatheringNode   = errors.New("unknown gathering node")
	ErrGatheringOutOfRange    = errors.New("gathering node is out of range")
	ErrGatheringNodeDepleted  = errors.New("gathering node is depleted")
	ErrInsufficientSkillLevel = errors.New("skill level is too low")
	ErrInventoryFull          = errors.New("inventory stack is full")
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
	Charges     int
	RespawnAt   time.Time
	Initialized bool
}

var gatheringNodes = []GatheringNodeDefinition{
	{ID: "tree-001", Name: "Young Frontier Tree", SkillID: "woodcutting", RequiredLevel: 1, ResourceID: "softwood_log", Experience: 25, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 2, Z: 0}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 30},
	{ID: "stone-001", Name: "Exposed Stone Deposit", SkillID: "mining", RequiredLevel: 1, ResourceID: "rough_stone", Experience: 20, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -2, Z: 0}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 30},
	{ID: "copper-001", Name: "Copper Vein", SkillID: "mining", RequiredLevel: 1, ResourceID: "copper_ore", Experience: 30, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: -8}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 45},
	{ID: "herb-001", Name: "Wild Herb Patch", SkillID: "foraging", RequiredLevel: 1, ResourceID: "wild_herb", Experience: 25, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 9, Z: 1}, InteractionRange: 3, MaxCharges: 4, RespawnSeconds: 35},
	{ID: "crop-001", Name: "Frontier Crop Plot", SkillID: "farming", RequiredLevel: 1, ResourceID: "frontier_grain", Experience: 25, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -9, Z: 4}, InteractionRange: 3, MaxCharges: 4, RespawnSeconds: 35},
	{ID: "scrap-001", Name: "Debris Scrap Pile", SkillID: "scavenging", RequiredLevel: 1, ResourceID: "salvaged_scrap", Experience: 30, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 6, Z: -8}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 40},
	// 6 Additional Mining Resource Nodes (Iron, Cobalt, Titanium, Aether Crystal, Durasteel, Luminite, Mythril)
	{ID: "iron-001", Name: "Deep Iron Vein", SkillID: "mining", RequiredLevel: 2, ResourceID: "iron_ore", Experience: 45, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -14, Z: -10}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 50},
	{ID: "cobalt-001", Name: "Cobalt Ridge Seam", SkillID: "mining", RequiredLevel: 2, ResourceID: "cobalt_ore", Experience: 50, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 14, Z: 9}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 50},
	{ID: "crystal-001", Name: "Aether Crystal Formation", SkillID: "mining", RequiredLevel: 2, ResourceID: "aether_crystal", Experience: 55, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -15, Z: 14}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 60},
	{ID: "titanium-001", Name: "Titanium Outcrop", SkillID: "mining", RequiredLevel: 3, ResourceID: "titanium_ore", Experience: 65, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 16, Z: -14}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 65},
	{ID: "durasteel-001", Name: "Durasteel Fissure", SkillID: "mining", RequiredLevel: 4, ResourceID: "durasteel_ore", Experience: 85, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -26, Z: -22}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 75},
	{ID: "luminite-001", Name: "Luminite Geode Cluster", SkillID: "mining", RequiredLevel: 4, ResourceID: "luminite_gem", Experience: 95, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 26, Z: 22}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 80},
	{ID: "mythril-001", Name: "Mythril Star-Vein", SkillID: "mining", RequiredLevel: 5, ResourceID: "mythril_ore", Experience: 125, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -30, Z: 26}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 95},
	{ID: "orichalcum-001", Name: "Orichalcum Vent Seam", SkillID: "mining", RequiredLevel: 5, ResourceID: "orichalcum_ore", Experience: 135, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 30, Z: -28}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 100},
	{ID: "obsidian-001", Name: "Volcanic Obsidian Outcrop", SkillID: "mining", RequiredLevel: 5, ResourceID: "obsidian_shard", Experience: 130, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -34, Z: -8}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 90},
	{ID: "void-quartz-001", Name: "Void Quartz Monolith", SkillID: "mining", RequiredLevel: 6, ResourceID: "void_quartz", Experience: 155, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 34, Z: 14}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 110},
	{ID: "neutronium-001", Name: "Neutronium Crater Vein", SkillID: "mining", RequiredLevel: 6, ResourceID: "neutronium_ore", Experience: 170, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -22, Z: 34}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 120},
	{ID: "starsteel-001", Name: "Fallen Star-Metal Meteorite", SkillID: "mining", RequiredLevel: 7, ResourceID: "starsteel_ore", Experience: 210, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 38, Z: -4}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 140},
	{ID: "electrum-001", Name: "Electrum Crest Vein", SkillID: "mining", RequiredLevel: 6, ResourceID: "electrum_ore", Experience: 165, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -40, Z: 4}, InteractionRange: 3, MaxCharges: 3, RespawnSeconds: 115},
	{ID: "solarite-001", Name: "Solarite Sun-Spire", SkillID: "mining", RequiredLevel: 7, ResourceID: "solarite_crystal", Experience: 195, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 8, Z: 40}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 130},
	{ID: "chronite-001", Name: "Chronite Rift Anomaly", SkillID: "mining", RequiredLevel: 8, ResourceID: "chronite_ore", Experience: 235, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -12, Z: -40}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 150},
	{ID: "abyssal-001", Name: "Abyssal Adamant Chasm", SkillID: "mining", RequiredLevel: 9, ResourceID: "abyssal_ore", Experience: 275, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: -44, Z: -32}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 165},
	{ID: "nova-prism-001", Name: "Nova Core Matrix", SkillID: "mining", RequiredLevel: 10, ResourceID: "nova_prism", Experience: 320, Position: WorldPosition{ZoneID: DefaultWorldZoneID, X: 44, Z: 36}, InteractionRange: 3, MaxCharges: 2, RespawnSeconds: 180},
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
	if err := p.CanAct(); err != nil {
		return GatheringResult{}, state, err
	}
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
