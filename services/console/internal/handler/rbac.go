package handler

import (
	"log/slog"

	"console/internal/auth"
	"console/internal/model"

	"github.com/gofiber/fiber/v2"
)

// RequireRole returns a Fiber middleware that rejects the request
// when the authenticated user does not have one of the allowed roles.
// The user's role is taken from JWT claims that JWTMiddleware has
// already attached to the request context as "claims".
func RequireRole(allowed ...model.Role) fiber.Handler {
	allowedSet := make(map[model.Role]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		raw := c.Locals("claims")
		claims, ok := raw.(*auth.Claims)
		if !ok || claims == nil {
			return c.Status(401).JSON(model.ErrorResponse{
				Error:   "missing_claims",
				Message: "Authentication required",
			})
		}

		role := model.Role(claims.Role)
		if _, ok := allowedSet[role]; !ok {
			slog.Warn("rbac_denied",
				slog.String("username", claims.Username),
				slog.String("role", string(role)),
				slog.String("path", c.Path()),
				slog.String("ip", c.IP()),
			)
			return c.Status(403).JSON(model.ErrorResponse{
				Error:   "forbidden",
				Message: "Insufficient role for this action",
			})
		}
		return c.Next()
	}
}

// roleRank assigns a small numeric rank to each role so the middleware
// can express "at least editor" style rules. Kept here rather than in
// the model package to keep the model free of behaviour.
func roleRank(r model.Role) int {
	switch r {
	case model.RoleAdmin:
		return 3
	case model.RoleEditor:
		return 2
	case model.RoleViewer:
		return 1
	}
	return 0
}

// RequireMinRole is a convenience wrapper around RequireRole that allows
// a role and any higher-ranked one. e.g. RequireMinRole(model.RoleEditor)
// permits editor and admin.
//
// Currently unused but exported for future fine-grained route protection.
func RequireMinRole(min model.Role) fiber.Handler {
	all := []model.Role{model.RoleAdmin, model.RoleEditor, model.RoleViewer}
	var allowed []model.Role
	for _, r := range all {
		if roleRank(r) >= roleRank(min) {
			allowed = append(allowed, r)
		}
	}
	return RequireRole(allowed...)
}
