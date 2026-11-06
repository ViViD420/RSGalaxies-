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

	create := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Wayfarer"}`))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", created.Code, created.Body.String())
	}

	var player game.Player
	if err := json.Unmarshal(created.Body.Bytes(), &player); err != nil {
		t.Fatalf("decode player: %v", err)
	}
	if player.ID == "" {
		t.Fatal("created player has empty ID")
	}

	get := httptest.NewRequest(http.MethodGet, "/api/v1/players/"+string(player.ID), nil)
	got := httptest.NewRecorder()
	handler.ServeHTTP(got, get)
	if got.Code != http.StatusOK {
		t.Fatalf("get status = %d, body = %s", got.Code, got.Body.String())
	}
}

func TestCreatePlayerRejectsUnknownFields(t *testing.T) {
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(`{"name":"Wayfarer","admin":true}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestMissingPlayerReturnsNotFound(t *testing.T) {
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/missing", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}


func TestCreatePlayerRejectsTrailingJSONData(t *testing.T) {
	handler := New(game.NewMemoryPlayerRepository()).Routes()
	for _, body := range []string{
		`{"name":"Wayfarer"} {"name":"Second"}`,
		`{"name":"Wayfarer"} trailing`,
	} {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/players", bytes.NewBufferString(body))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body %q: status = %d, want %d", body, response.Code, http.StatusBadRequest)
		}
	}
}
