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
