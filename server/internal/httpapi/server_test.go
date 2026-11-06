package httpapi

import (
	"time"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
	"net/url"
	"fmt"
	"strings"
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
	craft := httptest.NewRecorder()
	handler.ServeHTTP(craft, httptest.NewRequest(http.MethodPost, "/api/v1/players/player/craft/basic-plank", nil))
	if craft.Code != http.StatusNotFound && craft.Code != http.StatusMethodNotAllowed {
		t.Fatalf("production crafting route status = %d, want route unavailable", craft.Code)
	}
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodGet, "/api/v1/players", nil),
		httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Wayfarer"}`)),
		httptest.NewRequest(http.MethodGet, "/api/v1/players/player", nil),
		httptest.NewRequest(http.MethodGet, "/api/v1/players/player/skills", nil),
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound && response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("production player API %s status = %d, want route unavailable", request.Method+" "+request.URL.Path, response.Code)
		}
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
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Headers"), "Authorization") {
		t.Fatal("development CORS preflight does not allow bearer authorization")
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


func TestCraftingCreatesOutputAndConsumesIngredients(t *testing.T) {
	repository := game.NewMemoryPlayerRepository()
	player, err := game.NewPlayer("p-craft", "Wayfarer", time.Now())
	if err != nil { t.Fatal(err) }
	player.Position.Z = 2
	player.Inventory["softwood_log"] = 2
	if err := repository.Create(context.Background(), player); err != nil { t.Fatal(err) }
	handler := New(repository).Routes()

	recipes := httptest.NewRecorder()
	handler.ServeHTTP(recipes, httptest.NewRequest(http.MethodGet, "/api/v1/crafting/recipes", nil))
	if recipes.Code != http.StatusOK { t.Fatalf("recipes status = %d", recipes.Code) }
	stations := httptest.NewRecorder()
	handler.ServeHTTP(stations, httptest.NewRequest(http.MethodGet, "/api/v1/crafting/stations", nil))
	if stations.Code != http.StatusOK { t.Fatalf("stations status = %d", stations.Code) }

	crafted := httptest.NewRecorder()
	handler.ServeHTTP(crafted, httptest.NewRequest(http.MethodPost, "/api/v1/players/p-craft/craft/basic-plank", nil))
	if crafted.Code != http.StatusOK { t.Fatalf("craft status = %d body = %s", crafted.Code, crafted.Body.String()) }
	var result struct { Player game.Player `json:"player"` }
	if err := json.Unmarshal(crafted.Body.Bytes(), &result); err != nil { t.Fatal(err) }
	if result.Player.Inventory["softwood_log"] != 0 || result.Player.Inventory["wooden_plank"] != 1 {
		t.Fatalf("craft inventory incorrect: %v", result.Player.Inventory)
	}
}


type testAccountStore struct {
	accounts map[string]auth.Account
	tokens map[string]auth.Account
	revoked map[string]bool
}

func newTestAccountStore() *testAccountStore {
	return &testAccountStore{accounts:map[string]auth.Account{},tokens:map[string]auth.Account{},revoked:map[string]bool{}}
}
func (s *testAccountStore) Register(_ context.Context,email,password string,now time.Time)(auth.Account,string,error){
	id:="account-"+email
	account:=auth.Account{ID:id,Email:email,CreatedAt:now}
	s.accounts[email]=account
	token:="token-"+id
	s.tokens[token]=account
	return account,token,nil
}
func (s *testAccountStore) Login(ctx context.Context,email,password string,now time.Time)(auth.Account,string,error){
	account,ok:=s.accounts[email];if !ok{return auth.Account{},"",auth.ErrInvalidCredentials}
	token:="login-"+account.ID;s.tokens[token]=account;return account,token,nil
}
func (s *testAccountStore) AuthenticateSession(_ context.Context,token string,_ time.Time)(auth.Account,error){
	account,ok:=s.tokens[token];if !ok||s.revoked[token]{return auth.Account{},auth.ErrSessionNotFound};return account,nil
}
func (s *testAccountStore) RevokeSession(_ context.Context,token string,_ time.Time)error{s.revoked[token]=true;return nil}

