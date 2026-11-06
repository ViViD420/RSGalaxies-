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
	"os"
	"strings"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
	"github.com/ViViD420/RSGalaxies-/server/internal/auth"
)

type characterStore interface {
	CreateCharacter(context.Context, string, game.Player) error
	GetCharacter(context.Context, string, game.PlayerID) (game.Player, error)
	ListCharacters(context.Context, string) ([]game.Player, error)
}

type Server struct {
	players game.PlayerRepository
	accounts auth.AccountStore
	characters characterStore
	now func() time.Time
}

func New(players game.PlayerRepository) *Server {
	return &Server{players: players, now: time.Now}
}

func NewWithAccounts(players game.PlayerRepository, accounts auth.AccountStore, characters characterStore) *Server {
	return &Server{players: players, accounts: accounts, characters: characters, now: time.Now}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/skills", s.listSkills)
	mux.HandleFunc("GET /api/v1/world/zones", s.listWorldZones)
	mux.HandleFunc("GET /api/v1/world/nodes", s.listGatheringNodes)
	mux.HandleFunc("GET /api/v1/items", s.listItems)
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
		mux.HandleFunc("POST /api/v1/auth/register", s.register)
		mux.HandleFunc("POST /api/v1/auth/login", s.login)
		mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
		mux.HandleFunc("GET /api/v1/auth/me", s.me)
		mux.HandleFunc("GET /api/v1/account/characters", s.listOwnedCharacters)
		mux.HandleFunc("POST /api/v1/account/characters", s.createOwnedCharacter)
		mux.HandleFunc("GET /api/v1/account/characters/{id}", s.getOwnedCharacter)
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
		w.Header().Set("Cache-Control", "no-store")
		if os.Getenv("RS_GALAXIES_ENV") != "production" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
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
