package handler

import (
	"errors"
	"log/slog"
	"time"

	"console/internal/auth"
	"console/internal/model"
	"console/internal/store"

	"github.com/gofiber/fiber/v2"
)

// AccessTokenTTL is the lifetime of a freshly minted access JWT. Keep
// it short so a leaked token is only useful for a bounded window.
const AccessTokenTTL = 24 * time.Hour

// AuthHandler handles authentication-related HTTP requests. The user
// store is backed by MySQL (see internal/store); the previous
// in-memory map seeded with a hard-coded admin has been removed.
type AuthHandler struct {
	jwtManager *auth.JWTManager
	users      *store.UserStore
}

// NewAuthHandler creates a new AuthHandler. The previous constructor
// signature took only a JWT secret; callers must now pass the user
// store wired to MySQL.
func NewAuthHandler(jwtSecret string, users *store.UserStore) *AuthHandler {
	return &AuthHandler{
		jwtManager: auth.NewJWTManager(jwtSecret),
		users:      users,
	}
}

// LoginRequest is the JSON body of POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the success body for /login.
type LoginResponse struct {
	Token     string             `json:"token"`
	User      model.UserResponse `json:"user"`
	ExpiresAt time.Time          `json:"expires_at"`
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("login_parse_error", slog.String("ip", c.IP()))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	user, err := h.users.GetByUsername(c.Context(), req.Username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			slog.Warn("login_user_not_found", slog.String("username", req.Username), slog.String("ip", c.IP()))
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid username or password",
			})
		}
		slog.Error("login_user_lookup_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to look up user",
		})
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		slog.Warn("login_password_mismatch", slog.String("username", req.Username), slog.String("ip", c.IP()))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	now := time.Now()
	expiresAt := now.Add(AccessTokenTTL)
	token, err := h.jwtManager.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		slog.Error("login_token_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to start session",
		})
	}

	slog.Info("login_success", slog.String("username", user.Username), slog.String("role", string(user.Role)))

	return c.JSON(LoginResponse{
		Token:     token,
		User:      user.ToResponse(),
		ExpiresAt: expiresAt,
	})
}

// Me handles GET /api/me. Returns the user that owns the access JWT.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*auth.Claims)
	if !ok || claims == nil {
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "missing_claims",
			Message: "Authentication required",
		})
	}

	user, err := h.users.GetByID(c.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "user_not_found",
				Message: "User no longer exists",
			})
		}
		slog.Error("me_lookup_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to look up user",
		})
	}

	return c.JSON(model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}

// JWTMiddleware returns a Fiber middleware that validates JWT tokens
// and stores the claims in the request locals.
func (h *AuthHandler) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "missing_token",
				Message: "Authorization header is required",
			})
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "invalid_token_format",
				Message: "Authorization header must be Bearer token",
			})
		}
		tokenString := authHeader[len(bearerPrefix):]

		claims, err := h.jwtManager.Validate(tokenString)
		if err != nil {
			slog.Warn("token_validation_failed", slog.String("ip", c.IP()), slog.Any("error", err))
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "invalid_token",
				Message: "Invalid or expired token",
			})
		}

		c.Locals("claims", claims)
		return c.Next()
	}
}
