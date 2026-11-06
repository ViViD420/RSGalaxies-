package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ViViD420/RSGalaxies-/server/internal/game"
)

type Server struct {
	players game.PlayerRepository
	now     func() time.Time
}

func New(players game.PlayerRepository) *Server {
	return &Server{players: players, now: time.Now}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/v1/players", s.listPlayers)
	mux.HandleFunc("POST /api/v1/players", s.createPlayer)
	mux.HandleFunc("GET /api/v1/players/{id}", s.getPlayer)
	return securityHeaders(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "rs-galaxies-world"})
}

func (s *Server) listPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := s.players.List(r.Context())
	if err != nil {
		log.Printf("list players: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"players": players})
}

func (s *Server) createPlayer(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4<<10)
	defer r.Body.Close()

	var request struct {
		Name string `json:"name"`
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

	id, err := newID()
	if err != nil {
		log.Printf("generate player id: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	player, err := game.NewPlayer(game.PlayerID(id), request.Name, s.now())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid_player")
		return
	}
	if err := s.players.Create(r.Context(), player); err != nil {
		if errors.Is(err, game.ErrPlayerAlreadyExists) {
			writeError(w, http.StatusConflict, "player_conflict")
			return
		}
		log.Printf("create player: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusCreated, player)
}

func (s *Server) getPlayer(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_player_id")
		return
	}
	player, err := s.players.Get(r.Context(), game.PlayerID(id))
	if err != nil {
		if errors.Is(err, game.ErrPlayerNotFound) {
			writeError(w, http.StatusNotFound, "player_not_found")
			return
		}
		log.Printf("get player: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, player)
}

func newID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