type testCharacterStore struct { players map[string]map[game.PlayerID]game.Player; enemies map[string]game.EnemyState; messages []game.ChatMessage; lastChat map[game.PlayerID]time.Time; nextChatID int64; presence map[game.PlayerID]game.RemotePlayer }
func newTestCharacterStore()*testCharacterStore{return &testCharacterStore{players:map[string]map[game.PlayerID]game.Player{},enemies:map[string]game.EnemyState{},lastChat:map[game.PlayerID]time.Time{},presence:map[game.PlayerID]game.RemotePlayer{}}}
func (s *testCharacterStore) CreateCharacter(_ context.Context,accountID string,player game.Player)error{
	if s.players[accountID]==nil{s.players[accountID]=map[game.PlayerID]game.Player{}}
	if _,ok:=s.players[accountID][player.ID];ok{return game.ErrPlayerAlreadyExists}
	s.players[accountID][player.ID]=player;return nil
}
func (s *testCharacterStore) GetCharacter(_ context.Context,accountID string,id game.PlayerID)(game.Player,error){
	player,ok:=s.players[accountID][id];if !ok{return game.Player{},game.ErrPlayerNotFound};return player,nil
}
func (s *testCharacterStore) ListCharacters(_ context.Context,accountID string)([]game.Player,error){
	players:=make([]game.Player,0,len(s.players[accountID]));for _,player:=range s.players[accountID]{players=append(players,player)};return players,nil
}

