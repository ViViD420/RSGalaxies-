package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

func TestPlayerLifecycle(t *testing.T) {
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	create := httptest.NewRequest(http.MethodPost,"/api/v1/players",bytes.NewBufferString(`{"name":"Wayfarer"}`))
	created:=httptest.NewRecorder();handler.ServeHTTP(created,create)
	if created.Code!=http.StatusCreated{t.Fatalf("create status=%d body=%s",created.Code,created.Body.String())}
	var player game.Player
	if err:=json.Unmarshal(created.Body.Bytes(),&player);err!=nil{t.Fatalf("decode player: %v",err)}
	if player.ID==""{t.Fatal("created player has empty ID")}
	if len(player.Skills)!=len(game.SkillDefinitions()){t.Fatalf("created player has %d skills",len(player.Skills))}
	get:=httptest.NewRequest(http.MethodGet,"/api/v1/players/"+string(player.ID),nil)
	got:=httptest.NewRecorder();handler.ServeHTTP(got,get)
	if got.Code!=http.StatusOK{t.Fatalf("get status=%d body=%s",got.Code,got.Body.String())}
}

func TestSkillCatalogAndExperienceAward(t *testing.T) {
	handler:=New(game.NewMemoryPlayerRepository()).Routes()
	catalog:=httptest.NewRecorder();handler.ServeHTTP(catalog,httptest.NewRequest(http.MethodGet,"/api/v1/skills",nil))
	if catalog.Code!=http.StatusOK{t.Fatalf("catalog status=%d",catalog.Code)}
	create:=httptest.NewRecorder();handler.ServeHTTP(create,httptest.NewRequest(http.MethodPost,"/api/v1/players",bytes.NewBufferString(`{"name":"Wayfarer"}`)))
	var player game.Player
	if err:=json.Unmarshal(create.Body.Bytes(),&player);err!=nil{t.Fatal(err)}
	path:="/api/v1/players/"+string(player.ID)+"/skills/woodcutting/experience"
	award:=httptest.NewRecorder();handler.ServeHTTP(award,httptest.NewRequest(http.MethodPost,path,bytes.NewBufferString(`{"amount":100}`)))
	if award.Code!=http.StatusOK{t.Fatalf("award status=%d body=%s",award.Code,award.Body.String())}
	var result struct{LeveledUp bool `json:"leveledUp"`;Skill game.SkillProgress `json:"skill"`}
	if err:=json.Unmarshal(award.Body.Bytes(),&result);err!=nil{t.Fatal(err)}
	if !result.LeveledUp||result.Skill.Level!=2||result.Skill.Experience!=100{t.Fatalf("unexpected award response: %+v",result)}
}

func TestExperienceAwardRejectsInvalidSkillAndAmount(t *testing.T) {
	handler:=New(game.NewMemoryPlayerRepository()).Routes()
	created:=httptest.NewRecorder();handler.ServeHTTP(created,httptest.NewRequest(http.MethodPost,"/api/v1/players",bytes.NewBufferString(`{"name":"Wayfarer"}`)))
	var player game.Player;if err:=json.Unmarshal(created.Body.Bytes(),&player);err!=nil{t.Fatal(err)}
	for _,tc:=range []struct{skill,body string;want int}{{"alchemy",`{"amount":10}`,http.StatusUnprocessableEntity},{"mining",`{"amount":0}`,http.StatusUnprocessableEntity},{"mining",`{"amount":10} {"amount":20}`,http.StatusBadRequest}}{
		path:="/api/v1/players/"+string(player.ID)+"/skills/"+tc.skill+"/experience"
		response:=httptest.NewRecorder();handler.ServeHTTP(response,httptest.NewRequest(http.MethodPost,path,bytes.NewBufferString(tc.body)))
		if response.Code!=tc.want{t.Errorf("%s %s: status=%d want=%d body=%s",tc.skill,tc.body,response.Code,tc.want,response.Body.String())}
	}
}

func TestCreatePlayerRejectsUnknownFields(t *testing.T) {
	handler:=New(game.NewMemoryPlayerRepository()).Routes()
	request:=httptest.NewRequest(http.MethodPost,"/api/v1/players",bytes.NewBufferString(`{"name":"Wayfarer","admin":true}`))
	response:=httptest.NewRecorder();handler.ServeHTTP(response,request)
	if response.Code!=http.StatusBadRequest{t.Fatalf("status=%d want=%d",response.Code,http.StatusBadRequest)}
}

func TestMissingPlayerReturnsNotFound(t *testing.T) {
	handler:=New(game.NewMemoryPlayerRepository()).Routes()
	request:=httptest.NewRequest(http.MethodGet,"/api/v1/players/missing",nil)
	response:=httptest.NewRecorder();handler.ServeHTTP(response,request)
	if response.Code!=http.StatusNotFound{t.Fatalf("status=%d want=%d",response.Code,http.StatusNotFound)}
}

