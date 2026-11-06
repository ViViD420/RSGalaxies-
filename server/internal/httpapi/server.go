package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net"
	"sync"
	"os"
	"strings"
	"strconv"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
)

type characterStore interface {
	CreateCharacter(context.Context, string, game.Player) error
	GetCharacter(context.Context, string, game.PlayerID) (game.Player, error)
	ListCharacters(context.Context, string) ([]game.Player, error)
	AwardExperienceForAccount(context.Context, string, game.PlayerID, string, int) (game.Player, bool, error)
	MovePlayerForAccount(context.Context, string, game.PlayerID, float64, float64, time.Time) (game.Player, error)
	GatherForAccount(context.Context, string, game.PlayerID, string, time.Time) (game.Player, game.GatheringResult, error)
	CraftForAccount(context.Context, string, game.PlayerID, string, time.Time) (game.Player, game.CraftingResult, error)
	TransferBankItemForAccount(context.Context, string, game.PlayerID, string, int, bool) (game.Player, error)
	TradeItemForAccount(context.Context, string, game.PlayerID, string, int, bool) (game.Player, int64, error)
	EquipItemForAccount(context.Context, string, game.PlayerID, string) (game.Player, error)
	UnequipItemForAccount(context.Context, string, game.PlayerID, string) (game.Player, error)
	AttackEnemyForAccount(context.Context, string, game.PlayerID, string, time.Time) (game.Player, game.CombatResult, error)
	RestForAccount(context.Context, string, game.PlayerID) (game.Player, error)
	SendLocalChatForAccount(context.Context, string, game.PlayerID, string, time.Time) (game.ChatMessage, error)
	ListLocalChatForAccount(context.Context, string, game.PlayerID, int64, int) ([]game.ChatMessage, error)
	UseItemForAccount(context.Context, string, game.PlayerID, string) (game.Player, int, error)
}

const maxAuthRateKeys = 10000

type authAttemptWindow struct { started time.Time; count int }

type Server struct {
	authRateMu sync.Mutex
	authAttempts map[string]authAttemptWindow
	players game.PlayerRepository
	accounts auth.AccountStore
	characters characterStore
	now func() time.Time
}

func New(players game.PlayerRepository) *Server {
	return &Server{players: players, now: time.Now, authAttempts: make(map[string]authAttemptWindow)}
}

func NewWithAccounts(players game.PlayerRepository, accounts auth.AccountStore, characters characterStore) *Server {
	return &Server{players: players, accounts: accounts, characters: characters, now: time.Now, authAttempts: make(map[string]authAttemptWindow)}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/skills", s.listSkills)
	mux.HandleFunc("GET /api/v1/world/zones", s.listWorldZones)
	mux.HandleFunc("GET /api/v1/world/nodes", s.listGatheringNodes)
	mux.HandleFunc("GET /api/v1/items", s.listItems)
	mux.HandleFunc("GET /api/v1/market/items", s.listMarketItems)
	mux.HandleFunc("GET /api/v1/combat/enemies", s.listEnemies)
	mux.HandleFunc("GET /api/v1/crafting/recipes", s.listCraftingRecipes)
	mux.HandleFunc("GET /api/v1/crafting/stations", s.listCraftingStations)
	// Player APIs are development-only until account authentication and character ownership are implemented.
	if os.Getenv("RS_GALAXIES_ENV") != "production" {
		mux.HandleFunc("GET /api/v1/players", s.listPlayers)
		mux.HandleFunc("POST /api/v1/players", s.createPlayer)
		mux.HandleFunc("GET /api/v1/players/{id}", s.getPlayer)
		mux.HandleFunc("GET /api/v1/players/{id}/skills", s.getPlayerSkills)
		mux.HandleFunc("POST /api/v1/players/{id}/skills/{skill}/experience", s.awardExperience)
		mux.HandleFunc("POST /api/v1/players/{id}/move", s.movePlayer)
		mux.HandleFunc("POST /api/v1/players/{id}/gather/{nodeID}", s.gather)
		mux.HandleFunc("POST /api/v1/players/{id}/craft/{recipeID}", s.craft)
	}
	if s.accounts != nil && s.characters != nil {
		if os.Getenv("RS_GALAXIES_ENV") != "production" || os.Getenv("RS_GALAXIES_ENABLE_REGISTRATION") == "true" {
			mux.HandleFunc("POST /api/v1/auth/register", s.register)
		}
		mux.HandleFunc("POST /api/v1/auth/login", s.login)
		mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
		mux.HandleFunc("GET /api/v1/auth/me", s.me)
		mux.HandleFunc("GET /api/v1/account/characters", s.listOwnedCharacters)
		mux.HandleFunc("POST /api/v1/account/characters", s.createOwnedCharacter)
		mux.HandleFunc("GET /api/v1/account/characters/{id}", s.getOwnedCharacter)
		mux.HandleFunc("GET /api/v1/account/characters/{id}/skills", s.getOwnedCharacterSkills)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/skills/{skill}/experience", s.awardOwnedExperience)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/move", s.moveOwnedCharacter)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/gather/{nodeID}", s.gatherOwnedCharacter)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/craft/{recipeID}", s.craftOwnedCharacter)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/bank/deposit", s.depositOwnedBankItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/bank/withdraw", s.withdrawOwnedBankItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/market/buy/{itemID}", s.buyOwnedMarketItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/market/sell/{itemID}", s.sellOwnedMarketItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/equipment/equip/{itemID}", s.equipOwnedItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/equipment/unequip/{slot}", s.unequipOwnedItem)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/combat/attack/{enemyID}", s.attackOwnedEnemy)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/rest", s.restOwnedCharacter)
		mux.HandleFunc("GET /api/v1/account/characters/{id}/chat/local", s.listOwnedLocalChat)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/chat/local", s.sendOwnedLocalChat)
		mux.HandleFunc("POST /api/v1/account/characters/{id}/inventory/use/{itemID}", s.useOwnedItem)
	}
	return securityHeaders(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) { writeJSON(w,200,map[string]string{"status":"ok","service":"rs-galaxies-world"}) }