func TestAuthenticatedCharacterLifecycleAndOwnership(t *testing.T) {
	t.Setenv("RS_GALAXIES_ENV","production")
	t.Setenv("RS_GALAXIES_ENABLE_REGISTRATION","true")
	accounts:=newTestAccountStore()
	characters:=newTestCharacterStore()
	handler:=NewWithAccounts(game.NewMemoryPlayerRepository(),accounts,characters).Routes()

	unauthorized:=httptest.NewRecorder()
	handler.ServeHTTP(unauthorized,httptest.NewRequest(http.MethodGet,"/api/v1/account/characters",nil))
	if unauthorized.Code!=http.StatusUnauthorized{t.Fatalf("unauthenticated list status=%d",unauthorized.Code)}

	register:=httptest.NewRecorder()
	handler.ServeHTTP(register,httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",bytes.NewBufferString(`{"email":"owner@example.test","password":"long-enough-password"}`)))
	if register.Code!=http.StatusCreated{t.Fatalf("register status=%d body=%s",register.Code,register.Body.String())}
	var registration struct{AccessToken string `json:"accessToken"`;Account auth.Account `json:"account"`}
	if err:=json.Unmarshal(register.Body.Bytes(),&registration);err!=nil{t.Fatal(err)}
	if registration.AccessToken==""||registration.Account.ID==""{t.Fatal("registration response omitted account or token")}
	authHeader:=func(token string)*http.Request{request:=httptest.NewRequest(http.MethodGet,"/",nil);request.Header.Set("Authorization","Bearer "+token);return request}

	create:=httptest.NewRecorder()
	createRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters",bytes.NewBufferString(`{"name":"OwnedWayfarer"}`))
	createRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(create,createRequest)
	if create.Code!=http.StatusCreated{t.Fatalf("create character status=%d body=%s",create.Code,create.Body.String())}
	var player game.Player
	if err:=json.Unmarshal(create.Body.Bytes(),&player);err!=nil{t.Fatal(err)}
	sendChat:=httptest.NewRecorder()
	sendChatRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/chat/local",bytes.NewBufferString(`{"message":"Hello, frontier!"}`))
	sendChatRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(sendChat,sendChatRequest)
	if sendChat.Code!=http.StatusCreated{t.Fatalf("send chat status=%d body=%s",sendChat.Code,sendChat.Body.String())}
	listChat:=httptest.NewRecorder()
	listChatRequest:=httptest.NewRequest(http.MethodGet,"/api/v1/account/characters/"+string(player.ID)+"/chat/local?after=0",nil)
	listChatRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(listChat,listChatRequest)
	if listChat.Code!=http.StatusOK||!strings.Contains(listChat.Body.String(),"Hello, frontier!"){t.Fatalf("list chat status=%d body=%s",listChat.Code,listChat.Body.String())}
	cooldownChat:=httptest.NewRecorder()
	cooldownChatRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/chat/local",bytes.NewBufferString(`{"message":"Too soon"}`))
	cooldownChatRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(cooldownChat,cooldownChatRequest)
	if cooldownChat.Code!=http.StatusTooManyRequests{t.Fatalf("chat cooldown status=%d body=%s",cooldownChat.Code,cooldownChat.Body.String())}
	characters.players[registration.Account.ID][player.ID].Inventory["softwood_log"]=3
	buy:=httptest.NewRecorder()
	buyRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/market/buy/rough_stone",bytes.NewBufferString(`{"quantity":2}`))
	buyRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(buy,buyRequest)
	if buy.Code!=http.StatusOK{t.Fatalf("market buy status=%d body=%s",buy.Code,buy.Body.String())}
	sell:=httptest.NewRecorder()
	sellRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/market/sell/rough_stone",bytes.NewBufferString(`{"quantity":1}`))
	sellRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(sell,sellRequest)
	if sell.Code!=http.StatusOK{t.Fatalf("market sell status=%d body=%s",sell.Code,sell.Body.String())}
	if characters.players[registration.Account.ID][player.ID].Coins!=95||characters.players[registration.Account.ID][player.ID].Inventory["rough_stone"]!=1{t.Fatal("market transactions did not update currency and inventory")}
	buyGear:=httptest.NewRecorder()
	buyGearRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/market/buy/training_sword",bytes.NewBufferString(`{"quantity":1}`))
	buyGearRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(buyGear,buyGearRequest)
	if buyGear.Code!=http.StatusOK{t.Fatalf("buy gear status=%d body=%s",buyGear.Code,buyGear.Body.String())}
	equip:=httptest.NewRecorder()
	equipRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/equipment/equip/training_sword",nil)
	equipRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(equip,equipRequest)
	if equip.Code!=http.StatusOK{t.Fatalf("equip status=%d body=%s",equip.Code,equip.Body.String())}
	combatPlayer:=characters.players[registration.Account.ID][player.ID]
	combatPlayer.Position.X=6
	combatPlayer.LastCombatAt=time.Now().Add(-2*time.Second)
	characters.players[registration.Account.ID][player.ID]=combatPlayer
	attack:=httptest.NewRecorder()
	attackRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/combat/attack/frontier-rat-001",nil)
	attackRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(attack,attackRequest)
	if attack.Code!=http.StatusOK{t.Fatalf("combat attack status=%d body=%s",attack.Code,attack.Body.String())}
	if characters.players[registration.Account.ID][player.ID].Health!=98{t.Fatalf("combat did not apply retaliation: health=%d",characters.players[registration.Account.ID][player.ID].Health)}
	restAway:=httptest.NewRecorder()
	restAwayRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/rest",nil)
	restAwayRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(restAway,restAwayRequest)
	if restAway.Code!=http.StatusConflict{t.Fatalf("rest away from spawn status=%d body=%s",restAway.Code,restAway.Body.String())}
	combatPlayer=characters.players[registration.Account.ID][player.ID]
	combatPlayer.Position.X=0
	characters.players[registration.Account.ID][player.ID]=combatPlayer
	rest:=httptest.NewRecorder()
	restRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/rest",nil)
	restRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(rest,restRequest)
	if rest.Code!=http.StatusOK||characters.players[registration.Account.ID][player.ID].Health!=game.MaxPlayerHealth{t.Fatalf("safe rest status=%d health=%d body=%s",rest.Code,characters.players[registration.Account.ID][player.ID].Health,rest.Body.String())}
	buyPotion:=httptest.NewRecorder()
	buyPotionRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/market/buy/minor_healing_potion",bytes.NewBufferString(`{"quantity":1}`))
	buyPotionRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(buyPotion,buyPotionRequest)
	if buyPotion.Code!=http.StatusOK{t.Fatalf("buy potion status=%d body=%s",buyPotion.Code,buyPotion.Body.String())}
	combatPlayer=characters.players[registration.Account.ID][player.ID]
	combatPlayer.Health=80
	characters.players[registration.Account.ID][player.ID]=combatPlayer
	usePotion:=httptest.NewRecorder()
	usePotionRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/inventory/use/minor_healing_potion",nil)
	usePotionRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(usePotion,usePotionRequest)
	if usePotion.Code!=http.StatusOK||characters.players[registration.Account.ID][player.ID].Health!=100||characters.players[registration.Account.ID][player.ID].Inventory["minor_healing_potion"]!=0{t.Fatalf("use potion status=%d player=%+v body=%s",usePotion.Code,characters.players[registration.Account.ID][player.ID],usePotion.Body.String())}
	unequip:=httptest.NewRecorder()
	unequipRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/equipment/unequip/main_hand",nil)
	unequipRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(unequip,unequipRequest)
	if unequip.Code!=http.StatusOK{t.Fatalf("unequip status=%d body=%s",unequip.Code,unequip.Body.String())}
	deposit:=httptest.NewRecorder()
	depositRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/bank/deposit",bytes.NewBufferString(`{"itemId":"softwood_log","quantity":2}`))
	depositRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(deposit,depositRequest)
	if deposit.Code!=http.StatusOK{t.Fatalf("bank deposit status=%d body=%s",deposit.Code,deposit.Body.String())}
	withdraw:=httptest.NewRecorder()
	withdrawRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(player.ID)+"/bank/withdraw",bytes.NewBufferString(`{"itemId":"softwood_log","quantity":1}`))
	withdrawRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(withdraw,withdrawRequest)
	if withdraw.Code!=http.StatusOK{t.Fatalf("bank withdraw status=%d body=%s",withdraw.Code,withdraw.Body.String())}
	if characters.players[registration.Account.ID][player.ID].Bank["softwood_log"]!=1{t.Fatal("bank balance did not persist in test store")}

	secondRegister:=httptest.NewRecorder()
	handler.ServeHTTP(secondRegister,httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",bytes.NewBufferString(`{"email":"other@example.test","password":"long-enough-password"}`)))
	var second struct{AccessToken string `json:"accessToken"`}
	if err:=json.Unmarshal(secondRegister.Body.Bytes(),&second);err!=nil{t.Fatal(err)}

	experiencePath:="/api/v1/account/characters/"+string(player.ID)+"/skills/mining/experience"
	xp:=httptest.NewRecorder()
	xpRequest:=httptest.NewRequest(http.MethodPost,experiencePath,bytes.NewBufferString(`{"amount":50}`))
	xpRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(xp,xpRequest)
	if xp.Code!=http.StatusOK{t.Fatalf("owned XP award status=%d body=%s",xp.Code,xp.Body.String())}
	foreignXP:=httptest.NewRecorder()
	foreignXPRequest:=httptest.NewRequest(http.MethodPost,experiencePath,bytes.NewBufferString(`{"amount":50}`))
	foreignXPRequest.Header.Set("Authorization","Bearer "+second.AccessToken)
	handler.ServeHTTP(foreignXP,foreignXPRequest)
	if foreignXP.Code!=http.StatusNotFound{t.Fatalf("cross-account XP award status=%d",foreignXP.Code)}
	foreign:=httptest.NewRecorder()
	foreignRequest:=authHeader(second.AccessToken)
	foreignRequest.Method=http.MethodGet
	foreignRequest.URL=mustParseTestURL("/api/v1/account/characters/"+string(player.ID))
	handler.ServeHTTP(foreign,foreignRequest)
	if foreign.Code!=http.StatusNotFound{t.Fatalf("cross-account character lookup status=%d",foreign.Code)}

	me:=httptest.NewRecorder()
	meRequest:=authHeader(registration.AccessToken);meRequest.URL=mustParseTestURL("/api/v1/auth/me")
	handler.ServeHTTP(me,meRequest)
	if me.Code!=http.StatusOK{t.Fatalf("authenticated account status=%d",me.Code)}

	logout:=httptest.NewRecorder()
	logoutRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/logout",nil)
	logoutRequest.Header.Set("Authorization","Bearer "+registration.AccessToken)
	handler.ServeHTTP(logout,logoutRequest)
	if logout.Code!=http.StatusNoContent{t.Fatalf("logout status=%d",logout.Code)}
	afterLogout:=httptest.NewRecorder()
	afterRequest:=authHeader(registration.AccessToken);afterRequest.URL=mustParseTestURL("/api/v1/auth/me")
	handler.ServeHTTP(afterLogout,afterRequest)
	if afterLogout.Code!=http.StatusUnauthorized{t.Fatalf("revoked token status=%d",afterLogout.Code)}
}

func mustParseTestURL(path string)*url.URL{
	parsed,err:=url.Parse(path);if err!=nil{panic(err)};return parsed
}

func (s *testCharacterStore) AwardExperienceForAccount(_ context.Context, accountID string, id game.PlayerID, skill string, amount int) (game.Player,bool,error) {
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},false,game.ErrPlayerNotFound}
 leveled,err:=player.AwardExperience(skill,amount);if err!=nil{return game.Player{},false,err};s.players[accountID][id]=player;return player,leveled,nil
}
func (s *testCharacterStore) MovePlayerForAccount(_ context.Context, accountID string, id game.PlayerID, x,z float64, now time.Time)(game.Player,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},game.ErrPlayerNotFound}
 if err:=player.MoveTo(x,z,now);err!=nil{return player,err};s.players[accountID][id]=player;return player,nil
}
func (s *testCharacterStore) GatherForAccount(_ context.Context, accountID string, id game.PlayerID, node string, now time.Time)(game.Player,game.GatheringResult,error){
 return game.Player{},game.GatheringResult{},game.ErrPlayerNotFound
}
func (s *testCharacterStore) CraftForAccount(_ context.Context, accountID string, id game.PlayerID, recipe string, now time.Time)(game.Player,game.CraftingResult,error){
 return game.Player{},game.CraftingResult{},game.ErrPlayerNotFound
}

