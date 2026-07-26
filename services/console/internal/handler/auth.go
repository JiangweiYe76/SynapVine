package handler

import (
	"errors"
	"log/slog"
	"time"

	"console/internal/auth"
	"console/internal/model"
	"console/internal/store"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AccessTokenTTL is the lifetime of a freshly minted access JWT. Keep
// it short so a leaked token is only useful for a bounded window.
const AccessTokenTTL = 24 * time.Hour

// RefreshTokenTTL is the lifetime of a refresh token. Refresh tokens
// are rotated on every successful /api/auth/refresh call.
const RefreshTokenTTL = 7 * 24 * time.Hour

// AuthHandler handles authentication-related HTTP requests. The user
// store is now backed by MySQL (see internal/store); the previous
// in-memory map seeded with a hard-coded admin has been removed.
type AuthHandler struct {
	jwtManager    *auth.JWTManager
	users         *store.UserStore
	refreshTokens *store.RefreshTokenStore
	audit         *store.AuditStore
	cookieSecure  bool
}

// NewAuthHandler creates a new AuthHandler. The previous constructor
// signature took only a JWT secret; callers must now pass the user
// store, refresh-token store, and audit store wired to the same MySQL
// database. cookieSecure controls the Secure attribute on the
// refresh-token cookie (false for dev over plain HTTP, true in prod).
func NewAuthHandler(jwtSecret string, cookieSecure bool, users *store.UserStore, refresh *store.RefreshTokenStore, audit *store.AuditStore) *AuthHandler {
	return &AuthHandler{
		jwtManager:    auth.NewJWTManager(jwtSecret),
		users:         users,
		refreshTokens: refresh,
		audit:         audit,
		cookieSecure:  cookieSecure,
	}
}

// LoginRequest is the JSON body of POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LogoutRequest is the JSON body of POST /api/auth/logout. The refresh
// token itself is read from the httpOnly cookie, not the body, so the
// body only carries the all_devices flag. Omitting the body still
// revokes the access JWT via token_ver bump and clears the cookie.
type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

// LoginResponse is the success body for both /login and /refresh. The
// refresh token is delivered via an HttpOnly Set-Cookie header and is
// intentionally absent from the JSON body so client-side JS can never
// read it.
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
		_ = h.audit.Log(c.Context(), store.AuditEvent{
			Username: user.Username,
			Action:   "login_failed",
			IP:       c.IP(),
		})
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid username or password",
		})
	}

	resp, err := h.issueSession(c, user)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to start session",
		})
	}

	slog.Info("login_success", slog.String("username", user.Username), slog.String("role", string(user.Role)))
	_ = h.audit.Log(c.Context(), store.AuditEvent{
		UserID:   user.ID,
		Username: user.Username,
		Action:   "login",
		IP:       c.IP(),
	})

	return c.JSON(resp)
}

// Refresh handles POST /api/auth/refresh. It rotates the refresh token:
// the presented cookie token is deleted and a new one is issued (and
// written back to the cookie) alongside a fresh access JWT. The request
// body is empty; the refresh token is read from the httpOnly cookie.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshID := c.Cookie("refresh_token")
	if refreshID == "" {
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_refresh_token",
			Message: "Refresh token is invalid or has been revoked",
		})
	}

	userID, expiresAt, err := h.refreshTokens.Lookup(c.Context(), refreshID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			h.clearRefreshCookie(c)
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "invalid_refresh_token",
				Message: "Refresh token is invalid or has been revoked",
			})
		}
		slog.Error("refresh_lookup_failed", slog.Any("error", err))
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to validate refresh token",
		})
	}
	if time.Now().After(expiresAt) {
		_ = h.refreshTokens.Delete(c.Context(), refreshID)
		h.clearRefreshCookie(c)
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "refresh_token_expired",
			Message: "Refresh token has expired, please sign in again",
		})
	}

	user, err := h.users.GetByID(c.Context(), userID)
	if err != nil {
		slog.Error("refresh_user_lookup_failed", slog.Any("error", err))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_refresh_token",
			Message: "Refresh token refers to a user that no longer exists",
		})
	}

	// Rotate: delete the old refresh token row before issuing a new
	// one. If the new insert fails, the user simply has to sign in
	// again. issueSession writes the new refresh token to the cookie as
	// a side effect.
	if err := h.refreshTokens.Delete(c.Context(), refreshID); err != nil {
		slog.Error("refresh_delete_old_failed", slog.Any("error", err))
	}

	resp, err := h.issueSession(c, user)
	if err != nil {
		return c.Status(500).JSON(model.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to rotate session",
		})
	}

	return c.JSON(resp)
}

// Logout handles POST /api/auth/logout. Requires a valid JWT (so the
// user is already authenticated). Deletes the refresh token row read
// from the httpOnly cookie, clears the cookie, and bumps token_ver so
// the current access JWT is rejected by future requests.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	claims, ok := c.Locals("claims").(*auth.Claims)
	if !ok || claims == nil {
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "missing_claims",
			Message: "Authentication required",
		})
	}

	var req LogoutRequest
	// Body is optional. BodyParser failure is ignored: a client can
	// log out by sending an empty body. Only all_devices is read from
	// the body; the refresh token comes from the cookie.
	_ = c.BodyParser(&req)

	refreshID := c.Cookie("refresh_token")

	if req.AllDevices {
		if err := h.refreshTokens.DeleteAllForUser(c.Context(), claims.UserID); err != nil {
			slog.Error("logout_revoke_all_failed", slog.Any("error", err))
		}
	} else if refreshID != "" {
		// Only delete the row if it actually belongs to this user;
		// otherwise an attacker with a stolen JWT could probe for
		// other users' refresh tokens.
		userID, _, lookupErr := h.refreshTokens.Lookup(c.Context(), refreshID)
		if lookupErr == nil && userID == claims.UserID {
			if err := h.refreshTokens.Delete(c.Context(), refreshID); err != nil {
				slog.Error("logout_revoke_failed", slog.Any("error", err))
			}
		}
	}

	// Always clear the cookie so the browser drops any stale refresh
	// token, even when no matching DB row was found.
	h.clearRefreshCookie(c)

	if err := h.users.BumpTokenVer(c.Context(), claims.UserID); err != nil {
		slog.Error("logout_bump_tokenver_failed", slog.Any("error", err))
	}

	_ = h.audit.Log(c.Context(), store.AuditEvent{
		UserID:   claims.UserID,
		Username: claims.Username,
		Action:   "logout",
		IP:       c.IP(),
	})

	return c.SendStatus(204)
}