func (s *Server) listSkills(w http.ResponseWriter, _ *http.Request) { writeJSON(w,200,map[string]any{"skills":game.SkillDefinitions(),"maxLevel":game.MaxSkillLevel,"maxExperience":game.MaxSkillExperience}) }
func (s *Server) listWorldZones(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"zones": game.WorldZones()})
}


func (s *Server) listGatheringNodes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"nodes": game.GatheringNodes()})
}

func (s *Server) listItems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": game.ItemDefinitions()})
}


func (s *Server) listCraftingRecipes(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"recipes": game.CraftingRecipes()})
}

func (s *Server) listCraftingStations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"stations": game.CraftingStations()})
}

func (s *Server) listPlayers(w http.ResponseWriter,r *http.Request) {
	players,err:=s.players.List(r.Context());if err!=nil{log.Printf("list players: %v",err);writeError(w,500,"internal_error");return};writeJSON(w,200,map[string]any{"players":players})
}
func (s *Server) createPlayer(w http.ResponseWriter,r *http.Request) {
	r.Body=http.MaxBytesReader(w,r.Body,4<<10);defer r.Body.Close()
	var request struct{Name string `json:"name"`}
	decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields()
	if err:=decoder.Decode(&request);err!=nil{writeError(w,400,"invalid_json");return}
	if err:=decoder.Decode(&struct{}{});err!=io.EOF{writeError(w,400,"invalid_json");return}
	id,err:=newID();if err!=nil{log.Printf("generate player id: %v",err);writeError(w,500,"internal_error");return}
	player,err:=game.NewPlayer(game.PlayerID(id),request.Name,s.now());if err!=nil{writeError(w,422,"invalid_player");return}
	if err:=s.players.Create(r.Context(),player);err!=nil{if errors.Is(err,game.ErrPlayerAlreadyExists){writeError(w,409,"player_conflict");return};log.Printf("create player: %v",err);writeError(w,500,"internal_error");return}
	writeJSON(w,201,player)
}
func (s *Server) getPlayer(w http.ResponseWriter,r *http.Request) {
	player,err:=s.players.Get(r.Context(),game.PlayerID(strings.TrimSpace(r.PathValue("id"))));if err!=nil{s.writePlayerError(w,err);return};writeJSON(w,200,player)
}
func (s *Server) getPlayerSkills(w http.ResponseWriter,r *http.Request) {
	player,err:=s.players.Get(r.Context(),game.PlayerID(strings.TrimSpace(r.PathValue("id"))));if err!=nil{s.writePlayerError(w,err);return}
	writeJSON(w,200,map[string]any{"playerId":player.ID,"skills":player.Skills})
}
func (s *Server) awardExperience(w http.ResponseWriter,r *http.Request) {
	r.Body=http.MaxBytesReader(w,r.Body,1<<10);defer r.Body.Close()
	var request struct{Amount int `json:"amount"`}
	decoder:=json.NewDecoder(r.Body);decoder.DisallowUnknownFields()
	if err:=decoder.Decode(&request);err!=nil{writeError(w,400,"invalid_json");return}
	if err:=decoder.Decode(&struct{}{});err!=io.EOF{writeError(w,400,"invalid_json");return}
	player,leveled,err:=s.players.AwardExperience(r.Context(),game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("skill"),request.Amount)
	if err!=nil{
		switch{case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"player_not_found")
		case errors.Is(err,game.ErrUnknownSkill):writeError(w,422,"unknown_skill")
		case errors.Is(err,game.ErrInvalidExperience):writeError(w,422,"invalid_experience")
		case errors.Is(err,game.ErrExperienceOverflow):writeError(w,409,"experience_cap_reached")
		default:log.Printf("award experience: %v",err);writeError(w,500,"internal_error")};return
	}
	writeJSON(w,200,map[string]any{"player":player,"skill":player.Skills[strings.ToLower(strings.TrimSpace(r.PathValue("skill")))],"leveledUp":leveled})
}