func TestProductionRegistrationIsDisabledByDefault(t *testing.T) {
 t.Setenv("RS_GALAXIES_ENV","production")
 t.Setenv("RS_GALAXIES_ENABLE_REGISTRATION","")
 handler:=NewWithAccounts(game.NewMemoryPlayerRepository(),newTestAccountStore(),newTestCharacterStore()).Routes()
 response:=httptest.NewRecorder()
 handler.ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",bytes.NewBufferString(`{"email":"blocked@example.test","password":"long-enough-password"}`)))
 if response.Code!=http.StatusNotFound&&response.Code!=http.StatusMethodNotAllowed{t.Fatalf("registration should be disabled by default in production, status=%d",response.Code)}
}

func TestAuthenticationRegistrationRateLimit(t *testing.T) {
 t.Setenv("RS_GALAXIES_ENV","development")
 handler:=NewWithAccounts(game.NewMemoryPlayerRepository(),newTestAccountStore(),newTestCharacterStore()).Routes()
 for i:=0;i<6;i++ {
  response:=httptest.NewRecorder()
  body:=fmt.Sprintf(`{"email":"limit-%d@example.test","password":"long-enough-password"}`,i)
  handler.ServeHTTP(response,httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",bytes.NewBufferString(body)))
  want:=http.StatusCreated;if i==5{want=http.StatusTooManyRequests}
  if response.Code!=want{t.Fatalf("attempt %d status=%d want=%d body=%s",i+1,response.Code,want,response.Body.String())}
 }
}

func TestProductionCORSUsesExplicitOriginAllowlist(t *testing.T) {
 t.Setenv("RS_GALAXIES_ENV","production")
 t.Setenv("RS_GALAXIES_ALLOWED_ORIGINS","https://game.example.test, https://admin.example.test")
 handler:=New(game.NewMemoryPlayerRepository()).Routes()
 allowed:=httptest.NewRequest(http.MethodOptions,"/api/v1/skills",nil)
 allowed.Header.Set("Origin","https://game.example.test")
 allowed.Header.Set("Access-Control-Request-Headers","authorization")
 response:=httptest.NewRecorder();handler.ServeHTTP(response,allowed)
 if response.Code!=http.StatusNoContent{t.Fatalf("allowed preflight status=%d",response.Code)}
 if response.Header().Get("Access-Control-Allow-Origin")!="https://game.example.test"{t.Fatal("configured production origin was not allowed")}
 if !strings.Contains(response.Header().Get("Access-Control-Allow-Headers"),"Authorization"){t.Fatal("production CORS omitted Authorization")}
 blocked:=httptest.NewRequest(http.MethodOptions,"/api/v1/skills",nil)
 blocked.Header.Set("Origin","https://untrusted.example.test")
 blockedResponse:=httptest.NewRecorder();handler.ServeHTTP(blockedResponse,blocked)
 if blockedResponse.Header().Get("Access-Control-Allow-Origin")!=""{t.Fatal("unconfigured production origin was allowed")}
}

func (s *testCharacterStore) TransferBankItemForAccount(_ context.Context,accountID string,id game.PlayerID,itemID string,quantity int,deposit bool)(game.Player,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},game.ErrPlayerNotFound}
 var err error
 if deposit{err=player.DepositItem(itemID,quantity)}else{err=player.WithdrawItem(itemID,quantity)}
 if err!=nil{return player,err};s.players[accountID][id]=player;return player,nil
}

