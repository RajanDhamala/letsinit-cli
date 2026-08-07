package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	sqlc "http-server/db/sqlc"
	"http-server/internal/middlewares"
	"http-server/internal/utils"
)

type registerInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (c *Controller) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var input registerInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Username == "" || input.Email == "" || len(input.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "username, email, and a password of at least 8 characters are required",
		})
		return
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "hash password failed"})
		return
	}

	user, err := c.queries.CreateUser(r.Context(), sqlc.CreateUserParams{
		Username: input.Username,
		Email:    input.Email,
		Password: hashedPassword,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create user failed"})
		return
	}

	accessToken, refreshToken, err := utils.CreateTokens(user.ID, user.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create tokens failed"})
		return
	}
	setAuthCookies(w, accessToken, refreshToken, os.Getenv("COOKIE_SECURE") == "true")

	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input loginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	user, err := c.queries.GetUserByEmail(r.Context(), strings.ToLower(strings.TrimSpace(input.Email)))
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && utils.ComparePassword(input.Password, user.Password) != nil) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid email or password"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "login failed"})
		return
	}

	accessToken, refreshToken, err := utils.CreateTokens(user.ID, user.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create tokens failed"})
		return
	}
	setAuthCookies(w, accessToken, refreshToken, os.Getenv("COOKIE_SECURE") == "true")

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (c *Controller) MeHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	user, err := c.queries.GetUserByID(r.Context(), claims.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "get user failed"})
		return
	}

	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (c *Controller) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	expires := time.Now().Add(-time.Hour)
	secure := os.Getenv("COOKIE_SECURE") == "true"
	http.SetCookie(w, &http.Cookie{Name: "accessToken", Value: "", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, Path: "/", Expires: expires})
	http.SetCookie(w, &http.Cookie{Name: "refreshToken", Value: "", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, Path: "/", Expires: expires})
	w.WriteHeader(http.StatusNoContent)
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(15 * time.Minute),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func toUserResponse(user sqlc.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