func (s *Server) craft(w http.ResponseWriter, r *http.Request) {
	player, result, err := s.players.Craft(r.Context(), game.PlayerID(strings.TrimSpace(r.PathValue("id"))), r.PathValue("recipeID"), s.now())
	if err != nil {
		switch {
		case errors.Is(err, game.ErrPlayerNotFound):
			writeError(w, http.StatusNotFound, "player_not_found")
		case errors.Is(err, game.ErrUnknownCraftingRecipe), errors.Is(err, game.ErrUnknownCraftingStation):
			writeError(w, http.StatusNotFound, "crafting_definition_not_found")
		case errors.Is(err, game.ErrCraftingOutOfRange):
			writeError(w, http.StatusUnprocessableEntity, "crafting_out_of_range")
		case errors.Is(err, game.ErrInsufficientIngredients):
			writeError(w, http.StatusUnprocessableEntity, "insufficient_ingredients")
		case errors.Is(err, game.ErrCraftingSkillTooLow):
			writeError(w, http.StatusUnprocessableEntity, "crafting_skill_too_low")
		case errors.Is(err, game.ErrCraftingOutputFull), errors.Is(err, game.ErrExperienceOverflow):
			writeError(w, http.StatusConflict, "crafting_capacity_reached")
		default:
			log.Printf("craft item: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"player": player, "crafting": result})
}

func (s *Server) gather(w http.ResponseWriter, r *http.Request) {
	player, result, err := s.players.Gather(r.Context(), game.PlayerID(strings.TrimSpace(r.PathValue("id"))), r.PathValue("nodeID"), s.now())
	if err != nil {
		switch {
		case errors.Is(err, game.ErrPlayerNotFound):
			writeError(w, http.StatusNotFound, "player_not_found")
		case errors.Is(err, game.ErrUnknownGatheringNode):
			writeError(w, http.StatusNotFound, "gathering_node_not_found")
		case errors.Is(err, game.ErrGatheringOutOfRange):
			writeError(w, http.StatusUnprocessableEntity, "gathering_out_of_range")
		case errors.Is(err, game.ErrGatheringNodeDepleted):
			writeError(w, http.StatusConflict, "gathering_node_depleted")
		case errors.Is(err, game.ErrInsufficientSkillLevel):
			writeError(w, http.StatusUnprocessableEntity, "insufficient_skill_level")
		case errors.Is(err, game.ErrInventoryFull), errors.Is(err, game.ErrExperienceOverflow):
			writeError(w, http.StatusConflict, "inventory_or_progression_limit")
		default:
			log.Printf("gather resource: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"player": player, "gathering": result})
}

func (s *Server) movePlayer(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<10)
	defer r.Body.Close()
	var request struct {
		X float64 `json:"x"`
		Z float64 `json:"z"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	player, err := s.players.MovePlayer(r.Context(), game.PlayerID(strings.TrimSpace(r.PathValue("id"))), request.X, request.Z, s.now())
	if err != nil {
		switch {
		case errors.Is(err, game.ErrPlayerNotFound):
			writeError(w, http.StatusNotFound, "player_not_found")
		case errors.Is(err, game.ErrMovementTooFast):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "movement_too_fast", "position": player.Position})
		case errors.Is(err, game.ErrOutsideWorld):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "outside_world_bounds", "position": player.Position})
		case errors.Is(err, game.ErrInvalidMovement), errors.Is(err, game.ErrUnknownWorldZone):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": "invalid_movement", "position": player.Position})
		default:
			log.Printf("move player: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"player": player, "position": player.Position})
}

func(s *Server)writePlayerError(w http.ResponseWriter,err error){if errors.Is(err,game.ErrPlayerNotFound){writeError(w,404,"player_not_found");return};log.Printf("player lookup: %v",err);writeError(w,500,"internal_error")}
func newID()(string,error){var b [16]byte;if _,err:=rand.Read(b[:]);err!=nil{return "",err};return hex.EncodeToString(b[:]),nil}
func writeJSON(w http.ResponseWriter,status int,value any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);if err:=json.NewEncoder(w).Encode(value);err!=nil{log.Printf("write response: %v",err)}}
func writeError(w http.ResponseWriter,status int,code string){writeJSON(w,status,map[string]string{"error":code})}
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		if os.Getenv("RS_GALAXIES_ENV") != "production" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		} else {
			origin := r.Header.Get("Origin")
			for _, allowed := range strings.Split(os.Getenv("RS_GALAXIES_ALLOWED_ORIGINS"), ",") {
				if origin != "" && strings.TrimSpace(allowed) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Max-Age", "600")
					if r.Method == http.MethodOptions {
						w.WriteHeader(http.StatusNoContent)
						return
					}
					break
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

type credentialsRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func decodeStrictJSON(w http.ResponseWriter, r *http.Request, limit int64, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_json")
		return false
	}
	return true
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(w, r, 5) { return }
	var request credentialsRequest
	if !decodeStrictJSON(w, r, 4<<10, &request) { return }
	account, token, err := s.accounts.Register(r.Context(), request.Email, request.Password, s.now())
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrInvalidPassword):
			writeError(w, http.StatusUnprocessableEntity, "invalid_registration")
		case errors.Is(err, auth.ErrAccountExists):
			writeError(w, http.StatusConflict, "account_exists")
		default:
			log.Printf("register account: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"account": account, "accessToken": token, "tokenType": "Bearer", "expiresIn": int64((30*24*time.Hour).Seconds())})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.allowAuthAttempt(w, r, 10) { return }
	var request credentialsRequest
	if !decodeStrictJSON(w, r, 4<<10, &request) { return }
	account, token, err := s.accounts.Login(r.Context(), request.Email, request.Password, s.now())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		log.Printf("login account: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"account": account, "accessToken": token, "tokenType": "Bearer", "expiresIn": int64((30*24*time.Hour).Seconds())})
}

func (s *Server) bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	parts := strings.Split(header, " ")
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" || strings.Contains(parts[1], " ") {
		return "", false
	}
	return parts[1], true
}

func (s *Server) authenticatedAccount(w http.ResponseWriter, r *http.Request) (auth.Account, string, bool) {
	token, ok := s.bearerToken(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication_required")
		return auth.Account{}, "", false
	}
	account, err := s.accounts.AuthenticateSession(r.Context(), token, s.now())
	if err != nil {
		if errors.Is(err, auth.ErrSessionNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid_session")
		} else {
			log.Printf("authenticate account session: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return auth.Account{}, "", false
	}
	return account, token, true
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_, token, ok := s.authenticatedAccount(w, r)
	if !ok { return }
	if err := s.accounts.RevokeSession(r.Context(), token, s.now()); err != nil {
		log.Printf("revoke account session: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	account, _, ok := s.authenticatedAccount(w, r)
	if !ok { return }
	writeJSON(w, http.StatusOK, map[string]any{"account": account})
}

func (s *Server) listOwnedCharacters(w http.ResponseWriter, r *http.Request) {
	account, _, ok := s.authenticatedAccount(w, r)
	if !ok { return }
	players, err := s.characters.ListCharacters(r.Context(), account.ID)
	if err != nil {
		log.Printf("list owned characters: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"characters": players})
}

func (s *Server) createOwnedCharacter(w http.ResponseWriter, r *http.Request) {
	account, _, ok := s.authenticatedAccount(w, r)
	if !ok { return }
	var request struct { Name string `json:"name"` }
	if !decodeStrictJSON(w, r, 4<<10, &request) { return }
	id, err := newID()
	if err != nil {
		log.Printf("generate character ID: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	player, err := game.NewPlayer(game.PlayerID(id), request.Name, s.now())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_character")
		return
	}
	if err := s.characters.CreateCharacter(r.Context(), account.ID, player); err != nil {
		if errors.Is(err, game.ErrPlayerAlreadyExists) {
			writeError(w, http.StatusConflict, "character_conflict")
		} else if errors.Is(err, auth.ErrAccountNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid_account")
		} else {
			log.Printf("create owned character: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusCreated, player)
}

func (s *Server) getOwnedCharacter(w http.ResponseWriter, r *http.Request) {
	account, _, ok := s.authenticatedAccount(w, r)
	if !ok { return }
	player, err := s.characters.GetCharacter(r.Context(), account.ID, game.PlayerID(strings.TrimSpace(r.PathValue("id"))))
	if err != nil {
		if errors.Is(err, game.ErrPlayerNotFound) {
			writeError(w, http.StatusNotFound, "character_not_found")
		} else {
			log.Printf("get owned character: %v", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
		}
		return
	}
	writeJSON(w, http.StatusOK, player)
}

func (s *Server) getOwnedCharacterSkills(w http.ResponseWriter, r *http.Request) {
 account, _, ok := s.authenticatedAccount(w, r); if !ok { return }
 player, err := s.characters.GetCharacter(r.Context(), account.ID, game.PlayerID(strings.TrimSpace(r.PathValue("id"))))
 if err != nil { if errors.Is(err, game.ErrPlayerNotFound) { writeError(w,404,"character_not_found") } else { log.Printf("get owned skills: %v",err); writeError(w,500,"internal_error") }; return }
 writeJSON(w,200,map[string]any{"characterId":player.ID,"skills":player.Skills})
}
func (s *Server) awardOwnedExperience(w http.ResponseWriter, r *http.Request) {
 account, _, ok := s.authenticatedAccount(w,r); if !ok { return }
 var request struct { Amount int }
 if !decodeStrictJSON(w,r,1<<10,&request) { return }
 player, leveled, err := s.characters.AwardExperienceForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("skill"),request.Amount)
 if err != nil {
  switch { case errors.Is(err,game.ErrPlayerNotFound): writeError(w,404,"character_not_found")
  case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
  case errors.Is(err,game.ErrUnknownSkill): writeError(w,422,"unknown_skill")
  case errors.Is(err,game.ErrInvalidExperience): writeError(w,422,"invalid_experience")
  case errors.Is(err,game.ErrExperienceOverflow): writeError(w,409,"experience_cap_reached")
  default: log.Printf("award owned XP: %v",err); writeError(w,500,"internal_error") }; return
 }
 writeJSON(w,200,map[string]any{"character":player,"skill":player.Skills[strings.ToLower(strings.TrimSpace(r.PathValue("skill")))],"leveledUp":leveled})
}
func (s *Server) moveOwnedCharacter(w http.ResponseWriter, r *http.Request) {
 account, _, ok := s.authenticatedAccount(w,r); if !ok { return }
 var request struct { X float64; Z float64 }
 if !decodeStrictJSON(w,r,1<<10,&request) { return }
 player,err:=s.characters.MovePlayerForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),request.X,request.Z,s.now())
 if err != nil {
  switch { case errors.Is(err,game.ErrPlayerNotFound): writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
  case errors.Is(err,game.ErrMovementTooFast): writeJSON(w,422,map[string]any{"error":"movement_too_fast","position":player.Position})
  case errors.Is(err,game.ErrOutsideWorld): writeJSON(w,422,map[string]any{"error":"outside_world_bounds","position":player.Position})
  case errors.Is(err,game.ErrInvalidMovement),errors.Is(err,game.ErrUnknownWorldZone): writeJSON(w,422,map[string]any{"error":"invalid_movement","position":player.Position})
  default: log.Printf("move owned character: %v",err); writeError(w,500,"internal_error") }; return
 }
 writeJSON(w,200,map[string]any{"character":player,"position":player.Position})
}
func (s *Server) gatherOwnedCharacter(w http.ResponseWriter, r *http.Request) {
 account, _, ok := s.authenticatedAccount(w,r); if !ok { return }
 player,result,err:=s.characters.GatherForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("nodeID"),s.now())
 if err != nil {
  switch { case errors.Is(err,game.ErrPlayerNotFound): writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
  case errors.Is(err,game.ErrUnknownGatheringNode): writeError(w,404,"gathering_node_not_found")
  case errors.Is(err,game.ErrGatheringOutOfRange): writeError(w,422,"gathering_out_of_range")
  case errors.Is(err,game.ErrGatheringNodeDepleted): writeError(w,409,"gathering_node_depleted")
  case errors.Is(err,game.ErrInsufficientSkillLevel): writeError(w,422,"insufficient_skill_level")
  case errors.Is(err,game.ErrInventoryFull),errors.Is(err,game.ErrExperienceOverflow): writeError(w,409,"inventory_or_progression_limit")
  default: log.Printf("gather owned: %v",err); writeError(w,500,"internal_error") }; return
 }
 writeJSON(w,200,map[string]any{"character":player,"gathering":result})
}
func (s *Server) craftOwnedCharacter(w http.ResponseWriter, r *http.Request) {
 account, _, ok := s.authenticatedAccount(w,r); if !ok { return }
 player,result,err:=s.characters.CraftForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("recipeID"),s.now())
 if err != nil {
  switch { case errors.Is(err,game.ErrPlayerNotFound): writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
  case errors.Is(err,game.ErrUnknownCraftingRecipe),errors.Is(err,game.ErrUnknownCraftingStation): writeError(w,404,"crafting_definition_not_found")
  case errors.Is(err,game.ErrCraftingOutOfRange): writeError(w,422,"crafting_out_of_range")
  case errors.Is(err,game.ErrInsufficientIngredients): writeError(w,422,"insufficient_ingredients")
  case errors.Is(err,game.ErrCraftingSkillTooLow): writeError(w,422,"crafting_skill_too_low")
  case errors.Is(err,game.ErrCraftingOutputFull),errors.Is(err,game.ErrExperienceOverflow): writeError(w,409,"crafting_capacity_reached")
  default: log.Printf("craft owned: %v",err); writeError(w,500,"internal_error") }; return
 }
 writeJSON(w,200,map[string]any{"character":player,"crafting":result})
}

func (s *Server) allowAuthAttempt(w http.ResponseWriter,r *http.Request,limit int) bool {
 host,_,err:=net.SplitHostPort(r.RemoteAddr)
 if err!=nil { host=r.RemoteAddr }
 key:=host+"|"+r.URL.Path
 now:=s.now()
 s.authRateMu.Lock()
 defer s.authRateMu.Unlock()
 for candidate,window:=range s.authAttempts {
  if now.Sub(window.started)>=time.Minute { delete(s.authAttempts,candidate) }
 }
 window,ok:=s.authAttempts[key]
 if !ok && len(s.authAttempts)>=maxAuthRateKeys {
  oldestKey:=""
  var oldest time.Time
  for candidate,candidateWindow:=range s.authAttempts {
   if oldestKey==""||candidateWindow.started.Before(oldest) { oldestKey=candidate;oldest=candidateWindow.started }
  }
  if oldestKey!="" { delete(s.authAttempts,oldestKey) }
 }
 if !ok||now.Sub(window.started)>=time.Minute { window=authAttemptWindow{started:now} }
 if window.count>=limit {
  retry:=int(window.started.Add(time.Minute).Sub(now).Seconds())
  if retry<1 { retry=1 }
  w.Header().Set("Retry-After",strconv.Itoa(retry))
  writeError(w,http.StatusTooManyRequests,"rate_limited")
  return false
 }
 window.count++
 s.authAttempts[key]=window
 return true
}

func (s *Server) depositOwnedBankItem(w http.ResponseWriter,r *http.Request) {
	s.transferOwnedBankItem(w,r,true)
}

func (s *Server) withdrawOwnedBankItem(w http.ResponseWriter,r *http.Request) {
	s.transferOwnedBankItem(w,r,false)
}

func (s *Server) transferOwnedBankItem(w http.ResponseWriter,r *http.Request,deposit bool) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	var request struct { ItemID string `json:"itemId"`; Quantity int `json:"quantity"` }
	if !decodeStrictJSON(w,r,1<<10,&request){return}
	player,err:=s.characters.TransferBankItemForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),request.ItemID,request.Quantity,deposit)
	if err!=nil {
		switch {
		case errors.Is(err,game.ErrPlayerNotFound): writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
		case errors.Is(err,game.ErrUnknownBankItem): writeError(w,404,"unknown_item")
		case errors.Is(err,game.ErrInvalidItemQuantity): writeError(w,422,"invalid_quantity")
		case errors.Is(err,game.ErrInsufficientItems): writeError(w,409,"insufficient_items")
		case errors.Is(err,game.ErrBankFull): writeError(w,409,"bank_full")
		case errors.Is(err,game.ErrBankStackFull): writeError(w,409,"bank_stack_full")
		case errors.Is(err,game.ErrInventoryStackFull): writeError(w,409,"inventory_stack_full")
		default: log.Printf("transfer owned bank item: %v",err); writeError(w,500,"internal_error")
		}
		return
	}
	writeJSON(w,200,map[string]any{"character":player,"bank":player.Bank,"inventory":player.Inventory})
}

func (s *Server) listMarketItems(w http.ResponseWriter,r *http.Request) {
	writeJSON(w,http.StatusOK,map[string]any{"items":game.ItemDefinitions(),"startingCoins":game.StartingCoins})
}

func (s *Server) buyOwnedMarketItem(w http.ResponseWriter,r *http.Request) {
	s.tradeOwnedMarketItem(w,r,true)
}

func (s *Server) sellOwnedMarketItem(w http.ResponseWriter,r *http.Request) {
	s.tradeOwnedMarketItem(w,r,false)
}

func (s *Server) tradeOwnedMarketItem(w http.ResponseWriter,r *http.Request,buy bool) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	var request struct { Quantity int `json:"quantity"` }
	if !decodeStrictJSON(w,r,1<<10,&request){return}
	player,amount,err:=s.characters.TradeItemForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("itemID"),request.Quantity,buy)
	if err!=nil {
		switch {
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
		case errors.Is(err,game.ErrUnknownMerchantItem):writeError(w,404,"item_not_traded")
		case errors.Is(err,game.ErrInvalidItemQuantity):writeError(w,422,"invalid_quantity")
		case errors.Is(err,game.ErrInsufficientCurrency):writeError(w,409,"insufficient_coins")
		case errors.Is(err,game.ErrInsufficientItems):writeError(w,409,"insufficient_items")
		case errors.Is(err,game.ErrInventoryStackFull),errors.Is(err,game.ErrMerchantStackFull):writeError(w,409,"stack_capacity_reached")
		case errors.Is(err,game.ErrCurrencyOverflow):writeError(w,409,"currency_limit_reached")
		default:log.Printf("trade with owned character: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	action:="bought";if !buy{action="sold"}
	writeJSON(w,200,map[string]any{"character":player,"trade":map[string]any{"action":action,"itemId":r.PathValue("itemID"),"quantity":request.Quantity,"coins":amount},"coins":player.Coins,"inventory":player.Inventory})
}

func (s *Server) equipOwnedItem(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	player,err:=s.characters.EquipItemForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("itemID"))
	s.writeEquipmentResult(w,player,err)
}

func (s *Server) unequipOwnedItem(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	player,err:=s.characters.UnequipItemForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("slot"))
	s.writeEquipmentResult(w,player,err)
}

func (s *Server) writeEquipmentResult(w http.ResponseWriter,player game.Player,err error) {
	if err!=nil {
		switch {
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated): writeError(w,409,"player_defeated")
		case errors.Is(err,game.ErrItemNotEquippable):writeError(w,422,"item_not_equippable")
		case errors.Is(err,game.ErrEquipmentSlotEmpty):writeError(w,409,"equipment_slot_empty")
		case errors.Is(err,game.ErrInvalidEquipmentSlot):writeError(w,422,"invalid_equipment_slot")
		case errors.Is(err,game.ErrInsufficientItems):writeError(w,409,"item_not_in_inventory")
		case errors.Is(err,game.ErrInventoryStackFull):writeError(w,409,"inventory_stack_full")
		default:log.Printf("change character equipment: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	attack,defense:=player.EquipmentBonuses()
	writeJSON(w,200,map[string]any{"character":player,"equipment":player.Equipment,"bonuses":map[string]int{"attack":attack,"defense":defense}})
}

func (s *Server) listEnemies(w http.ResponseWriter,r *http.Request) {
	writeJSON(w,http.StatusOK,map[string]any{"enemies":game.EnemyDefinitions()})
}

func (s *Server) attackOwnedEnemy(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	player,result,err:=s.characters.AttackEnemyForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("enemyID"),s.now())
	if err!=nil {
		switch {
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrUnknownEnemy):writeError(w,404,"unknown_enemy")
		case errors.Is(err,game.ErrEnemyOutOfRange):writeError(w,409,"enemy_out_of_range")
		case errors.Is(err,game.ErrEnemyRespawning):writeError(w,409,"enemy_respawning")
		case errors.Is(err,game.ErrCombatCooldown):w.Header().Set("Retry-After","1");writeError(w,429,"combat_cooldown")
		case errors.Is(err,game.ErrPlayerDefeated):writeError(w,409,"player_defeated")
		default:log.Printf("attack with owned character: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	attack,defense:=player.EquipmentBonuses()
	writeJSON(w,200,map[string]any{"character":player,"combat":result,"bonuses":map[string]int{"attack":attack,"defense":defense}})
}

func (s *Server) restOwnedCharacter(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	player,err:=s.characters.RestForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))))
	if err!=nil {
		switch {
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrRestOutOfRange):writeError(w,409,"rest_out_of_range")
		default:log.Printf("rest owned character: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	writeJSON(w,200,map[string]any{"character":player,"health":player.Health})
}

func (s *Server) listOwnedLocalChat(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	afterID:=int64(0)
	if raw:=r.URL.Query().Get("after");raw!=""{
		value,err:=strconv.ParseInt(raw,10,64);if err!=nil||value<0{writeError(w,422,"invalid_chat_cursor");return};afterID=value
	}
	limit:=50
	if raw:=r.URL.Query().Get("limit");raw!=""{
		value,err:=strconv.Atoi(raw);if err!=nil||value<1||value>50{writeError(w,422,"invalid_chat_limit");return};limit=value
	}
	messages,err:=s.characters.ListLocalChatForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),afterID,limit)
	if err!=nil{
		if errors.Is(err,game.ErrPlayerNotFound){writeError(w,404,"character_not_found")}else{log.Printf("list local chat: %v",err);writeError(w,500,"internal_error")}
		return
	}
	writeJSON(w,200,map[string]any{"messages":messages,"after":afterID})
}

func (s *Server) sendOwnedLocalChat(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	var request struct{Message string `json:"message"`}
	if !decodeStrictJSON(w,r,2<<10,&request){return}
	message,err:=s.characters.SendLocalChatForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),request.Message,s.now())
	if err!=nil{
		switch{
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrInvalidChatMessage):writeError(w,422,"invalid_chat_message")
		case errors.Is(err,game.ErrChatCooldown):w.Header().Set("Retry-After","2");writeError(w,429,"chat_cooldown")
		default:log.Printf("send local chat: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	writeJSON(w,201,map[string]any{"message":message})
}

func (s *Server) useOwnedItem(w http.ResponseWriter,r *http.Request) {
	account,_,ok:=s.authenticatedAccount(w,r);if !ok{return}
	player,healed,err:=s.characters.UseItemForAccount(r.Context(),account.ID,game.PlayerID(strings.TrimSpace(r.PathValue("id"))),r.PathValue("itemID"))
	if err!=nil{
		switch{
		case errors.Is(err,game.ErrPlayerNotFound):writeError(w,404,"character_not_found")
		case errors.Is(err,game.ErrPlayerDefeated):writeError(w,409,"player_defeated")
		case errors.Is(err,game.ErrItemNotUsable):writeError(w,422,"item_not_usable")
		case errors.Is(err,game.ErrInsufficientItems):writeError(w,409,"item_not_in_inventory")
		case errors.Is(err,game.ErrHealthFull):writeError(w,409,"health_full")
		default:log.Printf("use item for owned character: %v",err);writeError(w,500,"internal_error")
		}
		return
	}
	writeJSON(w,200,map[string]any{"character":player,"itemId":r.PathValue("itemID"),"healed":healed,"health":player.Health,"inventory":player.Inventory})
}