func (s *testCharacterStore) TradeItemForAccount(_ context.Context,accountID string,id game.PlayerID,itemID string,quantity int,buy bool)(game.Player,int64,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},0,game.ErrPlayerNotFound}
 var amount int64;var err error
 if buy{amount,err=player.BuyItem(itemID,quantity)}else{amount,err=player.SellItem(itemID,quantity)}
 if err!=nil{return player,0,err};s.players[accountID][id]=player;return player,amount,nil
}

func TestMarketCatalogIsPublicAndContainsServerPrices(t *testing.T) {
 t.Setenv("RS_GALAXIES_ENV","production")
 handler:=New(game.NewMemoryPlayerRepository()).Routes()
 response:=httptest.NewRecorder()
 handler.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/api/v1/market/items",nil))
 if response.Code!=http.StatusOK{t.Fatalf("market catalog status=%d",response.Code)}
 var payload struct{Items []game.ItemDefinition `json:"items"`;StartingCoins int64 `json:"startingCoins"`}
 if err:=json.Unmarshal(response.Body.Bytes(),&payload);err!=nil{t.Fatal(err)}
 if len(payload.Items)!=12||payload.StartingCoins!=game.StartingCoins{t.Fatalf("unexpected market catalog: %+v",payload)}
 if payload.Items[0].BuyPrice<=payload.Items[0].SellPrice{t.Fatal("merchant catalog should expose a grounded buy/sell spread")}
}