// Me handles GET /api/me. Returns the user that owns the access JWT,
// re-validated against the current token_ver so a revoked token is
// surfaced as 401 rather than stale data.
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

	if user.TokenVer != claims.TokenVer {
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "token_revoked",
			Message: "Session has been revoked, please sign in again",
		})
	}

	return c.JSON(model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}

// validateAndSetClaims is the shared core logic for JWT validation.
// It validates the token, checks token_ver against the database, and
// stores the claims in c.Locals("claims").
func (h *AuthHandler) validateAndSetClaims(c *fiber.Ctx, tokenString string) error {
	claims, err := h.jwtManager.Validate(tokenString)
	if err != nil {
		slog.Warn("token_validation_failed", slog.String("ip", c.IP()), slog.Any("error", err))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid or expired token",
		})
	}

	// Cross-check token_ver against the live user row. This is
	// the revocation channel: bumping token_ver on logout or
	// password change invalidates every outstanding JWT for
	// that user.
	user, err := h.users.GetByID(c.Context(), claims.UserID)
	if err != nil {
		slog.Warn("token_user_lookup_failed", slog.String("user_id", claims.UserID), slog.Any("error", err))
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid or expired token",
		})
	}
	if user.TokenVer != claims.TokenVer {
		return c.Status(401).JSON(model.ErrorResponse{
			Error:   "token_revoked",
			Message: "Session has been revoked, please sign in again",
		})
	}

	c.Locals("claims", claims)
	return nil
}

// JWTMiddleware returns a Fiber middleware that validates JWT tokens
// from the Authorization header. It ensures the embedded token_ver
// matches the user's current token_ver in the database. Mismatches
// (logout, password change) cause a 401 with error="token_revoked".
func (h *AuthHandler) JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "missing_token",
				Message: "Authorization header is required",
			})
		}
		tokenString := authHeader[len(bearerPrefix):]

		if err := h.validateAndSetClaims(c, tokenString); err != nil {
			return err
		}
		return c.Next()
	}
}

// issueSession mints a new access JWT and refresh token pair for the
// given user. Used by /login and /refresh. The refresh token is written
// to the httpOnly cookie as a side effect; the returned LoginResponse
// carries only the access token, user, and expiry.
func (h *AuthHandler) issueSession(c *fiber.Ctx, user *model.User) (LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(AccessTokenTTL)
	refreshExpires := now.Add(RefreshTokenTTL)

	token, err := h.jwtManager.Generate(user.ID, user.Username, string(user.Role), user.TokenVer)
	if err != nil {
		slog.Error("issue_token_failed", slog.Any("error", err))
		return LoginResponse{}, err
	}

	refreshID, err := newRefreshID()
	if err != nil {
		slog.Error("issue_refresh_id_failed", slog.Any("error", err))
		return LoginResponse{}, err
	}
	if err := h.refreshTokens.Create(c.Context(), refreshID, user.ID, refreshExpires); err != nil {
		slog.Error("issue_refresh_persist_failed", slog.Any("error", err))
		return LoginResponse{}, err
	}

	// Deliver the refresh token via an HttpOnly cookie so client-side JS
	// can never read it. The JSON body intentionally omits it.
	h.setRefreshCookie(c, refreshID, refreshExpires)

	return LoginResponse{
		Token:     token,
		User:      user.ToResponse(),
		ExpiresAt: expiresAt,
	}, nil
}

// setRefreshCookie writes the refresh token as an HttpOnly cookie. Path
// is scoped to /api/auth so the cookie is sent only to auth endpoints,
// avoiding unnecessary transmission on every API call. SameSite=Strict
// blocks cross-site CSRF on auth endpoints. Secure is env-controlled
// (COOKIE_SECURE) so dev over plain HTTP still works.
func (h *AuthHandler) setRefreshCookie(c *fiber.Ctx, id string, expiresAt time.Time) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    id,
		Path:     "/api/auth",
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: "Strict",
	})
}

// clearRefreshCookie expires the refresh-token cookie immediately so the
// browser removes it. The attributes (Path, SameSite, Secure) must
// match the set cookie for the browser to honour the deletion.
func (h *AuthHandler) clearRefreshCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   h.cookieSecure,
		SameSite: "Strict",
	})
}

// newRefreshID returns a UUID v4 string. The DB column is VARCHAR(36),
// which is the canonical UUID length (8-4-4-4-12 with hyphens), so
// using uuid.NewString keeps the value within range. The uuid package
// uses crypto/rand internally, so the value is cryptographically
// random.
//
// The (string, error) signature is kept for forward compatibility: the
// underlying uuid.New() can in theory fail if the system RNG fails,
// and a future revision may want to surface that.
func newRefreshID() (string, error) {
	return uuid.NewString(), nil
}
