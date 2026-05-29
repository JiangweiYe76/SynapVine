package handler

import (
	"log/slog"
	"time"

	"console/internal/auth"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	jwtManager *auth.JWTManager
	users      map[string]*model.User
}

// NewAuthHandler creates a new AuthHandler with the given JWT secret
func NewAuthHandler(jwtSecret string) *AuthHandler {
	h := &AuthHandler{
		jwtManager: auth.NewJWTManager(jwtSecret),
		users:      make(map[string]*model.User),
	}

	hashedPassword, _ := auth.HashPassword("admin123")
	h.users["admin"] = &model.User{
		ID:        "1",
		Username:  "admin",
		Password:  hashedPassword,
		Role:      model.RoleAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return h
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Warn("login_parse_error", slog.String("ip", c.IP()))
		return c.Status(400).JSON(model.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	user, exists := h.users[req.Username]
	if !exists {
		slog.Warn("login_user_not_found", slog.String("username", req.Username), slog.String("ip", c.IP()))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		slog.Warn("login_password_mismatch", slog.String("username", req.Username), slog.String("ip", c.IP()))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	token, err := h.jwtManager.Generate(user.ID, user.Username, string(user.Role))
	if err != nil {
		slog.Error("login_token_generation_failed", slog.String("username", req.Username), slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate authentication token",
		})
	}

	slog.Info("login_success", slog.String("username", req.Username), slog.String("role", string(user.Role)))

	return c.JSON(model.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	})
}

// Me handles GET /api/me
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	claims := c.Locals("claims").(*auth.Claims)
	return c.JSON(model.UserResponse{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     model.Role(claims.Role),
	})
}

// JWTMiddleware returns a Fiber middleware that validates JWT tokens
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