func (s *testCharacterStore) EquipItemForAccount(_ context.Context, accountID string, id game.PlayerID, itemID string) (game.Player,error) { player,ok:=s.players[accountID][id]; if !ok{return game.Player{},game.ErrPlayerNotFound}; if err:=player.EquipItem(itemID);err!=nil{return player,err};s.players[accountID][id]=player;return player,nil }
func (s *testCharacterStore) UnequipItemForAccount(_ context.Context, accountID string, id game.PlayerID, slot string) (game.Player,error) { player,ok:=s.players[accountID][id]; if !ok{return game.Player{},game.ErrPlayerNotFound}; if err:=player.UnequipItem(slot);err!=nil{return player,err};s.players[accountID][id]=player;return player,nil }

func (s *testCharacterStore) AttackEnemyForAccount(_ context.Context,accountID string,id game.PlayerID,enemyID string,now time.Time)(game.Player,game.CombatResult,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},game.CombatResult{},game.ErrPlayerNotFound}
 enemy,ok:=game.FindEnemy(enemyID);if !ok{return game.Player{},game.CombatResult{},game.ErrUnknownEnemy}
 result,state,err:=player.AttackEnemy(enemy,s.enemies[enemyID],now);if err!=nil{return player,game.CombatResult{},err}
 s.enemies[enemyID]=state;s.players[accountID][id]=player;return player,result,nil
}
func (s *testCharacterStore) RestForAccount(_ context.Context,accountID string,id game.PlayerID)(game.Player,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},game.ErrPlayerNotFound}
 if err:=player.Rest();err!=nil{return player,err};s.players[accountID][id]=player;return player,nil
}

func TestEnemyCatalogIsPublic(t *testing.T) {
 t.Setenv("RS_GALAXIES_ENV","production")
 handler:=New(game.NewMemoryPlayerRepository()).Routes()
 response:=httptest.NewRecorder()
 handler.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"/api/v1/combat/enemies",nil))
 if response.Code!=http.StatusOK{t.Fatalf("enemy catalog status=%d body=%s",response.Code,response.Body.String())}
 var payload struct{Enemies []game.EnemyDefinition `json:"enemies"`}
 if err:=json.Unmarshal(response.Body.Bytes(),&payload);err!=nil{t.Fatal(err)}
 if len(payload.Enemies)!=2{t.Fatalf("enemy definitions=%d want=2",len(payload.Enemies))}
}

