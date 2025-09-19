package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama/v3/src/app/battles"
	"github.com/heroiclabs/nakama/v3/src/app/bot"
	"github.com/heroiclabs/nakama/v3/src/app/groups"
	leaderboardsvc "github.com/heroiclabs/nakama/v3/src/app/leaderboard"
	"github.com/heroiclabs/nakama/v3/src/domain/shared"
)

type AuthLoginRequest struct {
	Strategy string            `json:"strategy"`
	DeviceID string            `json:"device_id"`
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Password string            `json:"password"`
	Vars     map[string]string `json:"vars"`
}

type AuthLoginResponse struct {
	UserID       string `json:"user_id"`
	SessionToken string `json:"session_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username"`
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req AuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Strategy == "device" {
		result, err := s.cfg.AuthService.AuthenticateDevice(r.Context(), req.DeviceID, req.Username, req.Vars)
		if err != nil {
			s.writeError(w, http.StatusUnauthorized, err)
			return
		}
		s.writeJSON(w, http.StatusOK, AuthLoginResponse{
			UserID:       string(result.UserID),
			SessionToken: result.SessionToken,
			RefreshToken: result.RefreshToken,
			Username:     result.Username,
		})
		return
	}
	result, err := s.cfg.AuthService.AuthenticateEmail(r.Context(), req.Email, req.Password, req.Vars)
	if err != nil {
		s.writeError(w, http.StatusUnauthorized, err)
		return
	}
	s.writeJSON(w, http.StatusOK, AuthLoginResponse{
		UserID:       string(result.UserID),
		SessionToken: result.SessionToken,
		RefreshToken: result.RefreshToken,
		Username:     result.Username,
	})
}

type CreateGroupRequest struct {
	CreatorID   string `json:"creator_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Open        bool   `json:"open"`
	AvatarURL   string `json:"avatar_url"`
	LangTag     string `json:"lang_tag"`
}

type CreateGroupResponse struct {
	GroupID string `json:"group_id"`
	Handle  string `json:"handle"`
}

func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	out, err := s.cfg.GroupService.CreateGroup(r.Context(), groups.CreateInput{
		CreatorID:   shared.PlayerID(req.CreatorID),
		Name:        req.Name,
		Description: req.Description,
		Open:        req.Open,
		AvatarURL:   req.AvatarURL,
		LangTag:     req.LangTag,
	})
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	s.writeJSON(w, http.StatusCreated, CreateGroupResponse{GroupID: string(out.GroupID), Handle: out.Handle})
}

type StartBattleRequest struct {
	LeaderID       string         `json:"leader_id"`
	IdempotencyKey string         `json:"idempotency_key"`
	Metadata       map[string]any `json:"metadata"`
	Preset         string         `json:"preset"`
}

type StartBattleResponse struct {
	BattleID string `json:"battle_id"`
	MatchID  string `json:"match_id"`
}

func (s *Server) handleStartBattle(w http.ResponseWriter, r *http.Request) {
	var req StartBattleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	out, err := s.cfg.BattleService.StartBattle(r.Context(), battles.StartCommand{
		LeaderID:       shared.PlayerID(req.LeaderID),
		IdempotencyKey: shared.IdempotencyKey(req.IdempotencyKey),
		Metadata:       req.Metadata,
		Preset:         req.Preset,
	})
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, StartBattleResponse{BattleID: string(out.BattleID), MatchID: out.MatchID})
}

type SubmitScoreRequest struct {
	PlayerID       string `json:"player_id"`
	Score          int64  `json:"score"`
	IdempotencyKey string `json:"idempotency_key"`
}

func (s *Server) handleSubmitScore(w http.ResponseWriter, r *http.Request) {
	seasonID := mux.Vars(r)["season"]
	var req SubmitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	_, err := s.cfg.LeaderboardService.Submit(r.Context(), leaderboardsvc.SubmitCommand{
		PlayerID:       shared.PlayerID(req.PlayerID),
		SeasonID:       shared.SeasonID(seasonID),
		Score:          req.Score,
		IdempotencyKey: shared.IdempotencyKey(req.IdempotencyKey),
	})
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type BotWebhookRequest struct {
	CommandID      string `json:"command_id"`
	Channel        string `json:"channel"`
	PlayerID       string `json:"player_id"`
	Payload        []byte `json:"payload"`
	IdempotencyKey string `json:"idempotency_key"`
}

type BotWebhookResponse struct {
	Accepted bool `json:"accepted"`
}

func (s *Server) handleBotWebhook(w http.ResponseWriter, r *http.Request) {
	var req BotWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	out, err := s.cfg.BotService.Handle(r.Context(), bot.CommandInput{
		CommandID:      shared.BotCommandID(req.CommandID),
		Channel:        req.Channel,
		PlayerID:       shared.PlayerID(req.PlayerID),
		Payload:        req.Payload,
		IdempotencyKey: shared.IdempotencyKey(req.IdempotencyKey),
	})
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err)
		return
	}
	s.writeJSON(w, http.StatusAccepted, BotWebhookResponse{Accepted: out.Accepted})
}
