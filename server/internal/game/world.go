package game

const DefaultWorldZoneID = "verdant-frontier"

type WorldBounds struct {
	MinX float64 `json:"minX"`
	MaxX float64 `json:"maxX"`
	MinZ float64 `json:"minZ"`
	MaxZ float64 `json:"maxZ"`
}

type WorldPosition struct {
	ZoneID string  `json:"zoneId"`
	X      float64 `json:"x"`
	Z      float64 `json:"z"`
}

type WorldZone struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Bounds      WorldBounds   `json:"bounds"`
	Spawn       WorldPosition `json:"spawn"`
}

var worldZones = []WorldZone{
	{
		ID: DefaultWorldZoneID,
		Name: "Verdant Frontier",
		Description: "A temperate frontier zone used for the first movement and exploration prototype.",
		Bounds: WorldBounds{MinX: -120, MaxX: 120, MinZ: -120, MaxZ: 120},
		Spawn: WorldPosition{ZoneID: DefaultWorldZoneID, X: 0, Z: 0},
	},
}

func WorldZones() []WorldZone {
	zones := make([]WorldZone, len(worldZones))
	copy(zones, worldZones)
	return zones
}

func FindWorldZone(id string) (WorldZone, bool) {
	for _, zone := range worldZones {
		if zone.ID == id {
			return zone, true
		}
	}
	return WorldZone{}, false
}