func (s *testCharacterStore) SendLocalChatForAccount(_ context.Context,accountID string,id game.PlayerID,message string,now time.Time)(game.ChatMessage,error){
 player,ok:=s.players[accountID][id];if !ok{return game.ChatMessage{},game.ErrPlayerNotFound}
 cleaned,err:=game.ValidateChatMessage(message);if err!=nil{return game.ChatMessage{},err}
 if last:=s.lastChat[id];!last.IsZero()&&now.Sub(last)<game.ChatCooldown{return game.ChatMessage{},game.ErrChatCooldown}
 s.nextChatID++
 result:=game.ChatMessage{ID:s.nextChatID,ZoneID:player.Position.ZoneID,CharacterID:id,CharacterName:player.Name,Message:cleaned,CreatedAt:now}
 s.messages=append(s.messages,result);s.lastChat[id]=now;return result,nil
}
func (s *testCharacterStore) ListLocalChatForAccount(_ context.Context,accountID string,id game.PlayerID,afterID int64,limit int)([]game.ChatMessage,error){
 player,ok:=s.players[accountID][id];if !ok{return nil,game.ErrPlayerNotFound}
 result:=make([]game.ChatMessage,0)
 for _,message:=range s.messages{if message.ZoneID==player.Position.ZoneID&&message.ID>afterID{result=append(result,message);if len(result)>=limit{break}}}
 return result,nil
}

func (s *testCharacterStore) UseItemForAccount(_ context.Context,accountID string,id game.PlayerID,itemID string)(game.Player,int,error){
 player,ok:=s.players[accountID][id];if !ok{return game.Player{},0,game.ErrPlayerNotFound}
 healed,err:=player.UseItem(itemID);if err!=nil{return player,0,err};s.players[accountID][id]=player;return player,healed,nil
}


func TestAuthenticationRateLimiterRejectsNewKeysWhenCapacityIsFull(t *testing.T) {
	now:=time.Date(2026,11,6,18,0,0,0,time.UTC)
	server:=New(game.NewMemoryPlayerRepository())
	server.now=func()time.Time{return now}
	server.authAttempts=make(map[string]authAttemptWindow,maxAuthRateKeys)
	for i:=0;i<maxAuthRateKeys;i++ {
		started:=now
		if i==0 { started=now.Add(-time.Second) }
		server.authAttempts[fmt.Sprintf("host-%d|/api/v1/auth/login",i)]=authAttemptWindow{started:started,count:1}
	}
	request:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/login",nil)
	request.RemoteAddr="new-host:12345"
	response:=httptest.NewRecorder()
	if server.allowAuthAttempt(response,request,5){t.Fatal("new key should be rejected while rate-limit capacity is full")}
	if response.Code!=http.StatusTooManyRequests||response.Header().Get("Retry-After")!="60"{t.Fatalf("capacity response status=%d retry-after=%q",response.Code,response.Header().Get("Retry-After"))}
	if len(server.authAttempts)!=maxAuthRateKeys{t.Fatalf("rate-limit key count=%d want=%d",len(server.authAttempts),maxAuthRateKeys)}
	if _,ok:=server.authAttempts["host-0|/api/v1/auth/login"];!ok{t.Fatal("active key was evicted under pressure")}
	if _,ok:=server.authAttempts["new-host|/api/v1/auth/login"];ok{t.Fatal("rejected key was retained")}
}


func TestReadinessSeparatesLivenessFromDependencyHealth(t *testing.T) {
	server:=New(game.NewMemoryPlayerRepository())
	handler:=server.Routes()
	live:=httptest.NewRecorder()
	handler.ServeHTTP(live,httptest.NewRequest(http.MethodGet,"/healthz",nil))
	if live.Code!=http.StatusOK{t.Fatalf("liveness status=%d body=%s",live.Code,live.Body.String())}
	ready:=httptest.NewRecorder()
	handler.ServeHTTP(ready,httptest.NewRequest(http.MethodGet,"/readyz",nil))
	if ready.Code!=http.StatusOK{t.Fatalf("development readiness status=%d body=%s",ready.Code,ready.Body.String())}

	probeCalls:=0
	server.SetReadinessCheck(func(ctx context.Context)error{probeCalls++;return fmt.Errorf("private database connection detail")})
	handler=server.Routes()
	notReady:=httptest.NewRecorder()
	handler.ServeHTTP(notReady,httptest.NewRequest(http.MethodGet,"/readyz",nil))
	if notReady.Code!=http.StatusServiceUnavailable{t.Fatalf("failed dependency readiness status=%d body=%s",notReady.Code,notReady.Body.String())}
	if strings.Contains(notReady.Body.String(),"private database"){t.Fatal("readiness response leaked dependency details")}
	if probeCalls!=1{t.Fatalf("readiness probe calls=%d want=1",probeCalls)}
	stillLive:=httptest.NewRecorder()
	handler.ServeHTTP(stillLive,httptest.NewRequest(http.MethodGet,"/healthz",nil))
	if stillLive.Code!=http.StatusOK{t.Fatalf("liveness should stay healthy during DB outage: status=%d",stillLive.Code)}
}