func TestCreatePlayerRejectsTrailingJSONData(t *testing.T) {
	handler:=New(game.NewMemoryPlayerRepository()).Routes()
	for _,body:=range []string{`{"name":"Wayfarer"} {"name":"Second"}`,`{"name":"Wayfarer"} trailing`}{
		request:=httptest.NewRequest(http.MethodPost,"/api/v1/players",bytes.NewBufferString(body))
		response:=httptest.NewRecorder();handler.ServeHTTP(response,request)
		if response.Code!=http.StatusBadRequest{t.Fatalf("body %q: status=%d want=%d",body,response.Code,http.StatusBadRequest)}
	}
}

func TestProductionDisablesDevelopmentAwardEndpoint(t *testing.T) {
	t.Setenv("RS_GALAXIES_ENV", "production")
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/player/skills/mining/experience", bytes.NewBufferString(`{"amount":10}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusMethodNotAllowed && response.Code != http.StatusNotFound {
		t.Fatalf("production development route status = %d, want route unavailable", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("production response unexpectedly enabled development CORS")
	}
	move := httptest.NewRecorder()
	handler.ServeHTTP(move, httptest.NewRequest(http.MethodPost, "/api/v1/players/player/move", bytes.NewBufferString(`{"x":1,"z":1}`)))
	if move.Code != http.StatusNotFound && move.Code != http.StatusMethodNotAllowed {
		t.Fatalf("production movement route status = %d, want route unavailable", move.Code)
	}
	gather := httptest.NewRecorder()
	handler.ServeHTTP(gather, httptest.NewRequest(http.MethodPost, "/api/v1/players/player/gather/tree-001", nil))
	if gather.Code != http.StatusNotFound && gather.Code != http.StatusMethodNotAllowed {
		t.Fatalf("production gathering route status = %d, want route unavailable", gather.Code)
	}
}

func TestDevelopmentCORSPreflight(t *testing.T) {
	t.Setenv("RS_GALAXIES_ENV", "development")
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/players", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("development CORS origin header missing")
	}
}

func TestWorldZonesAndValidatedMovement(t *testing.T) {
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	zones := httptest.NewRecorder()
	handler.ServeHTTP(zones, httptest.NewRequest(http.MethodGet, "/api/v1/world/zones", nil))
	if zones.Code != http.StatusOK {
		t.Fatalf("world zones status = %d", zones.Code)
	}

	created := httptest.NewRecorder()
	handler.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Wayfarer"}`)))
	var player game.Player
	if err := json.Unmarshal(created.Body.Bytes(), &player); err != nil {
		t.Fatal(err)
	}
	if player.Position.ZoneID != game.DefaultWorldZoneID {
		t.Fatalf("spawn zone = %q", player.Position.ZoneID)
	}

	path := "/api/v1/players/" + string(player.ID) + "/move"
	moved := httptest.NewRecorder()
	handler.ServeHTTP(moved, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{"x":0.1,"z":0}`)))
	if moved.Code != http.StatusOK {
		t.Fatalf("valid move status = %d body = %s", moved.Code, moved.Body.String())
	}

	outside := httptest.NewRecorder()
	handler.ServeHTTP(outside, httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(`{"x":200,"z":0}`)))
	if outside.Code != http.StatusUnprocessableEntity {
		t.Fatalf("out-of-bounds move status = %d, want %d", outside.Code, http.StatusUnprocessableEntity)
	}
}


func TestGatheringCreatesInventoryAndAwardsSkillExperience(t *testing.T) {
	repository := game.NewMemoryPlayerRepository()
	handler := New(repository).Routes()
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Wayfarer"}`)))
	if created.Code != http.StatusCreated { t.Fatalf("create status = %d", created.Code) }
	var player game.Player
	if err := json.Unmarshal(created.Body.Bytes(), &player); err != nil { t.Fatal(err) }

	nodes := httptest.NewRecorder()
	handler.ServeHTTP(nodes, httptest.NewRequest(http.MethodGet, "/api/v1/world/nodes", nil))
	if nodes.Code != http.StatusOK { t.Fatalf("nodes status = %d", nodes.Code) }
	items := httptest.NewRecorder()
	handler.ServeHTTP(items, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))
	if items.Code != http.StatusOK { t.Fatalf("items status = %d", items.Code) }

	path := "/api/v1/players/" + string(player.ID) + "/gather/tree-001"
	gathered := httptest.NewRecorder()
	handler.ServeHTTP(gathered, httptest.NewRequest(http.MethodPost, path, nil))
	if gathered.Code != http.StatusOK { t.Fatalf("gather status = %d body = %s", gathered.Code, gathered.Body.String()) }
	var result struct {
		Player game.Player `json:"player"`
		Gathering game.GatheringResult `json:"gathering"`
	}
	if err := json.Unmarshal(gathered.Body.Bytes(), &result); err != nil { t.Fatal(err) }
	if result.Player.Inventory["softwood_log"] != 1 || result.Player.Skills["woodcutting"].Experience != 25 {
		t.Fatalf("gather transaction incomplete: inventory=%v skill=%+v", result.Player.Inventory, result.Player.Skills["woodcutting"])
	}
}