func (s *testCharacterStore) HeartbeatPresenceForAccount(_ context.Context,accountID string,id game.PlayerID,now time.Time)error {
	player,ok:=s.players[accountID][id];if !ok{return game.ErrPlayerNotFound}
	s.presence[id]=game.RemotePlayer{CharacterID:id,Name:player.Name,ZoneID:player.Position.ZoneID,X:player.Position.X,Z:player.Position.Z,LastSeenAt:now}
	return nil
}
func (s *testCharacterStore) ListZonePresenceForAccount(_ context.Context,accountID string,id game.PlayerID,now time.Time)([]game.RemotePlayer,error) {
	player,ok:=s.players[accountID][id];if !ok{return nil,game.ErrPlayerNotFound}
	result:=make([]game.RemotePlayer,0)
	for otherID,remote:=range s.presence{if otherID!=id&&remote.ZoneID==player.Position.ZoneID&&now.Sub(remote.LastSeenAt)<=game.PresenceLease{result=append(result,remote)}}
	return result,nil
}

func TestAuthenticatedZonePresenceHeartbeatAndRoster(t *testing.T) {
	t.Setenv("RS_GALAXIES_ENV","development")
	characters:=newTestCharacterStore()
	handler:=NewWithAccounts(game.NewMemoryPlayerRepository(),newTestAccountStore(),characters).Routes()
	first:=registerTestAccount(t,handler,"presence-one@example.test")
	second:=registerTestAccount(t,handler,"presence-two@example.test")
	firstPlayer:=createTestCharacter(t,handler,first.AccessToken,"PresenceOne")
	secondPlayer:=createTestCharacter(t,handler,second.AccessToken,"PresenceTwo")
	heartbeat:=httptest.NewRecorder()
	heartbeatRequest:=httptest.NewRequest(http.MethodPost,"/api/v1/account/characters/"+string(secondPlayer.ID)+"/presence/heartbeat",nil)
	heartbeatRequest.Header.Set("Authorization","Bearer "+second.AccessToken)
	handler.ServeHTTP(heartbeat,heartbeatRequest)
	if heartbeat.Code!=http.StatusNoContent{t.Fatalf("presence heartbeat status=%d body=%s",heartbeat.Code,heartbeat.Body.String())}
	roster:=httptest.NewRecorder()
	rosterRequest:=httptest.NewRequest(http.MethodGet,"/api/v1/account/characters/"+string(firstPlayer.ID)+"/presence",nil)
	rosterRequest.Header.Set("Authorization","Bearer "+first.AccessToken)
	handler.ServeHTTP(roster,rosterRequest)
	if roster.Code!=http.StatusOK{t.Fatalf("presence roster status=%d body=%s",roster.Code,roster.Body.String())}
	var payload struct{Players []game.RemotePlayer `json:"players"`;LeaseSeconds int `json:"leaseSeconds"`}
	if err:=json.Unmarshal(roster.Body.Bytes(),&payload);err!=nil{t.Fatal(err)}
	if len(payload.Players)!=1||payload.Players[0].CharacterID!=secondPlayer.ID||payload.Players[0].Name!="PresenceTwo"{t.Fatalf("presence roster=%+v",payload)}
	if payload.LeaseSeconds!=int(game.PresenceLease.Seconds()){t.Fatalf("lease seconds=%d",payload.LeaseSeconds)}
	if strings.Contains(roster.Body.String(),first.Account.ID)||strings.Contains(roster.Body.String(),second.Account.ID){t.Fatal("presence roster leaked account identifiers")}
	foreign:=httptest.NewRecorder()
	foreignRequest:=httptest.NewRequest(http.MethodGet,"/api/v1/account/characters/"+string(secondPlayer.ID)+"/presence",nil)
	foreignRequest.Header.Set("Authorization","Bearer "+first.AccessToken)
	handler.ServeHTTP(foreign,foreignRequest)
	if foreign.Code!=http.StatusNotFound{t.Fatalf("foreign presence status=%d body=%s",foreign.Code,foreign.Body.String())}
}
